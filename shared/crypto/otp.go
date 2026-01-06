// Package crypto provides cryptographic utilities for the application.
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"math/big"
)

// GenerateOTP generates a cryptographically secure N-digit OTP.
// The digits parameter must be between 4 and 10.
func GenerateOTP(digits int) (string, error) {
	if digits < 4 || digits > 10 {
		return "", fmt.Errorf("OTP digits must be between 4 and 10, got %d", digits)
	}

	// Calculate max value (10^digits)
	max := new(big.Int)
	max.Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)

	// Generate random number
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n), nil
}

// GenerateOTP6 generates a cryptographically secure 6-digit OTP.
// This is the most common OTP format.
func GenerateOTP6() (string, error) {
	return GenerateOTP(6)
}

// GenerateOTP4 generates a cryptographically secure 4-digit OTP.
func GenerateOTP4() (string, error) {
	return GenerateOTP(4)
}

// ValidateOTPFormat validates that the OTP has the correct format.
// It checks that the OTP is exactly expectedDigits long and contains only digits.
// Note: This does NOT verify correctness, only format.
func ValidateOTPFormat(otp string, expectedDigits int) bool {
	if len(otp) != expectedDigits {
		return false
	}
	for _, c := range otp {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// SecureCompare performs constant-time string comparison.
// This prevents timing attacks when comparing secrets like OTP codes.
func SecureCompare(a, b string) bool {
	// Convert to bytes for comparison
	aBytes := []byte(a)
	bBytes := []byte(b)

	// Use constant time compare
	return subtle.ConstantTimeCompare(aBytes, bBytes) == 1
}

// SecureCompareBytes performs constant-time byte slice comparison.
func SecureCompareBytes(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
