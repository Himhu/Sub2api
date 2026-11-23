//go:build unit

package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func costTestAccount(id int64, rate float64) accountWithLoad {
	r := rate
	return accountWithLoad{
		account:  &Account{ID: id, RateMultiplier: &r},
		loadInfo: &AccountLoadInfo{AccountID: id},
	}
}

func costTestIDs(accounts []accountWithLoad) []int64 {
	ids := make([]int64, 0, len(accounts))
	for _, acc := range accounts {
		ids = append(ids, acc.account.ID)
	}
	return ids
}

func costTestConfig() config.GatewaySchedulingConfig {
	return config.GatewaySchedulingConfig{
		PreferLowerCost:    true,
		CostTTFTGuardAbsMS: 800,
		CostTTFTGuardRatio: 0.5,
	}
}

func costTestMs(v int) *int { return &v }

func TestAccountTTFTStats_EWMA(t *testing.T) {
	stats := newAccountTTFTStats()

	_, ok := stats.snapshot(1)
	require.False(t, ok, "no samples yet")

	stats.report(1, costTestMs(1000))
	got, ok := stats.snapshot(1)
	require.True(t, ok)
	require.InDelta(t, 1000, got, 0.001, "first sample seeds the EWMA")

	stats.report(1, costTestMs(2000))
	got, _ = stats.snapshot(1)
	require.InDelta(t, 0.2*2000+0.8*1000, got, 0.001)

	// nil / non-positive samples (non-stream requests) must not move the value
	stats.report(1, nil)
	stats.report(1, costTestMs(0))
	again, _ := stats.snapshot(1)
	require.Equal(t, got, again)

	stats.forget(1)
	_, ok = stats.snapshot(1)
	require.False(t, ok)
}

func TestFilterByLowestCost_PicksCheapestTier(t *testing.T) {
	accounts := []accountWithLoad{
		costTestAccount(1, 1.0),
		costTestAccount(2, 0.45),
		costTestAccount(3, 0.45),
		costTestAccount(4, 0.8),
	}
	got := filterByLowestCost(accounts, newAccountTTFTStats(), costTestConfig())
	require.ElementsMatch(t, []int64{2, 3}, costTestIDs(got), "both accounts of the cheapest tier survive for load/LRU selection")
}

func TestFilterByLowestCost_NilRateCountsAsOne(t *testing.T) {
	accounts := []accountWithLoad{
		{account: &Account{ID: 1}, loadInfo: &AccountLoadInfo{AccountID: 1}},
		costTestAccount(2, 0.9),
		costTestAccount(3, 1.0),
	}
	got := filterByLowestCost(accounts, newAccountTTFTStats(), costTestConfig())
	require.Equal(t, []int64{2}, costTestIDs(got))

	// An explicit 0 (free line) beats everything.
	accounts = append(accounts, costTestAccount(4, 0))
	got = filterByLowestCost(accounts, newAccountTTFTStats(), costTestConfig())
	require.Equal(t, []int64{4}, costTestIDs(got))
}

func TestFilterByLowestCost_TTFTGuardDemotesSlowCheapLine(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(600))  // expensive but fast
	stats.report(2, costTestMs(2500)) // cheap but slow: +1900ms and +316% → tripped
	accounts := []accountWithLoad{
		costTestAccount(1, 1.0),
		costTestAccount(2, 0.45),
	}
	got := filterByLowestCost(accounts, stats, costTestConfig())
	require.Equal(t, []int64{1}, costTestIDs(got), "slow cheap line is demoted behind the fast tier")
}

func TestFilterByLowestCost_GuardNeedsBothAbsoluteAndRelative(t *testing.T) {
	cfg := costTestConfig()

	// Case A: big absolute gap but small relative gap (baseline is already slow) → stays.
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(5000))
	stats.report(2, costTestMs(6000)) // +1000ms (> 800) but only +20% (< 50%)
	accounts := []accountWithLoad{costTestAccount(1, 1.0), costTestAccount(2, 0.45)}
	require.Equal(t, []int64{2}, costTestIDs(filterByLowestCost(accounts, stats, cfg)))

	// Case B: big relative gap but small absolute gap (both fast) → stays.
	stats = newAccountTTFTStats()
	stats.report(1, costTestMs(200))
	stats.report(2, costTestMs(700)) // +500ms (< 800) but +250%
	require.Equal(t, []int64{2}, costTestIDs(filterByLowestCost(accounts, stats, cfg)))

	// Case C: both exceeded → demoted.
	stats = newAccountTTFTStats()
	stats.report(1, costTestMs(500))
	stats.report(2, costTestMs(1400)) // +900ms and +180%
	require.Equal(t, []int64{1}, costTestIDs(filterByLowestCost(accounts, stats, cfg)))
}

