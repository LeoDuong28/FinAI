package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nghiaduong/finai/internal/domain"
	apperr "github.com/nghiaduong/finai/internal/errors"
)

const minSyncInterval = 5 * time.Minute

// AccountService handles bank account business logic.
type AccountService struct {
	accountRepo domain.BankAccountRepository
	plaidSvc    *PlaidService
	enc         *EncryptionService
	auditRepo   domain.AuditRepository
	txnSync     *TransactionSyncService
}

// NewAccountService creates a new account service.
func NewAccountService(
	accountRepo domain.BankAccountRepository,
	plaidSvc *PlaidService,
	enc *EncryptionService,
	auditRepo domain.AuditRepository,
	txnSync *TransactionSyncService,
) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
		plaidSvc:    plaidSvc,
		enc:         enc,
		auditRepo:   auditRepo,
		txnSync:     txnSync,
	}
}

// LinkMetadata contains institution info from the Plaid Link callback.
type LinkMetadata struct {
	InstitutionID   string
	InstitutionName string
}

// CreateLinkToken generates a Plaid Link token for the user.
func (s *AccountService) CreateLinkToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := s.plaidSvc.CreateLinkToken(ctx, userID.String())
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create link token")
		return "", &apperr.DomainError{Code: apperr.CodePlaidError, Message: "Failed to initialize bank connection"}
	}
	return token, nil
}

// LinkAccount exchanges a Plaid public token, fetches accounts, encrypts credentials,
// stores them, and triggers an initial transaction sync.
func (s *AccountService) LinkAccount(ctx context.Context, userID uuid.UUID, publicToken string, metadata LinkMetadata) ([]domain.BankAccount, error) {
	// Exchange public token for access token
	accessToken, itemID, err := s.plaidSvc.ExchangePublicToken(ctx, publicToken)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to exchange public token")
		return nil, &apperr.DomainError{Code: apperr.CodePlaidError, Message: "Failed to link bank account"}
	}

	// Encrypt credentials before storing
	encAccessToken, err := s.enc.Encrypt(accessToken)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to secure credentials")
	}
	encItemID, err := s.enc.Encrypt(itemID)
	if err != nil {
		return nil, apperr.NewInternalError("Failed to secure credentials")
	}

	// Upsert institution
	inst := &domain.Institution{
		PlaidID: metadata.InstitutionID,
		Name:    metadata.InstitutionName,
	}
	if err := s.accountRepo.CreateInstitution(ctx, inst); err != nil {
		log.Error().Err(err).Msg("failed to upsert institution")
		return nil, apperr.NewInternalError("Failed to save institution")
	}

	// Fetch accounts from Plaid
	plaidAccounts, err := s.plaidSvc.GetAccounts(ctx, accessToken)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to fetch accounts from Plaid")
		return nil, &apperr.DomainError{Code: apperr.CodePlaidError, Message: "Failed to fetch account details"}
	}

	// Store each account
	var accounts []domain.BankAccount
	for _, pa := range plaidAccounts {
		encAccountID, err := s.enc.Encrypt(pa.PlaidAccountID)
		if err != nil {
			return nil, apperr.NewInternalError("Failed to secure credentials")
		}

		account := &domain.BankAccount{
			UserID:           userID,
			InstitutionID:    &inst.ID,
			PlaidAccountID:   &encAccountID,
			PlaidAccessToken: &encAccessToken,
			PlaidItemID:      &encItemID,
			Name:             pa.Name,
			OfficialName:     pa.OfficialName,
			Type:             pa.Type,
			Subtype:          pa.Subtype,
			Mask:             pa.Mask,
			CurrentBalance:   pa.CurrentBalance,
			AvailableBalance: pa.AvailableBalance,
			CreditLimit:      pa.CreditLimit,
			CurrencyCode:     pa.CurrencyCode,
			IsActive:         true,
			IsAsset:          pa.IsAsset,
		}

		if err := s.accountRepo.Create(ctx, account); err != nil {
			log.Error().Err(err).Str("user_id", userID.String()).Msg("failed to create bank account")
			return nil, apperr.NewInternalError("Failed to save account")
		}

		accounts = append(accounts, *account)
	}

	// Audit log
	entityType := "bank_account"
	s.logAudit(ctx, &userID, "link_account", &entityType, nil, fmt.Sprintf("Linked %d accounts from %s", len(accounts), metadata.InstitutionName))

	// Trigger initial transaction sync in background (detached from HTTP context).
	// Use a 5-minute timeout to prevent goroutine leaks on large accounts.
	syncAccounts := make([]domain.BankAccount, len(accounts))
	copy(syncAccounts, accounts)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		for i := range syncAccounts {
			if err := s.txnSync.SyncAccount(bgCtx, &syncAccounts[i]); err != nil {
				log.Warn().Err(err).
					Str("account_id", syncAccounts[i].ID.String()).
					Msg("initial transaction sync failed, will retry later")
			}
		}
	}()

	return accounts, nil
}

