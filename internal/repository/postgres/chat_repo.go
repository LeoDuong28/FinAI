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

type ChatRepo struct {
	queries *generated.Queries
}

func NewChatRepo(pool *pgxpool.Pool) *ChatRepo {
	return &ChatRepo{queries: generated.New(pool)}
}

func (r *ChatRepo) Create(ctx context.Context, msg *domain.ChatMessage) error {
	var contextData *json.RawMessage
	if msg.ContextData != nil {
		data, err := json.Marshal(msg.ContextData)
		if err == nil {
			raw := json.RawMessage(data)
			contextData = &raw
		}
	}
	row, err := r.queries.CreateChatMessage(ctx, generated.CreateChatMessageParams{
		UserID:      msg.UserID,
		SessionID:   msg.SessionID,
		Role:        msg.Role,
		Content:     msg.Content,
		ContextData: contextData,
		TokensUsed:  msg.TokensUsed,
	})
	if err != nil {
		return err
	}
	msg.ID = row.ID
	msg.CreatedAt = row.CreatedAt
	return nil
}

func (r *ChatRepo) ListBySession(ctx context.Context, userID, sessionID uuid.UUID) ([]domain.ChatMessage, error) {
	rows, err := r.queries.ListChatMessages(ctx, generated.ListChatMessagesParams{
		UserID:    userID,
		SessionID: sessionID,
	})
	if err != nil {
		return nil, err
	}

	messages := make([]domain.ChatMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, domain.ChatMessage{
			ID:        row.ID,
			UserID:    row.UserID,
			SessionID: row.SessionID,
			Role:      row.Role,
			Content:   row.Content,
			CreatedAt: row.CreatedAt,
		})
	}
	return messages, nil
}

func (r *ChatRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.queries.DeleteChatHistory(ctx, userID)
	return err
}

// DeleteOlderThan deletes chat messages older than 90 days.
// Note: the before parameter is unused — the underlying SQL query uses a
// hardcoded 90-day interval. This satisfies the domain.ChatRepository
// interface but the retention window is not caller-configurable.
func (r *ChatRepo) DeleteOlderThan(ctx context.Context, _ time.Time) (int64, error) {
	return r.queries.DeleteOldChatMessages(ctx)
}
