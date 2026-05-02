package service_test

import (
	"testing"

	"github.com/nghiaduong/finai/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKey = "01234567890123456789012345678901" // exactly 32 bytes

func TestEncryptionService_RoundTrip(t *testing.T) {
	svc := service.NewEncryptionService(testKey)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"unicode", "こんにちは世界"},
		{"special chars", "p@$$w0rd!#%&*"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := svc.Encrypt(tt.plaintext)
			require.NoError(t, err)
			assert.NotEqual(t, tt.plaintext, encrypted)

			decrypted, err := svc.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptionService_DifferentCiphertexts(t *testing.T) {
	svc := service.NewEncryptionService(testKey)
	plaintext := "same input"

	enc1, err := svc.Encrypt(plaintext)
	require.NoError(t, err)
	enc2, err := svc.Encrypt(plaintext)
	require.NoError(t, err)

	// Different nonces produce different ciphertexts
	assert.NotEqual(t, enc1, enc2)
}

func TestEncryptionService_WrongKey(t *testing.T) {
	svc1 := service.NewEncryptionService(testKey)
	svc2 := service.NewEncryptionService("98765432109876543210987654321098")

	encrypted, err := svc1.Encrypt("secret data")
	require.NoError(t, err)

	_, err = svc2.Decrypt(encrypted)
	assert.Error(t, err)
}

func TestEncryptionService_TamperedCiphertext(t *testing.T) {
	svc := service.NewEncryptionService(testKey)

	encrypted, err := svc.Encrypt("secret data")
	require.NoError(t, err)

	// Tamper with the ciphertext
	tampered := encrypted[:len(encrypted)-2] + "XX"
	_, err = svc.Decrypt(tampered)
	assert.Error(t, err)
}

func TestEncryptionService_InvalidBase64(t *testing.T) {
	svc := service.NewEncryptionService(testKey)
	_, err := svc.Decrypt("not-valid-base64!!!")
	assert.Error(t, err)
}

func TestEncryptionService_OptionalNil(t *testing.T) {
	svc := service.NewEncryptionService(testKey)

	enc, err := svc.EncryptOptional(nil)
	assert.NoError(t, err)
	assert.Nil(t, enc)

	dec, err := svc.DecryptOptional(nil)
	assert.NoError(t, err)
	assert.Nil(t, dec)
}

func TestEncryptionService_OptionalRoundTrip(t *testing.T) {
	svc := service.NewEncryptionService(testKey)
	val := "optional value"

	enc, err := svc.EncryptOptional(&val)
	require.NoError(t, err)
	require.NotNil(t, enc)

	dec, err := svc.DecryptOptional(enc)
	require.NoError(t, err)
	require.NotNil(t, dec)
	assert.Equal(t, val, *dec)
}

func TestNewEncryptionService_PanicsOnWrongKeyLength(t *testing.T) {
	assert.Panics(t, func() {
		service.NewEncryptionService("short-key")
	})
}
