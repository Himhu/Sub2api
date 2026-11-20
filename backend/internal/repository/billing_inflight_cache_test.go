package repository

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newInflightTestEnv(t *testing.T, enabled bool, ttlSeconds int) (*miniredis.Miniredis, *billingCache, *service.BillingCacheService) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := &billingCache{rdb: rdb}
	cfg := &config.Config{}
	cfg.Billing.InflightReservation = config.InflightReservationConfig{
		Enabled:          enabled,
		TTLSeconds:       ttlSeconds,
		DefaultMaxTokens: 8192,
	}
	svc := service.NewBillingCacheService(cache, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)
	return mr, cache, svc
}

func inflightCount(t *testing.T, cache *billingCache, userID int64) int64 {
	t.Helper()
	zkey, _ := billingInflightKeys(userID)
	n, err := cache.rdb.ZCard(context.Background(), zkey).Result()
	require.NoError(t, err)
	return n
}

func TestInflightReservation_DisabledKeepsLegacyBehavior(t *testing.T) {
	_, cache, svc := newInflightTestEnv(t, false, 60)
	ctx := context.Background()
	user := &service.User{ID: 11}
	require.NoError(t, cache.SetUserBalance(ctx, user.ID, 0.01))
	require.False(t, svc.InflightReservationEnabled())
	for i := 0; i < 5; i++ {
		_, err := svc.ReserveInflightBalance(ctx, user, nil, 10)
		require.NoError(t, err)
	}
	require.Equal(t, int64(0), inflightCount(t, cache, user.ID))
}

func TestInflightReservation_ZeroEstimateSkips(t *testing.T) {
	_, cache, svc := newInflightTestEnv(t, true, 60)
	ctx := context.Background()
	user := &service.User{ID: 13}
	require.NoError(t, cache.SetUserBalance(ctx, user.ID, 1))
	_, err := svc.ReserveInflightBalance(ctx, user, nil, 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), inflightCount(t, cache, user.ID))
}

// downReserveCache 余额读取正常，但预留走一个已关闭的 Redis。
type downReserveCache struct {
	service.BillingCache
	down *billingCache
}

func (d *downReserveCache) ReserveInflightBalance(ctx context.Context, userID int64, requestID string, amount, balance float64, ttl time.Duration) (bool, float64, error) {
	return d.down.ReserveInflightBalance(ctx, userID, requestID, amount, balance, ttl)
}

func (d *downReserveCache) ReleaseInflightBalance(ctx context.Context, userID int64, requestID string) error {
	return d.down.ReleaseInflightBalance(ctx, userID, requestID)
}

func TestInflightReservation_RedisDownFailsOpen(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	up := &billingCache{rdb: rdb}
	require.NoError(t, up.SetUserBalance(context.Background(), 5, 0.01))

	downMR := miniredis.RunT(t)
	downRDB := redis.NewClient(&redis.Options{Addr: downMR.Addr(), MaxRetries: -1, DialTimeout: 100 * time.Millisecond})
	t.Cleanup(func() { _ = downRDB.Close() })
	downMR.Close()

	cfg := &config.Config{}
	cfg.Billing.InflightReservation = config.InflightReservationConfig{Enabled: true, TTLSeconds: 60}
	svc := service.NewBillingCacheService(&downReserveCache{BillingCache: up, down: &billingCache{rdb: downRDB}}, nil, nil, nil, nil, cfg, nil)
	t.Cleanup(svc.Stop)

	user := &service.User{ID: 5}
	for i := 0; i < 3; i++ {
		release, err := svc.ReserveInflightBalance(context.Background(), user, nil, 100)
		require.NoError(t, err)
		release()
	}
}

func TestInflightReservation_ScriptErrorSurfaces(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	cache := &billingCache{rdb: rdb}
	mr.Close()
	_, _, err := cache.ReserveInflightBalance(context.Background(), 1, "r", 1, 1, time.Second)
	require.Error(t, err)
	require.False(t, errors.Is(err, context.Canceled))
	_ = rdb.Close()
}

