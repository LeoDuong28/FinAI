package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	plaid "github.com/plaid/plaid-go/v31/plaid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/config"
)

// PlaidAccount represents a bank account from Plaid.
type PlaidAccount struct {
	PlaidAccountID   string
	Name             string
	OfficialName     *string
	Type             string
	Subtype          *string
	Mask             *string
	CurrentBalance   decimal.Decimal
	AvailableBalance *decimal.Decimal
	CreditLimit      *decimal.Decimal
	CurrencyCode     string
	IsAsset          bool
}

// PlaidTransaction represents a transaction from Plaid's sync endpoint.
type PlaidTransaction struct {
	PlaidTxnID   string
	AccountID    string // Plaid account ID (not our DB ID)
	Amount       decimal.Decimal
	CurrencyCode string
	Date         time.Time
	Name         string
	MerchantName *string
	Pending      bool
	Type         string // debit | credit
}

// TransactionSyncResult holds results from Plaid's /transactions/sync.
type TransactionSyncResult struct {
	Added      []PlaidTransaction
	Modified   []PlaidTransaction
	Removed    []string // plaid_txn_ids
	NextCursor string
	HasMore    bool
}

// PlaidService wraps the Plaid API client.
type PlaidService struct {
	client *plaid.APIClient
	cfg    *config.PlaidConfig
}

// NewPlaidService creates a new Plaid API client.
func NewPlaidService(cfg *config.PlaidConfig) *PlaidService {
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.ClientID)
	configuration.AddDefaultHeader("PLAID-SECRET", cfg.Secret)

	switch cfg.Env {
	case "production":
		configuration.UseEnvironment(plaid.Production)
	default:
		configuration.UseEnvironment(plaid.Sandbox)
	}

	return &PlaidService{
		client: plaid.NewAPIClient(configuration),
		cfg:    cfg,
	}
}

// CreateLinkToken generates a Plaid Link token for the given user.
func (s *PlaidService) CreateLinkToken(ctx context.Context, userID string) (string, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID,
	}

	req := plaid.NewLinkTokenCreateRequest("FinAI", "en", []plaid.CountryCode{plaid.COUNTRYCODE_US}, user)
	req.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	resp, _, err := s.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return "", fmt.Errorf("plaid link token: %w", err)
	}

	return resp.GetLinkToken(), nil
}

// ExchangePublicToken exchanges a Plaid public_token for an access_token and item_id.
func (s *PlaidService) ExchangePublicToken(ctx context.Context, publicToken string) (accessToken string, itemID string, err error) {
	req := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	resp, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*req).Execute()
	if err != nil {
		return "", "", fmt.Errorf("plaid token exchange: %w", err)
	}

	return resp.GetAccessToken(), resp.GetItemId(), nil
}

