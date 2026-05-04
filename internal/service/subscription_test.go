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

func setupSubscriptionService() (*service.SubscriptionService, *mocks.MockSubscriptionRepository) {
	repo := new(mocks.MockSubscriptionRepository)
	return service.NewSubscriptionService(repo), repo
}

func TestSubscriptionService_Create_Success(t *testing.T) {
	svc, repo := setupSubscriptionService()
	ctx := context.Background()
	userID := uuid.New()

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil)

	sub, err := svc.Create(ctx, userID, service.CreateSubscriptionInput{
		Name:      "Netflix",
		Amount:    decimal.NewFromFloat(15.99),
		Frequency: "monthly",
	})

	require.NoError(t, err)
	assert.Equal(t, "Netflix", sub.Name)
	assert.Equal(t, "USD", sub.CurrencyCode) // default
	assert.Equal(t, "active", sub.Status)    // default
}

func TestSubscriptionService_Create_ValidationErrors(t *testing.T) {
	svc, _ := setupSubscriptionService()
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name  string
		input service.CreateSubscriptionInput
	}{
		{"empty name", service.CreateSubscriptionInput{Amount: decimal.NewFromInt(10), Frequency: "monthly"}},
		{"zero amount", service.CreateSubscriptionInput{Name: "Test", Amount: decimal.Zero, Frequency: "monthly"}},
		{"invalid frequency", service.CreateSubscriptionInput{Name: "Test", Amount: decimal.NewFromInt(10), Frequency: "daily"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, userID, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestSubscriptionService_Cancel_Success(t *testing.T) {
	svc, repo := setupSubscriptionService()
	ctx := context.Background()
	userID := uuid.New()
	sub := testutil.NewTestSubscription(func(s *domain.Subscription) { s.UserID = userID })

	repo.On("GetByID", ctx, userID, sub.ID).Return(sub, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil)

	err := svc.Cancel(ctx, userID, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "cancelled", sub.Status)
}

func TestSubscriptionService_Cancel_NotFound(t *testing.T) {
	svc, repo := setupSubscriptionService()
	ctx := context.Background()
	userID := uuid.New()
	subID := uuid.New()

	repo.On("GetByID", ctx, userID, subID).Return(nil, domain.ErrNotFound)

	err := svc.Cancel(ctx, userID, subID)
	assert.Error(t, err)
}

func TestSubscriptionService_Delete_Success(t *testing.T) {
	svc, repo := setupSubscriptionService()
	ctx := context.Background()
	userID := uuid.New()
	sub := testutil.NewTestSubscription(func(s *domain.Subscription) { s.UserID = userID })

	repo.On("GetByID", ctx, userID, sub.ID).Return(sub, nil)
	repo.On("Delete", ctx, userID, sub.ID).Return(nil)

	err := svc.Delete(ctx, userID, sub.ID)
	require.NoError(t, err)
}
