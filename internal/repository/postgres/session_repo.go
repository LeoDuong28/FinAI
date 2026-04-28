package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// SessionRepo implements domain.SessionRepository using PostgreSQL.
type SessionRepo struct {
	q *generated.Queries
}

// NewSessionRepo creates a new session repository.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{
		q: generated.New(pool),
	}
}

func (r *SessionRepo) Create(ctx context.Context, session *domain.Session) error {
	row, err := r.q.CreateSession(ctx, generated.CreateSessionParams{
		UserID:       session.UserID,
		RefreshToken: session.RefreshToken,
		UserAgent:    &session.UserAgent,
		IpAddress:    &session.IPAddress,
		ExpiresAt:    session.ExpiresAt,
	})
	if err != nil {
		return err
	}
	session.ID = row.ID
	session.CreatedAt = row.CreatedAt
	return nil
}

func (r *SessionRepo) GetByToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	row, err := r.q.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	s := &domain.Session{
		ID:           row.ID,
		UserID:       row.UserID,
		RefreshToken: row.RefreshToken,
		ExpiresAt:    row.ExpiresAt,
		CreatedAt:    row.CreatedAt,
	}
	if row.UserAgent != nil {
		s.UserAgent = *row.UserAgent
	}
	if row.IpAddress != nil {
		s.IPAddress = *row.IpAddress
	}
	return s, nil
}

func (r *SessionRepo) DeleteByID(ctx context.Context, userID, id uuid.UUID) error {
	return r.q.DeleteSessionByID(ctx, generated.DeleteSessionByIDParams{
		ID:     id,
		UserID: userID,
	})
}

func (r *SessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.q.DeleteSessionsByUserID(ctx, userID)
}

func (r *SessionRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return r.q.DeleteExpiredSessions(ctx)
}
