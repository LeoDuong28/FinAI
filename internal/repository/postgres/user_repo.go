package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// UserRepo implements domain.UserRepository using PostgreSQL.
type UserRepo struct {
	q *generated.Queries
}

// NewUserRepo creates a new user repository.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		q: generated.New(pool),
	}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	row, err := r.q.CreateUser(ctx, generated.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Currency:     user.Currency,
		Timezone:     user.Timezone,
	})
	if err != nil {
		return err
	}
	user.ID = row.ID
	user.CreatedAt = row.CreatedAt
	user.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToUser(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToUser(row), nil
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.q.UpdateUser(ctx, generated.UpdateUserParams{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.AvatarURL,
		Theme:     user.Theme,
		Currency:  user.Currency,
		Timezone:  user.Timezone,
	})
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteUser(ctx, id)
}

func (r *UserRepo) IncrementFailedLogin(ctx context.Context, id uuid.UUID) (int32, error) {
	return r.q.IncrementFailedLogin(ctx, id)
}

func (r *UserRepo) ResetFailedLogin(ctx context.Context, id uuid.UUID) error {
	return r.q.ResetFailedLogin(ctx, id)
}

func (r *UserRepo) LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error {
	return r.q.LockAccount(ctx, generated.LockAccountParams{
		ID:          id,
		LockedUntil: &until,
	})
}

func rowToUser(row generated.User) *domain.User {
	return &domain.User{
		ID:                  row.ID,
		Email:               row.Email,
		PasswordHash:        row.PasswordHash,
		FirstName:           row.FirstName,
		LastName:            row.LastName,
		Role:                row.Role,
		EmailVerified:       row.EmailVerified,
		AvatarURL:           row.AvatarUrl,
		TOTPSecret:          row.TotpSecret,
		TOTPEnabled:         row.TotpEnabled,
		FailedLoginAttempts: int(row.FailedLoginAttempts),
		LockedUntil:         row.LockedUntil,
		OnboardingCompleted: row.OnboardingCompleted,
		Theme:               row.Theme,
		Currency:            row.Currency,
		Timezone:            row.Timezone,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
