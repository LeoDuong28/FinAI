package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Institution struct {
	ID      uuid.UUID `json:"id"`
	PlaidID string    `json:"plaid_id"`
	Name    string    `json:"name"`
	LogoURL *string   `json:"logo_url,omitempty"`
	Color   *string   `json:"color,omitempty"`
}

type BankAccount struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	InstitutionID    *uuid.UUID      `json:"institution_id,omitempty"`
	PlaidAccountID   *string         `json:"-"` // encrypted
	PlaidAccessToken *string         `json:"-"` // encrypted
	PlaidItemID      *string         `json:"-"` // encrypted
	Name             string          `json:"name"`
	OfficialName     *string         `json:"official_name,omitempty"`
	Type             string          `json:"type"` // checking|savings|credit|investment|loan
	Subtype          *string         `json:"subtype,omitempty"`
	Mask             *string         `json:"mask,omitempty"`
	CurrentBalance   decimal.Decimal `json:"current_balance"`
	AvailableBalance *decimal.Decimal `json:"available_balance,omitempty"`
	CreditLimit      *decimal.Decimal `json:"credit_limit,omitempty"`
	CurrencyCode     string          `json:"currency_code"`
	IsActive         bool            `json:"is_active"`
	IsAsset          bool            `json:"is_asset"`
	LastSyncedAt     *time.Time      `json:"last_synced_at,omitempty"`
	SyncCursor       *string         `json:"-"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`

	// Joined data
	Institution *Institution `json:"institution,omitempty"`
}

type BankAccountRepository interface {
	Create(ctx context.Context, account *BankAccount) error
	GetByID(ctx context.Context, userID, accountID uuid.UUID) (*BankAccount, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]BankAccount, error)
	Update(ctx context.Context, account *BankAccount) error
	Delete(ctx context.Context, userID, accountID uuid.UUID) error
	CreateInstitution(ctx context.Context, inst *Institution) error
	GetForSync(ctx context.Context, userID, accountID uuid.UUID) (*BankAccount, error)
	UpdateSyncCursor(ctx context.Context, userID, accountID uuid.UUID, cursor string) error
}
