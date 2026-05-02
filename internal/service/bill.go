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

type BillService struct {
	billRepo domain.BillRepository
}

func NewBillService(billRepo domain.BillRepository) *BillService {
	return &BillService{billRepo: billRepo}
}

type CreateBillInput struct {
	Name         string
	Amount       decimal.Decimal
	DueDate      time.Time
	Frequency    string // once | monthly | quarterly | yearly
	CategoryID   *uuid.UUID
	IsAutopay    bool
	ReminderDays int
}

func (s *BillService) Create(ctx context.Context, userID uuid.UUID, input CreateBillInput) (*domain.Bill, error) {
	if input.Name == "" {
		return nil, apperr.NewValidationError("name is required")
	}
	if input.Amount.IsNegative() || input.Amount.IsZero() {
		return nil, apperr.NewValidationError("amount must be positive")
	}
	validFreq := map[string]bool{"once": true, "monthly": true, "quarterly": true, "yearly": true}
	if !validFreq[input.Frequency] {
		return nil, apperr.NewValidationError("frequency must be once, monthly, quarterly, or yearly")
	}
	if input.DueDate.IsZero() {
		return nil, apperr.NewValidationError("due date is required")
	}
	if input.ReminderDays <= 0 {
		input.ReminderDays = 3
	}

	bill := &domain.Bill{
		UserID:       userID,
		Name:         input.Name,
		Amount:       input.Amount,
		DueDate:      input.DueDate,
		Frequency:    input.Frequency,
		CategoryID:   input.CategoryID,
		IsAutopay:    input.IsAutopay,
		ReminderDays: input.ReminderDays,
	}

	if err := s.billRepo.Create(ctx, bill); err != nil {
		return nil, apperr.NewInternalError("failed to create bill")
	}
	return bill, nil
}

func (s *BillService) List(ctx context.Context, userID uuid.UUID) ([]domain.Bill, error) {
	return s.billRepo.ListByUserID(ctx, userID)
}

func (s *BillService) GetByID(ctx context.Context, userID, billID uuid.UUID) (*domain.Bill, error) {
	bill, err := s.billRepo.GetByID(ctx, userID, billID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("bill")
		}
		return nil, apperr.NewInternalError("failed to get bill")
	}
	return bill, nil
}

func (s *BillService) Update(ctx context.Context, userID, billID uuid.UUID, input CreateBillInput) (*domain.Bill, error) {
	bill, err := s.billRepo.GetByID(ctx, userID, billID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("bill")
		}
		return nil, apperr.NewInternalError("failed to get bill")
	}

	if input.Name != "" {
		bill.Name = input.Name
	}
	if !input.Amount.IsZero() {
		if input.Amount.IsNegative() {
			return nil, apperr.NewValidationError("amount must be positive")
		}
		bill.Amount = input.Amount
	}
	if !input.DueDate.IsZero() {
		bill.DueDate = input.DueDate
	}
	if input.Frequency != "" {
		validFreq := map[string]bool{"once": true, "monthly": true, "quarterly": true, "yearly": true}
		if !validFreq[input.Frequency] {
			return nil, apperr.NewValidationError("frequency must be once, monthly, quarterly, or yearly")
		}
		bill.Frequency = input.Frequency
	}
	bill.CategoryID = input.CategoryID
	bill.IsAutopay = input.IsAutopay
	if input.ReminderDays > 0 {
		bill.ReminderDays = input.ReminderDays
	}

	if err := s.billRepo.Update(ctx, bill); err != nil {
		return nil, apperr.NewInternalError("failed to update bill")
	}
	return bill, nil
}

func (s *BillService) MarkPaid(ctx context.Context, userID, billID uuid.UUID) error {
	bill, err := s.billRepo.GetByID(ctx, userID, billID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("bill")
		}
		return apperr.NewInternalError("failed to get bill")
	}

	if err := s.billRepo.UpdateStatus(ctx, userID, billID, "paid"); err != nil {
		return apperr.NewInternalError("failed to mark bill as paid")
	}

	// For recurring bills, advance the due date to the next period.
	if bill.Frequency != "once" {
		nextDue := advanceDueDate(bill.DueDate, bill.Frequency)
		bill.DueDate = nextDue
		bill.Status = "upcoming"
		if err := s.billRepo.Update(ctx, bill); err != nil {
			return apperr.NewInternalError("failed to advance bill due date")
		}
	}

	return nil
}

// advanceDueDate moves a due date forward by the bill's frequency.
func advanceDueDate(current time.Time, frequency string) time.Time {
	switch frequency {
	case "monthly":
		return current.AddDate(0, 1, 0)
	case "quarterly":
		return current.AddDate(0, 3, 0)
	case "yearly":
		return current.AddDate(1, 0, 0)
	default:
		return current
	}
}

func (s *BillService) Delete(ctx context.Context, userID, billID uuid.UUID) error {
	_, err := s.billRepo.GetByID(ctx, userID, billID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("bill")
		}
		return apperr.NewInternalError("failed to get bill")
	}
	return s.billRepo.Delete(ctx, userID, billID)
}
