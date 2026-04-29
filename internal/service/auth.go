package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nghiaduong/finai/internal/config"
	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/middleware"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo     domain.UserRepository
	sessionRepo  domain.SessionRepository
	tokenRevoker domain.TokenRevoker
	cfg          *config.AuthConfig
	hasher       *PasswordHasher
}

// AuthOption configures an AuthService.
type AuthOption func(*AuthService)

// WithPasswordHasher overrides the default password hasher (useful for tests).
func WithPasswordHasher(h *PasswordHasher) AuthOption {
	return func(s *AuthService) { s.hasher = h }
}

// NewAuthService creates a new auth service.
func NewAuthService(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	tokenRevoker domain.TokenRevoker,
	cfg *config.AuthConfig,
	opts ...AuthOption,
) *AuthService {
	s := &AuthService{
		userRepo:     userRepo,
		sessionRepo:  sessionRepo,
		tokenRevoker: tokenRevoker,
		cfg:          cfg,
		hasher:       NewPasswordHasher(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// RegisterInput is the input for user registration.
type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Currency  string
	Timezone  string
}

// LoginInput is the input for user login.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// TokenPair represents access and refresh tokens.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if err := ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	// Always hash the password before checking email existence.
	// This ensures consistent CPU work regardless of whether the email
	// exists, preventing a timing side-channel for email enumeration.
	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to process password")
	}

	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, apperr.NewInternalError("Failed to check existing account")
	}
	if existing != nil {
		return nil, apperr.NewValidationError("Registration failed. Please check your details and try again.")
	}

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}
	timezone := input.Timezone
	if timezone == "" {
		timezone = "America/New_York"
	}

	user := &domain.User{
		Email:        input.Email,
		PasswordHash: hash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Role:         "user",
		Currency:     currency,
		Timezone:     timezone,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, apperr.NewInternalError("Failed to create account")
	}

	return user, nil
}

// Login authenticates a user and returns a token pair.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Perform a dummy hash verification to prevent timing side-channel
			// that could reveal whether an email exists in the system.
			s.hasher.dummyVerify(input.Password)
			return nil, apperr.NewUnauthorizedError("Invalid email or password")
		}
		return nil, apperr.NewInternalError("Login failed")
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, apperr.NewAccountLockedError("Account is temporarily locked. Try again later.")
	}

	// Check if email is verified
	// TODO: enforce once email verification flow is implemented
	// if !user.EmailVerified {
	// 	return nil, apperr.NewForbiddenError("Please verify your email address before logging in")
	// }

	// Verify password
	match, err := s.hasher.Verify(input.Password, user.PasswordHash)
	if err != nil || !match {
		// Increment failed login attempts and get current count from DB
		failedAttempts, incrErr := s.userRepo.IncrementFailedLogin(ctx, user.ID)
		if incrErr != nil {
			log.Error().Err(incrErr).Str("user_id", user.ID.String()).Msg("failed to increment login attempts")
			return nil, apperr.NewUnauthorizedError("Invalid email or password")
		}

		// Lock account if too many failures (use DB count, not stale in-memory value)
		if int(failedAttempts) >= s.cfg.MaxLoginAttempts {
			lockUntil := time.Now().Add(s.cfg.LockoutDuration)
			if lockErr := s.userRepo.LockAccount(ctx, user.ID, lockUntil); lockErr != nil {
				log.Error().Err(lockErr).Str("user_id", user.ID.String()).Msg("failed to lock account")
			}
			return nil, apperr.NewAccountLockedError(
				"Too many failed attempts. Account temporarily locked. Try again later.",
			)
		}

		return nil, apperr.NewUnauthorizedError("Invalid email or password")
	}

	// Reset failed login counter on success
	if err := s.userRepo.ResetFailedLogin(ctx, user.ID); err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("failed to reset login attempts")
	}

	// Generate token pair
	tokens, err := s.generateTokenPair(ctx, user, input.UserAgent, input.IPAddress)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to create session")
	}

	return tokens, nil
}

// Logout invalidates the user's session and revokes the current access token.
func (s *AuthService) Logout(ctx context.Context) error {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return apperr.NewUnauthorizedError("Not authenticated")
	}

	// Revoke the current JWT access token so it can't be reused
	if jti, ok := middleware.GetJTI(ctx); ok {
		jtiUUID, err := uuid.Parse(jti)
		if err == nil {
			if revokeErr := s.tokenRevoker.RevokeToken(ctx, jtiUUID, userID, time.Now().Add(s.cfg.AccessTokenTTL)); revokeErr != nil {
				log.Error().Err(revokeErr).Str("user_id", userID.String()).Msg("failed to revoke access token during logout")
			}
		}
	}

	// Delete all sessions for this user (intentional: full logout across all devices).
	// TODO: for single-device logout, pass refresh token and call DeleteByID instead.
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *domain.User, userAgent, ipAddress string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.cfg.AccessTokenTTL)

	// Generate access token
	jti := uuid.New()
	accessClaims := middleware.FinAIClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti.String(),
			Subject:   user.ID.String(),
			Issuer:    "finai",
			Audience:  jwt.ClaimStrings{"finai-app"},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenStr, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshBytes := make([]byte, 32)
	if _, err := rand.Read(refreshBytes); err != nil {
		return nil, err
	}
	refreshToken := hex.EncodeToString(refreshBytes)

	// Store session
	session := &domain.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    now.Add(s.cfg.RefreshTokenTTL),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenStr,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}
