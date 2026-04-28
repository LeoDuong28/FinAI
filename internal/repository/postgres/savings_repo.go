package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

type SavingsRepo struct {
	q *generated.Queries
}

func NewSavingsRepo(pool *pgxpool.Pool) *SavingsRepo {
	return &SavingsRepo{q: generated.New(pool)}
}

func (r *SavingsRepo) Create(ctx context.Context, goal *domain.SavingsGoal) error {
	row, err := r.q.CreateSavingsGoal(ctx, generated.CreateSavingsGoalParams{
		UserID:              goal.UserID,
		Name:                goal.Name,
		TargetAmount:        goal.TargetAmount,
		TargetDate:          goal.TargetDate,
		MonthlyContribution: goal.MonthlyContribution,
		Icon:                goal.Icon,
		Color:               goal.Color,
	})
	if err != nil {
		return err
	}
	goal.ID = row.ID
	goal.Status = row.Status
	goal.CurrentAmount = row.CurrentAmount
	goal.CreatedAt = row.CreatedAt
	goal.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *SavingsRepo) GetByID(ctx context.Context, userID, goalID uuid.UUID) (*domain.SavingsGoal, error) {
	row, err := r.q.GetSavingsGoalByID(ctx, generated.GetSavingsGoalByIDParams{
		ID:     goalID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return savingsRowToDomain(row), nil
}

func (r *SavingsRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.SavingsGoal, error) {
	rows, err := r.q.ListSavingsGoalsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	goals := make([]domain.SavingsGoal, len(rows))
	for i, row := range rows {
		goals[i] = *savingsRowToDomain(row)
	}
	return goals, nil
}

func (r *SavingsRepo) Update(ctx context.Context, goal *domain.SavingsGoal) error {
	_, err := r.q.UpdateSavingsGoal(ctx, generated.UpdateSavingsGoalParams{
		ID:                  goal.ID,
		UserID:              goal.UserID,
		Name:                goal.Name,
		TargetAmount:        goal.TargetAmount,
		TargetDate:          goal.TargetDate,
		MonthlyContribution: goal.MonthlyContribution,
		Icon:                goal.Icon,
		Color:               goal.Color,
	})
	return err
}

func (r *SavingsRepo) AddFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*domain.SavingsGoal, error) {
	row, err := r.q.AddFundsToSavingsGoal(ctx, generated.AddFundsToSavingsGoalParams{
		ID:            goalID,
		UserID:        userID,
		CurrentAmount: amount,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return savingsRowToDomain(row), nil
}

func (r *SavingsRepo) WithdrawFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*domain.SavingsGoal, error) {
	row, err := r.q.WithdrawFundsFromSavingsGoal(ctx, generated.WithdrawFundsFromSavingsGoalParams{
		ID:            goalID,
		UserID:        userID,
		CurrentAmount: amount,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return savingsRowToDomain(row), nil
}

func (r *SavingsRepo) Delete(ctx context.Context, userID, goalID uuid.UUID) error {
	return r.q.DeleteSavingsGoal(ctx, generated.DeleteSavingsGoalParams{
		ID:     goalID,
		UserID: userID,
	})
}

func savingsRowToDomain(row generated.SavingsGoal) *domain.SavingsGoal {
	return &domain.SavingsGoal{
		ID: row.ID, UserID: row.UserID, Name: row.Name,
		TargetAmount: row.TargetAmount, CurrentAmount: row.CurrentAmount,
		TargetDate: row.TargetDate, MonthlyContribution: row.MonthlyContribution,
		Icon: row.Icon, Color: row.Color, Status: row.Status,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
