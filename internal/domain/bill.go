package domain

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Bill struct {
	ID              uuid.UUID       `json:"id"`
	UserID          uuid.UUID       `json:"user_id"`
	Name            string          `json:"name"`
	Amount          decimal.Decimal `json:"amount"`
	DueDate         time.Time       `json:"due_date"`
	Frequency       string          `json:"frequency"` // once|monthly|quarterly|yearly
	CategoryID      *uuid.UUID      `json:"category_id,omitempty"`
	IsAutopay       bool            `json:"is_autopay"`
	Status          string          `json:"status"` // upcoming|paid|overdue
	ReminderDays    int             `json:"reminder_days"`
	NegotiationTip  *string         `json:"negotiation_tip,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`

	Category *Category `json:"category,omitempty"`
}

func (b *Bill) IsOverdue() bool {
	return b.Status != "paid" && time.Now().UTC().After(b.DueDate)
}

func (b *Bill) DaysUntilDue() int {
	hours := b.DueDate.Sub(time.Now().UTC()).Hours()
	if hours <= 0 {
		return 0
	}
	return int(math.Ceil(hours / 24))
}

func (b *Bill) ShouldRemind() bool {
	return b.Status == "upcoming" && b.DaysUntilDue() <= b.ReminderDays
}

type BillRepository interface {
	Create(ctx context.Context, bill *Bill) error
	GetByID(ctx context.Context, userID, billID uuid.UUID) (*Bill, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Bill, error)
	ListUpcoming(ctx context.Context, userID uuid.UUID, days int) ([]Bill, error)
	ListOverdue(ctx context.Context, userID uuid.UUID) ([]Bill, error)
	Update(ctx context.Context, bill *Bill) error
	UpdateStatus(ctx context.Context, userID, billID uuid.UUID, status string) error
	Delete(ctx context.Context, userID, billID uuid.UUID) error
}
