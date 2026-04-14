package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes — machine-readable, consistent across all endpoints.
const (
	CodeValidation     = "VALIDATION_ERROR"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeConflict       = "CONFLICT"
	CodeRateLimited    = "RATE_LIMITED"
	CodeAccountLocked  = "ACCOUNT_LOCKED"
	CodeBudgetExceeded = "BUDGET_EXCEEDED"
	CodeAIUnavailable  = "AI_SERVICE_UNAVAILABLE"
	CodePlaidError     = "PLAID_ERROR"
	CodeTimeout        = "TIMEOUT"
	CodePayloadTooLarge = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMedia = "UNSUPPORTED_MEDIA_TYPE"
	CodeInternal       = "INTERNAL_ERROR"
)

// DomainError represents a business-level error with a machine-readable code.
type DomainError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details []FieldError   `json:"details,omitempty"`
}

// FieldError represents a validation error on a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// HTTPStatus maps error codes to HTTP status codes.
func (e *DomainError) HTTPStatus() int {
	switch e.Code {
	case CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeAccountLocked:
		return http.StatusLocked
	case CodeBudgetExceeded:
		return http.StatusUnprocessableEntity
	case CodeAIUnavailable:
		return http.StatusServiceUnavailable
	case CodePlaidError:
		return http.StatusBadGateway
	case CodeTimeout:
		return http.StatusRequestTimeout
	case CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeUnsupportedMedia:
		return http.StatusUnsupportedMediaType
	default:
		return http.StatusInternalServerError
	}
}

// Constructor helpers

func NewValidationError(message string, details ...FieldError) *DomainError {
	return &DomainError{Code: CodeValidation, Message: message, Details: details}
}

func NewUnauthorizedError(message string) *DomainError {
	return &DomainError{Code: CodeUnauthorized, Message: message}
}

func NewForbiddenError(message string) *DomainError {
	return &DomainError{Code: CodeForbidden, Message: message}
}

func NewNotFoundError(resource string) *DomainError {
	return &DomainError{Code: CodeNotFound, Message: fmt.Sprintf("%s not found", resource)}
}

func NewConflictError(message string) *DomainError {
	return &DomainError{Code: CodeConflict, Message: message}
}

func NewAccountLockedError(message string) *DomainError {
	return &DomainError{Code: CodeAccountLocked, Message: message}
}

func NewInternalError(message string) *DomainError {
	return &DomainError{Code: CodeInternal, Message: message}
}

// IsDomainError checks if an error is a DomainError.
func IsDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}
