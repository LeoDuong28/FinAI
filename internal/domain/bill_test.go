package domain_test

import (
	"testing"
	"time"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBill_IsOverdue(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		dueDate time.Time
		want    bool
	}{
		{"paid is never overdue", "paid", time.Now().Add(-48 * time.Hour), false},
		{"upcoming past due is overdue", "upcoming", time.Now().Add(-48 * time.Hour), true},
		{"upcoming future due is not overdue", "upcoming", time.Now().Add(48 * time.Hour), false},
		{"overdue status past due", "overdue", time.Now().Add(-48 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Bill{Status: tt.status, DueDate: tt.dueDate}
			assert.Equal(t, tt.want, b.IsOverdue())
		})
	}
}

func TestBill_DaysUntilDue(t *testing.T) {
	tests := []struct {
		name    string
		dueDate time.Time
		want    int
	}{
		{"past due returns 0", time.Now().Add(-48 * time.Hour), 0},
		{"due today returns 0 or 1", time.Now().Add(1 * time.Hour), 1},
		{"due in 3 days", time.Now().Add(72 * time.Hour), 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Bill{DueDate: tt.dueDate}
			got := b.DaysUntilDue()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBill_ShouldRemind(t *testing.T) {
	tests := []struct {
		name         string
		status       string
		dueDate      time.Time
		reminderDays int
		want         bool
	}{
		{
			"upcoming within reminder window",
			"upcoming",
			time.Now().Add(48 * time.Hour),
			3,
			true,
		},
		{
			"upcoming outside reminder window",
			"upcoming",
			time.Now().Add(120 * time.Hour),
			3,
			false,
		},
		{
			"paid does not remind",
			"paid",
			time.Now().Add(24 * time.Hour),
			3,
			false,
		},
		{
			"overdue does not remind",
			"overdue",
			time.Now().Add(-24 * time.Hour),
			3,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Bill{
				Status:       tt.status,
				DueDate:      tt.dueDate,
				ReminderDays: tt.reminderDays,
			}
			assert.Equal(t, tt.want, b.ShouldRemind())
		})
	}
}
