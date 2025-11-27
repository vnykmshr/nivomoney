package repository

import (
	"testing"
)

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "percent sign",
			input:    "test%value",
			expected: `test\%value`,
		},
		{
			name:     "underscore",
			input:    "test_value",
			expected: `test\_value`,
		},
		{
			name:     "backslash",
			input:    `test\value`,
			expected: `test\\value`,
		},
		{
			name:     "all special characters",
			input:    `test%_\value`,
			expected: `test\%\_\\value`,
		},
		{
			name:     "multiple percent signs",
			input:    "%%test%%",
			expected: `\%\%test\%\%`,
		},
		{
			name:     "SQL injection attempt",
			input:    "%' OR '1'='1",
			expected: `\%' OR '1'='1`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeLikePattern(tt.input)
			if result != tt.expected {
				t.Errorf("escapeLikePattern(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
