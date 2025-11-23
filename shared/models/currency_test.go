package models

import (
	"testing"
)

func TestCurrency_Validate(t *testing.T) {
	tests := []struct {
		name        string
		currency    Currency
		expectError bool
	}{
		{"valid USD", USD, false},
		{"valid EUR", EUR, false},
		{"valid GBP", GBP, false},
		{"empty currency", "", true},
		{"invalid currency", "XXX", true},
		{"lowercase valid", "usd", true}, // Should be uppercase
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currency.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCurrency_String(t *testing.T) {
	if USD.String() != "USD" {
		t.Errorf("Expected 'USD', got '%s'", USD.String())
	}
}

func TestCurrency_IsSupported(t *testing.T) {
	if !USD.IsSupported() {
		t.Error("USD should be supported")
	}
	if Currency("XXX").IsSupported() {
		t.Error("XXX should not be supported")
	}
}

func TestParseCurrency(t *testing.T) {
	tests := []struct {
		input       string
		expected    Currency
		expectError bool
	}{
		{"USD", USD, false},
		{"usd", USD, false},     // Should be case-insensitive
		{"  EUR  ", EUR, false}, // Should trim spaces
		{"XXX", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseCurrency(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestGetSupportedCurrencies(t *testing.T) {
	currencies := GetSupportedCurrencies()
	if len(currencies) == 0 {
		t.Error("Should return at least one supported currency")
	}

	// Check that USD is in the list
	found := false
	for _, c := range currencies {
		if c == USD {
			found = true
			break
		}
	}
	if !found {
		t.Error("USD should be in supported currencies")
	}
}

func TestCurrency_GetDecimalPlaces(t *testing.T) {
	tests := []struct {
		currency Currency
		expected int
	}{
		{USD, 2},
		{EUR, 2},
		{JPY, 0},
		{GBP, 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.currency), func(t *testing.T) {
			if got := tt.currency.GetDecimalPlaces(); got != tt.expected {
				t.Errorf("Expected %d decimal places, got %d", tt.expected, got)
			}
		})
	}
}

func TestCurrency_GetSymbol(t *testing.T) {
	tests := []struct {
		currency Currency
		expected string
	}{
		{USD, "$"},
		{EUR, "€"},
		{GBP, "£"},
		{JPY, "¥"},
		{INR, "₹"},
	}

	for _, tt := range tests {
		t.Run(string(tt.currency), func(t *testing.T) {
			if got := tt.currency.GetSymbol(); got != tt.expected {
				t.Errorf("Expected symbol '%s', got '%s'", tt.expected, got)
			}
		})
	}
}