// ListAccounts returns all active accounts for a user.
func (s *AccountService) ListAccounts(ctx context.Context, userID uuid.UUID) ([]domain.BankAccount, error) {
	return s.accountRepo.ListByUserID(ctx, userID)
}

// GetAccount returns a single account for a user.
func (s *AccountService) GetAccount(ctx context.Context, userID, accountID uuid.UUID) (*domain.BankAccount, error) {
	account, err := s.accountRepo.GetByID(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, apperr.NewNotFoundError("bank account")
		}
		return nil, apperr.NewInternalError("failed to get bank account")
	}
	return account, nil
}

// SyncAccount triggers a transaction sync for a specific account.
func (s *AccountService) SyncAccount(ctx context.Context, userID, accountID uuid.UUID) error {
	account, err := s.accountRepo.GetForSync(ctx, userID, accountID)
	if err != nil {
		return err
	}

	// Rate limit: don't sync more often than every 5 minutes.
	if account.LastSyncedAt != nil && time.Since(*account.LastSyncedAt) < minSyncInterval {
		return apperr.NewValidationError("Account was synced recently, please wait a few minutes")
	}

	return s.txnSync.SyncAccount(ctx, account)
}

// UnlinkAccount removes a Plaid item and deactivates the bank account.
func (s *AccountService) UnlinkAccount(ctx context.Context, userID, accountID uuid.UUID) error {
	account, err := s.accountRepo.GetByID(ctx, userID, accountID)
	if err != nil {
		return err
	}

	// Decrypt access token and revoke Plaid access
	if account.PlaidAccessToken != nil {
		decrypted, err := s.enc.Decrypt(*account.PlaidAccessToken)
		if err != nil {
			log.Error().Err(err).Msg("failed to decrypt access token for revocation")
		} else {
			if err := s.plaidSvc.RemoveItem(ctx, decrypted); err != nil {
				log.Warn().Err(err).Str("account_id", accountID.String()).Msg("failed to revoke Plaid access")
				// Continue with deactivation even if Plaid revocation fails
			}
		}
	}

	// Soft delete the account
	if err := s.accountRepo.Delete(ctx, userID, accountID); err != nil {
		return apperr.NewInternalError("Failed to unlink account")
	}

	// Audit log
	entityType := "bank_account"
	s.logAudit(ctx, &userID, "unlink_account", &entityType, &accountID, "Unlinked bank account")

	return nil
}

func (s *AccountService) logAudit(ctx context.Context, userID *uuid.UUID, action string, entityType *string, entityID *uuid.UUID, message string) {
	auditLog := &domain.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
	}
	if message != "" {
		auditLog.Metadata = map[string]any{"message": message}
	}
	if err := s.auditRepo.Create(ctx, auditLog); err != nil {
		log.Error().Err(err).Str("action", action).Msg("failed to create audit log")
	}
}
