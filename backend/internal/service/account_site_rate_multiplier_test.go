//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAccountUserBillingRateMultiplier(t *testing.T) {
	half, zero, negative := 0.5, 0.0, -1.0

	var nilAccount *Account
	_, ok := nilAccount.UserBillingRateMultiplier()
	require.False(t, ok, "nil account follows the group multiplier")

	_, ok = (&Account{}).UserBillingRateMultiplier()
	require.False(t, ok, "unset site multiplier follows the group multiplier")

	_, ok = (&Account{UserRateMultiplier: &negative}).UserBillingRateMultiplier()
	require.False(t, ok, "negative site multiplier is treated as unset")

	got, ok := (&Account{UserRateMultiplier: &half}).UserBillingRateMultiplier()
	require.True(t, ok)
	require.InDelta(t, 0.5, got, 1e-9)

	got, ok = (&Account{UserRateMultiplier: &zero}).UserBillingRateMultiplier()
	require.True(t, ok, "zero is a valid site multiplier (free line)")
	require.InDelta(t, 0, got, 1e-9)
}

func TestRecordUsageMultipliers_AccountSiteMultiplierOverridesGroup(t *testing.T) {
	svc := &GatewayService{cfg: &config.Config{}}
	groupID := int64(7)
	apiKey := &APIKey{GroupID: &groupID, Group: &Group{ID: groupID, RateMultiplier: 2}}
	pricingAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// 未填写本站倍率：沿用分组倍率。
	multiplier, _, _ := svc.recordUsageMultipliers(context.Background(), apiKey, nil, &Account{}, pricingAt)
	require.InDelta(t, 2.0, multiplier, 1e-9)

	// 填写本站倍率：覆盖分组倍率。
	site := 0.5
	multiplier, _, _ = svc.recordUsageMultipliers(context.Background(), apiKey, nil, &Account{UserRateMultiplier: &site}, pricingAt)
	require.InDelta(t, 0.5, multiplier, 1e-9)

	// 账号为空（如报价路径未带账号）：沿用分组倍率，不 panic。
	multiplier, _, _ = svc.recordUsageMultipliers(context.Background(), apiKey, nil, nil, pricingAt)
	require.InDelta(t, 2.0, multiplier, 1e-9)
}

func TestBuildAccountForCreate_SiteRateMultiplier(t *testing.T) {
	site, negative := 1.5, -1.0

	account, err := buildAccountForCreate(&CreateAccountInput{Name: "a", Platform: PlatformAnthropic, Type: AccountTypeAPIKey, UserRateMultiplier: &site}, nil)
	require.NoError(t, err)
	require.NotNil(t, account.UserRateMultiplier)
	require.InDelta(t, 1.5, *account.UserRateMultiplier, 1e-9)

	account, err = buildAccountForCreate(&CreateAccountInput{Name: "a", Platform: PlatformAnthropic, Type: AccountTypeAPIKey}, nil)
	require.NoError(t, err)
	require.Nil(t, account.UserRateMultiplier, "omitted site multiplier follows the group")

	account, err = buildAccountForCreate(&CreateAccountInput{Name: "a", Platform: PlatformAnthropic, Type: AccountTypeAPIKey, UserRateMultiplier: &negative}, nil)
	require.NoError(t, err)
	require.Nil(t, account.UserRateMultiplier, "negative site multiplier is ignored on create")
}

func TestAdminServiceUpdateAccount_SiteRateMultiplier(t *testing.T) {
	existing := 0.8
	newRepo := func() *longContextBillingRepoStub {
		value := existing
		return &longContextBillingRepoStub{account: &Account{
			ID: 1, Platform: PlatformAnthropic, Type: AccountTypeAPIKey, UserRateMultiplier: &value,
		}}
	}

	// 未提供：保持原值。
	account, err := (&adminServiceImpl{accountRepo: newRepo()}).UpdateAccount(context.Background(), 1, &UpdateAccountInput{Name: "renamed"})
	require.NoError(t, err)
	require.NotNil(t, account.UserRateMultiplier)
	require.InDelta(t, 0.8, *account.UserRateMultiplier, 1e-9)

	// 提供新值：更新。
	updated := 2.0
	account, err = (&adminServiceImpl{accountRepo: newRepo()}).UpdateAccount(context.Background(), 1, &UpdateAccountInput{UserRateMultiplier: &updated})
	require.NoError(t, err)
	require.NotNil(t, account.UserRateMultiplier)
	require.InDelta(t, 2.0, *account.UserRateMultiplier, 1e-9)

	// 负数：清除，回到沿用分组倍率。
	clear := -1.0
	account, err = (&adminServiceImpl{accountRepo: newRepo()}).UpdateAccount(context.Background(), 1, &UpdateAccountInput{UserRateMultiplier: &clear})
	require.NoError(t, err)
	require.Nil(t, account.UserRateMultiplier)
}
