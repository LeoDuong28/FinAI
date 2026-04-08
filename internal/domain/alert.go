package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Alert types
const (
	AlertBudgetWarning        = "budget_warning"
	AlertBudgetExceeded       = "budget_exceeded"
	AlertLargeTransaction     = "large_txn"
	AlertLowBalance           = "low_balance"
	AlertBillDue              = "bill_due"
	AlertBillOverdue          = "bill_overdue"
	AlertUnusualSpending      = "unusual_spending"
	AlertSubscriptionDetected = "subscription_detected"
	AlertSubscriptionChange   = "subscription_price_change"
	AlertSavingsMilestone     = "savings_milestone"
	AlertWeeklySummary        = "weekly_summary"
)

// Alert severities
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

type Alert struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Message       string     `json:"message"`
	Severity      string     `json:"severity"`
	IsRead        bool       `json:"is_read"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AlertRepository interface {
	Create(ctx context.Context, alert *Alert) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]Alert, error)
	ListUnread(ctx context.Context, userID uuid.UUID) ([]Alert, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkAsRead(ctx context.Context, userID, alertID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}
