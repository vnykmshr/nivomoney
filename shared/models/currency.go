package models

import (
	"fmt"
	"strings"
)

// Currency represents an ISO 4217 currency code.
type Currency string

// Supported currencies
// Note: INR (Indian Rupee) is listed first as the primary currency for India-centric operations
const (
	INR Currency = "INR" // Indian Rupee (primary)
	USD Currency = "USD" // US Dollar
	EUR Currency = "EUR" // Euro
	GBP Currency = "GBP" // British Pound
	JPY Currency = "JPY" // Japanese Yen
	CNY Currency = "CNY" // Chinese Yuan
	CAD Currency = "CAD" // Canadian Dollar
	AUD Currency = "AUD" // Australian Dollar
	CHF Currency = "CHF" // Swiss Franc
	SGD Currency = "SGD" // Singapore Dollar
)

// DefaultCurrency is the default currency for the system (India-centric)
const DefaultCurrency = INR

// supportedCurrencies is a map of all supported currencies.
var supportedCurrencies = map[Currency]bool{
	INR: true, // Primary currency
	USD: true,
	EUR: true,
	GBP: true,
	JPY: true,
	CNY: true,
	CAD: true,
	AUD: true,
	CHF: true,
	SGD: true,
}

// Validate checks if the currency is supported.
func (c Currency) Validate() error {
	if c == "" {
		return fmt.Errorf("currency code is required")
	}
	if !supportedCurrencies[c] {
		return fmt.Errorf("unsupported currency: %s", c)
	}
	return nil
}

// String returns the currency code as a string.
func (c Currency) String() string {
	return string(c)
}

// IsSupported returns true if the currency is supported.
func (c Currency) IsSupported() bool {
	return supportedCurrencies[c]
}

// ParseCurrency parses a string into a Currency.
func ParseCurrency(s string) (Currency, error) {
	c := Currency(strings.ToUpper(strings.TrimSpace(s)))
	if err := c.Validate(); err != nil {
		return "", err
	}
	return c, nil
}

// GetSupportedCurrencies returns a list of all supported currencies.
func GetSupportedCurrencies() []Currency {
	currencies := make([]Currency, 0, len(supportedCurrencies))
	for c := range supportedCurrencies {
		currencies = append(currencies, c)
	}
	return currencies
}

// GetDecimalPlaces returns the number of decimal places for a currency.
// Most currencies use 2 decimal places, but some (like JPY) use 0.
func (c Currency) GetDecimalPlaces() int {
	switch c {
	case JPY:
		return 0
	default:
		return 2
	}
}

// GetSymbol returns the currency symbol.
func (c Currency) GetSymbol() string {
	switch c {
	case INR:
		return "₹" // Primary currency
	case USD:
		return "$"
	case EUR:
		return "€"
	case GBP:
		return "£"
	case JPY:
		return "¥"
	case CNY:
		return "¥"
	case CAD:
		return "C$"
	case AUD:
		return "A$"
	case CHF:
		return "CHF"
	case SGD:
		return "S$"
	default:
		return string(c)
	}
}
