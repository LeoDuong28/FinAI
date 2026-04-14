package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Category struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	Slug     string     `json:"slug"`
	Icon     *string    `json:"icon,omitempty"`
	Color    *string    `json:"color,omitempty"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
	IsSystem *bool      `json:"is_system"`
}

type Transaction struct {
	ID                   uuid.UUID        `json:"id"`
	UserID               uuid.UUID        `json:"user_id"`
	AccountID            *uuid.UUID       `json:"account_id,omitempty"`
	CategoryID           *uuid.UUID       `json:"category_id,omitempty"`
	PlaidTxnID           *string          `json:"-"`
	Amount               decimal.Decimal  `json:"amount"`
	CurrencyCode         string           `json:"currency_code"`
	Date                 time.Time        `json:"date"`
	Name                 string           `json:"name"`
	MerchantName         *string          `json:"merchant_name,omitempty"`
	Pending              bool             `json:"pending"`
	Type                 string           `json:"type"` // debit | credit
	Notes                *string          `json:"notes,omitempty"`
	IsExcluded           bool             `json:"is_excluded"`
	IsRecurring          bool             `json:"is_recurring"`
	AICategoryConfidence *decimal.Decimal `json:"ai_category_confidence,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`

	// Joined data
	Category *Category `json:"category,omitempty"`
}

// TransactionFilter holds query parameters for listing transactions.
type TransactionFilter struct {
	CategoryID *uuid.UUID
	AccountID  *uuid.UUID
	DateFrom   *time.Time
	DateTo     *time.Time
	AmountMin  *decimal.Decimal
	AmountMax  *decimal.Decimal
	Search     *string
	Type       *string // debit | credit
	Limit      int
	Cursor     *TransactionCursor
}

// TransactionCursor for cursor-based pagination.
type TransactionCursor struct {
	Date time.Time
	ID   uuid.UUID
}

type TransactionRepository interface {
	Create(ctx context.Context, txn *Transaction) error
	GetByID(ctx context.Context, userID, txnID uuid.UUID) (*Transaction, error)
	List(ctx context.Context, userID uuid.UUID, filter TransactionFilter) ([]Transaction, *TransactionCursor, error)
	Update(ctx context.Context, txn *Transaction) error
	UpdateNotes(ctx context.Context, userID, txnID uuid.UUID, notes *string) error
	Delete(ctx context.Context, userID, txnID uuid.UUID) error
	Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]Transaction, error)
}

// TransactionSyncRepository defines the subset of transaction operations needed for Plaid sync.
type TransactionSyncRepository interface {
	UpsertByPlaidID(ctx context.Context, params UpsertTransactionParams) (uuid.UUID, error)
	DeleteByPlaidID(ctx context.Context, userID uuid.UUID, plaidTxnID string) error
}

// UpsertTransactionParams holds parameters for upserting a Plaid transaction.
type UpsertTransactionParams struct {
	UserID       uuid.UUID
	AccountID    *uuid.UUID
	PlaidTxnID   *string
	Amount       decimal.Decimal
	CurrencyCode string
	Date         time.Time
	Name         string
	MerchantName *string
	Pending      bool
	Type         string
}

type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
}
