package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"github.com/nghiaduong/finai/internal/domain"
)

// ── UserRepository ───────────────────────────────────────────────

type MockUserRepository struct{ mock.Mock }

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if u := args.Get(0); u != nil {
		return u.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepository) IncrementFailedLogin(ctx context.Context, id uuid.UUID) (int32, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int32), args.Error(1)
}
func (m *MockUserRepository) ResetFailedLogin(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepository) LockAccount(ctx context.Context, id uuid.UUID, until time.Time) error {
	return m.Called(ctx, id, until).Error(0)
}

// ── SessionRepository ────────────────────────────────────────────

type MockSessionRepository struct{ mock.Mock }

func (m *MockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}
func (m *MockSessionRepository) GetByToken(ctx context.Context, refreshToken string) (*domain.Session, error) {
	args := m.Called(ctx, refreshToken)
	if s := args.Get(0); s != nil {
		return s.(*domain.Session), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSessionRepository) DeleteByID(ctx context.Context, userID, id uuid.UUID) error {
	return m.Called(ctx, userID, id).Error(0)
}
func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *MockSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// ── TokenRevoker ─────────────────────────────────────────────────

type MockTokenRevoker struct{ mock.Mock }

func (m *MockTokenRevoker) RevokeToken(ctx context.Context, jti, userID uuid.UUID, expiresAt time.Time) error {
	return m.Called(ctx, jti, userID, expiresAt).Error(0)
}
func (m *MockTokenRevoker) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

// ── BankAccountRepository ────────────────────────────────────────

type MockBankAccountRepository struct{ mock.Mock }

func (m *MockBankAccountRepository) Create(ctx context.Context, account *domain.BankAccount) error {
	return m.Called(ctx, account).Error(0)
}
func (m *MockBankAccountRepository) GetByID(ctx context.Context, userID, accountID uuid.UUID) (*domain.BankAccount, error) {
	args := m.Called(ctx, userID, accountID)
	if a := args.Get(0); a != nil {
		return a.(*domain.BankAccount), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBankAccountRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.BankAccount, error) {
	args := m.Called(ctx, userID)
	if a := args.Get(0); a != nil {
		return a.([]domain.BankAccount), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBankAccountRepository) Update(ctx context.Context, account *domain.BankAccount) error {
	return m.Called(ctx, account).Error(0)
}
func (m *MockBankAccountRepository) Delete(ctx context.Context, userID, accountID uuid.UUID) error {
	return m.Called(ctx, userID, accountID).Error(0)
}
func (m *MockBankAccountRepository) CreateInstitution(ctx context.Context, inst *domain.Institution) error {
	return m.Called(ctx, inst).Error(0)
}
func (m *MockBankAccountRepository) GetForSync(ctx context.Context, userID, accountID uuid.UUID) (*domain.BankAccount, error) {
	args := m.Called(ctx, userID, accountID)
	if a := args.Get(0); a != nil {
		return a.(*domain.BankAccount), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBankAccountRepository) UpdateSyncCursor(ctx context.Context, userID, accountID uuid.UUID, cursor string) error {
	return m.Called(ctx, userID, accountID, cursor).Error(0)
}

// ── TransactionRepository ────────────────────────────────────────

type MockTransactionRepository struct{ mock.Mock }

func (m *MockTransactionRepository) Create(ctx context.Context, txn *domain.Transaction) error {
	return m.Called(ctx, txn).Error(0)
}
func (m *MockTransactionRepository) GetByID(ctx context.Context, userID, txnID uuid.UUID) (*domain.Transaction, error) {
	args := m.Called(ctx, userID, txnID)
	if t := args.Get(0); t != nil {
		return t.(*domain.Transaction), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTransactionRepository) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, *domain.TransactionCursor, error) {
	args := m.Called(ctx, userID, filter)
	var txns []domain.Transaction
	if t := args.Get(0); t != nil {
		txns = t.([]domain.Transaction)
	}
	var cursor *domain.TransactionCursor
	if c := args.Get(1); c != nil {
		cursor = c.(*domain.TransactionCursor)
	}
	return txns, cursor, args.Error(2)
}
func (m *MockTransactionRepository) Update(ctx context.Context, txn *domain.Transaction) error {
	return m.Called(ctx, txn).Error(0)
}
func (m *MockTransactionRepository) UpdateNotes(ctx context.Context, userID, txnID uuid.UUID, notes *string) error {
	return m.Called(ctx, userID, txnID, notes).Error(0)
}
func (m *MockTransactionRepository) Delete(ctx context.Context, userID, txnID uuid.UUID) error {
	return m.Called(ctx, userID, txnID).Error(0)
}
func (m *MockTransactionRepository) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]domain.Transaction, error) {
	args := m.Called(ctx, userID, query, limit)
	if t := args.Get(0); t != nil {
		return t.([]domain.Transaction), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── TransactionSyncRepository ────────────────────────────────────

type MockTransactionSyncRepository struct{ mock.Mock }

func (m *MockTransactionSyncRepository) UpsertByPlaidID(ctx context.Context, params domain.UpsertTransactionParams) (uuid.UUID, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(uuid.UUID), args.Error(1)
}
func (m *MockTransactionSyncRepository) DeleteByPlaidID(ctx context.Context, userID uuid.UUID, plaidTxnID string) error {
	return m.Called(ctx, userID, plaidTxnID).Error(0)
}

// ── CategoryRepository ───────────────────────────────────────────

type MockCategoryRepository struct{ mock.Mock }

func (m *MockCategoryRepository) List(ctx context.Context) ([]domain.Category, error) {
	args := m.Called(ctx)
	if c := args.Get(0); c != nil {
		return c.([]domain.Category), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockCategoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	args := m.Called(ctx, slug)
	if c := args.Get(0); c != nil {
		return c.(*domain.Category), args.Error(1)
	}
	return nil, args.Error(1)
}

// ── BudgetRepository ─────────────────────────────────────────────

type MockBudgetRepository struct{ mock.Mock }

func (m *MockBudgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	return m.Called(ctx, budget).Error(0)
}
func (m *MockBudgetRepository) GetByID(ctx context.Context, userID, budgetID uuid.UUID) (*domain.Budget, error) {
	args := m.Called(ctx, userID, budgetID)
	if b := args.Get(0); b != nil {
		return b.(*domain.Budget), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBudgetRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Budget, error) {
	args := m.Called(ctx, userID)
	if b := args.Get(0); b != nil {
		return b.([]domain.Budget), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBudgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	return m.Called(ctx, budget).Error(0)
}
func (m *MockBudgetRepository) Delete(ctx context.Context, userID, budgetID uuid.UUID) error {
	return m.Called(ctx, userID, budgetID).Error(0)
}

// ── BillRepository ───────────────────────────────────────────────

type MockBillRepository struct{ mock.Mock }

func (m *MockBillRepository) Create(ctx context.Context, bill *domain.Bill) error {
	return m.Called(ctx, bill).Error(0)
}
func (m *MockBillRepository) GetByID(ctx context.Context, userID, billID uuid.UUID) (*domain.Bill, error) {
	args := m.Called(ctx, userID, billID)
	if b := args.Get(0); b != nil {
		return b.(*domain.Bill), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Bill, error) {
	args := m.Called(ctx, userID)
	if b := args.Get(0); b != nil {
		return b.([]domain.Bill), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillRepository) ListUpcoming(ctx context.Context, userID uuid.UUID, days int) ([]domain.Bill, error) {
	args := m.Called(ctx, userID, days)
	if b := args.Get(0); b != nil {
		return b.([]domain.Bill), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillRepository) ListOverdue(ctx context.Context, userID uuid.UUID) ([]domain.Bill, error) {
	args := m.Called(ctx, userID)
	if b := args.Get(0); b != nil {
		return b.([]domain.Bill), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockBillRepository) Update(ctx context.Context, bill *domain.Bill) error {
	return m.Called(ctx, bill).Error(0)
}
func (m *MockBillRepository) UpdateStatus(ctx context.Context, userID, billID uuid.UUID, status string) error {
	return m.Called(ctx, userID, billID, status).Error(0)
}
func (m *MockBillRepository) Delete(ctx context.Context, userID, billID uuid.UUID) error {
	return m.Called(ctx, userID, billID).Error(0)
}

// ── SubscriptionRepository ───────────────────────────────────────

type MockSubscriptionRepository struct{ mock.Mock }

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockSubscriptionRepository) GetByID(ctx context.Context, userID, subID uuid.UUID) (*domain.Subscription, error) {
	args := m.Called(ctx, userID, subID)
	if s := args.Get(0); s != nil {
		return s.(*domain.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	args := m.Called(ctx, userID)
	if s := args.Get(0); s != nil {
		return s.([]domain.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) ListActive(ctx context.Context, userID uuid.UUID) ([]domain.Subscription, error) {
	args := m.Called(ctx, userID)
	if s := args.Get(0); s != nil {
		return s.([]domain.Subscription), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *MockSubscriptionRepository) Delete(ctx context.Context, userID, subID uuid.UUID) error {
	return m.Called(ctx, userID, subID).Error(0)
}

// ── SavingsGoalRepository ────────────────────────────────────────

type MockSavingsGoalRepository struct{ mock.Mock }

func (m *MockSavingsGoalRepository) Create(ctx context.Context, goal *domain.SavingsGoal) error {
	return m.Called(ctx, goal).Error(0)
}
func (m *MockSavingsGoalRepository) GetByID(ctx context.Context, userID, goalID uuid.UUID) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, userID, goalID)
	if g := args.Get(0); g != nil {
		return g.(*domain.SavingsGoal), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSavingsGoalRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.SavingsGoal, error) {
	args := m.Called(ctx, userID)
	if g := args.Get(0); g != nil {
		return g.([]domain.SavingsGoal), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSavingsGoalRepository) Update(ctx context.Context, goal *domain.SavingsGoal) error {
	return m.Called(ctx, goal).Error(0)
}
func (m *MockSavingsGoalRepository) AddFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, userID, goalID, amount)
	if g := args.Get(0); g != nil {
		return g.(*domain.SavingsGoal), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSavingsGoalRepository) WithdrawFunds(ctx context.Context, userID, goalID uuid.UUID, amount decimal.Decimal) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, userID, goalID, amount)
	if g := args.Get(0); g != nil {
		return g.(*domain.SavingsGoal), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSavingsGoalRepository) Delete(ctx context.Context, userID, goalID uuid.UUID) error {
	return m.Called(ctx, userID, goalID).Error(0)
}

// ── AlertRepository ──────────────────────────────────────────────

type MockAlertRepository struct{ mock.Mock }

func (m *MockAlertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	return m.Called(ctx, alert).Error(0)
}
func (m *MockAlertRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]domain.Alert, error) {
	args := m.Called(ctx, userID, limit)
	if a := args.Get(0); a != nil {
		return a.([]domain.Alert), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAlertRepository) ListUnread(ctx context.Context, userID uuid.UUID) ([]domain.Alert, error) {
	args := m.Called(ctx, userID)
	if a := args.Get(0); a != nil {
		return a.([]domain.Alert), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAlertRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockAlertRepository) MarkAsRead(ctx context.Context, userID, alertID uuid.UUID) error {
	return m.Called(ctx, userID, alertID).Error(0)
}
func (m *MockAlertRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *MockAlertRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

// ── ChatRepository ───────────────────────────────────────────────

type MockChatRepository struct{ mock.Mock }

func (m *MockChatRepository) Create(ctx context.Context, msg *domain.ChatMessage) error {
	return m.Called(ctx, msg).Error(0)
}
func (m *MockChatRepository) ListBySession(ctx context.Context, userID, sessionID uuid.UUID) ([]domain.ChatMessage, error) {
	args := m.Called(ctx, userID, sessionID)
	if msgs := args.Get(0); msgs != nil {
		return msgs.([]domain.ChatMessage), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockChatRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *MockChatRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

// ── AuditRepository ──────────────────────────────────────────────

type MockAuditRepository struct{ mock.Mock }

func (m *MockAuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *MockAuditRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]domain.AuditLog, error) {
	args := m.Called(ctx, userID, limit)
	if a := args.Get(0); a != nil {
		return a.([]domain.AuditLog), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockAuditRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}
