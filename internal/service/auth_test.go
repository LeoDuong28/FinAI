package service_test

import (
	"context"
	"testing"
	"time"

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

// fastHasher returns a PasswordHasher with minimal Argon2 params for fast tests.
func fastHasher() *service.PasswordHasher {
	return service.NewPasswordHasherWithParams(&argon2id.Params{
		Memory:      64,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  8,
		KeyLength:   16,
	})
}

func setupAuthService() (*service.AuthService, *mocks.MockUserRepository, *mocks.MockSessionRepository, *mocks.MockTokenRevoker) {
	userRepo := new(mocks.MockUserRepository)
	sessionRepo := new(mocks.MockSessionRepository)
	tokenRevoker := new(mocks.MockTokenRevoker)
	cfg := testutil.NewTestAuthConfig()

	svc := service.NewAuthService(userRepo, sessionRepo, tokenRevoker, cfg, service.WithPasswordHasher(fastHasher()))
	return svc, userRepo, sessionRepo, tokenRevoker
}

func TestAuthService_Register_Success(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "new@example.com").Return(nil, domain.ErrNotFound)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := svc.Register(ctx, service.RegisterInput{
		Email:     "new@example.com",
		Password:  "StrongP@ss1",
		FirstName: "Test",
		LastName:  "User",
	})

	require.NoError(t, err)
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "USD", user.Currency)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	svc, _, _, _ := setupAuthService()

	_, err := svc.Register(context.Background(), service.RegisterInput{
		Email:    "test@example.com",
		Password: "weak",
	})

	assert.Error(t, err)
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()
	existing := testutil.NewTestUser()

	userRepo.On("GetByEmail", ctx, "existing@example.com").Return(existing, nil)

	_, err := svc.Register(ctx, service.RegisterInput{
		Email:    "existing@example.com",
		Password: "StrongP@ss1",
	})

	assert.Error(t, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	svc, userRepo, sessionRepo, _ := setupAuthService()
	ctx := context.Background()

	h := fastHasher()
	hash, _ := h.Hash("StrongP@ss1")
	user := testutil.NewTestUser(func(u *domain.User) {
		u.PasswordHash = hash
	})

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	userRepo.On("ResetFailedLogin", ctx, user.ID).Return(nil)
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)

	tokens, err := svc.Login(ctx, service.LoginInput{
		Email:    "test@example.com",
		Password: "StrongP@ss1",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	h := fastHasher()
	hash, _ := h.Hash("StrongP@ss1")
	user := testutil.NewTestUser(func(u *domain.User) {
		u.PasswordHash = hash
	})

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	userRepo.On("IncrementFailedLogin", ctx, user.ID).Return(int32(1), nil)

	_, err := svc.Login(ctx, service.LoginInput{
		Email:    "test@example.com",
		Password: "WrongP@ss1",
	})

	assert.Error(t, err)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	userRepo.On("GetByEmail", ctx, "nobody@example.com").Return(nil, domain.ErrNotFound)

	_, err := svc.Login(ctx, service.LoginInput{
		Email:    "nobody@example.com",
		Password: "StrongP@ss1",
	})

	assert.Error(t, err)
}

func TestAuthService_Login_AccountLocked(t *testing.T) {
	svc, userRepo, _, _ := setupAuthService()
	ctx := context.Background()

	h := fastHasher()
	hash, _ := h.Hash("StrongP@ss1")
	user := testutil.NewTestUser(func(u *domain.User) {
		u.PasswordHash = hash
		locked := time.Now().UTC().Add(24 * time.Hour) // locked until tomorrow
		u.LockedUntil = &locked
	})

	userRepo.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

	_, err := svc.Login(ctx, service.LoginInput{
		Email:    "test@example.com",
		Password: "StrongP@ss1",
	})

	assert.Error(t, err)
}

func TestAuthService_Logout(t *testing.T) {
	svc, _, sessionRepo, tokenRevoker := setupAuthService()

	userID := uuid.New()
	jti := uuid.New().String()
	ctx := testutil.ContextWithUserID(userID)
	ctx = testutil.ContextWithJTI(ctx, jti)

	tokenRevoker.On("RevokeToken", ctx, mock.AnythingOfType("uuid.UUID"), userID, mock.AnythingOfType("time.Time")).Return(nil)
	sessionRepo.On("DeleteByUserID", ctx, userID).Return(nil)

	err := svc.Logout(ctx)
	require.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}
