package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		pass    string
		wantErr bool
	}{
		{"valid password", "MyP@ssw0rd!", false},
		{"too short", "Ab1!", true},
		{"exactly 8 chars valid", "Abcde1@x", false},
		{"missing uppercase", "abcdef1@", true},
		{"missing lowercase", "ABCDEF1@", true},
		{"missing digit", "Abcdefg@", true},
		{"missing special", "Abcdefg1", true},
		{"empty password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pass)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword_TooLong(t *testing.T) {
	// 129 characters
	long := make([]byte, 129)
	for i := range long {
		long[i] = 'A'
	}
	err := ValidatePassword(string(long))
	assert.Error(t, err)
}

func TestPasswordHasher_HashAndVerify(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "TestP@ss1"

	hash, err := hasher.Hash(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	match, err := hasher.Verify(password, hash)
	require.NoError(t, err)
	assert.True(t, match)

	match, err = hasher.Verify("wrong", hash)
	require.NoError(t, err)
	assert.False(t, match)
}
