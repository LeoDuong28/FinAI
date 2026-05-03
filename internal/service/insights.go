package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// CategorySpending represents spending in a single category.
type CategorySpending struct {
	CategoryID   uuid.UUID       `json:"category_id"`
	CategoryName string          `json:"category_name"`
	Icon         *string         `json:"icon,omitempty"`
	Color        *string         `json:"color,omitempty"`
	Total        decimal.Decimal `json:"total"`
	TxnCount     int64           `json:"txn_count"`
}

// SavingsProgressResult holds aggregate savings data.
type SavingsProgressResult struct {
	TotalSaved  decimal.Decimal `json:"total_saved"`
	TotalTarget decimal.Decimal `json:"total_target"`
}

// InsightsService provides aggregated financial insights.
type InsightsService struct {
	queries   generated.Querier
	aiService *AIService
}

// NewInsightsService creates a new insights service.
func NewInsightsService(queries generated.Querier, aiService *AIService) *InsightsService {
	return &InsightsService{queries: queries, aiService: aiService}
}

// CurrentMonthBounds returns the first instant of the current month and the
// last instant of the last day (23:59:59.999999999) so that queries using
// <= comparisons include all transactions on the final day.
func CurrentMonthBounds() (time.Time, time.Time) {
	now := time.Now().UTC()
	year, month, _ := now.Date()
	from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return from, to
}

// MonthlySpending returns total debit spending for the current month.
func (s *InsightsService) MonthlySpending(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	from, to := CurrentMonthBounds()
	total, err := s.queries.SumSpendingByUser(ctx, generated.SumSpendingByUserParams{
		UserID: userID,
		Date:   from,
		Date_2: to,
	})
	if err != nil {
		return decimal.Zero, apperr.NewInternalError("failed to calculate spending")
	}
	return total, nil
}

// MonthlyIncome returns total credit income for the current month.
func (s *InsightsService) MonthlyIncome(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	from, to := CurrentMonthBounds()
	total, err := s.queries.SumIncomeByUser(ctx, generated.SumIncomeByUserParams{
		UserID: userID,
		Date:   from,
		Date_2: to,
	})
	if err != nil {
		return decimal.Zero, apperr.NewInternalError("failed to calculate income")
	}
	return total, nil
}

// SubscriptionsTotal returns normalized monthly cost of all active subscriptions.
func (s *InsightsService) SubscriptionsTotal(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	total, err := s.queries.SumActiveSubscriptions(ctx, userID)
	if err != nil {
		return decimal.Zero, apperr.NewInternalError("failed to calculate subscriptions total")
	}
	return total, nil
}

// SavingsProgress returns aggregate savings amounts.
func (s *InsightsService) SavingsProgress(ctx context.Context, userID uuid.UUID) (*SavingsProgressResult, error) {
	row, err := s.queries.SumSavingsProgress(ctx, userID)
	if err != nil {
		return nil, apperr.NewInternalError("failed to calculate savings progress")
	}
	return &SavingsProgressResult{
		TotalSaved:  row.TotalSaved,
		TotalTarget: row.TotalTarget,
	}, nil
}

// SpendingByCategory returns spending grouped by category for a date range.
func (s *InsightsService) SpendingByCategory(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]CategorySpending, error) {
	rows, err := s.queries.SpendingByCategory(ctx, generated.SpendingByCategoryParams{
		UserID: userID,
		Date:   from,
		Date_2: to,
	})
	if err != nil {
		return nil, apperr.NewInternalError("failed to get spending by category")
	}

	result := make([]CategorySpending, 0, len(rows))
	for _, r := range rows {
		result = append(result, CategorySpending{
			CategoryID:   r.CategoryID,
			CategoryName: r.CategoryName,
			Icon:         r.Icon,
			Color:        r.Color,
			Total:        r.Total,
			TxnCount:     r.TxnCount,
		})
	}
	return result, nil
}

// SpendingForecast gathers daily spending history and calls the AI service for a forecast.
func (s *InsightsService) SpendingForecast(ctx context.Context, userID uuid.UUID, days int) (*ForecastResult, error) {
	if days <= 0 || days > 90 {
		days = 30
	}

	// Get 90 days of history for the forecast model
	to := time.Now().UTC()
	from := to.AddDate(0, 0, -90)

	rows, err := s.queries.DailySpendingHistory(ctx, generated.DailySpendingHistoryParams{
		UserID: userID,
		Date:   from,
		Date_2: to,
	})
	if err != nil {
		return nil, apperr.NewInternalError("failed to get spending history")
	}

	if len(rows) < 7 {
		return nil, apperr.NewValidationError("not enough transaction history for forecast (need at least 7 days)")
	}

	history := make([]DailySpendingInput, 0, len(rows))
	for _, r := range rows {
		history = append(history, DailySpendingInput{
			Date:   r.Date.Format("2006-01-02"),
			Amount: DecimalToFloat(r.Total),
		})
	}

	forecast, err := s.aiService.ForecastSpending(ctx, history, days)
	if err != nil {
		log.Error().Err(err).Msg("AI forecast failed")
		return nil, apperr.NewInternalError("forecast temporarily unavailable")
	}
	return forecast, nil
}
