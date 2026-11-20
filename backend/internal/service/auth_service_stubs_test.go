//go:build unit

package service_test

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type emailBindRefreshTokenCacheStub struct {
	mu       sync.Mutex
	tokens   map[string]*service.RefreshTokenData
	userSets map[int64]map[string]struct{}
	families map[string]map[string]struct{}
}

func newEmailBindRefreshTokenCacheStub() *emailBindRefreshTokenCacheStub {
	return &emailBindRefreshTokenCacheStub{
		tokens:   make(map[string]*service.RefreshTokenData),
		userSets: make(map[int64]map[string]struct{}),
		families: make(map[string]map[string]struct{}),
	}
}

func (s *emailBindRefreshTokenCacheStub) StoreRefreshToken(_ context.Context, tokenHash string, data *service.RefreshTokenData, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cloned := *data
	s.tokens[tokenHash] = &cloned
	return nil
}

func (s *emailBindRefreshTokenCacheStub) GetRefreshToken(_ context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.tokens[tokenHash]
	if !ok {
		return nil, service.ErrRefreshTokenNotFound
	}
	cloned := *data
	return &cloned, nil
}

func (s *emailBindRefreshTokenCacheStub) DeleteRefreshToken(_ context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, tokenHash)
	for _, tokenSet := range s.userSets {
		delete(tokenSet, tokenHash)
	}
	for _, tokenSet := range s.families {
		delete(tokenSet, tokenHash)
	}
	return nil
}

func (s *emailBindRefreshTokenCacheStub) ConsumeRefreshToken(_ context.Context, tokenHash string) (*service.RefreshTokenData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.tokens[tokenHash]
	if !ok {
		return nil, service.ErrRefreshTokenNotFound
	}
	delete(s.tokens, tokenHash)
	for _, tokenSet := range s.userSets {
		delete(tokenSet, tokenHash)
	}
	for _, tokenSet := range s.families {
		delete(tokenSet, tokenHash)
	}
	cloned := *data
	return &cloned, nil
}

func (s *emailBindRefreshTokenCacheStub) DeleteUserRefreshTokens(_ context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for tokenHash := range s.userSets[userID] {
		delete(s.tokens, tokenHash)
		for _, tokenSet := range s.families {
			delete(tokenSet, tokenHash)
		}
	}
	delete(s.userSets, userID)
	return nil
}

func (s *emailBindRefreshTokenCacheStub) DeleteTokenFamily(_ context.Context, familyID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for tokenHash := range s.families[familyID] {
		delete(s.tokens, tokenHash)
		for _, tokenSet := range s.userSets {
			delete(tokenSet, tokenHash)
		}
	}
	delete(s.families, familyID)
	return nil
}

func (s *emailBindRefreshTokenCacheStub) AddToUserTokenSet(_ context.Context, userID int64, tokenHash string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.userSets[userID] == nil {
		s.userSets[userID] = make(map[string]struct{})
	}
	s.userSets[userID][tokenHash] = struct{}{}
	return nil
}

func (s *emailBindRefreshTokenCacheStub) AddToFamilyTokenSet(_ context.Context, familyID string, tokenHash string, _ time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.families[familyID] == nil {
		s.families[familyID] = make(map[string]struct{})
	}
	s.families[familyID][tokenHash] = struct{}{}
	return nil
}

func (s *emailBindRefreshTokenCacheStub) GetUserTokenHashes(_ context.Context, userID int64) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokenSet := s.userSets[userID]
	out := make([]string, 0, len(tokenSet))
	for tokenHash := range tokenSet {
		out = append(out, tokenHash)
	}
	return out, nil
}

func (s *emailBindRefreshTokenCacheStub) GetFamilyTokenHashes(_ context.Context, familyID string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tokenSet := s.families[familyID]
	out := make([]string, 0, len(tokenSet))
	for tokenHash := range tokenSet {
		out = append(out, tokenHash)
	}
	return out, nil
}

func (s *emailBindRefreshTokenCacheStub) IsTokenInFamily(_ context.Context, familyID string, tokenHash string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.families[familyID][tokenHash]
	return ok, nil
}

type emailBindUserRepoStub struct {
	mu           sync.Mutex
	usersByID    map[int64]*service.User
	usersByEmail map[string]*service.User
}

func newEmailBindUserRepoStub(user *service.User) *emailBindUserRepoStub {
	cloned := cloneEmailBindUser(user)
	return &emailBindUserRepoStub{
		usersByID: map[int64]*service.User{
			cloned.ID: cloned,
		},
		usersByEmail: map[string]*service.User{
			cloned.Email: cloned,
		},
	}
}

func (s *emailBindUserRepoStub) Create(context.Context, *service.User) error { return nil }

func (s *emailBindUserRepoStub) CreateWithEmailAliasGuard(ctx context.Context, user *service.User) error {
	return s.Create(ctx, user)
}

