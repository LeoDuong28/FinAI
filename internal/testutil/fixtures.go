package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/nghiaduong/finai/internal/config"
	"github.com/nghiaduong/finai/internal/domain"
)

// FixedTime is a deterministic timestamp for test fixtures to avoid flaky tests.
var FixedTime = time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC)

func NewTestUser(opts ...func(*domain.User)) *domain.User {
	u := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$RWh6VVdPMllTNE1KMjdBRmxMWWxMT0JLcUo3cGVFSQ",
		FirstName:    "Test",
		LastName:     "User",
		Role:         "user",
		Currency:     "USD",
		Timezone:     "America/New_York",
		Theme:        "light",
		CreatedAt:    FixedTime,
		UpdatedAt:    FixedTime,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

func NewTestBudget(opts ...func(*domain.Budget)) *domain.Budget {
	b := &domain.Budget{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Name:           "Groceries",
		AmountLimit:    decimal.NewFromInt(500),
		Period:         "monthly",
		StartDate:      FixedTime.AddDate(0, -1, 0),
		AlertThreshold: decimal.NewFromFloat(0.80),
		IsActive:       true,
		CreatedAt:      FixedTime,
		UpdatedAt:      FixedTime,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func NewTestBill(opts ...func(*domain.Bill)) *domain.Bill {
	b := &domain.Bill{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Name:         "Electric Bill",
		Amount:       decimal.NewFromInt(120),
		DueDate:      FixedTime.AddDate(0, 0, 7),
		Frequency:    "monthly",
		IsAutopay:    false,
		Status:       "upcoming",
		ReminderDays: 3,
		CreatedAt:    FixedTime,
		UpdatedAt:    FixedTime,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func NewTestSavingsGoal(opts ...func(*domain.SavingsGoal)) *domain.SavingsGoal {
	g := &domain.SavingsGoal{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		Name:          "Emergency Fund",
		TargetAmount:  decimal.NewFromInt(10000),
		CurrentAmount: decimal.NewFromInt(2500),
		Status:        "active",
		CreatedAt:     FixedTime,
		UpdatedAt:     FixedTime,
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

func NewTestTransaction(opts ...func(*domain.Transaction)) *domain.Transaction {
	t := &domain.Transaction{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Amount:       decimal.NewFromFloat(42.50),
		CurrencyCode: "USD",
		Date:         FixedTime,
		Name:         "Coffee Shop",
		Type:         "debit",
		CreatedAt:    FixedTime,
		UpdatedAt:    FixedTime,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func NewTestSubscription(opts ...func(*domain.Subscription)) *domain.Subscription {
	s := &domain.Subscription{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		Name:         "Netflix",
		Amount:       decimal.NewFromFloat(15.99),
		CurrencyCode: "USD",
		Frequency:    "monthly",
		Status:       "active",
		CreatedAt:    FixedTime,
		UpdatedAt:    FixedTime,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewTestAuthConfig() *config.AuthConfig {
	return &config.AuthConfig{
		JWTSecret:        "test-jwt-secret-that-is-at-least-32-chars-long",
		EncryptionKey:    "01234567890123456789012345678901", // exactly 32 bytes
		AccessTokenTTL:   15 * time.Minute,
		RefreshTokenTTL:  7 * 24 * time.Hour,
		MaxLoginAttempts: 5,
		LockoutDuration:  15 * time.Minute,
		MaxTOTPAttempts:  5,
	}
}
