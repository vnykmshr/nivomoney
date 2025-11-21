package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeNotFound, "resource not found")

	if err.Code != ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Message != "resource not found" {
		t.Errorf("Expected message 'resource not found', got '%s'", err.Message)
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "error without underlying error",
			err:      New(ErrCodeNotFound, "user not found"),
			expected: "[NOT_FOUND] user not found",
		},
		{
			name:     "error with underlying error",
			err:      Wrap(errors.New("db error"), ErrCodeDatabaseError, "failed to query"),
			expected: "[DATABASE_ERROR] failed to query: db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := Wrap(underlying, ErrCodeInternal, "wrapped")

	unwrapped := err.Unwrap()
	if unwrapped != underlying {
		t.Errorf("Unwrap() did not return underlying error")
	}
}

func TestError_WithDetails(t *testing.T) {
	err := New(ErrCodeValidation, "validation failed")
	details := map[string]interface{}{
		"field": "email",
		"issue": "invalid format",
	}

	err.WithDetails(details)

	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}
	if err.Details["field"] != "email" {
		t.Errorf("Expected field='email', got '%v'", err.Details["field"])
	}
}

func TestError_AddDetail(t *testing.T) {
	err := New(ErrCodeValidation, "validation failed")
	err.AddDetail("field", "username")
	err.AddDetail("min_length", 3)

	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}
	if err.Details["field"] != "username" {
		t.Errorf("Expected field='username', got '%v'", err.Details["field"])
	}
	if err.Details["min_length"] != 3 {
		t.Errorf("Expected min_length=3, got '%v'", err.Details["min_length"])
	}
}

func TestError_HTTPStatusCode(t *testing.T) {
	tests := []struct {
		code           ErrorCode
		expectedStatus int
	}{
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeBadRequest, http.StatusBadRequest},
		{ErrCodeValidation, http.StatusBadRequest},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeConflict, http.StatusConflict},
		{ErrCodeRateLimit, http.StatusTooManyRequests},
		{ErrCodePrecondition, http.StatusPreconditionFailed},
		{ErrCodeInternal, http.StatusInternalServerError},
		{ErrCodeUnavailable, http.StatusServiceUnavailable},
		{ErrCodeTimeout, http.StatusGatewayTimeout},
		{ErrCodeDatabaseError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := New(tt.code, "test error")
			if got := err.HTTPStatusCode(); got != tt.expectedStatus {
				t.Errorf("HTTPStatusCode() = %v, want %v", got, tt.expectedStatus)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	underlying := errors.New("db connection failed")
	err := Wrap(underlying, ErrCodeDatabaseError, "failed to connect")

	if err.Code != ErrCodeDatabaseError {
		t.Errorf("Expected code %s, got %s", ErrCodeDatabaseError, err.Code)
	}
	if err.Err != underlying {
		t.Error("Underlying error not preserved")
	}

	// Test wrapping nil
	nilErr := Wrap(nil, ErrCodeInternal, "test")
	if nilErr != nil {
		t.Error("Wrap(nil) should return nil")
	}
}

func TestWrapf(t *testing.T) {
	underlying := errors.New("connection refused")
	err := Wrapf(underlying, ErrCodeUnavailable, "failed to connect to %s:%d", "localhost", 5432)

	expected := "failed to connect to localhost:5432"
	if err.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, err.Message)
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("user")
	if err.Code != ErrCodeNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("Expected 'user not found', got '%s'", err.Message)
	}
}

func TestNotFoundWithID(t *testing.T) {
	err := NotFoundWithID("wallet", "wallet-123")
	if err.Message != "wallet with id wallet-123 not found" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}

func TestValidationWithFields(t *testing.T) {
	fields := map[string]string{
		"email":    "invalid format",
		"password": "too short",
	}
	err := ValidationWithFields("validation failed", fields)

	if err.Code != ErrCodeValidation {
		t.Errorf("Expected code %s, got %s", ErrCodeValidation, err.Code)
	}
	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}
}

