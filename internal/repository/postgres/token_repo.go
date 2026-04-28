package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// TokenRepo implements domain.TokenRevoker using PostgreSQL.
type TokenRepo struct {
	q *generated.Queries
}

// NewTokenRepo creates a new token repository.
func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{
		q: generated.New(pool),
	}
}

func (r *TokenRepo) RevokeToken(ctx context.Context, jti, userID uuid.UUID, expiresAt time.Time) error {
	return r.q.RevokeToken(ctx, generated.RevokeTokenParams{
		Jti:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})
}

func (r *TokenRepo) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return r.q.RevokeAllUserTokens(ctx, userID)
}