// GetAccounts fetches accounts associated with an access token.
func (s *PlaidService) GetAccounts(ctx context.Context, accessToken string) ([]PlaidAccount, error) {
	req := plaid.NewAccountsGetRequest(accessToken)
	resp, _, err := s.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid get accounts: %w", err)
	}

	accounts := make([]PlaidAccount, 0, len(resp.GetAccounts()))
	for _, a := range resp.GetAccounts() {
		account := PlaidAccount{
			PlaidAccountID: a.GetAccountId(),
			Name:           a.GetName(),
			Type:           string(a.GetType()),
			CurrencyCode:   "USD",
			IsAsset:        isAssetType(a.GetType()),
		}

		if name := a.GetOfficialName(); name != "" {
			account.OfficialName = &name
		}
		if subtype, ok := a.GetSubtypeOk(); ok && subtype != nil {
			s := string(*subtype)
			account.Subtype = &s
		}
		if mask := a.GetMask(); mask != "" {
			account.Mask = &mask
		}

		bal := a.GetBalances()
		account.CurrentBalance = decimalFromFloat64(bal.GetCurrent())

		if avail, ok := bal.GetAvailableOk(); ok && avail != nil {
			d := decimalFromFloat64(*avail)
			account.AvailableBalance = &d
		}
		if limit, ok := bal.GetLimitOk(); ok && limit != nil {
			d := decimalFromFloat64(*limit)
			account.CreditLimit = &d
		}
		if iso := bal.GetIsoCurrencyCode(); iso != "" {
			account.CurrencyCode = iso
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// SyncTransactions calls Plaid's /transactions/sync with cursor-based pagination.
func (s *PlaidService) SyncTransactions(ctx context.Context, accessToken string, cursor string) (*TransactionSyncResult, error) {
	req := plaid.NewTransactionsSyncRequest(accessToken)
	if cursor != "" {
		req.SetCursor(cursor)
	}

	resp, _, err := s.client.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*req).Execute()
	if err != nil {
		return nil, fmt.Errorf("plaid transactions sync: %w", err)
	}

	result := &TransactionSyncResult{
		NextCursor: resp.GetNextCursor(),
		HasMore:    resp.GetHasMore(),
	}

	for _, t := range resp.GetAdded() {
		txn, err := mapPlaidTransaction(t)
		if err != nil {
			log.Warn().Err(err).Msg("skipping added transaction with bad date")
			continue
		}
		result.Added = append(result.Added, txn)
	}
	for _, t := range resp.GetModified() {
		txn, err := mapPlaidTransaction(t)
		if err != nil {
			log.Warn().Err(err).Msg("skipping modified transaction with bad date")
			continue
		}
		result.Modified = append(result.Modified, txn)
	}
	for _, t := range resp.GetRemoved() {
		result.Removed = append(result.Removed, t.GetTransactionId())
	}

	return result, nil
}

// RemoveItem removes a Plaid Item (revokes access token).
func (s *PlaidService) RemoveItem(ctx context.Context, accessToken string) error {
	req := plaid.NewItemRemoveRequest(accessToken)
	_, _, err := s.client.PlaidApi.ItemRemove(ctx).ItemRemoveRequest(*req).Execute()
	if err != nil {
		return fmt.Errorf("plaid remove item: %w", err)
	}
	return nil
}

func mapPlaidTransaction(t plaid.Transaction) (PlaidTransaction, error) {
	date, err := parseDate(t.GetDate())
	if err != nil {
		return PlaidTransaction{}, fmt.Errorf("transaction %s: %w", t.GetTransactionId(), err)
	}

	txn := PlaidTransaction{
		PlaidTxnID: t.GetTransactionId(),
		AccountID:  t.GetAccountId(),
		Amount:     decimalFromFloat64(t.GetAmount()),
		Date:       date,
		Name:       t.GetName(),
		Pending:    t.GetPending(),
	}

	// Plaid: positive = money out (debit), negative = money in (credit)
	if txn.Amount.IsNegative() {
		txn.Type = "credit"
		txn.Amount = txn.Amount.Abs()
	} else {
		txn.Type = "debit"
	}

	if iso := t.GetIsoCurrencyCode(); iso != "" {
		txn.CurrencyCode = iso
	} else {
		txn.CurrencyCode = "USD"
	}

	if merchant := t.GetMerchantName(); merchant != "" {
		txn.MerchantName = &merchant
	}

	return txn, nil
}

func parseDate(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", date, err)
	}
	return t, nil
}

// decimalFromFloat64 converts a float64 to decimal.Decimal via string
// to avoid the precision loss inherent in decimal.NewFromFloat.
func decimalFromFloat64(f float64) decimal.Decimal {
	d, _ := decimal.NewFromString(strconv.FormatFloat(f, 'f', 2, 64))
	return d
}

func isAssetType(t plaid.AccountType) bool {
	switch t {
	case plaid.ACCOUNTTYPE_DEPOSITORY, plaid.ACCOUNTTYPE_INVESTMENT, plaid.ACCOUNTTYPE_BROKERAGE:
		return true
	default:
		return false // credit, loan = liability
	}
}
