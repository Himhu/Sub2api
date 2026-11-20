//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// totpVMUserRepoStub 仅实现 TOTP 验证方式测试所需方法；未桩方法调用即 panic（嵌入 nil 接口）。
type totpVMUserRepoStub struct {
	UserRepository
	user          *User
	totpDisabled  bool
	disableCalled bool
}

func (s *totpVMUserRepoStub) GetByID(ctx context.Context, id int64) (*User, error) {
	if s.user == nil {
		return nil, errors.New("user not found")
	}
	return s.user, nil
}

func (s *totpVMUserRepoStub) DisableTotp(ctx context.Context, userID int64) error {
	s.disableCalled = true
	s.totpDisabled = true
	return nil
}

type totpVMSettingRepoStub struct {
	SettingRepository
	values map[string]string
}

func (s *totpVMSettingRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	v, ok := s.values[key]
	if !ok {
		return "", errors.New("setting not found")
	}
	return v, nil
}

func newTotpVMService(t *testing.T, user *User, emailVerifyEnabled bool) (*TotpService, *totpVMUserRepoStub) {
	t.Helper()
	userRepo := &totpVMUserRepoStub{user: user}
	values := map[string]string{}
	if emailVerifyEnabled {
		// 历史上的 email_verify_enabled 开关；自托管裁剪后该设置不再影响 TOTP 身份校验方式。
		values["email_verify_enabled"] = "true"
	}
	settingSvc := NewSettingService(&totpVMSettingRepoStub{values: values}, nil)
	return NewTotpService(userRepo, nil, nil, settingSvc), userRepo
}

func TestTotpDisableAdminUsesPasswordEvenWithEmailVerifyEnabled(t *testing.T) {
	admin := &User{ID: 1, Email: "admin@example.com", Role: RoleAdmin, TotpEnabled: true}
	require.NoError(t, admin.SetPassword("correct-password"))
	svc, userRepo := newTotpVMService(t, admin, true)

	// 缺密码 → 要求密码（而非邮箱验证码）。
	err := svc.Disable(context.Background(), admin.ID, "", "")
	require.ErrorIs(t, err, ErrPasswordRequired)

	// 密码错误 → 拒绝。
	err = svc.Disable(context.Background(), admin.ID, "", "wrong-password")
	require.ErrorIs(t, err, ErrPasswordIncorrect)

	// 密码正确 → 成功停用；全程不需要邮箱验证码（emailService 为 nil，走到邮箱分支会 panic）。
	err = svc.Disable(context.Background(), admin.ID, "", "correct-password")
	require.NoError(t, err)
	require.True(t, userRepo.disableCalled)
}

func TestTotpDisableRegularUserUsesPassword(t *testing.T) {
	user := &User{ID: 2, Email: "user@example.com", Role: RoleUser, TotpEnabled: true}
	require.NoError(t, user.SetPassword("whatever"))
	svc, userRepo := newTotpVMService(t, user, true)

	err := svc.Disable(context.Background(), user.ID, "", "")
	require.ErrorIs(t, err, ErrPasswordRequired)

	err = svc.Disable(context.Background(), user.ID, "", "whatever")
	require.NoError(t, err)
	require.True(t, userRepo.disableCalled)
}
