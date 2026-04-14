package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Subscription struct {
	ID                  uuid.UUID        `json:"id"`
	UserID              uuid.UUID        `json:"user_id"`
	Name                string           `json:"name"`
	MerchantName        *string          `json:"merchant_name,omitempty"`
	Amount              decimal.Decimal  `json:"amount"`
	CurrencyCode        string           `json:"currency_code"`
	Frequency           string           `json:"frequency"` // weekly|monthly|quarterly|yearly
	CategoryID          *uuid.UUID       `json:"category_id,omitempty"`
	NextBilling         *time.Time       `json:"next_billing,omitempty"`
	LastCharged         *time.Time       `json:"last_charged,omitempty"`
	Status              string           `json:"status"` // active|paused|cancelled
	AutoDetected        bool             `json:"auto_detected"`
	DetectionConfidence *decimal.Decimal `json:"detection_confidence,omitempty"`
	LogoURL             *string          `json:"logo_url,omitempty"`
	CancellationURL     *string          `json:"cancellation_url,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`

	Category *Category `json:"category,omitempty"`
}

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *Subscription) error
	GetByID(ctx context.Context, userID, subID uuid.UUID) (*Subscription, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	ListActive(ctx context.Context, userID uuid.UUID) ([]Subscription, error)
	Update(ctx context.Context, sub *Subscription) error
	Delete(ctx context.Context, userID, subID uuid.UUID) error
}
