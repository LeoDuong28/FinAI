package service_test

import (
	"context"
	"testing"
	"time"

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

func setupBillService() (*service.BillService, *mocks.MockBillRepository) {
	repo := new(mocks.MockBillRepository)
	return service.NewBillService(repo), repo
}

func TestBillService_Create_Success(t *testing.T) {
	svc, repo := setupBillService()
	ctx := context.Background()
	userID := uuid.New()

	repo.On("Create", ctx, mock.AnythingOfType("*domain.Bill")).Return(nil)

	bill, err := svc.Create(ctx, userID, service.CreateBillInput{
		Name:      "Electric",
		Amount:    decimal.NewFromInt(120),
		DueDate:   time.Now().Add(7 * 24 * time.Hour),
		Frequency: "monthly",
	})

	require.NoError(t, err)
	assert.Equal(t, "Electric", bill.Name)
	assert.Equal(t, 3, bill.ReminderDays) // default
}

func TestBillService_Create_ValidationErrors(t *testing.T) {
	svc, _ := setupBillService()
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name  string
		input service.CreateBillInput
	}{
		{"empty name", service.CreateBillInput{Amount: decimal.NewFromInt(100), Frequency: "monthly", DueDate: time.Now()}},
		{"zero amount", service.CreateBillInput{Name: "Test", Amount: decimal.Zero, Frequency: "monthly", DueDate: time.Now()}},
		{"invalid frequency", service.CreateBillInput{Name: "Test", Amount: decimal.NewFromInt(100), Frequency: "daily", DueDate: time.Now()}},
		{"zero due date", service.CreateBillInput{Name: "Test", Amount: decimal.NewFromInt(100), Frequency: "monthly"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(ctx, userID, tt.input)
			assert.Error(t, err)
		})
	}
}

func TestBillService_MarkPaid_OneTime(t *testing.T) {
	svc, repo := setupBillService()
	ctx := context.Background()
	userID := uuid.New()
	bill := testutil.NewTestBill(func(b *domain.Bill) {
		b.UserID = userID
		b.Frequency = "once"
	})

	repo.On("GetByID", ctx, userID, bill.ID).Return(bill, nil)
	repo.On("UpdateStatus", ctx, userID, bill.ID, "paid").Return(nil)

	err := svc.MarkPaid(ctx, userID, bill.ID)
	require.NoError(t, err)
	// Should NOT call Update for one-time bills
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestBillService_MarkPaid_Recurring(t *testing.T) {
	svc, repo := setupBillService()
	ctx := context.Background()
	userID := uuid.New()
	originalDue := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	bill := testutil.NewTestBill(func(b *domain.Bill) {
		b.UserID = userID
		b.Frequency = "monthly"
		b.DueDate = originalDue
	})

	repo.On("GetByID", ctx, userID, bill.ID).Return(bill, nil)
	repo.On("UpdateStatus", ctx, userID, bill.ID, "paid").Return(nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.Bill")).Return(nil)

	err := svc.MarkPaid(ctx, userID, bill.ID)
	require.NoError(t, err)
	repo.AssertCalled(t, "Update", ctx, mock.AnythingOfType("*domain.Bill"))

	// Verify the due date was advanced by one month
	expectedDue := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDue, bill.DueDate)
	assert.Equal(t, "upcoming", bill.Status)
}

func TestBillService_Delete_NotFound(t *testing.T) {
	svc, repo := setupBillService()
	ctx := context.Background()
	userID := uuid.New()
	billID := uuid.New()

	repo.On("GetByID", ctx, userID, billID).Return(nil, domain.ErrNotFound)

	err := svc.Delete(ctx, userID, billID)
	assert.Error(t, err)
}
