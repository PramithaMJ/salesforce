// Package types provides shared types for the Salesforce SDK.
package types

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Version constants
const (
	DefaultAPIVersion = "59.0"
	DefaultTimeout    = 30 * time.Second
	DefaultMaxRetries = 3
)

// Token represents an OAuth access token.
type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	InstanceURL  string    `json:"instance_url"`
	ID           string    `json:"id,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	Signature    string    `json:"signature,omitempty"`
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return time.Since(t.IssuedAt) > 2*time.Hour
	}
	return time.Now().After(t.ExpiresAt.Add(-5 * time.Minute))
}

// Logger defines the logging interface.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// ErrorCode represents Salesforce error codes.
type ErrorCode string

const (
	ErrorCodeInvalidSession       ErrorCode = "INVALID_SESSION_ID"
	ErrorCodeSessionExpired       ErrorCode = "SESSION_EXPIRED"
	ErrorCodeInvalidField         ErrorCode = "INVALID_FIELD"
	ErrorCodeMalformedQuery       ErrorCode = "MALFORMED_QUERY"
	ErrorCodeInvalidType          ErrorCode = "INVALID_TYPE"
	ErrorCodeEntityDeleted        ErrorCode = "ENTITY_IS_DELETED"
	ErrorCodeDuplicateValue       ErrorCode = "DUPLICATE_VALUE"
	ErrorCodeRequiredFieldMissing ErrorCode = "REQUIRED_FIELD_MISSING"
	ErrorCodeInvalidCrossRef      ErrorCode = "INVALID_CROSS_REFERENCE_KEY"
	ErrorCodeInsufficientAccess   ErrorCode = "INSUFFICIENT_ACCESS_OR_READONLY"
	ErrorCodeRequestLimit         ErrorCode = "REQUEST_LIMIT_EXCEEDED"
	ErrorCodeStorageLimit         ErrorCode = "STORAGE_LIMIT_EXCEEDED"
	ErrorCodeNotFound             ErrorCode = "NOT_FOUND"
	ErrorCodeFieldCustomValidation ErrorCode = "FIELD_CUSTOM_VALIDATION_EXCEPTION"
	ErrorCodeFieldIntegrity       ErrorCode = "FIELD_INTEGRITY_EXCEPTION"
	ErrorCodeUnableToLockRow      ErrorCode = "UNABLE_TO_LOCK_ROW"
	ErrorCodeProcessingHalt       ErrorCode = "PROCESSING_HALTED"
)

// APIError represents a Salesforce API error.
type APIError struct {
	Message    string    `json:"message"`
	ErrorCode  ErrorCode `json:"errorCode"`
	Fields     []string  `json:"fields,omitempty"`
	StatusCode int       `json:"-"`
}

func (e *APIError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf("[%s] %s (fields: %s)", e.ErrorCode, e.Message, strings.Join(e.Fields, ", "))
	}
	return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
}

// IsRetryable determines if the error can be retried.
func (e *APIError) IsRetryable() bool {
	switch e.ErrorCode {
	case ErrorCodeRequestLimit, ErrorCodeUnableToLockRow:
		return true
	}
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

// APIErrors represents multiple API errors.
type APIErrors []APIError

func (e APIErrors) Error() string {
	if len(e) == 0 {
		return "unknown error"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return strings.Join(msgs, "; ")
}

// AuthError represents an authentication error.
type AuthError struct {
	ErrorType   string `json:"error"`
	Description string `json:"error_description"`
	StatusCode  int    `json:"-"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorType, e.Description)
}

// RateLimitError indicates API rate limit exceeded.
type RateLimitError struct {
	RetryAfter int
	Message    string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %d seconds: %s", e.RetryAfter, e.Message)
}

// ValidationError represents input validation errors.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// NotFoundError indicates a resource was not found.
type NotFoundError struct {
	ObjectType string
	ID         string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.ObjectType, e.ID)
}

// ParseAPIError parses a Salesforce API error response.
func ParseAPIError(statusCode int, body []byte) error {
	var errs APIErrors
	if err := json.Unmarshal(body, &errs); err == nil && len(errs) > 0 {
		for i := range errs {
			errs[i].StatusCode = statusCode
		}
		return errs
	}

	var singleErr APIError
	if err := json.Unmarshal(body, &singleErr); err == nil && singleErr.ErrorCode != "" {
		singleErr.StatusCode = statusCode
		return &singleErr
	}

	return &APIError{
		Message:    string(body),
		ErrorCode:  ErrorCode(fmt.Sprintf("HTTP_%d", statusCode)),
		StatusCode: statusCode,
	}
}

// IsNotFoundError checks if the error is a not found error.
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.ErrorCode == ErrorCodeNotFound || apiErr.StatusCode == http.StatusNotFound
	}
	if apiErrs, ok := err.(APIErrors); ok && len(apiErrs) > 0 {
		return apiErrs[0].ErrorCode == ErrorCodeNotFound || apiErrs[0].StatusCode == http.StatusNotFound
	}
	_, ok := err.(*NotFoundError)
	return ok
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	if _, ok := err.(*AuthError); ok {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.ErrorCode == ErrorCodeInvalidSession || apiErr.ErrorCode == ErrorCodeSessionExpired
	}
	return false
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	if _, ok := err.(*RateLimitError); ok {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.ErrorCode == ErrorCodeRequestLimit || apiErr.StatusCode == http.StatusTooManyRequests
	}
	return false
}

// IsRetryableError checks if the error can be retried.
func IsRetryableError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsRetryable()
	}
	if apiErrs, ok := err.(APIErrors); ok && len(apiErrs) > 0 {
		return apiErrs[0].IsRetryable()
	}
	return IsRateLimitError(err)
}
