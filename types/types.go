// Package types contains shared types for the Salesforce SDK to avoid import cycles.
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

// Token represents an authentication token.
type Token struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	InstanceURL  string
	IssuedAt     time.Time
	ExpiresAt    time.Time
}

// IsExpired checks if the token has expired.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(t.ExpiresAt)
}

// ErrorCode represents a Salesforce error code.
type ErrorCode string

// Common Salesforce error codes
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
	ErrorCodeInsufficientAccess   ErrorCode = "INSUFFICIENT_ACCESS_ON_CROSS_REFERENCE_ENTITY"
	ErrorCodeRequestLimit         ErrorCode = "REQUEST_LIMIT_EXCEEDED"
	ErrorCodeStorageLimit         ErrorCode = "STORAGE_LIMIT_EXCEEDED"
)

// APIError represents an error returned by the Salesforce API.
type APIError struct {
	Message    string            `json:"message"`
	ErrorCode  ErrorCode         `json:"errorCode"`
	Fields     []string          `json:"fields,omitempty"`
	StatusCode int               `json:"-"`
	Raw        []byte            `json:"-"`
	Details    map[string]string `json:"extended,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf("[%s] %s (fields: %s)", e.ErrorCode, e.Message, strings.Join(e.Fields, ", "))
	}
	return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
}

// IsSessionInvalid checks if the error is due to an invalid or expired session.
func (e *APIError) IsSessionInvalid() bool {
	return e.ErrorCode == ErrorCodeInvalidSession || e.ErrorCode == ErrorCodeSessionExpired
}

// IsRetryable checks if the request should be retried.
func (e *APIError) IsRetryable() bool {
	return e.StatusCode == http.StatusServiceUnavailable ||
		e.StatusCode == http.StatusTooManyRequests ||
		e.StatusCode == http.StatusGatewayTimeout
}

// APIErrors represents multiple API errors.
type APIErrors []APIError

// Error implements the error interface.
func (e APIErrors) Error() string {
	if len(e) == 0 {
		return "unknown error"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// First returns the first error or nil if empty.
func (e APIErrors) First() *APIError {
	if len(e) == 0 {
		return nil
	}
	return &e[0]
}

// AuthError represents an authentication error.
type AuthError struct {
	ErrorType   string `json:"error"`
	Description string `json:"error_description"`
	StatusCode  int    `json:"-"`
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication failed: %s - %s", e.ErrorType, e.Description)
}

// IsInvalidGrant checks if the error is due to an invalid grant.
func (e *AuthError) IsInvalidGrant() bool {
	return e.ErrorType == "invalid_grant"
}

// ValidationError represents a client-side validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// RateLimitError represents a rate limit exceeded error.
type RateLimitError struct {
	RetryAfter int
	Message    string
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (retry after %d seconds)", e.Message, e.RetryAfter)
}

// ParseAPIError parses an API error from the response body.
func ParseAPIError(statusCode int, body []byte) error {
	var apiErrors APIErrors
	if err := json.Unmarshal(body, &apiErrors); err == nil && len(apiErrors) > 0 {
		for i := range apiErrors {
			apiErrors[i].StatusCode = statusCode
			apiErrors[i].Raw = body
		}
		return apiErrors
	}

	var apiError APIError
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.ErrorCode != "" {
		apiError.StatusCode = statusCode
		apiError.Raw = body
		return &apiError
	}

	var authError AuthError
	if err := json.Unmarshal(body, &authError); err == nil && authError.ErrorType != "" {
		authError.StatusCode = statusCode
		return &authError
	}

	return &APIError{
		Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
		StatusCode: statusCode,
		Raw:        body,
	}
}

// IsNotFoundError checks if the error is a 404 not found error.
func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	if apiErrs, ok := err.(APIErrors); ok && len(apiErrs) > 0 {
		return apiErrs[0].StatusCode == http.StatusNotFound
	}
	return false
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	if _, ok := err.(*AuthError); ok {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.IsSessionInvalid()
	}
	if apiErrs, ok := err.(APIErrors); ok && len(apiErrs) > 0 {
		return apiErrs[0].IsSessionInvalid()
	}
	return false
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	if _, ok := err.(*RateLimitError); ok {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusTooManyRequests ||
			apiErr.ErrorCode == ErrorCodeRequestLimit
	}
	return false
}

// Logger defines the interface for logging.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
