package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// SettingsService handles user profile and password management.
type SettingsService struct {
	userRepo domain.UserRepository
	queries  generated.Querier
	hasher   *PasswordHasher
}

// SettingsOption configures the SettingsService.
type SettingsOption func(*SettingsService)

// WithSettingsHasher overrides the default password hasher (useful for testing).
func WithSettingsHasher(h *PasswordHasher) SettingsOption {
	return func(s *SettingsService) { s.hasher = h }
}

// NewSettingsService creates a new settings service.
func NewSettingsService(userRepo domain.UserRepository, queries generated.Querier, opts ...SettingsOption) *SettingsService {
	s := &SettingsService{
		userRepo: userRepo,
		queries:  queries,
		hasher:   NewPasswordHasher(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// GetProfile returns the user's profile.
func (s *SettingsService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("user")
		}
		return nil, apperr.NewInternalError("failed to get profile")
	}
	return user, nil
}

// UpdateProfileInput holds the fields that can be updated.
type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Currency  string
	Timezone  string
	Theme     string
}

// UpdateProfile updates the user's profile fields.
func (s *SettingsService) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("user")
		}
		return nil, apperr.NewInternalError("failed to get profile")
	}

	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Currency != "" {
		user.Currency = input.Currency
	}
	if input.Timezone != "" {
		user.Timezone = input.Timezone
	}
	if input.Theme != "" {
		user.Theme = input.Theme
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, apperr.NewInternalError("failed to update profile")
	}
	return user, nil
}

// ChangePasswordInput holds old and new passwords.
type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

// ChangePassword verifies the old password and updates to the new one.
func (s *SettingsService) ChangePassword(ctx context.Context, userID uuid.UUID, input ChangePasswordInput) error {
	if input.OldPassword == "" || input.NewPassword == "" {
		return apperr.NewValidationError("old and new passwords are required")
	}

	if err := ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("user")
		}
		return apperr.NewInternalError("failed to get user")
	}

	match, err := s.hasher.Verify(input.OldPassword, user.PasswordHash)
	if err != nil || !match {
		return apperr.NewValidationError("current password is incorrect")
	}

	newHash, err := s.hasher.Hash(input.NewPassword)
	if err != nil {
		return apperr.NewInternalError("failed to process password")
	}

	if err := s.queries.UpdateUserPassword(ctx, generated.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: newHash,
	}); err != nil {
		return apperr.NewInternalError("failed to update password")
	}
	return nil
}
