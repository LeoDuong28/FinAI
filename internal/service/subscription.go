package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
)

type SubscriptionService struct {
	subRepo domain.SubscriptionRepository
}

func NewSubscriptionService(subRepo domain.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{subRepo: subRepo}
}

type CreateSubscriptionInput struct {
	Name         string
	MerchantName *string
	Amount       decimal.Decimal
	CurrencyCode string
	Frequency    string // weekly | monthly | quarterly | yearly
	CategoryID   *uuid.UUID
	Status       string
}

func (s *SubscriptionService) Create(ctx context.Context, userID uuid.UUID, input CreateSubscriptionInput) (*domain.Subscription, error) {
	if input.Name == "" {
		return nil, apperr.NewValidationError("name is required")
	}
	if input.Amount.IsNegative() || input.Amount.IsZero() {
		return nil, apperr.NewValidationError("amount must be positive")
	}
	validFreq := map[string]bool{"weekly": true, "monthly": true, "quarterly": true, "yearly": true}
	if !validFreq[input.Frequency] {
		return nil, apperr.NewValidationError("frequency must be weekly, monthly, quarterly, or yearly")
	}
	if input.CurrencyCode == "" {
		input.CurrencyCode = "USD"
	}
	if input.Status == "" {
		input.Status = "active"
	}

	sub := &domain.Subscription{
		UserID:       userID,
		Name:         input.Name,
		MerchantName: input.MerchantName,
		Amount:       input.Amount,
		CurrencyCode: input.CurrencyCode,
		Frequency:    input.Frequency,
		CategoryID:   input.CategoryID,
		Status:       input.Status,
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, apperr.NewInternalError("failed to create subscription")
	}
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	return s.subRepo.ListByUserID(ctx, userID)
}

func (s *SubscriptionService) GetByID(ctx context.Context, userID, subID uuid.UUID) (*domain.Subscription, error) {
	sub, err := s.subRepo.GetByID(ctx, userID, subID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("subscription")
		}
		return nil, apperr.NewInternalError("failed to get subscription")
	}
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, userID, subID uuid.UUID, input CreateSubscriptionInput) (*domain.Subscription, error) {
	sub, err := s.subRepo.GetByID(ctx, userID, subID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("subscription")
		}
		return nil, apperr.NewInternalError("failed to get subscription")
	}

	if input.Name != "" {
		sub.Name = input.Name
	}
	if !input.Amount.IsZero() {
		if input.Amount.IsNegative() {
			return nil, apperr.NewValidationError("amount must be positive")
		}
		sub.Amount = input.Amount
	}
	if input.Frequency != "" {
		validFreq := map[string]bool{"weekly": true, "monthly": true, "quarterly": true, "yearly": true}
		if !validFreq[input.Frequency] {
			return nil, apperr.NewValidationError("frequency must be weekly, monthly, quarterly, or yearly")
		}
		sub.Frequency = input.Frequency
	}
	if input.Status != "" {
		sub.Status = input.Status
	}
	if input.CategoryID != nil {
		sub.CategoryID = input.CategoryID
	}

	if err := s.subRepo.Update(ctx, sub); err != nil {
		return nil, apperr.NewInternalError("failed to update subscription")
	}
	return sub, nil
}

func (s *SubscriptionService) Cancel(ctx context.Context, userID, subID uuid.UUID) error {
	sub, err := s.subRepo.GetByID(ctx, userID, subID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("subscription")
		}
		return apperr.NewInternalError("failed to get subscription")
	}
	sub.Status = "cancelled"
	return s.subRepo.Update(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, userID, subID uuid.UUID) error {
	_, err := s.subRepo.GetByID(ctx, userID, subID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return apperr.NewNotFoundError("subscription")
		}
		return apperr.NewInternalError("failed to get subscription")
	}
	return s.subRepo.Delete(ctx, userID, subID)
}
