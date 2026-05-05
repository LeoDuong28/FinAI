package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// MockQuerier implements the generated.Querier interface.
// Only methods used in service tests have real implementations;
// all others return zero values to satisfy the interface.
type MockQuerier struct{ mock.Mock }

// ── Methods used by services (with mock wiring) ─────────────────

func (m *MockQuerier) SumTransactionsByUserAndDateRange(ctx context.Context, arg generated.SumTransactionsByUserAndDateRangeParams) (decimal.Decimal, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockQuerier) SumSpendingByUser(ctx context.Context, arg generated.SumSpendingByUserParams) (decimal.Decimal, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockQuerier) SumIncomeByUser(ctx context.Context, arg generated.SumIncomeByUserParams) (decimal.Decimal, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockQuerier) SumActiveSubscriptions(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockQuerier) SumSavingsProgress(ctx context.Context, userID uuid.UUID) (generated.SumSavingsProgressRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(generated.SumSavingsProgressRow), args.Error(1)
}

func (m *MockQuerier) SpendingByCategory(ctx context.Context, arg generated.SpendingByCategoryParams) ([]generated.SpendingByCategoryRow, error) {
	args := m.Called(ctx, arg)
	if r := args.Get(0); r != nil {
		return r.([]generated.SpendingByCategoryRow), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuerier) DailySpendingHistory(ctx context.Context, arg generated.DailySpendingHistoryParams) ([]generated.DailySpendingHistoryRow, error) {
	args := m.Called(ctx, arg)
	if r := args.Get(0); r != nil {
		return r.([]generated.DailySpendingHistoryRow), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuerier) SumAssetBalances(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockQuerier) SumLiabilityBalances(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(decimal.Decimal), args.Error(1)
}

func (m *MockQuerier) CreateNetworthSnapshot(ctx context.Context, arg generated.CreateNetworthSnapshotParams) (generated.NetworthSnapshot, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(generated.NetworthSnapshot), args.Error(1)
}

func (m *MockQuerier) GetLatestNetworthSnapshot(ctx context.Context, userID uuid.UUID) (generated.NetworthSnapshot, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(generated.NetworthSnapshot), args.Error(1)
}

func (m *MockQuerier) ListNetworthSnapshots(ctx context.Context, arg generated.ListNetworthSnapshotsParams) ([]generated.NetworthSnapshot, error) {
	args := m.Called(ctx, arg)
	if r := args.Get(0); r != nil {
		return r.([]generated.NetworthSnapshot), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQuerier) GetUserByID(ctx context.Context, id uuid.UUID) (generated.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(generated.User), args.Error(1)
}

func (m *MockQuerier) UpdateUser(ctx context.Context, arg generated.UpdateUserParams) (generated.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(generated.User), args.Error(1)
}

func (m *MockQuerier) UpdateUserPassword(ctx context.Context, arg generated.UpdateUserPasswordParams) error {
	return m.Called(ctx, arg).Error(0)
}

// ── Stub methods (satisfy interface, not used in tests) ─────────

func (m *MockQuerier) AddFundsToSavingsGoal(ctx context.Context, arg generated.AddFundsToSavingsGoalParams) (generated.SavingsGoal, error) {
	return generated.SavingsGoal{}, nil
}
func (m *MockQuerier) CountTransactionsByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CountUnreadAlerts(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) CreateAlert(ctx context.Context, arg generated.CreateAlertParams) (generated.Alert, error) {
	return generated.Alert{}, nil
}
func (m *MockQuerier) CreateAuditLog(ctx context.Context, arg generated.CreateAuditLogParams) (generated.AuditLog, error) {
	return generated.AuditLog{}, nil
}
func (m *MockQuerier) CreateBankAccount(ctx context.Context, arg generated.CreateBankAccountParams) (generated.BankAccount, error) {
	return generated.BankAccount{}, nil
}
func (m *MockQuerier) CreateBill(ctx context.Context, arg generated.CreateBillParams) (generated.Bill, error) {
	return generated.Bill{}, nil
}
func (m *MockQuerier) CreateBudget(ctx context.Context, arg generated.CreateBudgetParams) (generated.Budget, error) {
	return generated.Budget{}, nil
}
func (m *MockQuerier) CreateChatMessage(ctx context.Context, arg generated.CreateChatMessageParams) (generated.ChatMessage, error) {
	return generated.ChatMessage{}, nil
}
func (m *MockQuerier) CreateInstitution(ctx context.Context, arg generated.CreateInstitutionParams) (generated.Institution, error) {
	return generated.Institution{}, nil
}
func (m *MockQuerier) CreateSavingsGoal(ctx context.Context, arg generated.CreateSavingsGoalParams) (generated.SavingsGoal, error) {
	return generated.SavingsGoal{}, nil
}
func (m *MockQuerier) CreateSession(ctx context.Context, arg generated.CreateSessionParams) (generated.Session, error) {
	return generated.Session{}, nil
}
func (m *MockQuerier) CreateSubscription(ctx context.Context, arg generated.CreateSubscriptionParams) (generated.Subscription, error) {
	return generated.Subscription{}, nil
}
func (m *MockQuerier) CreateTransaction(ctx context.Context, arg generated.CreateTransactionParams) (generated.Transaction, error) {
	return generated.Transaction{}, nil
}
func (m *MockQuerier) CreateUser(ctx context.Context, arg generated.CreateUserParams) (generated.User, error) {
	return generated.User{}, nil
}
func (m *MockQuerier) DeactivateBankAccount(ctx context.Context, arg generated.DeactivateBankAccountParams) error {
	return nil
}
func (m *MockQuerier) DeactivateBudget(ctx context.Context, arg generated.DeactivateBudgetParams) error {
	return nil
}
func (m *MockQuerier) DeleteBankAccount(ctx context.Context, arg generated.DeleteBankAccountParams) error {
	return nil
}
func (m *MockQuerier) DeleteBill(ctx context.Context, arg generated.DeleteBillParams) error {
	return nil
}
func (m *MockQuerier) DeleteBudget(ctx context.Context, arg generated.DeleteBudgetParams) error {
	return nil
}
func (m *MockQuerier) DeleteChatHistory(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteExpiredIdempotencyKeys(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteExpiredRevokedTokens(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldAlerts(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldAuditLogs(ctx context.Context, createdAt time.Time) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldChatMessages(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteOldWebhooks(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) DeleteSavingsGoal(ctx context.Context, arg generated.DeleteSavingsGoalParams) error {
	return nil
}
func (m *MockQuerier) DeleteSessionByID(ctx context.Context, arg generated.DeleteSessionByIDParams) error {
	return nil
}
func (m *MockQuerier) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *MockQuerier) DeleteSubscription(ctx context.Context, arg generated.DeleteSubscriptionParams) error {
	return nil
}
func (m *MockQuerier) DeleteTransaction(ctx context.Context, arg generated.DeleteTransactionParams) error {
	return nil
}
func (m *MockQuerier) DeleteTransactionByPlaidID(ctx context.Context, arg generated.DeleteTransactionByPlaidIDParams) error {
	return nil
}
func (m *MockQuerier) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockQuerier) GetBankAccountByID(ctx context.Context, arg generated.GetBankAccountByIDParams) (generated.BankAccount, error) {
	return generated.BankAccount{}, nil
}
func (m *MockQuerier) GetBankAccountForSync(ctx context.Context, arg generated.GetBankAccountForSyncParams) (generated.BankAccount, error) {
	return generated.BankAccount{}, nil
}
func (m *MockQuerier) GetBillByID(ctx context.Context, arg generated.GetBillByIDParams) (generated.GetBillByIDRow, error) {
	return generated.GetBillByIDRow{}, nil
}
func (m *MockQuerier) GetBudgetByID(ctx context.Context, arg generated.GetBudgetByIDParams) (generated.GetBudgetByIDRow, error) {
	return generated.GetBudgetByIDRow{}, nil
}
func (m *MockQuerier) GetCategoryByID(ctx context.Context, id uuid.UUID) (generated.Category, error) {
	return generated.Category{}, nil
}
func (m *MockQuerier) GetCategoryBySlug(ctx context.Context, slug string) (generated.Category, error) {
	return generated.Category{}, nil
}
func (m *MockQuerier) GetIdempotencyKey(ctx context.Context, key string) (generated.IdempotencyKey, error) {
	return generated.IdempotencyKey{}, nil
}
func (m *MockQuerier) GetInstitutionByPlaidID(ctx context.Context, plaidID string) (generated.Institution, error) {
	return generated.Institution{}, nil
}
func (m *MockQuerier) GetSavingsGoalByID(ctx context.Context, arg generated.GetSavingsGoalByIDParams) (generated.SavingsGoal, error) {
	return generated.SavingsGoal{}, nil
}
func (m *MockQuerier) GetSessionByToken(ctx context.Context, refreshToken string) (generated.Session, error) {
	return generated.Session{}, nil
}
func (m *MockQuerier) GetSubscriptionByID(ctx context.Context, arg generated.GetSubscriptionByIDParams) (generated.GetSubscriptionByIDRow, error) {
	return generated.GetSubscriptionByIDRow{}, nil
}
func (m *MockQuerier) GetTransactionByID(ctx context.Context, arg generated.GetTransactionByIDParams) (generated.GetTransactionByIDRow, error) {
	return generated.GetTransactionByIDRow{}, nil
}
func (m *MockQuerier) GetTransactionByPlaidID(ctx context.Context, arg generated.GetTransactionByPlaidIDParams) (generated.Transaction, error) {
	return generated.Transaction{}, nil
}
func (m *MockQuerier) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
	return generated.User{}, nil
}
func (m *MockQuerier) IncrementFailedLogin(ctx context.Context, id uuid.UUID) (int32, error) {
	return 0, nil
}
func (m *MockQuerier) IsTokenRevoked(ctx context.Context, jti uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockQuerier) IsWebhookProcessed(ctx context.Context, webhookID string) (bool, error) {
	return false, nil
}
func (m *MockQuerier) ListActiveSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.ListActiveSubscriptionsByUserIDRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListAlertsByUserID(ctx context.Context, arg generated.ListAlertsByUserIDParams) ([]generated.Alert, error) {
	return nil, nil
}
func (m *MockQuerier) ListAuditLogsByUserID(ctx context.Context, arg generated.ListAuditLogsByUserIDParams) ([]generated.AuditLog, error) {
	return nil, nil
}
func (m *MockQuerier) ListBankAccountsByPlaidItemID(ctx context.Context, plaidItemID *string) ([]generated.BankAccount, error) {
	return nil, nil
}
func (m *MockQuerier) ListBankAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.ListBankAccountsByUserIDRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListBillsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.ListBillsByUserIDRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListBudgetsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.ListBudgetsByUserIDRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListCategories(ctx context.Context) ([]generated.Category, error) {
	return nil, nil
}
func (m *MockQuerier) ListChatMessages(ctx context.Context, arg generated.ListChatMessagesParams) ([]generated.ChatMessage, error) {
	return nil, nil
}
func (m *MockQuerier) ListChildCategories(ctx context.Context, parentID *uuid.UUID) ([]generated.Category, error) {
	return nil, nil
}
func (m *MockQuerier) ListOverdueBills(ctx context.Context, userID uuid.UUID) ([]generated.ListOverdueBillsRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListParentCategories(ctx context.Context) ([]generated.Category, error) {
	return nil, nil
}
func (m *MockQuerier) ListRecentChatSessions(ctx context.Context, userID uuid.UUID) ([]generated.ListRecentChatSessionsRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListSavingsGoalsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.SavingsGoal, error) {
	return nil, nil
}
func (m *MockQuerier) ListSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.ListSubscriptionsByUserIDRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListTransactions(ctx context.Context, arg generated.ListTransactionsParams) ([]generated.ListTransactionsRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListTransactionsByCategory(ctx context.Context, arg generated.ListTransactionsByCategoryParams) ([]generated.ListTransactionsByCategoryRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListTransactionsByDateRange(ctx context.Context, arg generated.ListTransactionsByDateRangeParams) ([]generated.ListTransactionsByDateRangeRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListTransactionsWithCursor(ctx context.Context, arg generated.ListTransactionsWithCursorParams) ([]generated.ListTransactionsWithCursorRow, error) {
	return nil, nil
}
func (m *MockQuerier) ListUnreadAlertsByUserID(ctx context.Context, userID uuid.UUID) ([]generated.Alert, error) {
	return nil, nil
}
func (m *MockQuerier) ListUpcomingBills(ctx context.Context, arg generated.ListUpcomingBillsParams) ([]generated.ListUpcomingBillsRow, error) {
	return nil, nil
}
func (m *MockQuerier) LockAccount(ctx context.Context, arg generated.LockAccountParams) error {
	return nil
}
func (m *MockQuerier) MarkAlertAsRead(ctx context.Context, arg generated.MarkAlertAsReadParams) error {
	return nil
}
func (m *MockQuerier) MarkAllAlertsAsRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) MarkOverdueBills(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *MockQuerier) MarkWebhookProcessed(ctx context.Context, arg generated.MarkWebhookProcessedParams) error {
	return nil
}
func (m *MockQuerier) ResetFailedLogin(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockQuerier) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *MockQuerier) RevokeToken(ctx context.Context, arg generated.RevokeTokenParams) error {
	return nil
}
func (m *MockQuerier) SaveIdempotencyKey(ctx context.Context, arg generated.SaveIdempotencyKeyParams) error {
	return nil
}
func (m *MockQuerier) SearchTransactions(ctx context.Context, arg generated.SearchTransactionsParams) ([]generated.SearchTransactionsRow, error) {
	return nil, nil
}
func (m *MockQuerier) SetTOTPSecret(ctx context.Context, arg generated.SetTOTPSecretParams) error {
	return nil
}
func (m *MockQuerier) UpdateBankAccountBalance(ctx context.Context, arg generated.UpdateBankAccountBalanceParams) error {
	return nil
}
func (m *MockQuerier) UpdateBankAccountSyncCursor(ctx context.Context, arg generated.UpdateBankAccountSyncCursorParams) error {
	return nil
}
func (m *MockQuerier) UpdateBill(ctx context.Context, arg generated.UpdateBillParams) (generated.Bill, error) {
	return generated.Bill{}, nil
}
func (m *MockQuerier) UpdateBillNegotiationTip(ctx context.Context, arg generated.UpdateBillNegotiationTipParams) error {
	return nil
}
func (m *MockQuerier) UpdateBillStatus(ctx context.Context, arg generated.UpdateBillStatusParams) error {
	return nil
}
func (m *MockQuerier) UpdateBudget(ctx context.Context, arg generated.UpdateBudgetParams) (generated.Budget, error) {
	return generated.Budget{}, nil
}
func (m *MockQuerier) UpdateSavingsGoal(ctx context.Context, arg generated.UpdateSavingsGoalParams) (generated.SavingsGoal, error) {
	return generated.SavingsGoal{}, nil
}
func (m *MockQuerier) UpdateSubscription(ctx context.Context, arg generated.UpdateSubscriptionParams) (generated.Subscription, error) {
	return generated.Subscription{}, nil
}
func (m *MockQuerier) UpdateSubscriptionLastCharged(ctx context.Context, arg generated.UpdateSubscriptionLastChargedParams) error {
	return nil
}
func (m *MockQuerier) UpdateTransactionCategory(ctx context.Context, arg generated.UpdateTransactionCategoryParams) error {
	return nil
}
func (m *MockQuerier) UpdateTransactionNotes(ctx context.Context, arg generated.UpdateTransactionNotesParams) error {
	return nil
}
func (m *MockQuerier) UpdateUserEmailVerified(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockQuerier) UpdateUserOnboardingCompleted(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *MockQuerier) UpsertTransactionByPlaidID(ctx context.Context, arg generated.UpsertTransactionByPlaidIDParams) (generated.Transaction, error) {
	return generated.Transaction{}, nil
}
func (m *MockQuerier) WithdrawFundsFromSavingsGoal(ctx context.Context, arg generated.WithdrawFundsFromSavingsGoalParams) (generated.SavingsGoal, error) {
	return generated.SavingsGoal{}, nil
}
