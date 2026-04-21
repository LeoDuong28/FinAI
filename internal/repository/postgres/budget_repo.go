package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

type BudgetRepo struct {
	q *generated.Queries
}

func NewBudgetRepo(pool *pgxpool.Pool) *BudgetRepo {
	return &BudgetRepo{q: generated.New(pool)}
}

func (r *BudgetRepo) Create(ctx context.Context, budget *domain.Budget) error {
	row, err := r.q.CreateBudget(ctx, generated.CreateBudgetParams{
		UserID:         budget.UserID,
		CategoryID:     budget.CategoryID,
		Name:           budget.Name,
		AmountLimit:    budget.AmountLimit,
		Period:         budget.Period,
		StartDate:      budget.StartDate,
		AlertThreshold: budget.AlertThreshold,
	})
	if err != nil {
		return err
	}
	budget.ID = row.ID
	budget.CreatedAt = row.CreatedAt
	budget.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *BudgetRepo) GetByID(ctx context.Context, userID, budgetID uuid.UUID) (*domain.Budget, error) {
	row, err := r.q.GetBudgetByID(ctx, generated.GetBudgetByIDParams{
		ID:     budgetID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return budgetRowToDomain(row), nil
}

func (r *BudgetRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	rows, err := r.q.ListBudgetsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	budgets := make([]domain.Budget, len(rows))
	for i, row := range rows {
		budgets[i] = domain.Budget{
			ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID,
			Name: row.Name, AmountLimit: row.AmountLimit, Period: row.Period,
			StartDate: row.StartDate, AlertThreshold: row.AlertThreshold,
			IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			budgets[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				budgets[i].Category.ID = *row.CategoryID
			}
		}
	}
	return budgets, nil
}

func (r *BudgetRepo) Update(ctx context.Context, budget *domain.Budget) error {
	row, err := r.q.UpdateBudget(ctx, generated.UpdateBudgetParams{
		ID:             budget.ID,
		UserID:         budget.UserID,
		Name:           budget.Name,
		AmountLimit:    budget.AmountLimit,
		Period:         budget.Period,
		StartDate:      budget.StartDate,
		AlertThreshold: budget.AlertThreshold,
	})
	if err != nil {
		return err
	}
	budget.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *BudgetRepo) Delete(ctx context.Context, userID, budgetID uuid.UUID) error {
	return r.q.DeactivateBudget(ctx, generated.DeactivateBudgetParams{
		ID:     budgetID,
		UserID: userID,
	})
}

func budgetRowToDomain(row generated.GetBudgetByIDRow) *domain.Budget {
	b := &domain.Budget{
		ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID,
		Name: row.Name, AmountLimit: row.AmountLimit, Period: row.Period,
		StartDate: row.StartDate, AlertThreshold: row.AlertThreshold,
		IsActive: row.IsActive, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.CategoryName != nil {
		b.Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
		if row.CategoryID != nil {
			b.Category.ID = *row.CategoryID
		}
	}
	return b
}
