package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type SavingsGoal struct {
	ID                   uuid.UUID        `json:"id"`
	UserID               uuid.UUID        `json:"user_id"`
	Name                 string           `json:"name"`
	TargetAmount         decimal.Decimal  `json:"target_amount"`
	CurrentAmount        decimal.Decimal  `json:"current_amount"`
	TargetDate           *time.Time       `json:"target_date,omitempty"`
	MonthlyContribution  *decimal.Decimal `json:"monthly_contribution,omitempty"`
	Icon                 *string          `json:"icon,omitempty"`
	Color                *string          `json:"color,omitempty"`
	Status               string           `json:"status"` // active|completed|cancelled
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

func (g *SavingsGoal) PercentComplete() decimal.Decimal {
	if g.TargetAmount.IsZero() {
		return decimal.Zero
	}
	return g.CurrentAmount.Div(g.TargetAmount).Mul(decimal.NewFromInt(100))
}

func (g *SavingsGoal) Remaining() decimal.Decimal {
	return g.TargetAmount.Sub(g.CurrentAmount)
}

func (g *SavingsGoal) IsComplete() bool {
	return g.CurrentAmount.GreaterThanOrEqual(g.TargetAmount)
}

type SavingsGoalRepository interface {
	Create(ctx context.Context, goal *SavingsGoal) error
	GetByID(ctx context.Context, userID, goalID uuid.UUID) (*SavingsGoal, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]SavingsGoal, error)
	Update(ctx context.Context, goal *SavingsGoal) error
	AddFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*SavingsGoal, error)
	WithdrawFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*SavingsGoal, error)
	Delete(ctx context.Context, userID, goalID uuid.UUID) error
}
