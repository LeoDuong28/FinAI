package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionService provides AES-256-GCM authenticated encryption.
type EncryptionService struct {
	aead cipher.AEAD
}

// NewEncryptionService creates a new encryption service with the given 32-byte key.
// Panics if the key is not exactly 32 bytes.
func NewEncryptionService(key string) *EncryptionService {
	if len(key) != 32 {
		panic(fmt.Sprintf("encryption key must be exactly 32 bytes, got %d", len(key)))
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(fmt.Sprintf("create AES cipher: %s", err))
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(fmt.Sprintf("create GCM: %s", err))
	}

	return &EncryptionService{aead: aead}
}

// Encrypt encrypts plaintext using AES-256-GCM and returns a base64-encoded string
// containing the nonce prepended to the ciphertext.
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := s.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decodes the base64 input, extracts the nonce, and decrypts using AES-256-GCM.
func (s *EncryptionService) Decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	nonceSize := s.aead.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := s.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptOptional encrypts a string pointer, returning nil if input is nil.
func (s *EncryptionService) EncryptOptional(plaintext *string) (*string, error) {
	if plaintext == nil {
		return nil, nil
	}
	encrypted, err := s.Encrypt(*plaintext)
	if err != nil {
		return nil, err
	}
	return &encrypted, nil
}

// DecryptOptional decrypts a string pointer, returning nil if input is nil.
func (s *EncryptionService) DecryptOptional(encoded *string) (*string, error) {
	if encoded == nil {
		return nil, nil
	}
	decrypted, err := s.Decrypt(*encoded)
	if err != nil {
		return nil, err
	}
	return &decrypted, nil
}