func (s *emailBindUserRepoStub) GetByID(_ context.Context, id int64) (*service.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.usersByID[id]
	if !ok {
		return nil, service.ErrUserNotFound
	}
	return cloneEmailBindUser(user), nil
}

func (s *emailBindUserRepoStub) GetByEmail(_ context.Context, email string) (*service.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user, ok := s.usersByEmail[email]
	if !ok {
		return nil, service.ErrUserNotFound
	}
	return cloneEmailBindUser(user), nil
}

func (s *emailBindUserRepoStub) GetFirstAdmin(context.Context) (*service.User, error) {
	panic("unexpected GetFirstAdmin call")
}

func (s *emailBindUserRepoStub) Update(_ context.Context, user *service.User, _ service.UserUpdateFields) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.usersByID[user.ID]
	if !ok {
		return service.ErrUserNotFound
	}
	delete(s.usersByEmail, existing.Email)
	cloned := cloneEmailBindUser(user)
	s.usersByID[user.ID] = cloned
	s.usersByEmail[cloned.Email] = cloned
	return nil
}

func (s *emailBindUserRepoStub) Delete(context.Context, int64) error { return nil }

func (s *emailBindUserRepoStub) GetUserAvatar(context.Context, int64) (*service.UserAvatar, error) {
	return nil, nil
}

func (s *emailBindUserRepoStub) UpsertUserAvatar(context.Context, int64, service.UpsertUserAvatarInput) (*service.UserAvatar, error) {
	panic("unexpected UpsertUserAvatar call")
}

func (s *emailBindUserRepoStub) DeleteUserAvatar(context.Context, int64) error {
	panic("unexpected DeleteUserAvatar call")
}

func (s *emailBindUserRepoStub) List(context.Context, pagination.PaginationParams) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *emailBindUserRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, service.UserListFilters) ([]service.User, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *emailBindUserRepoStub) GetLatestUsedAtByUserIDs(context.Context, []int64) (map[int64]*time.Time, error) {
	return map[int64]*time.Time{}, nil
}

func (s *emailBindUserRepoStub) GetLatestUsedAtByUserID(context.Context, int64) (*time.Time, error) {
	return nil, nil
}

func (s *emailBindUserRepoStub) UpdateUserLastActiveAt(context.Context, int64, time.Time) error {
	return nil
}

func (s *emailBindUserRepoStub) UpdateBalance(context.Context, int64, float64) error { return nil }
func (s *emailBindUserRepoStub) DeductBalance(context.Context, int64, float64) error { return nil }
func (s *emailBindUserRepoStub) UpdateConcurrency(context.Context, int64, int) error { return nil }

func (s *emailBindUserRepoStub) ExistsByEmail(_ context.Context, email string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.usersByEmail[email]
	return ok, nil
}

func (s *emailBindUserRepoStub) AdjustBalance(ctx context.Context, id int64, delta float64) (service.BalanceChange, error) {
	panic("unexpected AdjustBalance call")
}

func (s *emailBindUserRepoStub) SetBalance(ctx context.Context, id int64, value float64) (service.BalanceChange, error) {
	panic("unexpected SetBalance call")
}

func (s *emailBindUserRepoStub) ExistsByEmailAlias(_ context.Context, email string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	identity := service.NormalizeEmailForAliasDedup(email)
	for stored := range s.usersByEmail {
		if service.NormalizeEmailForAliasDedup(stored) == identity {
			return true, nil
		}
	}
	return false, nil
}

func (s *emailBindUserRepoStub) BatchSetConcurrency(context.Context, []int64, int) (int, error) {
	return 0, nil
}
func (s *emailBindUserRepoStub) BatchAddConcurrency(context.Context, []int64, int) (int, error) {
	return 0, nil
}
func (s *emailBindUserRepoStub) BatchUpdateLimits(context.Context, []int64, *int, *int) (int, error) {
	return 0, nil
}

func (s *emailBindUserRepoStub) RemoveGroupFromAllowedGroups(context.Context, int64) (int64, error) {
	return 0, nil
}

func (s *emailBindUserRepoStub) AddGroupToAllowedGroups(context.Context, int64, int64) error {
	return nil
}

func (s *emailBindUserRepoStub) RemoveGroupFromUserAllowedGroups(context.Context, int64, int64) error {
	return nil
}

func (s *emailBindUserRepoStub) UnbindUserAuthProvider(context.Context, int64, string) error {
	return nil
}

func (s *emailBindUserRepoStub) UpdateTotpSecret(context.Context, int64, *string) error { return nil }
func (s *emailBindUserRepoStub) EnableTotp(context.Context, int64) error                { return nil }
func (s *emailBindUserRepoStub) DisableTotp(context.Context, int64) error               { return nil }
func (s *emailBindUserRepoStub) GetByIDIncludeDeleted(ctx context.Context, id int64) (*service.User, error) {
	return s.GetByID(ctx, id)
}

func cloneEmailBindUser(user *service.User) *service.User {
	if user == nil {
		return nil
	}
	cloned := *user
	return &cloned
}
