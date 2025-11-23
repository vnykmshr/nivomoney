// Package validator provides input validation for Nivo using gopantic.
package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/models"
)

func init() {
	// Register all custom validators on package initialization
	registerStandardValidators()
	registerFintechValidators()
	registerIndiaValidators()
}

// ParseAndValidate parses JSON/YAML data into a struct with validation.
// This is the primary API for validating HTTP request bodies.
func ParseAndValidate[T any](data []byte) (T, error) {
	result, err := model.ParseInto[T](data)
	if err != nil {
		return result, convertError(err)
	}
	return result, nil
}

// convertError converts gopantic validation errors to our error format.
func convertError(err error) *errors.Error {
	if err == nil {
		return nil
	}

	// Check if it's a validation error
	validationErr, ok := err.(*model.ValidationError)
	if ok {
		return errors.Validation(validationErr.Error()).
			AddDetail("field", validationErr.Field).
			AddDetail("value", validationErr.Value)
	}

	// Check for multiple errors
	if strings.Contains(err.Error(), "multiple errors:") {
		// Parse multiple validation errors
		details := make(map[string]interface{})
		errMsg := err.Error()

		// Extract field errors from the error message
		// This is a simple parser - gopantic aggregates errors in the message
		lines := strings.Split(errMsg, ";")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "validation error on field") {
				// Extract field name and error
				parts := strings.Split(line, "'")
				if len(parts) >= 2 {
					fieldName := parts[1]
					if len(parts) >= 4 {
						errorMsg := strings.TrimSpace(parts[3])
						details[fieldName] = errorMsg
					}
				}
			}
		}

		if len(details) > 0 {
			return errors.Validation("validation failed").WithDetails(details)
		}
	}

	// Generic validation error
	return errors.Validation(err.Error())
}

// registerStandardValidators registers standard comparison and format validators.
func registerStandardValidators() {
	// Greater than (gt) - numeric comparison
	model.RegisterGlobalFunc("gt", func(fieldName string, value interface{}, params map[string]interface{}) error {
		threshold, ok := params["value"].(float64)
		if !ok {
			return model.NewValidationError(fieldName, value, "gt", "gt parameter must be a number")
		}

		numValue, err := toFloat64(value)
		if err != nil {
			return model.NewValidationError(fieldName, value, "gt", "value must be numeric")
		}

		if numValue <= threshold {
			return model.NewValidationError(fieldName, value, "gt",
				fmt.Sprintf("must be greater than %g", threshold))
		}
		return nil
	})

	// Greater than or equal (gte)
	model.RegisterGlobalFunc("gte", func(fieldName string, value interface{}, params map[string]interface{}) error {
		threshold, ok := params["value"].(float64)
		if !ok {
			return model.NewValidationError(fieldName, value, "gte", "gte parameter must be a number")
		}

		numValue, err := toFloat64(value)
		if err != nil {
			return model.NewValidationError(fieldName, value, "gte", "value must be numeric")
		}

		if numValue < threshold {
			return model.NewValidationError(fieldName, value, "gte",
				fmt.Sprintf("must be greater than or equal to %g", threshold))
		}
		return nil
	})

	// Less than (lt)
	model.RegisterGlobalFunc("lt", func(fieldName string, value interface{}, params map[string]interface{}) error {
		threshold, ok := params["value"].(float64)
		if !ok {
			return model.NewValidationError(fieldName, value, "lt", "lt parameter must be a number")
		}

		numValue, err := toFloat64(value)
		if err != nil {
			return model.NewValidationError(fieldName, value, "lt", "value must be numeric")
		}

		if numValue >= threshold {
			return model.NewValidationError(fieldName, value, "lt",
				fmt.Sprintf("must be less than %g", threshold))
		}
		return nil
	})

	// Less than or equal (lte)
	model.RegisterGlobalFunc("lte", func(fieldName string, value interface{}, params map[string]interface{}) error {
		threshold, ok := params["value"].(float64)
		if !ok {
			return model.NewValidationError(fieldName, value, "lte", "lte parameter must be a number")
		}

		numValue, err := toFloat64(value)
		if err != nil {
			return model.NewValidationError(fieldName, value, "lte", "value must be numeric")
		}

		if numValue > threshold {
			return model.NewValidationError(fieldName, value, "lte",
				fmt.Sprintf("must be less than or equal to %g", threshold))
		}
		return nil
	})

	// One of (oneof) - enum validation
	model.RegisterGlobalFunc("oneof", func(fieldName string, value interface{}, params map[string]interface{}) error {
		allowedStr, ok := params["value"].(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "oneof", "oneof parameter must be a string")
		}

		allowed := strings.Fields(allowedStr)
		valueStr := fmt.Sprintf("%v", value)

		for _, a := range allowed {
			if valueStr == a {
				return nil
			}
		}

		return model.NewValidationError(fieldName, value, "oneof",
			fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
	})

	// UUID validation
	model.RegisterGlobalFunc("uuid", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "uuid", "value must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(strings.ToLower(str)) {
			return model.NewValidationError(fieldName, value, "uuid", "must be a valid UUID")
		}
		return nil
	})

	// URL validation
	model.RegisterGlobalFunc("url", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "url", "value must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "url", "must be a valid URL")
		}
		return nil
	})

	// Numeric string validation
	model.RegisterGlobalFunc("numeric", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "numeric", "value must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		numericRegex := regexp.MustCompile(`^[0-9]+$`)
		if !numericRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "numeric", "must contain only numbers")
		}
		return nil
	})
}

