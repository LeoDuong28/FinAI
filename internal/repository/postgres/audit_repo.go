package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// AuditRepo implements domain.AuditRepository using PostgreSQL.
type AuditRepo struct {
	q *generated.Queries
}

// NewAuditRepo creates a new audit repository.
func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{q: generated.New(pool)}
}

func (r *AuditRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	var metadata *json.RawMessage
	if log.Metadata != nil {
		data, err := json.Marshal(log.Metadata)
		if err != nil {
			return err
		}
		raw := json.RawMessage(data)
		metadata = &raw
	}

	row, err := r.q.CreateAuditLog(ctx, generated.CreateAuditLogParams{
		UserID:     log.UserID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		IpAddress:  log.IPAddress,
		UserAgent:  log.UserAgent,
		Metadata:   metadata,
	})
	if err != nil {
		return err
	}
	log.ID = row.ID
	log.CreatedAt = row.CreatedAt
	return nil
}

func (r *AuditRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	rows, err := r.q.ListAuditLogsByUserID(ctx, generated.ListAuditLogsByUserIDParams{
		UserID: &userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	logs := make([]domain.AuditLog, len(rows))
	for i, row := range rows {
		var metadata any
		if row.Metadata != nil {
			_ = json.Unmarshal(*row.Metadata, &metadata)
		}
		logs[i] = domain.AuditLog{
			ID:         row.ID,
			UserID:     row.UserID,
			Action:     row.Action,
			EntityType: row.EntityType,
			EntityID:   row.EntityID,
			IPAddress:  row.IpAddress,
			UserAgent:  row.UserAgent,
			Metadata:   metadata,
			CreatedAt:  row.CreatedAt,
		}
	}
	return logs, nil
}

func (r *AuditRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	return r.q.DeleteOldAuditLogs(ctx, before)
}
