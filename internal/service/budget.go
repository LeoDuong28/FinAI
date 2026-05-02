package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

type BudgetService struct {
	budgetRepo domain.BudgetRepository
	queries    generated.Querier
}

func NewBudgetService(budgetRepo domain.BudgetRepository, queries generated.Querier) *BudgetService {
	return &BudgetService{budgetRepo: budgetRepo, queries: queries}
}

type CreateBudgetInput struct {
	CategoryID     *uuid.UUID
	Name           string
	AmountLimit    decimal.Decimal
	Period         string // weekly | monthly | yearly
	AlertThreshold decimal.Decimal
}

func (s *BudgetService) Create(ctx context.Context, userID uuid.UUID, input CreateBudgetInput) (*domain.Budget, error) {
	if input.Name == "" {
		return nil, apperr.NewValidationError("name is required")
	}
	if input.AmountLimit.IsNegative() || input.AmountLimit.IsZero() {
		return nil, apperr.NewValidationError("amount limit must be positive")
	}
	if input.Period != "weekly" && input.Period != "monthly" && input.Period != "yearly" {
		return nil, apperr.NewValidationError("period must be weekly, monthly, or yearly")
	}
	if input.AlertThreshold.IsZero() {
		input.AlertThreshold = decimal.NewFromFloat(0.80)
	}

	budget := &domain.Budget{
		UserID:         userID,
		CategoryID:     input.CategoryID,
		Name:           input.Name,
		AmountLimit:    input.AmountLimit,
		Period:         input.Period,
		StartDate:      time.Now().UTC(),
		AlertThreshold: input.AlertThreshold,
		IsActive:       true,
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, apperr.NewInternalError("failed to create budget")
	}
	return budget, nil
}

func (s *BudgetService) List(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	budgets, err := s.budgetRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, apperr.NewInternalError("failed to list budgets")
	}

	// Calculate spent amount for each budget.
	now := time.Now().UTC()
	for i := range budgets {
		from, to := periodBounds(budgets[i].Period, budgets[i].StartDate, now)
		spent, err := s.queries.SumTransactionsByUserAndDateRange(ctx, generated.SumTransactionsByUserAndDateRangeParams{
			UserID:     userID,
			CategoryID: budgets[i].CategoryID,
			Date:       from,
			Date_2:     to,
		})
		if err != nil {
			log.Error().Err(err).Str("budget_id", budgets[i].ID.String()).Msg("failed to calculate budget spent")
		} else {
			budgets[i].Spent = spent
		}
	}

	return budgets, nil
}

func (s *BudgetService) GetByID(ctx context.Context, userID, budgetID uuid.UUID) (*domain.Budget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, userID, budgetID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("budget")
		}
		return nil, apperr.NewInternalError("failed to get budget")
	}

	now := time.Now().UTC()
	from, to := periodBounds(budget.Period, budget.StartDate, now)
	spent, err := s.queries.SumTransactionsByUserAndDateRange(ctx, generated.SumTransactionsByUserAndDateRangeParams{
		UserID:     userID,
		CategoryID: budget.CategoryID,
		Date:       from,
		Date_2:     to,
	})
	if err != nil {
		log.Error().Err(err).Str("budget_id", budgetID.String()).Msg("failed to calculate budget spent")
	} else {
		budget.Spent = spent
	}

	return budget, nil
}

func (s *BudgetService) Update(ctx context.Context, userID, budgetID uuid.UUID, input CreateBudgetInput) (*domain.Budget, error) {
	budget, err := s.budgetRepo.GetByID(ctx, userID, budgetID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("budget")
		}
		return nil, apperr.NewInternalError("failed to get budget")
	}

	if input.Name != "" {
		budget.Name = input.Name
	}
	if !input.AmountLimit.IsZero() {
		if input.AmountLimit.IsNegative() {
			return nil, apperr.NewValidationError("amount limit must be positive")
		}
		budget.AmountLimit = input.AmountLimit
	}
	if input.Period != "" {
		if input.Period != "weekly" && input.Period != "monthly" && input.Period != "yearly" {
			return nil, apperr.NewValidationError("period must be weekly, monthly, or yearly")
		}
		budget.Period = input.Period
	}
	if !input.AlertThreshold.IsZero() {
		budget.AlertThreshold = input.AlertThreshold
	}

	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, apperr.NewInternalError("failed to update budget")
	}
	return budget, nil
}

func (s *BudgetService) Delete(ctx context.Context, userID, budgetID uuid.UUID) error {
	_, err := s.budgetRepo.GetByID(ctx, userID, budgetID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("budget")
		}
		return apperr.NewInternalError("failed to get budget")
	}
	return s.budgetRepo.Delete(ctx, userID, budgetID)
}

// endOfDay returns 23:59:59.999999999 on the given date.
func endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// periodBounds returns the date range for the current budget period.
// The returned `to` is end-of-day so that queries with <= comparisons include the full day.
func periodBounds(period string, startDate, now time.Time) (time.Time, time.Time) {
	switch period {
	case "weekly":
		// Find the most recent start-of-week aligned to startDate.
		// Use integer division on truncated days to avoid floating-point drift.
		daysSinceStart := int(now.Sub(startDate).Truncate(24*time.Hour) / (24 * time.Hour))
		if daysSinceStart < 0 {
			// Budget hasn't started yet; use the start date as the period.
			return startDate, endOfDay(startDate.AddDate(0, 0, 6))
		}
		weekNum := daysSinceStart / 7
		from := startDate.AddDate(0, 0, weekNum*7)
		to := from.AddDate(0, 0, 6)
		return from, endOfDay(to)
	case "yearly":
		year := now.Year()
		// Clamp day to valid range for the target month (handles Feb 29 → Feb 28 in non-leap years).
		day := startDate.Day()
		from := time.Date(year, startDate.Month(), day, 0, 0, 0, 0, time.UTC)
		// time.Date normalizes overflow (e.g. Feb 30 → Mar 2), so check and clamp.
		if from.Month() != startDate.Month() {
			// Day overflowed; use the last day of the intended month.
			from = time.Date(year, startDate.Month()+1, 0, 0, 0, 0, 0, time.UTC)
		}
		if now.Before(from) {
			from = from.AddDate(-1, 0, 0)
		}
		to := from.AddDate(1, 0, -1)
		return from, endOfDay(to)
	default: // monthly
		year, month, _ := now.Date()
		from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, -1)
		return from, endOfDay(to)
	}
}
