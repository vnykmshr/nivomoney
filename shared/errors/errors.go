// Package errors provides custom error types and error handling utilities for Nivo services.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents a specific error type that can be used by clients.
type ErrorCode string

// Error codes for common error scenarios.
const (
	// Client errors (4xx)
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeBadRequest        ErrorCode = "BAD_REQUEST"
	ErrCodeValidation        ErrorCode = "VALIDATION_ERROR"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeRateLimit         ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodePrecondition      ErrorCode = "PRECONDITION_FAILED"
	ErrCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"

	// Server errors (5xx)
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout       ErrorCode = "TIMEOUT"
	ErrCodeDatabaseError ErrorCode = "DATABASE_ERROR"

	// Domain-specific errors
	ErrCodeInvalidAmount      ErrorCode = "INVALID_AMOUNT"
	ErrCodeInvalidCurrency    ErrorCode = "INVALID_CURRENCY"
	ErrCodeAccountFrozen      ErrorCode = "ACCOUNT_FROZEN"
	ErrCodeTransactionFailed  ErrorCode = "TRANSACTION_FAILED"
	ErrCodeDuplicateIdempotencyKey ErrorCode = "DUPLICATE_IDEMPOTENCY_KEY"
)

// Error represents a structured error with code, message, and details.
type Error struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Err     error                  `json:"-"` // Underlying error (not exposed in JSON)
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error wrapping support.
func (e *Error) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error.
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	e.Details = details
	return e
}

// AddDetail adds a single detail to the error.
func (e *Error) AddDetail(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// HTTPStatusCode returns the appropriate HTTP status code for this error.
func (e *Error) HTTPStatusCode() int {
	switch e.Code {
	// 4xx Client Errors
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeBadRequest, ErrCodeValidation, ErrCodeInvalidAmount, ErrCodeInvalidCurrency:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict, ErrCodeDuplicateIdempotencyKey:
		return http.StatusConflict
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodePrecondition, ErrCodeInsufficientFunds, ErrCodeAccountFrozen:
		return http.StatusPreconditionFailed

	// 5xx Server Errors
	case ErrCodeInternal, ErrCodeDatabaseError, ErrCodeTransactionFailed:
		return http.StatusInternalServerError
	case ErrCodeUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout:
		return http.StatusGatewayTimeout

	default:
		return http.StatusInternalServerError
	}
}

// New creates a new Error with the given code and message.
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a code and message.
func Wrap(err error, code ErrorCode, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
	}
}

// Common error constructors

// NotFound creates a not found error.
func NotFound(resource string) *Error {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

// NotFoundWithID creates a not found error with resource ID.
func NotFoundWithID(resource, id string) *Error {
	return New(ErrCodeNotFound, fmt.Sprintf("%s with id %s not found", resource, id))
}

// BadRequest creates a bad request error.
func BadRequest(message string) *Error {
	return New(ErrCodeBadRequest, message)
}

// Validation creates a validation error.
func Validation(message string) *Error {
	return New(ErrCodeValidation, message)
}

// ValidationWithFields creates a validation error with field details.
func ValidationWithFields(message string, fields map[string]string) *Error {
	details := make(map[string]interface{})
	for k, v := range fields {
		details[k] = v
	}
	return New(ErrCodeValidation, message).WithDetails(details)
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *Error {
	return New(ErrCodeUnauthorized, message)
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *Error {
	return New(ErrCodeForbidden, message)
}

// Conflict creates a conflict error.
func Conflict(message string) *Error {
	return New(ErrCodeConflict, message)
}

// Internal creates an internal server error.
func Internal(message string) *Error {
	return New(ErrCodeInternal, message)
}

// InternalWrap wraps an error as an internal server error.
func InternalWrap(err error, message string) *Error {
	return Wrap(err, ErrCodeInternal, message)
}

// Database creates a database error.
func Database(message string) *Error {
	return New(ErrCodeDatabaseError, message)
}

// DatabaseWrap wraps a database error.
func DatabaseWrap(err error, message string) *Error {
	return Wrap(err, ErrCodeDatabaseError, message)
}

// Unavailable creates a service unavailable error.
func Unavailable(message string) *Error {
	return New(ErrCodeUnavailable, message)
}

// Timeout creates a timeout error.
func Timeout(message string) *Error {
	return New(ErrCodeTimeout, message)
}

// InsufficientFunds creates an insufficient funds error.
func InsufficientFunds(message string) *Error {
	return New(ErrCodeInsufficientFunds, message)
}

// AccountFrozen creates an account frozen error.
func AccountFrozen(message string) *Error {
	return New(ErrCodeAccountFrozen, message)
}

// TransactionFailed creates a transaction failed error.
func TransactionFailed(message string) *Error {
	return New(ErrCodeTransactionFailed, message)
}

// DuplicateIdempotencyKey creates a duplicate idempotency key error.
func DuplicateIdempotencyKey(key string) *Error {
	return New(ErrCodeDuplicateIdempotencyKey, fmt.Sprintf("duplicate idempotency key: %s", key))
}

// Utility functions for error checking

// Is checks if an error is of a specific type using errors.Is.
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetErrorCode extracts the error code from an error if it's a custom Error.
func GetErrorCode(err error) ErrorCode {
	var e *Error
	if As(err, &e) {
		return e.Code
	}
	return ErrCodeInternal
}

// GetHTTPStatus extracts the HTTP status code from an error.
func GetHTTPStatus(err error) int {
	var e *Error
	if As(err, &e) {
		return e.HTTPStatusCode()
	}
	return http.StatusInternalServerError
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return GetErrorCode(err) == ErrCodeNotFound
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return GetErrorCode(err) == ErrCodeValidation
}

// IsUnauthorized checks if an error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return GetErrorCode(err) == ErrCodeUnauthorized
}

// IsForbidden checks if an error is a forbidden error.
func IsForbidden(err error) bool {
	return GetErrorCode(err) == ErrCodeForbidden
}

// IsInternal checks if an error is an internal server error.
func IsInternal(err error) bool {
	code := GetErrorCode(err)
	return code == ErrCodeInternal || code == ErrCodeDatabaseError || code == ErrCodeTransactionFailed
}
