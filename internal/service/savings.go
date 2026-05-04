package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
)

type SavingsService struct {
	savingsRepo domain.SavingsGoalRepository
}

func NewSavingsService(savingsRepo domain.SavingsGoalRepository) *SavingsService {
	return &SavingsService{savingsRepo: savingsRepo}
}

type CreateSavingsInput struct {
	Name                string
	TargetAmount        decimal.Decimal
	TargetDate          *time.Time
	MonthlyContribution *decimal.Decimal
	Icon                *string
	Color               *string
}

func (s *SavingsService) Create(ctx context.Context, userID uuid.UUID, input CreateSavingsInput) (*domain.SavingsGoal, error) {
	if input.Name == "" {
		return nil, apperr.NewValidationError("name is required")
	}
	if input.TargetAmount.IsNegative() || input.TargetAmount.IsZero() {
		return nil, apperr.NewValidationError("target amount must be positive")
	}

	goal := &domain.SavingsGoal{
		UserID:              userID,
		Name:                input.Name,
		TargetAmount:        input.TargetAmount,
		TargetDate:          input.TargetDate,
		MonthlyContribution: input.MonthlyContribution,
		Icon:                input.Icon,
		Color:               input.Color,
	}

	if err := s.savingsRepo.Create(ctx, goal); err != nil {
		return nil, apperr.NewInternalError("failed to create savings goal")
	}
	return goal, nil
}

func (s *SavingsService) List(ctx context.Context, userID uuid.UUID) ([]domain.SavingsGoal, error) {
	return s.savingsRepo.ListByUserID(ctx, userID)
}

func (s *SavingsService) GetByID(ctx context.Context, userID, goalID uuid.UUID) (*domain.SavingsGoal, error) {
	goal, err := s.savingsRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("savings goal")
		}
		return nil, apperr.NewInternalError("failed to get savings goal")
	}
	return goal, nil
}

func (s *SavingsService) Update(ctx context.Context, userID, goalID uuid.UUID, input CreateSavingsInput) (*domain.SavingsGoal, error) {
	goal, err := s.savingsRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("savings goal")
		}
		return nil, apperr.NewInternalError("failed to get savings goal")
	}

	if input.Name != "" {
		goal.Name = input.Name
	}
	if !input.TargetAmount.IsZero() {
		if input.TargetAmount.IsNegative() {
			return nil, apperr.NewValidationError("target amount must be positive")
		}
		goal.TargetAmount = input.TargetAmount
	}
	goal.TargetDate = input.TargetDate
	goal.MonthlyContribution = input.MonthlyContribution
	goal.Icon = input.Icon
	goal.Color = input.Color

	if err := s.savingsRepo.Update(ctx, goal); err != nil {
		return nil, apperr.NewInternalError("failed to update savings goal")
	}
	return goal, nil
}

func (s *SavingsService) AddFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*domain.SavingsGoal, error) {
	if amount.IsNegative() || amount.IsZero() {
		return nil, apperr.NewValidationError("amount must be positive")
	}
	goal, err := s.savingsRepo.AddFunds(ctx, userID, goalID, amount)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("savings goal")
		}
		return nil, apperr.NewInternalError("failed to add funds")
	}
	return goal, nil
}

func (s *SavingsService) WithdrawFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*domain.SavingsGoal, error) {
	if amount.IsNegative() || amount.IsZero() {
		return nil, apperr.NewValidationError("amount must be positive")
	}

	// Check existence first to distinguish not-found from insufficient funds.
	_, err := s.savingsRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("savings goal")
		}
		return nil, apperr.NewInternalError("failed to get savings goal")
	}

	// Atomic withdrawal: SQL WHERE clause ensures current_amount >= amount,
	// preventing overdraw even under concurrent requests.
	goal, err := s.savingsRepo.WithdrawFunds(ctx, userID, goalID, amount)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Goal exists but WHERE current_amount >= $3 failed → insufficient funds.
			return nil, apperr.NewValidationError("insufficient funds")
		}
		return nil, apperr.NewInternalError("failed to withdraw funds")
	}
	return goal, nil
}

func (s *SavingsService) Delete(ctx context.Context, userID, goalID uuid.UUID) error {
	_, err := s.savingsRepo.GetByID(ctx, userID, goalID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("savings goal")
		}
		return apperr.NewInternalError("failed to get savings goal")
	}
	return s.savingsRepo.Delete(ctx, userID, goalID)
}