func TestCommonConstructors(t *testing.T) {
	tests := []struct {
		name         string
		constructor  func() *Error
		expectedCode ErrorCode
	}{
		{"BadRequest", func() *Error { return BadRequest("bad input") }, ErrCodeBadRequest},
		{"Validation", func() *Error { return Validation("invalid") }, ErrCodeValidation},
		{"Unauthorized", func() *Error { return Unauthorized("no token") }, ErrCodeUnauthorized},
		{"Forbidden", func() *Error { return Forbidden("no access") }, ErrCodeForbidden},
		{"Conflict", func() *Error { return Conflict("duplicate") }, ErrCodeConflict},
		{"Internal", func() *Error { return Internal("server error") }, ErrCodeInternal},
		{"Database", func() *Error { return Database("query failed") }, ErrCodeDatabaseError},
		{"Unavailable", func() *Error { return Unavailable("down") }, ErrCodeUnavailable},
		{"Timeout", func() *Error { return Timeout("too slow") }, ErrCodeTimeout},
		{"InsufficientFunds", func() *Error { return InsufficientFunds("not enough") }, ErrCodeInsufficientFunds},
		{"AccountFrozen", func() *Error { return AccountFrozen("frozen") }, ErrCodeAccountFrozen},
		{"TransactionFailed", func() *Error { return TransactionFailed("tx failed") }, ErrCodeTransactionFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()
			if err.Code != tt.expectedCode {
				t.Errorf("Expected code %s, got %s", tt.expectedCode, err.Code)
			}
		})
	}
}

func TestInternalWrap(t *testing.T) {
	underlying := errors.New("panic recovered")
	err := InternalWrap(underlying, "internal server error")

	if err.Code != ErrCodeInternal {
		t.Errorf("Expected code %s, got %s", ErrCodeInternal, err.Code)
	}
	if err.Err != underlying {
		t.Error("Underlying error not preserved")
	}
}

func TestDatabaseWrap(t *testing.T) {
	underlying := errors.New("connection pool exhausted")
	err := DatabaseWrap(underlying, "database error")

	if err.Code != ErrCodeDatabaseError {
		t.Errorf("Expected code %s, got %s", ErrCodeDatabaseError, err.Code)
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode ErrorCode
	}{
		{
			name:         "custom error",
			err:          NotFound("user"),
			expectedCode: ErrCodeNotFound,
		},
		{
			name:         "standard error",
			err:          errors.New("standard error"),
			expectedCode: ErrCodeInternal,
		},
		{
			name:         "wrapped custom error",
			err:          Wrap(NotFound("user"), ErrCodeInternal, "wrapped"),
			expectedCode: ErrCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetErrorCode(tt.err)
			if code != tt.expectedCode {
				t.Errorf("GetErrorCode() = %s, want %s", code, tt.expectedCode)
			}
		})
	}
}

func TestGetHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "not found error",
			err:            NotFound("user"),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "validation error",
			err:            Validation("invalid input"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "standard error",
			err:            errors.New("standard error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := GetHTTPStatus(tt.err)
			if status != tt.expectedStatus {
				t.Errorf("GetHTTPStatus() = %d, want %d", status, tt.expectedStatus)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(NotFound("user")) {
		t.Error("IsNotFound() should return true for NotFound error")
	}
	if IsNotFound(Internal("error")) {
		t.Error("IsNotFound() should return false for non-NotFound error")
	}
}

func TestIsValidation(t *testing.T) {
	if !IsValidation(Validation("invalid")) {
		t.Error("IsValidation() should return true for Validation error")
	}
	if IsValidation(Internal("error")) {
		t.Error("IsValidation() should return false for non-Validation error")
	}
}

func TestIsUnauthorized(t *testing.T) {
	if !IsUnauthorized(Unauthorized("no token")) {
		t.Error("IsUnauthorized() should return true for Unauthorized error")
	}
	if IsUnauthorized(Internal("error")) {
		t.Error("IsUnauthorized() should return false for non-Unauthorized error")
	}
}

func TestIsForbidden(t *testing.T) {
	if !IsForbidden(Forbidden("no access")) {
		t.Error("IsForbidden() should return true for Forbidden error")
	}
	if IsForbidden(Internal("error")) {
		t.Error("IsForbidden() should return false for non-Forbidden error")
	}
}

func TestIsInternal(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"internal error", Internal("error"), true},
		{"database error", Database("error"), true},
		{"transaction failed", TransactionFailed("error"), true},
		{"not found error", NotFound("user"), false},
		{"validation error", Validation("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInternal(tt.err); got != tt.expected {
				t.Errorf("IsInternal() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDuplicateIdempotencyKey(t *testing.T) {
	err := DuplicateIdempotencyKey("key-123")
	if err.Code != ErrCodeDuplicateIdempotencyKey {
		t.Errorf("Expected code %s, got %s", ErrCodeDuplicateIdempotencyKey, err.Code)
	}
	if err.Message != "duplicate idempotency key: key-123" {
		t.Errorf("Unexpected message: %s", err.Message)
	}
}
