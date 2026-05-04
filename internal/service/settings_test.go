package service_test

import (
	"context"
	"testing"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/internal/testutil"
	"github.com/nghiaduong/finai/internal/testutil/mocks"
)

func testHasher() *service.PasswordHasher {
	return service.NewPasswordHasherWithParams(&argon2id.Params{
		Memory: 64, Iterations: 1, Parallelism: 1, SaltLength: 8, KeyLength: 16,
	})
}

func setupSettingsService() (*service.SettingsService, *mocks.MockUserRepository, *mocks.MockQuerier) {
	userRepo := new(mocks.MockUserRepository)
	querier := new(mocks.MockQuerier)
	h := testHasher()
	return service.NewSettingsService(userRepo, querier, service.WithSettingsHasher(h)), userRepo, querier
}

func TestSettingsService_GetProfile_Success(t *testing.T) {
	svc, userRepo, _ := setupSettingsService()
	ctx := context.Background()
	user := testutil.NewTestUser()

	userRepo.On("GetByID", ctx, user.ID).Return(user, nil)

	result, err := svc.GetProfile(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Email, result.Email)
}

func TestSettingsService_GetProfile_NotFound(t *testing.T) {
	svc, userRepo, _ := setupSettingsService()
	ctx := context.Background()
	userID := uuid.New()

	userRepo.On("GetByID", ctx, userID).Return(nil, domain.ErrNotFound)

	_, err := svc.GetProfile(ctx, userID)
	assert.Error(t, err)
}

func TestSettingsService_UpdateProfile(t *testing.T) {
	svc, userRepo, _ := setupSettingsService()
	ctx := context.Background()
	user := testutil.NewTestUser()

	userRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	result, err := svc.UpdateProfile(ctx, user.ID, service.UpdateProfileInput{
		FirstName: "Updated",
		Currency:  "EUR",
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated", result.FirstName)
	assert.Equal(t, "EUR", result.Currency)
}

func TestSettingsService_ChangePassword_Success(t *testing.T) {
	svc, userRepo, querier := setupSettingsService()
	ctx := context.Background()

	// Hash with the same fast hasher that the service uses (injected via WithSettingsHasher)
	h := testHasher()
	hash, err := h.Hash("OldP@ssw0rd")
	require.NoError(t, err)
	user := testutil.NewTestUser(func(u *domain.User) { u.PasswordHash = hash })

	userRepo.On("GetByID", ctx, user.ID).Return(user, nil)
	querier.On("UpdateUserPassword", ctx, mock.Anything).Return(nil)

	err = svc.ChangePassword(ctx, user.ID, service.ChangePasswordInput{
		OldPassword: "OldP@ssw0rd",
		NewPassword: "NewStr0ng!Pass",
	})

	require.NoError(t, err)
	querier.AssertCalled(t, "UpdateUserPassword", ctx, mock.Anything)
}

func TestSettingsService_ChangePassword_EmptyPasswords(t *testing.T) {
	svc, _, _ := setupSettingsService()

	err := svc.ChangePassword(context.Background(), uuid.New(), service.ChangePasswordInput{})
	assert.Error(t, err)
}

func TestSettingsService_ChangePassword_WeakNewPassword(t *testing.T) {
	svc, _, _ := setupSettingsService()

	err := svc.ChangePassword(context.Background(), uuid.New(), service.ChangePasswordInput{
		OldPassword: "OldP@ss1",
		NewPassword: "weak",
	})
	assert.Error(t, err)
}