// registerFintechValidators registers fintech-specific validators.
func registerFintechValidators() {
	// Currency validator - ISO 4217 currency codes
	model.RegisterGlobalFunc("currency", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "currency", "currency must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		currency := models.Currency(strings.ToUpper(str))
		if !currency.IsSupported() {
			return model.NewValidationError(fieldName, value, "currency",
				fmt.Sprintf("unsupported currency code: %s", str))
		}
		return nil
	})

	// Money amount validator - positive amounts in cents
	model.RegisterGlobalFunc("money_amount", func(fieldName string, value interface{}, params map[string]interface{}) error {
		amount, err := toInt64(value)
		if err != nil {
			return model.NewValidationError(fieldName, value, "money_amount", "amount must be a number")
		}

		if amount <= 0 {
			return model.NewValidationError(fieldName, value, "money_amount",
				"amount must be positive (in cents)")
		}
		return nil
	})

	// Account number validator - 10-20 alphanumeric characters
	model.RegisterGlobalFunc("account_number", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "account_number", "account number must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		if len(str) < 10 || len(str) > 20 {
			return model.NewValidationError(fieldName, value, "account_number",
				"account number must be 10-20 characters")
		}

		accountRegex := regexp.MustCompile(`^[A-Z0-9]+$`)
		if !accountRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "account_number",
				"account number must contain only uppercase letters and numbers")
		}
		return nil
	})

	// IBAN validator - International Bank Account Number
	model.RegisterGlobalFunc("iban", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "iban", "IBAN must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Remove spaces
		iban := strings.ReplaceAll(str, " ", "")

		if len(iban) < 15 || len(iban) > 34 {
			return model.NewValidationError(fieldName, value, "iban",
				"IBAN must be 15-34 characters")
		}

		// First 2 characters must be letters (country code)
		if !isLetter(rune(iban[0])) || !isLetter(rune(iban[1])) {
			return model.NewValidationError(fieldName, value, "iban",
				"IBAN must start with 2-letter country code")
		}

		// Next 2 characters must be digits (check digits)
		if !isDigit(rune(iban[2])) || !isDigit(rune(iban[3])) {
			return model.NewValidationError(fieldName, value, "iban",
				"IBAN check digits must be numeric")
		}

		return nil
	})

	// Sort code validator - UK bank sort codes (6 digits)
	model.RegisterGlobalFunc("sort_code", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "sort_code", "sort code must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Remove hyphens
		sortCode := strings.ReplaceAll(str, "-", "")

		if len(sortCode) != 6 {
			return model.NewValidationError(fieldName, value, "sort_code",
				"sort code must be 6 digits")
		}

		sortCodeRegex := regexp.MustCompile(`^[0-9]{6}$`)
		if !sortCodeRegex.MatchString(sortCode) {
			return model.NewValidationError(fieldName, value, "sort_code",
				"sort code must contain only digits")
		}
		return nil
	})

	// Routing number validator - US bank routing numbers (9 digits)
	model.RegisterGlobalFunc("routing_number", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "routing_number", "routing number must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		if len(str) != 9 {
			return model.NewValidationError(fieldName, value, "routing_number",
				"routing number must be 9 digits")
		}

		routingRegex := regexp.MustCompile(`^[0-9]{9}$`)
		if !routingRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "routing_number",
				"routing number must contain only digits")
		}
		return nil
	})
}

