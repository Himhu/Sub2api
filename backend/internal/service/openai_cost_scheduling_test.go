//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func openAICostTestService(accounts []Account, loads map[int64]*AccountLoadInfo, preferLowerCost bool, stats *accountTTFTStats) *OpenAIGatewayService {
	return &OpenAIGatewayService{
		accountRepo: stubOpenAIAccountRepo{accounts: accounts},
		cache:       &stubGatewayCache{},
		concurrencyService: NewConcurrencyService(stubConcurrencyCache{
			loadMap: loads,
		}),
		cfg: &config.Config{Gateway: config.GatewayConfig{Scheduling: config.GatewaySchedulingConfig{
			PreferLowerCost:     preferLowerCost,
			CostTTFTGuardAbsMS:  1500,
			CostTTFTGuardRatio:  2,
			FallbackWaitTimeout: time.Second,
		}}},
		ttftStats: stats,
	}
}

func openAICostTestAccounts() []Account {
	rate := func(v float64) *float64 { return &v }
	return []Account{
		{ID: 1, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 1, RateMultiplier: rate(1.0)},
		{ID: 2, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 1, RateMultiplier: rate(0.45)},
		{ID: 3, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true, Concurrency: 1, Priority: 1, RateMultiplier: rate(0.8)},
	}
}

func openAICostTestLoads() map[int64]*AccountLoadInfo {
	return map[int64]*AccountLoadInfo{
		1: {AccountID: 1, LoadRate: 10},
		2: {AccountID: 2, LoadRate: 10},
		3: {AccountID: 3, LoadRate: 10},
	}
}

func TestOpenAISelectAccountWithLoadAwareness_CostTTFTGuardDemotesSlowCheapLine(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(300))
	stats.report(2, costTestMs(2500)) // 最便宜，但首字明显更慢，被守门降档
	stats.report(3, costTestMs(400))

	svc := openAICostTestService(openAICostTestAccounts(), openAICostTestLoads(), true, stats)
	groupID := int64(1)
	selection, err := svc.SelectAccountWithLoadAwareness(context.Background(), &groupID, "cost", "gpt-4", nil)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.Equal(t, int64(3), selection.Account.ID, "the guarded cheap line falls behind the next cost tier")
}

func TestOpenAISelectAccountWithLoadAwareness_UnknownTTFTIsNeutral(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(300))
	// 账号 2 最便宜且没有首字样本，按中性处理，不应被降档

	svc := openAICostTestService(openAICostTestAccounts(), openAICostTestLoads(), true, stats)
	groupID := int64(1)
	selection, err := svc.SelectAccountWithLoadAwareness(context.Background(), &groupID, "cost", "gpt-4", nil)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.Equal(t, int64(2), selection.Account.ID)
}

func TestOpenAISelectAccountWithLoadAwareness_CostSchedulingDisabledKeepsLoadOrder(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(300))
	stats.report(2, costTestMs(2500))
	loads := openAICostTestLoads()
	loads[2].LoadRate = 90 // 最便宜的账号负载最高

	svc := openAICostTestService(openAICostTestAccounts(), loads, false, stats)
	groupID := int64(1)
	selection, err := svc.SelectAccountWithLoadAwareness(context.Background(), &groupID, "cost", "gpt-4", nil)
	require.NoError(t, err)
	require.NotNil(t, selection)
	require.Equal(t, int64(1), selection.Account.ID, "with the switch off the legacy load order is unchanged")
}
