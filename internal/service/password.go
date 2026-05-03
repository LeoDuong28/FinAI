package service

import (
	"unicode"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"

	apperr "github.com/nghiaduong/finai/internal/errors"
)

// PasswordHasher wraps Argon2id password hashing.
type PasswordHasher struct {
	params *argon2id.Params
}

// NewPasswordHasherWithParams creates a password hasher with custom parameters (for tests).
func NewPasswordHasherWithParams(params *argon2id.Params) *PasswordHasher {
	return &PasswordHasher{params: params}
}

// NewPasswordHasher creates a new password hasher with production parameters.
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		params: &argon2id.Params{
			Memory:      64 * 1024, // 64 MB
			Iterations:  3,
			Parallelism: 4,
			SaltLength:  16,
			KeyLength:   32,
		},
	}
}

// Hash generates an Argon2id hash from a plaintext password.
func (h *PasswordHasher) Hash(password string) (string, error) {
	return argon2id.CreateHash(password, h.params)
}

// Verify checks a plaintext password against an Argon2id hash.
func (h *PasswordHasher) Verify(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

// dummyVerify performs a hash operation to equalize timing when a user is not found.
// This prevents timing side-channels that reveal whether an email exists.
func (h *PasswordHasher) dummyVerify(password string) {
	// Use a syntactically valid Argon2id hash that will never match.
	dummyHash := "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$RWh6VVdPMllTNE1KMjdBRmxMWWxMT0JLcUo3cGVFSQ"
	_, _ = argon2id.ComparePasswordAndHash(password, dummyHash)
}

// ValidatePassword checks password strength requirements.
func ValidatePassword(password string) error {
	runeCount := utf8.RuneCountInString(password)
	if runeCount < 8 {
		return apperr.NewValidationError("Password must be at least 8 characters",
			apperr.FieldError{Field: "password", Message: "must be at least 8 characters"})
	}
	if runeCount > 128 {
		return apperr.NewValidationError("Password must be at most 128 characters",
			apperr.FieldError{Field: "password", Message: "must be at most 128 characters"})
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	var details []apperr.FieldError
	if !hasUpper {
		details = append(details, apperr.FieldError{Field: "password", Message: "must include an uppercase letter"})
	}
	if !hasLower {
		details = append(details, apperr.FieldError{Field: "password", Message: "must include a lowercase letter"})
	}
	if !hasDigit {
		details = append(details, apperr.FieldError{Field: "password", Message: "must include a digit"})
	}
	if !hasSpecial {
		details = append(details, apperr.FieldError{Field: "password", Message: "must include a special character"})
	}

	if len(details) > 0 {
		return apperr.NewValidationError("Password does not meet requirements", details...)
	}

	return nil
}
