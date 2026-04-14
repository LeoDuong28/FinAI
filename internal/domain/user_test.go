package domain_test

import (
	"testing"
	"time"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestUser_IsLocked(t *testing.T) {
	tests := []struct {
		name        string
		lockedUntil *time.Time
		want        bool
	}{
		{
			name:        "nil locked_until returns false",
			lockedUntil: nil,
			want:        false,
		},
		{
			name:        "past locked_until returns false",
			lockedUntil: timePtr(time.Now().Add(-1 * time.Hour)),
			want:        false,
		},
		{
			name:        "future locked_until returns true",
			lockedUntil: timePtr(time.Now().Add(1 * time.Hour)),
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.User{LockedUntil: tt.lockedUntil}
			assert.Equal(t, tt.want, user.IsLocked())
		})
	}
}

func TestUser_FullName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		want      string
	}{
		{"both names", "John", "Doe", "John Doe"},
		{"empty first", "", "Doe", " Doe"},
		{"empty last", "John", "", "John "},
		{"both empty", "", "", " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &domain.User{FirstName: tt.firstName, LastName: tt.lastName}
			assert.Equal(t, tt.want, user.FullName())
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
