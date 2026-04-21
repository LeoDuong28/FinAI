package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// AccountRepo implements domain.BankAccountRepository using PostgreSQL.
type AccountRepo struct {
	q *generated.Queries
}

// NewAccountRepo creates a new account repository.
func NewAccountRepo(pool *pgxpool.Pool) *AccountRepo {
	return &AccountRepo{q: generated.New(pool)}
}

func (r *AccountRepo) Create(ctx context.Context, account *domain.BankAccount) error {
	row, err := r.q.CreateBankAccount(ctx, generated.CreateBankAccountParams{
		UserID:           account.UserID,
		InstitutionID:    account.InstitutionID,
		PlaidAccountID:   account.PlaidAccountID,
		PlaidAccessToken: account.PlaidAccessToken,
		PlaidItemID:      account.PlaidItemID,
		Name:             account.Name,
		OfficialName:     account.OfficialName,
		Type:             account.Type,
		Subtype:          account.Subtype,
		Mask:             account.Mask,
		CurrentBalance:   account.CurrentBalance,
		AvailableBalance: account.AvailableBalance,
		CreditLimit:      account.CreditLimit,
		CurrencyCode:     account.CurrencyCode,
		IsAsset:          account.IsAsset,
	})
	if err != nil {
		return err
	}
	account.ID = row.ID
	account.CreatedAt = row.CreatedAt
	account.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *AccountRepo) GetByID(ctx context.Context, userID, accountID uuid.UUID) (*domain.BankAccount, error) {
	row, err := r.q.GetBankAccountByID(ctx, generated.GetBankAccountByIDParams{
		ID:     accountID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToBankAccount(row), nil
}

func (r *AccountRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.BankAccount, error) {
	rows, err := r.q.ListBankAccountsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	accounts := make([]domain.BankAccount, len(rows))
	for i, row := range rows {
		accounts[i] = domain.BankAccount{
			ID:               row.ID,
			UserID:           row.UserID,
			InstitutionID:    row.InstitutionID,
			PlaidAccountID:   row.PlaidAccountID,
			PlaidAccessToken: row.PlaidAccessToken,
			PlaidItemID:      row.PlaidItemID,
			Name:             row.Name,
			OfficialName:     row.OfficialName,
			Type:             row.Type,
			Subtype:          row.Subtype,
			Mask:             row.Mask,
			CurrentBalance:   row.CurrentBalance,
			AvailableBalance: row.AvailableBalance,
			CreditLimit:      row.CreditLimit,
			CurrencyCode:     row.CurrencyCode,
			IsActive:         row.IsActive,
			IsAsset:          row.IsAsset,
			LastSyncedAt:     row.LastSyncedAt,
			SyncCursor:       row.SyncCursor,
			CreatedAt:        row.CreatedAt,
			UpdatedAt:        row.UpdatedAt,
		}
		if row.InstitutionName != nil {
			accounts[i].Institution = &domain.Institution{
				Name:    *row.InstitutionName,
				LogoURL: row.InstitutionLogo,
				Color:   row.InstitutionColor,
			}
		}
	}
	return accounts, nil
}

// Update persists balance changes for the given account.
// Note: only CurrentBalance and AvailableBalance are updated; other fields
// (name, type, etc.) are managed by Plaid sync and should not be user-editable.
func (r *AccountRepo) Update(ctx context.Context, account *domain.BankAccount) error {
	return r.q.UpdateBankAccountBalance(ctx, generated.UpdateBankAccountBalanceParams{
		ID:               account.ID,
		UserID:           account.UserID,
		CurrentBalance:   account.CurrentBalance,
		AvailableBalance: account.AvailableBalance,
	})
}

func (r *AccountRepo) Delete(ctx context.Context, userID, accountID uuid.UUID) error {
	return r.q.DeactivateBankAccount(ctx, generated.DeactivateBankAccountParams{
		ID:     accountID,
		UserID: userID,
	})
}

// CreateInstitution upserts an institution and returns it.
func (r *AccountRepo) CreateInstitution(ctx context.Context, inst *domain.Institution) error {
	row, err := r.q.CreateInstitution(ctx, generated.CreateInstitutionParams{
		PlaidID: inst.PlaidID,
		Name:    inst.Name,
		LogoUrl: inst.LogoURL,
		Color:   inst.Color,
	})
	if err != nil {
		return err
	}
	inst.ID = row.ID
	return nil
}

// GetForSync returns an active account for syncing.
func (r *AccountRepo) GetForSync(ctx context.Context, userID, accountID uuid.UUID) (*domain.BankAccount, error) {
	row, err := r.q.GetBankAccountForSync(ctx, generated.GetBankAccountForSyncParams{
		ID:     accountID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return rowToBankAccount(row), nil
}

// UpdateSyncCursor updates the sync cursor and last_synced_at timestamp.
func (r *AccountRepo) UpdateSyncCursor(ctx context.Context, userID, accountID uuid.UUID, cursor string) error {
	return r.q.UpdateBankAccountSyncCursor(ctx, generated.UpdateBankAccountSyncCursorParams{
		ID:         accountID,
		UserID:     userID,
		SyncCursor: &cursor,
	})
}

// UpdateBalance updates account balance fields.
func (r *AccountRepo) UpdateBalance(ctx context.Context, userID, accountID uuid.UUID, current decimal.Decimal, available *decimal.Decimal) error {
	return r.q.UpdateBankAccountBalance(ctx, generated.UpdateBankAccountBalanceParams{
		ID:               accountID,
		UserID:           userID,
		CurrentBalance:   current,
		AvailableBalance: available,
	})
}

func rowToBankAccount(row generated.BankAccount) *domain.BankAccount {
	return &domain.BankAccount{
		ID:               row.ID,
		UserID:           row.UserID,
		InstitutionID:    row.InstitutionID,
		PlaidAccountID:   row.PlaidAccountID,
		PlaidAccessToken: row.PlaidAccessToken,
		PlaidItemID:      row.PlaidItemID,
		Name:             row.Name,
		OfficialName:     row.OfficialName,
		Type:             row.Type,
		Subtype:          row.Subtype,
		Mask:             row.Mask,
		CurrentBalance:   row.CurrentBalance,
		AvailableBalance: row.AvailableBalance,
		CreditLimit:      row.CreditLimit,
		CurrencyCode:     row.CurrencyCode,
		IsActive:         row.IsActive,
		IsAsset:          row.IsAsset,
		LastSyncedAt:     row.LastSyncedAt,
		SyncCursor:       row.SyncCursor,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}
