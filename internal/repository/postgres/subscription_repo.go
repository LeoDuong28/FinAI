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

type SubscriptionRepo struct {
	q *generated.Queries
}

func NewSubscriptionRepo(pool *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{q: generated.New(pool)}
}

func (r *SubscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	row, err := r.q.CreateSubscription(ctx, generated.CreateSubscriptionParams{
		UserID:              sub.UserID,
		Name:                sub.Name,
		MerchantName:        sub.MerchantName,
		Amount:              sub.Amount,
		CurrencyCode:        sub.CurrencyCode,
		Frequency:           sub.Frequency,
		CategoryID:          sub.CategoryID,
		NextBilling:         sub.NextBilling,
		LastCharged:         sub.LastCharged,
		Status:              sub.Status,
		AutoDetected:        sub.AutoDetected,
		DetectionConfidence: sub.DetectionConfidence,
		LogoUrl:             sub.LogoURL,
		CancellationUrl:     sub.CancellationURL,
	})
	if err != nil {
		return err
	}
	sub.ID = row.ID
	sub.CreatedAt = row.CreatedAt
	sub.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, userID, subID uuid.UUID) (*domain.Subscription, error) {
	row, err := r.q.GetSubscriptionByID(ctx, generated.GetSubscriptionByIDParams{
		ID:     subID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return subRowToDomain(row), nil
}

func (r *SubscriptionRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	rows, err := r.q.ListSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return subListToDomain(rows), nil
}

func (r *SubscriptionRepo) ListActive(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	rows, err := r.q.ListActiveSubscriptionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	subs := make([]domain.Subscription, len(rows))
	for i, row := range rows {
		subs[i] = domain.Subscription{
			ID: row.ID, UserID: row.UserID, Name: row.Name, MerchantName: row.MerchantName,
			Amount: row.Amount, CurrencyCode: row.CurrencyCode, Frequency: row.Frequency,
			CategoryID: row.CategoryID, NextBilling: row.NextBilling, LastCharged: row.LastCharged,
			Status: row.Status, AutoDetected: row.AutoDetected, DetectionConfidence: row.DetectionConfidence,
			LogoURL: row.LogoUrl, CancellationURL: row.CancellationUrl,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			subs[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				subs[i].Category.ID = *row.CategoryID
			}
		}
	}
	return subs, nil
}

func (r *SubscriptionRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	_, err := r.q.UpdateSubscription(ctx, generated.UpdateSubscriptionParams{
		ID:          sub.ID,
		UserID:      sub.UserID,
		Name:        sub.Name,
		Amount:      sub.Amount,
		Frequency:   sub.Frequency,
		CategoryID:  sub.CategoryID,
		NextBilling: sub.NextBilling,
		Status:      sub.Status,
	})
	return err
}

func (r *SubscriptionRepo) Delete(ctx context.Context, userID, subID uuid.UUID) error {
	return r.q.DeleteSubscription(ctx, generated.DeleteSubscriptionParams{
		ID:     subID,
		UserID: userID,
	})
}

func (r *SubscriptionRepo) SumMonthly(ctx context.Context, userID uuid.UUID) (string, error) {
	total, err := r.q.SumActiveSubscriptions(ctx, userID)
	if err != nil {
		return "0.00", err
	}
	return total.StringFixed(2), nil
}

func subRowToDomain(row generated.GetSubscriptionByIDRow) *domain.Subscription {
	s := &domain.Subscription{
		ID: row.ID, UserID: row.UserID, Name: row.Name, MerchantName: row.MerchantName,
		Amount: row.Amount, CurrencyCode: row.CurrencyCode, Frequency: row.Frequency,
		CategoryID: row.CategoryID, NextBilling: row.NextBilling, LastCharged: row.LastCharged,
		Status: row.Status, AutoDetected: row.AutoDetected, DetectionConfidence: row.DetectionConfidence,
		LogoURL: row.LogoUrl, CancellationURL: row.CancellationUrl,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.CategoryName != nil {
		s.Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
		if row.CategoryID != nil {
			s.Category.ID = *row.CategoryID
		}
	}
	return s
}

func subListToDomain(rows []generated.ListSubscriptionsByUserIDRow) []domain.Subscription {
	subs := make([]domain.Subscription, len(rows))
	for i, row := range rows {
		subs[i] = domain.Subscription{
			ID: row.ID, UserID: row.UserID, Name: row.Name, MerchantName: row.MerchantName,
			Amount: row.Amount, CurrencyCode: row.CurrencyCode, Frequency: row.Frequency,
			CategoryID: row.CategoryID, NextBilling: row.NextBilling, LastCharged: row.LastCharged,
			Status: row.Status, AutoDetected: row.AutoDetected, DetectionConfidence: row.DetectionConfidence,
			LogoURL: row.LogoUrl, CancellationURL: row.CancellationUrl,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			subs[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				subs[i].Category.ID = *row.CategoryID
			}
		}
	}
	return subs
}