func TestInflightReservation_RenewDoesNotResurrectExpiredMember(t *testing.T) {
	_, cache, _ := newInflightTestEnv(t, true, 60)
	ctx := context.Background()
	userID := int64(79)
	zkey, hkey := billingInflightKeys(userID)
	// 已过期但尚未被惰性清理的成员（score 在过去）。
	require.NoError(t, cache.rdb.ZAdd(ctx, zkey, redis.Z{Score: float64(time.Now().UnixMilli() - 1000), Member: "stale"}).Err())
	require.NoError(t, cache.rdb.HSet(ctx, hkey, "stale", "0.5").Err())

	ok, err := cache.RenewInflightBalance(ctx, userID, "stale", time.Minute)
	require.NoError(t, err)
	require.False(t, ok, "renew must not resurrect an expired member")
	require.Equal(t, int64(0), inflightCount(t, cache, userID))
	exists, err := cache.rdb.HExists(ctx, hkey, "stale").Result()
	require.NoError(t, err)
	require.False(t, exists)
}

func TestInflightScopeReservation_ReturnsOthersAndReleases(t *testing.T) {
	mr, cache, _ := newInflightTestEnv(t, true, 60)
	ctx := context.Background()

	others, err := cache.ReserveInflightScope(ctx, "apikey:7", "r1", 3, time.Minute)
	require.NoError(t, err)
	require.Zero(t, others)
	others, err = cache.ReserveInflightScope(ctx, "apikey:7", "r2", 2, time.Minute)
	require.NoError(t, err)
	require.InDelta(t, 3, others, 1e-9, "返回登记前其他在途预留的合计")
	others, err = cache.ReserveInflightScope(ctx, "sub:7", "r3", 5, time.Minute)
	require.NoError(t, err)
	require.Zero(t, others, "不同维度互不影响")

	require.NoError(t, cache.ReleaseInflightScope(ctx, "apikey:7", "r1"))
	require.NoError(t, cache.ReleaseInflightScope(ctx, "apikey:7", "r1"))
	others, err = cache.ReserveInflightScope(ctx, "apikey:7", "r4", 1, time.Minute)
	require.NoError(t, err)
	require.InDelta(t, 2, others, 1e-9)

	mr.FastForward(2 * time.Minute)
	others, err = cache.ReserveInflightScope(ctx, "apikey:7", "r5", 1, time.Minute)
	require.NoError(t, err)
	require.Zero(t, others, "过期预留不再计入")
}

func TestReserveFixedCharge_ConcurrentChargesAdmitOnlyWhatAPIKeyQuotaCovers(t *testing.T) {
	_, cache, svc := newInflightTestEnv(t, true, 60)
	ctx := context.Background()
	user := &service.User{ID: 42}
	apiKey := &service.APIKey{ID: 9, UserID: user.ID, Quota: 1.0}
	require.NoError(t, cache.SetUserBalance(ctx, user.ID, 100))

	const n = 20
	var admitted atomic.Int32
	reservations := make(chan *service.InflightReservation, n)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			res, err := svc.ReserveFixedCharge(ctx, user, apiKey, nil, 0.3)
			if err != nil {
				if !errors.Is(err, service.ErrAPIKeyQuotaExhausted) {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			admitted.Add(1)
			reservations <- res
		}()
	}
	close(start)
	wg.Wait()
	close(reservations)

	require.Equal(t, int32(3), admitted.Load(), "Key 额度 1.0、每笔 0.3：同时在途的不超过 3 笔")
	for res := range reservations {
		res.Release()
	}
	zkey, _ := billingInflightScopeKeys("apikey:9")
	remaining, err := cache.rdb.ZCard(ctx, zkey).Result()
	require.NoError(t, err)
	require.Zero(t, remaining)
	require.Equal(t, int64(0), inflightCount(t, cache, user.ID))
}