// registerIndiaValidators registers India-specific validators for financial compliance.
func registerIndiaValidators() {
	// IFSC code validator - Indian Financial System Code (11 characters)
	// Format: 4 bank code + 0 (reserved) + 6 branch code
	// Example: SBIN0001234
	model.RegisterGlobalFunc("ifsc", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "ifsc", "IFSC code must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Must be exactly 11 characters
		if len(str) != 11 {
			return model.NewValidationError(fieldName, value, "ifsc",
				"IFSC code must be exactly 11 characters")
		}

		// First 4 characters must be letters (bank code)
		for i := 0; i < 4; i++ {
			if !isLetter(rune(str[i])) {
				return model.NewValidationError(fieldName, value, "ifsc",
					"IFSC code must start with 4 letters (bank code)")
			}
		}

		// 5th character must be 0 (reserved)
		if str[4] != '0' {
			return model.NewValidationError(fieldName, value, "ifsc",
				"IFSC code 5th character must be 0")
		}

		// Last 6 characters must be alphanumeric (branch code)
		ifscRegex := regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
		if !ifscRegex.MatchString(strings.ToUpper(str)) {
			return model.NewValidationError(fieldName, value, "ifsc",
				"invalid IFSC code format (expected: BANK0BRANCH)")
		}

		return nil
	})

	// UPI ID validator - Unified Payments Interface ID
	// Format: username@bankcode
	// Example: user@paytm, john.doe@okaxis
	model.RegisterGlobalFunc("upi_id", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "upi_id", "UPI ID must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Must contain exactly one @
		parts := strings.Split(str, "@")
		if len(parts) != 2 {
			return model.NewValidationError(fieldName, value, "upi_id",
				"UPI ID must be in format username@bankcode")
		}

		username, bankcode := parts[0], parts[1]

		// Username validation (alphanumeric, dots, hyphens, underscores)
		if username == "" {
			return model.NewValidationError(fieldName, value, "upi_id",
				"UPI ID username cannot be empty")
		}

		usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
		if !usernameRegex.MatchString(username) {
			return model.NewValidationError(fieldName, value, "upi_id",
				"UPI ID username must contain only letters, numbers, dots, hyphens, or underscores")
		}

		// Bank code validation (alphanumeric only)
		if bankcode == "" {
			return model.NewValidationError(fieldName, value, "upi_id",
				"UPI ID bank code cannot be empty")
		}

		bankcodeRegex := regexp.MustCompile(`^[a-z0-9]+$`)
		if !bankcodeRegex.MatchString(bankcode) {
			return model.NewValidationError(fieldName, value, "upi_id",
				"UPI ID bank code must contain only lowercase letters and numbers")
		}

		return nil
	})

	// PAN card validator - Permanent Account Number
	// Format: 5 letters + 4 digits + 1 letter (10 characters total)
	// Example: ABCDE1234F
	model.RegisterGlobalFunc("pan", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "pan", "PAN must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Must be exactly 10 characters
		if len(str) != 10 {
			return model.NewValidationError(fieldName, value, "pan",
				"PAN must be exactly 10 characters")
		}

		// Validate PAN format: 5 letters + 4 digits + 1 letter (must be uppercase)
		panRegex := regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)
		if !panRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "pan",
				"invalid PAN format (expected: ABCDE1234F in uppercase)")
		}

		return nil
	})

	// Aadhaar validator - Unique Identification Number
	// Format: 12 digits
	// Example: 123456789012
	model.RegisterGlobalFunc("aadhaar", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "aadhaar", "Aadhaar must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Remove spaces if any (Aadhaar is sometimes formatted as XXXX XXXX XXXX)
		str = strings.ReplaceAll(str, " ", "")

		// Must be exactly 12 digits
		if len(str) != 12 {
			return model.NewValidationError(fieldName, value, "aadhaar",
				"Aadhaar must be exactly 12 digits")
		}

		// All characters must be digits
		aadhaarRegex := regexp.MustCompile(`^[0-9]{12}$`)
		if !aadhaarRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "aadhaar",
				"Aadhaar must contain only digits")
		}

		// Aadhaar cannot start with 0 or 1
		if str[0] == '0' || str[0] == '1' {
			return model.NewValidationError(fieldName, value, "aadhaar",
				"Aadhaar cannot start with 0 or 1")
		}

		return nil
	})

	// Indian phone number validator
	// Format: +91 followed by 10 digits (optional hyphens/spaces)
	// Examples: +919876543210, +91-9876543210, +91 98765 43210
	model.RegisterGlobalFunc("indian_phone", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "indian_phone", "phone number must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Remove spaces and hyphens for validation
		cleaned := strings.ReplaceAll(strings.ReplaceAll(str, " ", ""), "-", "")

		// Check for +91 prefix
		if !strings.HasPrefix(cleaned, "+91") {
			return model.NewValidationError(fieldName, value, "indian_phone",
				"Indian phone number must start with +91")
		}

		// Remove +91 prefix
		number := strings.TrimPrefix(cleaned, "+91")

		// Must be exactly 10 digits
		if len(number) != 10 {
			return model.NewValidationError(fieldName, value, "indian_phone",
				"Indian phone number must have 10 digits after +91")
		}

		// All characters must be digits
		phoneRegex := regexp.MustCompile(`^[6-9][0-9]{9}$`)
		if !phoneRegex.MatchString(number) {
			return model.NewValidationError(fieldName, value, "indian_phone",
				"Indian mobile number must start with 6-9 and contain 10 digits")
		}

		return nil
	})

	// PIN code validator - Indian Postal Index Number
	// Format: 6 digits
	// Example: 560001, 110001
	model.RegisterGlobalFunc("pincode", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "pincode", "PIN code must be a string")
		}

		if str == "" {
			return nil // Empty handled by required
		}

		// Must be exactly 6 digits
		if len(str) != 6 {
			return model.NewValidationError(fieldName, value, "pincode",
				"PIN code must be exactly 6 digits")
		}

		// All characters must be digits
		pincodeRegex := regexp.MustCompile(`^[0-9]{6}$`)
		if !pincodeRegex.MatchString(str) {
			return model.NewValidationError(fieldName, value, "pincode",
				"PIN code must contain only digits")
		}

		// First digit cannot be 0
		if str[0] == '0' {
			return model.NewValidationError(fieldName, value, "pincode",
				"PIN code cannot start with 0")
		}

		return nil
	})
}

// Helper functions

func toFloat64(value interface{}) (float64, error) {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

func toInt64(value interface{}) (int64, error) {
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(val.Float()), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

func isLetter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
