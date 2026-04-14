package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	SessionID   uuid.UUID  `json:"session_id"`
	Role        string     `json:"role"` // user | assistant
	Content     string     `json:"content"`
	ContextData any        `json:"context_data,omitempty"`
	TokensUsed  *int32     `json:"tokens_used,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ChatRepository interface {
	Create(ctx context.Context, msg *ChatMessage) error
	ListBySession(ctx context.Context, userID, sessionID uuid.UUID) ([]ChatMessage, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

type AuditLog struct {
	ID         uuid.UUID  `json:"id"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	Action     string     `json:"action"`
	EntityType *string    `json:"entity_type,omitempty"`
	EntityID   *uuid.UUID `json:"entity_id,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	Metadata   any        `json:"metadata,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]AuditLog, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
