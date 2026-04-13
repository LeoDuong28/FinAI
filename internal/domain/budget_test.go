package domain_test

import (
	"testing"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestBudget_PercentUsed(t *testing.T) {
	tests := []struct {
		name  string
		spent decimal.Decimal
		limit decimal.Decimal
		want  string
	}{
		{"zero limit returns zero", decimal.NewFromInt(50), decimal.Zero, "0"},
		{"zero spent returns zero", decimal.Zero, decimal.NewFromInt(100), "0"},
		{"50 of 100 = 50%", decimal.NewFromInt(50), decimal.NewFromInt(100), "50"},
		{"100 of 100 = 100%", decimal.NewFromInt(100), decimal.NewFromInt(100), "100"},
		{"150 of 100 = 150%", decimal.NewFromInt(150), decimal.NewFromInt(100), "150"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Budget{Spent: tt.spent, AmountLimit: tt.limit}
			assert.Equal(t, tt.want, b.PercentUsed().StringFixed(0))
		})
	}
}

func TestBudget_IsOverBudget(t *testing.T) {
	tests := []struct {
		name  string
		spent decimal.Decimal
		limit decimal.Decimal
		want  bool
	}{
		{"under budget", decimal.NewFromInt(50), decimal.NewFromInt(100), false},
		{"exactly at limit", decimal.NewFromInt(100), decimal.NewFromInt(100), false},
		{"over budget", decimal.NewFromInt(150), decimal.NewFromInt(100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Budget{Spent: tt.spent, AmountLimit: tt.limit}
			assert.Equal(t, tt.want, b.IsOverBudget())
		})
	}
}

func TestBudget_IsApproachingLimit(t *testing.T) {
	tests := []struct {
		name      string
		spent     decimal.Decimal
		limit     decimal.Decimal
		threshold decimal.Decimal
		want      bool
	}{
		{
			"below threshold",
			decimal.NewFromInt(70),
			decimal.NewFromInt(100),
			decimal.NewFromFloat(0.80),
			false,
		},
		{
			"at threshold",
			decimal.NewFromInt(80),
			decimal.NewFromInt(100),
			decimal.NewFromFloat(0.80),
			true,
		},
		{
			"above threshold but not over",
			decimal.NewFromInt(90),
			decimal.NewFromInt(100),
			decimal.NewFromFloat(0.80),
			true,
		},
		{
			"over budget returns false",
			decimal.NewFromInt(150),
			decimal.NewFromInt(100),
			decimal.NewFromFloat(0.80),
			false,
		},
		{
			"zero limit returns false",
			decimal.NewFromInt(50),
			decimal.Zero,
			decimal.NewFromFloat(0.80),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Budget{
				Spent:          tt.spent,
				AmountLimit:    tt.limit,
				AlertThreshold: tt.threshold,
			}
			assert.Equal(t, tt.want, b.IsApproachingLimit())
		})
	}
}
