package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Budget struct {
	ID             uuid.UUID       `json:"id"`
	UserID         uuid.UUID       `json:"user_id"`
	CategoryID     *uuid.UUID      `json:"category_id,omitempty"`
	Name           string          `json:"name"`
	AmountLimit    decimal.Decimal `json:"amount_limit"`
	Period         string          `json:"period"` // weekly|monthly|yearly
	StartDate      time.Time       `json:"start_date"`
	AlertThreshold decimal.Decimal `json:"alert_threshold"` // 0.80 = 80%
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`

	// Computed
	Category *Category       `json:"category,omitempty"`
	Spent    decimal.Decimal `json:"spent"`
}

func (b *Budget) PercentUsed() decimal.Decimal {
	if b.AmountLimit.IsZero() {
		return decimal.Zero
	}
	return b.Spent.Div(b.AmountLimit).Mul(decimal.NewFromInt(100))
}

func (b *Budget) IsOverBudget() bool {
	return b.Spent.GreaterThan(b.AmountLimit)
}

func (b *Budget) IsApproachingLimit() bool {
	if b.AmountLimit.IsZero() {
		return false
	}
	ratio := b.Spent.Div(b.AmountLimit)
	return ratio.GreaterThanOrEqual(b.AlertThreshold) && !b.IsOverBudget()
}

type BudgetRepository interface {
	Create(ctx context.Context, budget *Budget) error
	GetByID(ctx context.Context, userID, budgetID uuid.UUID) (*Budget, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, userID, budgetID uuid.UUID) error
}
