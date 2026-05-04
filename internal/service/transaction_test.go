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

func setupTransactionService() (*service.TransactionService, *mocks.MockTransactionRepository) {
	repo := new(mocks.MockTransactionRepository)
	return service.NewTransactionService(repo), repo
}

func TestTransactionService_Create_Success(t *testing.T) {
	svc, repo := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	txn, err := svc.Create(ctx, userID, service.CreateTransactionInput{
		Name:   "Coffee Shop",
		Amount: decimal.NewFromFloat(5.50),
		Type:   "debit",
	})

	require.NoError(t, err)
	assert.Equal(t, "Coffee Shop", txn.Name)
	assert.Equal(t, "USD", txn.CurrencyCode) // default
}

func TestTransactionService_Create_ValidationErrors(t *testing.T) {
	svc, _ := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name  string
		input service.CreateTransactionInput
	}{
		{"empty name", service.CreateTransactionInput{Amount: decimal.NewFromInt(10), Type: "debit"}},
		{"zero amount", service.CreateTransactionInput{Name: "Test", Amount: decimal.Zero, Type: "debit"}},
		{"invalid type", service.CreateTransactionInput{Name: "Test", Amount: decimal.NewFromInt(10), Type: "transfer"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, userID, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestTransactionService_List_DefaultLimit(t *testing.T) {
	svc, repo := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()

	// Filter with invalid limit should default to 20
	filter := domain.TransactionFilter{Limit: 0}
	repo.On("List", ctx, userID, mock.MatchedBy(func(f domain.TransactionFilter) bool {
		return f.Limit == 20
	})).Return([]domain.Transaction{}, (*domain.TransactionCursor)(nil), nil)

	_, _, err := svc.List(ctx, userID, filter)
	require.NoError(t, err)
}

func TestTransactionService_List_CapsLimit(t *testing.T) {
	svc, repo := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()

	filter := domain.TransactionFilter{Limit: 500}
	repo.On("List", ctx, userID, mock.MatchedBy(func(f domain.TransactionFilter) bool {
		return f.Limit == 100
	})).Return([]domain.Transaction{}, (*domain.TransactionCursor)(nil), nil)

	_, _, err := svc.List(ctx, userID, filter)
	require.NoError(t, err)
}

func TestTransactionService_UpdateCategory(t *testing.T) {
	svc, repo := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()
	txn := testutil.NewTestTransaction(func(tx *domain.Transaction) { tx.UserID = userID })
	catID := uuid.New()

	repo.On("GetByID", ctx, userID, txn.ID).Return(txn, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Transaction")).Return(nil)

	err := svc.UpdateCategory(ctx, userID, txn.ID, &catID)
	require.NoError(t, err)
	assert.Nil(t, txn.AICategoryConfidence) // manual override clears AI confidence
}

func TestTransactionService_Delete_NotFound(t *testing.T) {
	svc, repo := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()
	txnID := uuid.New()

	repo.On("GetByID", ctx, userID, txnID).Return(nil, domain.ErrNotFound)

	err := svc.Delete(ctx, userID, txnID)
	assert.Error(t, err)
}

func TestTransactionService_Search_EmptyQuery(t *testing.T) {
	svc, _ := setupTransactionService()

	_, err := svc.Search(context.Background(), uuid.New(), "")
	assert.Error(t, err)
}

func TestTransactionService_Search_Success(t *testing.T) {
	svc, repo := setupTransactionService()
	ctx := context.Background()
	userID := uuid.New()

	txns := []domain.Transaction{*testutil.NewTestTransaction()}
	repo.On("Search", ctx, userID, "coffee", 20).Return(txns, nil)

	result, err := svc.Search(ctx, userID, "coffee")
	require.NoError(t, err)
	assert.Len(t, result, 1)
}