func TestFilterByLowestCost_UnknownTTFTIsNeutral(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(300)) // only the expensive line has data
	accounts := []accountWithLoad{costTestAccount(1, 1.0), costTestAccount(2, 0.45)}
	got := filterByLowestCost(accounts, stats, costTestConfig())
	require.Equal(t, []int64{2}, costTestIDs(got), "a cheap line without samples is never demoted")
}

func TestFilterByLowestCost_GuardDisabledIsPureCost(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(100))
	stats.report(2, costTestMs(9000))
	cfg := costTestConfig()
	cfg.CostTTFTGuardAbsMS = 0
	accounts := []accountWithLoad{costTestAccount(1, 1.0), costTestAccount(2, 0.45)}
	require.Equal(t, []int64{2}, costTestIDs(filterByLowestCost(accounts, stats, cfg)))
}

func TestFilterByLowestCost_FallsThroughTiers(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(1, costTestMs(300))  // tier 1.0 fast
	stats.report(2, costTestMs(3000)) // tier 0.45 slow → demoted
	stats.report(3, costTestMs(400))  // tier 0.8 fine → chosen
	stats.report(4, costTestMs(4000)) // tier 0.45 slow → demoted
	accounts := []accountWithLoad{
		costTestAccount(1, 1.0),
		costTestAccount(2, 0.45),
		costTestAccount(3, 0.8),
		costTestAccount(4, 0.45),
	}
	require.Equal(t, []int64{3}, costTestIDs(filterByLowestCost(accounts, stats, costTestConfig())))
}

func TestFilterByLowestCost_AllTrippedReturnsOriginal(t *testing.T) {
	// Only one account has TTFT and it is the baseline; everything else is unknown → nothing
	// trips. Construct the opposite: baseline exists but every account in every tier trips is
	// impossible (baseline itself never trips), so verify the single-element and empty cases.
	require.Empty(t, filterByLowestCost(nil, newAccountTTFTStats(), costTestConfig()))
	one := []accountWithLoad{costTestAccount(1, 1.0)}
	require.Equal(t, one, filterByLowestCost(one, newAccountTTFTStats(), costTestConfig()))
}

func TestFilterByLowestCost_NilStats(t *testing.T) {
	accounts := []accountWithLoad{costTestAccount(1, 1.0), costTestAccount(2, 0.45)}
	require.Equal(t, []int64{2}, costTestIDs(filterByLowestCost(accounts, nil, costTestConfig())))
}

func costTestAccountWithPriority(id int64, rate float64, priority int) accountWithLoad {
	acc := costTestAccount(id, rate)
	acc.account.Priority = priority
	return acc
}

func TestOrderByLowestCost_PriorityGroupsThenCostTiers(t *testing.T) {
	stats := newAccountTTFTStats()
	stats.report(11, costTestMs(300))
	stats.report(12, costTestMs(2500)) // 便宜但慢，被守门降档
	accounts := []accountWithLoad{
		costTestAccountWithPriority(11, 1.0, 1),
		costTestAccountWithPriority(12, 0.45, 1),
		costTestAccountWithPriority(21, 0.3, 2), // 更便宜，但优先级更低，必须留在后面
		costTestAccountWithPriority(13, 0.8, 1),
	}
	got := costTestIDs(orderByLowestCost(accounts, stats, costTestConfig()))
	require.Equal(t, []int64{13, 11, 12, 21}, got, "priority 1 first, within it the guarded cheap tier falls behind, priority 2 stays last")
}

func TestOrderByLowestCost_KeepsRelativeOrderWithinTier(t *testing.T) {
	accounts := []accountWithLoad{
		costTestAccount(3, 0.5),
		costTestAccount(1, 0.5),
		costTestAccount(2, 1.0),
	}
	require.Equal(t, []int64{3, 1, 2}, costTestIDs(orderByLowestCost(accounts, newAccountTTFTStats(), costTestConfig())),
		"the cheaper tier keeps the incoming (load/LRU) order")
}

func TestOrderAccountsByLowestCost_NilEntries(t *testing.T) {
	accounts := []*Account{nil, costTestAccount(2, 1.0).account, costTestAccount(1, 0.4).account}
	got := orderAccountsByLowestCost(accounts, newAccountTTFTStats(), costTestConfig())
	require.Equal(t, []int64{1, 2}, []int64{got[0].ID, got[1].ID})
}
