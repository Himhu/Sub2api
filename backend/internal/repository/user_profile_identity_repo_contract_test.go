//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/suite"
)

type UserProfileIdentityRepoSuite struct {
	suite.Suite
	ctx    context.Context
	client *dbent.Client
	repo   *userRepository
}

func TestUserProfileIdentityRepoSuite(t *testing.T) {
	suite.Run(t, new(UserProfileIdentityRepoSuite))
}

func (s *UserProfileIdentityRepoSuite) SetupTest() {
	s.ctx = context.Background()
	s.client = testEntClient(s.T())
	s.repo = newUserRepositoryWithSQL(s.client, integrationDB)

	_, err := integrationDB.ExecContext(s.ctx, `TRUNCATE TABLE user_avatars RESTART IDENTITY`)
	s.Require().NoError(err)
}

func (s *UserProfileIdentityRepoSuite) mustCreateUser(label string) *dbent.User {
	s.T().Helper()

	user, err := s.client.User.Create().
		SetEmail(fmt.Sprintf("%s-%d@example.com", label, time.Now().UnixNano())).
		SetPasswordHash("test-password-hash").
		SetRole("user").
		SetStatus("active").
		Save(s.ctx)
	s.Require().NoError(err)
	return user
}

func (s *UserProfileIdentityRepoSuite) TestWithUserProfileIdentityTx_AllowsAvatarOnlyProfileUpdate() {
	user := s.mustCreateUser("avatar-only-update")

	model, err := s.repo.GetByID(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(model)

	err = s.repo.WithUserProfileIdentityTx(s.ctx, func(txCtx context.Context) error {
		_, err := s.repo.UpsertUserAvatar(txCtx, user.ID, service.UpsertUserAvatarInput{
			StorageProvider: "remote_url",
			URL:             "https://cdn.example.com/avatar.png",
		})
		if err != nil {
			return err
		}
		// 只改头像时用户行没有任何列需要写，掩码为空——与 UserService.updateProfile 一致。
		return s.repo.Update(txCtx, model, service.UserUpdateFields{})
	})
	s.Require().NoError(err)

	avatar, err := s.repo.GetUserAvatar(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(avatar)
	s.Require().Equal("https://cdn.example.com/avatar.png", avatar.URL)
}

func (s *UserProfileIdentityRepoSuite) TestUserAvatarCRUDAndUserLookup() {
	user := s.mustCreateUser("avatar")

	inlineAvatar, err := s.repo.UpsertUserAvatar(s.ctx, user.ID, service.UpsertUserAvatarInput{
		StorageProvider: "inline",
		URL:             "data:image/png;base64,QUJD",
		ContentType:     "image/png",
		ByteSize:        3,
		SHA256:          "902fbdd2b1df0c4f70b4a5d23525e932",
	})
	s.Require().NoError(err)
	s.Require().Equal("inline", inlineAvatar.StorageProvider)
	s.Require().Equal("data:image/png;base64,QUJD", inlineAvatar.URL)

	loadedAvatar, err := s.repo.GetUserAvatar(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(loadedAvatar)
	s.Require().Equal("image/png", loadedAvatar.ContentType)
	s.Require().Equal(3, loadedAvatar.ByteSize)

	_, err = s.repo.UpsertUserAvatar(s.ctx, user.ID, service.UpsertUserAvatarInput{
		StorageProvider: "remote_url",
		URL:             "https://cdn.example.com/avatar.png",
	})
	s.Require().NoError(err)

	loadedAvatar, err = s.repo.GetUserAvatar(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().NotNil(loadedAvatar)
	s.Require().Equal("remote_url", loadedAvatar.StorageProvider)
	s.Require().Equal("https://cdn.example.com/avatar.png", loadedAvatar.URL)
	s.Require().Zero(loadedAvatar.ByteSize)

	s.Require().NoError(s.repo.DeleteUserAvatar(s.ctx, user.ID))
	loadedAvatar, err = s.repo.GetUserAvatar(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().Nil(loadedAvatar)
}

func (s *UserProfileIdentityRepoSuite) TestUpdateUserLastActiveAt_UsesDedicatedColumn() {
	user := s.mustCreateUser("activity")
	loginAt := time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC)
	activeAt := loginAt.Add(5 * time.Minute)

	s.Require().NoError(s.repo.UpdateUserLastActiveAt(s.ctx, user.ID, activeAt))

	var storedActiveAt sqlNullTime
	s.Require().NoError(integrationDB.QueryRowContext(s.ctx, `
SELECT last_active_at
FROM users
WHERE id = $1`,
		user.ID,
	).Scan(&storedActiveAt))
	s.Require().True(storedActiveAt.Valid)
	s.Require().True(storedActiveAt.Time.Equal(activeAt))
}

type sqlNullTime struct {
	Time  time.Time
	Valid bool
}

func (t *sqlNullTime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		t.Valid = true
		return nil
	case nil:
		t.Time = time.Time{}
		t.Valid = false
		return nil
	default:
		return fmt.Errorf("unsupported scan type %T", value)
	}
}

func stringPtr(v string) *string {
	return &v
}
