package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Common domain errors.
var (
	ErrNotFound = errors.New("not found")
)

type User struct {
	ID                  uuid.UUID  `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Role                string     `json:"role"`
	EmailVerified       bool       `json:"email_verified"`
	AvatarURL           *string    `json:"avatar_url,omitempty"`
	TOTPSecret          *string    `json:"-"`
	TOTPEnabled         bool       `json:"totp_enabled"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	OnboardingCompleted bool       `json:"onboarding_completed"`
	Theme               string     `json:"theme"`
	Currency            string     `json:"currency"`
	Timezone            string     `json:"timezone"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*u.LockedUntil)
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

type Session struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	RefreshToken string    `json:"-"`
	UserAgent    string    `json:"user_agent"`
	IPAddress    string    `json:"ip_address"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementFailedLogin(ctx context.Context, id uuid.UUID) (int32, error)
	ResetFailedLogin(ctx context.Context, id uuid.UUID) error
	LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, refreshToken string) (*Session, error)
	DeleteByID(ctx context.Context, userID, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// TokenRevoker revokes JWT access tokens.
type TokenRevoker interface {
	RevokeToken(ctx context.Context, jti, userID uuid.UUID, expiresAt time.Time) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}
