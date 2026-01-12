package config

import "time"

// API Limits - used across handlers for consistent validation
const (
	// MaxSearchQueryLength is the maximum length for search query parameters.
	MaxSearchQueryLength = 200

	// DefaultPageLimit is the default number of items returned in paginated responses.
	DefaultPageLimit = 50

	// MaxPageLimit is the maximum number of items that can be requested in a single page.
	MaxPageLimit = 100

	// MaxStatementDays is the maximum date range for statement exports.
	MaxStatementDays = 365

	// MaxStatementDuration is MaxStatementDays as a time.Duration.
	MaxStatementDuration = MaxStatementDays * 24 * time.Hour
)
