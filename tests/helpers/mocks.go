// Package helpers provides test utilities including mock implementations.
package helpers

import (
	"time"

	"github.com/vnykmshr/nivo/shared/errors"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// ============================================================
// Shared Mock Helpers
// ============================================================

// MockError creates a mock error for testing.
func MockError(code string, message string) *errors.Error {
	switch code {
	case "NOT_FOUND":
		return errors.NotFound(message)
	case "CONFLICT":
		return errors.Conflict(message)
	case "UNAUTHORIZED":
		return errors.Unauthorized(message)
	case "FORBIDDEN":
		return errors.Forbidden(message)
	case "BAD_REQUEST":
		return errors.BadRequest(message)
	case "VALIDATION":
		return errors.Validation(message)
	default:
		return errors.Internal(message)
	}
}

// ============================================================
// Test Data Generators
// ============================================================

var testCounter int

// GenerateTestID generates a unique test ID with a prefix.
func GenerateTestID(prefix string) string {
	testCounter++
	return prefix + "-test-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+testCounter%10))
}

// GenerateTestEmail generates a unique test email.
func GenerateTestEmail() string {
	testCounter++
	return "test" + time.Now().Format("150405") + string(rune('0'+testCounter%10)) + "@nivo.test"
}

// GenerateTestPhone generates a unique test phone number.
func GenerateTestPhone() string {
	testCounter++
	ts := time.Now().UnixNano() % 1000000000
	return "+919" + string(rune('0'+testCounter%10)) + time.Now().Format("150405") + string(rune('0'+int(ts)%10))
}

// GenerateTestName generates a unique test name.
func GenerateTestName() string {
	testCounter++
	return "Test User " + time.Now().Format("150405") + string(rune('0'+testCounter%10))
}

// TestMoney creates a Money value for testing.
func TestMoney(amount int64, currency string) sharedModels.Money {
	return sharedModels.Money{
		Amount:   amount,
		Currency: sharedModels.Currency(currency),
	}
}

// ============================================================
// JWT Test Helpers
// ============================================================

// TestJWTSecret is a secret used for testing JWT tokens.
const TestJWTSecret = "test-jwt-secret-for-testing-only-32chars!"

// TestUserClaims represents claims for test JWT tokens.
type TestUserClaims struct {
	UserID      string
	Email       string
	AccountType string
}
