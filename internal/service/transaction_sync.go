package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
)

// TransactionSyncService handles syncing transactions from Plaid.
type TransactionSyncService struct {
	txnRepo     domain.TransactionSyncRepository
	accountRepo domain.BankAccountRepository
	plaidSvc    *PlaidService
	enc         *EncryptionService
}

// NewTransactionSyncService creates a new transaction sync service.
func NewTransactionSyncService(
	txnRepo domain.TransactionSyncRepository,
	accountRepo domain.BankAccountRepository,
	plaidSvc *PlaidService,
	enc *EncryptionService,
) *TransactionSyncService {
	return &TransactionSyncService{
		txnRepo:     txnRepo,
		accountRepo: accountRepo,
		plaidSvc:    plaidSvc,
		enc:         enc,
	}
}

// SyncAccount syncs transactions for a single bank account using Plaid's cursor-based sync.
func (s *TransactionSyncService) SyncAccount(ctx context.Context, account *domain.BankAccount) error {
	if account.PlaidAccessToken == nil || account.PlaidAccountID == nil {
		return apperr.NewValidationError("Account has no Plaid credentials")
	}

	// Decrypt access token
	accessToken, err := s.enc.Decrypt(*account.PlaidAccessToken)
	if err != nil {
		return apperr.NewInternalError("Failed to decrypt credentials")
	}

	// Get current cursor (empty string for initial sync)
	cursor := ""
	if account.SyncCursor != nil {
		cursor = *account.SyncCursor
	}

	// Decrypt Plaid account ID for filtering transactions
	var plaidAccountID string
	if account.PlaidAccountID != nil {
		decrypted, err := s.enc.Decrypt(*account.PlaidAccountID)
		if err != nil {
			return apperr.NewInternalError("Failed to decrypt account ID")
		}
		plaidAccountID = decrypted
	}

	var totalAdded, totalModified, totalRemoved int

	// Paginate through all available updates.
	// Cap iterations to prevent infinite loops from misbehaving upstream APIs.
	const maxSyncPages = 100
	for page := 0; page < maxSyncPages; page++ {
		result, err := s.plaidSvc.SyncTransactions(ctx, accessToken, cursor)
		if err != nil {
			log.Error().Err(err).
				Str("account_id", account.ID.String()).
				Msg("plaid transaction sync failed")
			return &apperr.DomainError{Code: apperr.CodePlaidError, Message: "Failed to sync transactions"}
		}

		// Track if any upserts in this batch failed — if so, don't advance cursor
		var batchFailed bool

		// Process added transactions
		for _, pt := range result.Added {
			if pt.AccountID != plaidAccountID {
				continue
			}
			if err := s.upsertTransaction(ctx, account.UserID, account.ID, pt); err != nil {
				log.Error().Err(err).
					Str("plaid_txn_id", pt.PlaidTxnID).
					Msg("failed to upsert added transaction")
				batchFailed = true
				continue
			}
			totalAdded++
		}

		// Process modified transactions
		for _, pt := range result.Modified {
			if pt.AccountID != plaidAccountID {
				continue
			}
			if err := s.upsertTransaction(ctx, account.UserID, account.ID, pt); err != nil {
				log.Error().Err(err).
					Str("plaid_txn_id", pt.PlaidTxnID).
					Msg("failed to upsert modified transaction")
				batchFailed = true
				continue
			}
			totalModified++
		}

		// Process removed transactions
		for _, plaidTxnID := range result.Removed {
			if err := s.txnRepo.DeleteByPlaidID(ctx, account.UserID, plaidTxnID); err != nil {
				log.Error().Err(err).
					Str("plaid_txn_id", plaidTxnID).
					Msg("failed to delete removed transaction")
				batchFailed = true
				continue
			}
			totalRemoved++
		}

		// If any operations in this batch failed, stop syncing and don't advance cursor.
		// The next sync attempt will replay this batch from Plaid.
		if batchFailed {
			log.Warn().
				Str("account_id", account.ID.String()).
				Msg("sync batch had failures, not advancing cursor — will retry on next sync")
			return apperr.NewInternalError("Some transactions failed to sync")
		}

		cursor = result.NextCursor

		// Update cursor after each fully-successful batch
		if err := s.accountRepo.UpdateSyncCursor(ctx, account.UserID, account.ID, cursor); err != nil {
			log.Error().Err(err).Msg("failed to update sync cursor")
			return apperr.NewInternalError("Failed to update sync state")
		}

		if !result.HasMore {
			break
		}
	}

	log.Info().
		Str("account_id", account.ID.String()).
		Int("added", totalAdded).
		Int("modified", totalModified).
		Int("removed", totalRemoved).
		Msg("transaction sync completed")

	return nil
}

func (s *TransactionSyncService) upsertTransaction(ctx context.Context, userID uuid.UUID, accountID uuid.UUID, pt PlaidTransaction) error {
	_, err := s.txnRepo.UpsertByPlaidID(ctx, domain.UpsertTransactionParams{
		UserID:       userID,
		AccountID:    &accountID,
		PlaidTxnID:   &pt.PlaidTxnID,
		Amount:       pt.Amount,
		CurrencyCode: pt.CurrencyCode,
		Date:         pt.Date,
		Name:         pt.Name,
		MerchantName: pt.MerchantName,
		Pending:      pt.Pending,
		Type:         pt.Type,
	})
	return err
}
