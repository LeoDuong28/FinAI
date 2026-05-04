package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPeriodBounds_Weekly(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)  // Wednesday
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)   // Wednesday, week 2

	from, to := periodBounds("weekly", start, now)

	assert.Equal(t, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), from)
	assert.Equal(t, 23, to.Hour())
	assert.Equal(t, 59, to.Minute())
}

func TestPeriodBounds_Weekly_BeforeStart(t *testing.T) {
	start := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC)

	from, to := periodBounds("weekly", start, now)

	assert.Equal(t, start, from)
	assert.True(t, to.After(from))
}

func TestPeriodBounds_Monthly(t *testing.T) {
	start := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	now := time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC)

	from, to := periodBounds("monthly", start, now)

	assert.Equal(t, time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), from)
	assert.Equal(t, 31, to.Day()) // March has 31 days
}

func TestPeriodBounds_Yearly(t *testing.T) {
	start := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	now := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)

	from, to := periodBounds("yearly", start, now)

	assert.Equal(t, time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), from)
	assert.Equal(t, 2026, to.Year())
}

func TestPeriodBounds_Yearly_Feb29(t *testing.T) {
	// Start date is Feb 29 (leap year), but current year is non-leap
	start := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	now := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

	from, _ := periodBounds("yearly", start, now)

	// Should clamp to Feb 28 in non-leap year
	assert.Equal(t, 2, int(from.Month()))
	assert.Equal(t, 28, from.Day())
}

func TestEndOfDay(t *testing.T) {
	d := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	eod := endOfDay(d)
	assert.Equal(t, 23, eod.Hour())
	assert.Equal(t, 59, eod.Minute())
	assert.Equal(t, 59, eod.Second())
}
