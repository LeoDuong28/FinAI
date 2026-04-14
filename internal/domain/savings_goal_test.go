package domain_test

import (
	"testing"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSavingsGoal_PercentComplete(t *testing.T) {
	tests := []struct {
		name    string
		current decimal.Decimal
		target  decimal.Decimal
		want    string
	}{
		{"zero target returns zero", decimal.NewFromInt(50), decimal.Zero, "0"},
		{"zero current returns zero", decimal.Zero, decimal.NewFromInt(100), "0"},
		{"half complete", decimal.NewFromInt(50), decimal.NewFromInt(100), "50"},
		{"fully complete", decimal.NewFromInt(100), decimal.NewFromInt(100), "100"},
		{"over target", decimal.NewFromInt(150), decimal.NewFromInt(100), "150"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domain.SavingsGoal{CurrentAmount: tt.current, TargetAmount: tt.target}
			assert.Equal(t, tt.want, g.PercentComplete().StringFixed(0))
		})
	}
}

func TestSavingsGoal_Remaining(t *testing.T) {
	tests := []struct {
		name    string
		current decimal.Decimal
		target  decimal.Decimal
		want    string
	}{
		{"remaining positive", decimal.NewFromInt(30), decimal.NewFromInt(100), "70"},
		{"remaining zero", decimal.NewFromInt(100), decimal.NewFromInt(100), "0"},
		{"over target negative", decimal.NewFromInt(150), decimal.NewFromInt(100), "-50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domain.SavingsGoal{CurrentAmount: tt.current, TargetAmount: tt.target}
			assert.Equal(t, tt.want, g.Remaining().StringFixed(0))
		})
	}
}

func TestSavingsGoal_IsComplete(t *testing.T) {
	tests := []struct {
		name    string
		current decimal.Decimal
		target  decimal.Decimal
		want    bool
	}{
		{"not complete", decimal.NewFromInt(50), decimal.NewFromInt(100), false},
		{"exactly at target", decimal.NewFromInt(100), decimal.NewFromInt(100), true},
		{"over target", decimal.NewFromInt(150), decimal.NewFromInt(100), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domain.SavingsGoal{CurrentAmount: tt.current, TargetAmount: tt.target}
			assert.Equal(t, tt.want, g.IsComplete())
		})
	}
}
