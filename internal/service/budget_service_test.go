package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/service"
	"github.com/nghiaduong/finai/internal/testutil"
	"github.com/nghiaduong/finai/internal/testutil/mocks"
)

func setupBudgetService() (*service.BudgetService, *mocks.MockBudgetRepository, *mocks.MockQuerier) {
	budgetRepo := new(mocks.MockBudgetRepository)
	querier := new(mocks.MockQuerier)
	svc := service.NewBudgetService(budgetRepo, querier)
	return svc, budgetRepo, querier
}

func TestBudgetService_Create_Success(t *testing.T) {
	svc, budgetRepo, _ := setupBudgetService()
	ctx := context.Background()
	userID := uuid.New()

	budgetRepo.On("Create", ctx, mock.AnythingOfType("*domain.Budget")).Return(nil)

	budget, err := svc.Create(ctx, userID, service.CreateBudgetInput{
		Name:        "Groceries",
		AmountLimit: decimal.NewFromInt(500),
		Period:      "monthly",
	})

	require.NoError(t, err)
	assert.Equal(t, "Groceries", budget.Name)
	assert.Equal(t, decimal.NewFromFloat(0.80).String(), budget.AlertThreshold.String()) // default
}

func TestBudgetService_Create_ValidationErrors(t *testing.T) {
	svc, _, _ := setupBudgetService()
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name  string
		input service.CreateBudgetInput
	}{
		{"empty name", service.CreateBudgetInput{AmountLimit: decimal.NewFromInt(100), Period: "monthly"}},
		{"zero amount", service.CreateBudgetInput{Name: "Test", AmountLimit: decimal.Zero, Period: "monthly"}},
		{"negative amount", service.CreateBudgetInput{Name: "Test", AmountLimit: decimal.NewFromInt(-10), Period: "monthly"}},
		{"invalid period", service.CreateBudgetInput{Name: "Test", AmountLimit: decimal.NewFromInt(100), Period: "daily"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, userID, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestBudgetService_GetByID_Success(t *testing.T) {
	svc, budgetRepo, querier := setupBudgetService()
	ctx := context.Background()
	userID := uuid.New()
	budget := testutil.NewTestBudget(func(b *domain.Budget) { b.UserID = userID })

	budgetRepo.On("GetByID", ctx, userID, budget.ID).Return(budget, nil)
	querier.On("SumTransactionsByUserAndDateRange", ctx, mock.Anything).Return(decimal.NewFromInt(200), nil)

	result, err := svc.GetByID(ctx, userID, budget.ID)
	require.NoError(t, err)
	assert.Equal(t, budget.ID, result.ID)
}

func TestBudgetService_GetByID_NotFound(t *testing.T) {
	svc, budgetRepo, _ := setupBudgetService()
	ctx := context.Background()
	userID := uuid.New()
	budgetID := uuid.New()

	budgetRepo.On("GetByID", ctx, userID, budgetID).Return(nil, domain.ErrNotFound)

	_, err := svc.GetByID(ctx, userID, budgetID)
	assert.Error(t, err)
}

func TestBudgetService_Delete_Success(t *testing.T) {
	svc, budgetRepo, _ := setupBudgetService()
	ctx := context.Background()
	userID := uuid.New()
	budget := testutil.NewTestBudget(func(b *domain.Budget) { b.UserID = userID })

	budgetRepo.On("GetByID", ctx, userID, budget.ID).Return(budget, nil)
	budgetRepo.On("Delete", ctx, userID, budget.ID).Return(nil)

	err := svc.Delete(ctx, userID, budget.ID)
	require.NoError(t, err)
}

func TestBudgetService_Delete_NotFound(t *testing.T) {
	svc, budgetRepo, _ := setupBudgetService()
	ctx := context.Background()
	userID := uuid.New()
	budgetID := uuid.New()

	budgetRepo.On("GetByID", ctx, userID, budgetID).Return(nil, domain.ErrNotFound)

	err := svc.Delete(ctx, userID, budgetID)
	assert.Error(t, err)
}
