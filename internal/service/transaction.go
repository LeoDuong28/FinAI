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

type TransactionService struct {
	txnRepo domain.TransactionRepository
}

func NewTransactionService(txnRepo domain.TransactionRepository) *TransactionService {
	return &TransactionService{txnRepo: txnRepo}
}

type CreateTransactionInput struct {
	AccountID    *uuid.UUID
	CategoryID   *uuid.UUID
	Amount       decimal.Decimal
	CurrencyCode string
	Date         time.Time
	Name         string
	MerchantName *string
	Type         string // debit | credit
	Notes        *string
}

func (s *TransactionService) Create(ctx context.Context, userID uuid.UUID, input CreateTransactionInput) (*domain.Transaction, error) {
	if input.Name == "" {
		return nil, apperr.NewValidationError("name is required")
	}
	if input.Amount.IsNegative() || input.Amount.IsZero() {
		return nil, apperr.NewValidationError("amount must be positive")
	}
	if input.Type != "debit" && input.Type != "credit" {
		return nil, apperr.NewValidationError("type must be debit or credit")
	}
	if input.CurrencyCode == "" {
		input.CurrencyCode = "USD"
	}
	if input.Date.IsZero() {
		input.Date = time.Now().UTC()
	}

	txn := &domain.Transaction{
		UserID:       userID,
		AccountID:    input.AccountID,
		CategoryID:   input.CategoryID,
		Amount:       input.Amount,
		CurrencyCode: input.CurrencyCode,
		Date:         input.Date,
		Name:         input.Name,
		MerchantName: input.MerchantName,
		Type:         input.Type,
		Notes:        input.Notes,
	}

	if err := s.txnRepo.Create(ctx, txn); err != nil {
		return nil, apperr.NewInternalError("failed to create transaction")
	}
	return txn, nil
}

func (s *TransactionService) GetByID(ctx context.Context, userID, txnID uuid.UUID) (*domain.Transaction, error) {
	txn, err := s.txnRepo.GetByID(ctx, userID, txnID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("transaction")
		}
		return nil, apperr.NewInternalError("failed to get transaction")
	}
	return txn, nil
}

func (s *TransactionService) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, *domain.TransactionCursor, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	return s.txnRepo.List(ctx, userID, filter)
}

func (s *TransactionService) UpdateCategory(ctx context.Context, userID, txnID uuid.UUID, categoryID *uuid.UUID) error {
	txn, err := s.txnRepo.GetByID(ctx, userID, txnID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("transaction")
		}
		return apperr.NewInternalError("failed to get transaction")
	}
	txn.CategoryID = categoryID
	txn.AICategoryConfidence = nil // manual override
	return s.txnRepo.Update(ctx, txn)
}

func (s *TransactionService) UpdateNotes(ctx context.Context, userID, txnID uuid.UUID, notes *string) error {
	_, err := s.txnRepo.GetByID(ctx, userID, txnID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("transaction")
		}
		return apperr.NewInternalError("failed to get transaction")
	}
	return s.txnRepo.UpdateNotes(ctx, userID, txnID, notes)
}

func (s *TransactionService) Delete(ctx context.Context, userID, txnID uuid.UUID) error {
	_, err := s.txnRepo.GetByID(ctx, userID, txnID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("transaction")
		}
		return apperr.NewInternalError("failed to get transaction")
	}
	return s.txnRepo.Delete(ctx, userID, txnID)
}

func (s *TransactionService) Search(ctx context.Context, userID uuid.UUID, query string) ([]domain.Transaction, error) {
	if query == "" {
		return nil, apperr.NewValidationError("search query is required")
	}
	return s.txnRepo.Search(ctx, userID, query, 20)
}
