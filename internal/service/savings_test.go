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

func setupSavingsService() (*service.SavingsService, *mocks.MockSavingsGoalRepository) {
	repo := new(mocks.MockSavingsGoalRepository)
	return service.NewSavingsService(repo), repo
}

func TestSavingsService_Create_Success(t *testing.T) {
	svc, repo := setupSavingsService()
	ctx := context.Background()
	userID := uuid.New()

	repo.On("Create", ctx, mock.AnythingOfType("*domain.SavingsGoal")).Return(nil)

	goal, err := svc.Create(ctx, userID, service.CreateSavingsInput{
		Name:         "Emergency Fund",
		TargetAmount: decimal.NewFromInt(10000),
	})

	require.NoError(t, err)
	assert.Equal(t, "Emergency Fund", goal.Name)
}

func TestSavingsService_Create_ValidationErrors(t *testing.T) {
	svc, _ := setupSavingsService()
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name  string
		input service.CreateSavingsInput
	}{
		{"empty name", service.CreateSavingsInput{TargetAmount: decimal.NewFromInt(100)}},
		{"zero target", service.CreateSavingsInput{Name: "Test", TargetAmount: decimal.Zero}},
		{"negative target", service.CreateSavingsInput{Name: "Test", TargetAmount: decimal.NewFromInt(-100)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, userID, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestSavingsService_AddFunds_Success(t *testing.T) {
	svc, repo := setupSavingsService()
	ctx := context.Background()
	userID := uuid.New()
	goalID := uuid.New()

	updated := testutil.NewTestSavingsGoal(func(g *domain.SavingsGoal) {
		g.ID = goalID
		g.CurrentAmount = decimal.NewFromInt(3000)
	})
	repo.On("AddFunds", ctx, userID, goalID, decimal.NewFromInt(500)).Return(updated, nil)

	goal, err := svc.AddFunds(ctx, userID, goalID, decimal.NewFromInt(500))
	require.NoError(t, err)
	assert.Equal(t, "3000", goal.CurrentAmount.StringFixed(0))
}

func TestSavingsService_AddFunds_NegativeAmount(t *testing.T) {
	svc, _ := setupSavingsService()

	_, err := svc.AddFunds(context.Background(), uuid.New(), uuid.New(), decimal.NewFromInt(-10))
	assert.Error(t, err)
}

func TestSavingsService_WithdrawFunds_Success(t *testing.T) {
	svc, repo := setupSavingsService()
	ctx := context.Background()
	userID := uuid.New()
	goalID := uuid.New()

	goal := testutil.NewTestSavingsGoal(func(g *domain.SavingsGoal) {
		g.ID = goalID
		g.CurrentAmount = decimal.NewFromInt(2500)
	})
	repo.On("GetByID", ctx, userID, goalID).Return(goal, nil)

	updated := testutil.NewTestSavingsGoal(func(g *domain.SavingsGoal) {
		g.ID = goalID
		g.CurrentAmount = decimal.NewFromInt(2000)
	})
	repo.On("WithdrawFunds", ctx, userID, goalID, decimal.NewFromInt(500)).Return(updated, nil)

	result, err := svc.WithdrawFunds(ctx, userID, goalID, decimal.NewFromInt(500))
	require.NoError(t, err)
	assert.Equal(t, "2000", result.CurrentAmount.StringFixed(0))
}

func TestSavingsService_WithdrawFunds_InsufficientFunds(t *testing.T) {
	svc, repo := setupSavingsService()
	ctx := context.Background()
	userID := uuid.New()
	goalID := uuid.New()

	goal := testutil.NewTestSavingsGoal(func(g *domain.SavingsGoal) {
		g.ID = goalID
		g.CurrentAmount = decimal.NewFromInt(100)
	})
	repo.On("GetByID", ctx, userID, goalID).Return(goal, nil)
	repo.On("WithdrawFunds", ctx, userID, goalID, decimal.NewFromInt(500)).Return(nil, domain.ErrNotFound)

	_, err := svc.WithdrawFunds(ctx, userID, goalID, decimal.NewFromInt(500))
	assert.Error(t, err)
}

func TestSavingsService_Delete_NotFound(t *testing.T) {
	svc, repo := setupSavingsService()
	ctx := context.Background()
	userID := uuid.New()
	goalID := uuid.New()

	repo.On("GetByID", ctx, userID, goalID).Return(nil, domain.ErrNotFound)

	err := svc.Delete(ctx, userID, goalID)
	assert.Error(t, err)
}
