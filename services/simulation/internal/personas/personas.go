package personas

import (
	"math/rand"
	"time"
)

// PersonaType represents different user behavior patterns
type PersonaType string

const (
	PersonaFrequentTrader PersonaType = "frequent_trader" // High-frequency small transactions
	PersonaSaver          PersonaType = "saver"           // Mostly deposits, rare withdrawals
	PersonaBillPayer      PersonaType = "bill_payer"      // Regular scheduled payments
	PersonaShopper        PersonaType = "shopper"         // Frequent payments, occasional deposits
	PersonaInvestor       PersonaType = "investor"        // Large deposits, strategic transfers
	PersonaCasual         PersonaType = "casual"          // Random sporadic activity
)

// Persona defines behavior patterns for simulation
type Persona struct {
	Type             PersonaType
	TransactionFreq  time.Duration  // How often transactions occur
	AmountRange      AmountRange    // Typical transaction amounts
	TransactionTypes map[string]int // Probability weights for transaction types
	ActiveHours      []int          // Hours of day when active (0-23)
	BalanceThreshold int64          // Minimum balance before deposits
}

// AmountRange defines min and max for transaction amounts
type AmountRange struct {
	MinPaise int64
	MaxPaise int64
}

// GetPersona returns a persona configuration by type
func GetPersona(pType PersonaType) *Persona {
	personas := map[PersonaType]*Persona{
		PersonaFrequentTrader: {
			Type:            PersonaFrequentTrader,
			TransactionFreq: 2 * time.Minute,
			AmountRange:     AmountRange{MinPaise: 1000, MaxPaise: 50000}, // ₹10 - ₹500
			TransactionTypes: map[string]int{
				"transfer":   60,
				"deposit":    30,
				"withdrawal": 10,
			},
			ActiveHours:      []int{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			BalanceThreshold: 100000, // ₹1000
		},
		PersonaSaver: {
			Type:            PersonaSaver,
			TransactionFreq: 24 * time.Hour,
			AmountRange:     AmountRange{MinPaise: 50000, MaxPaise: 1000000}, // ₹500 - ₹10,000
			TransactionTypes: map[string]int{
				"deposit":    80,
				"transfer":   15,
				"withdrawal": 5,
			},
			ActiveHours:      []int{10, 11, 14, 15, 19, 20},
			BalanceThreshold: 500000, // ₹5000
		},
		PersonaBillPayer: {
			Type:            PersonaBillPayer,
			TransactionFreq: 6 * time.Hour,
			AmountRange:     AmountRange{MinPaise: 50000, MaxPaise: 500000}, // ₹500 - ₹5,000
			TransactionTypes: map[string]int{
				"transfer":   70,
				"deposit":    25,
				"withdrawal": 5,
			},
			ActiveHours:      []int{9, 10, 11, 17, 18, 19},
			BalanceThreshold: 200000, // ₹2000
		},
		PersonaShopper: {
			Type:            PersonaShopper,
			TransactionFreq: 4 * time.Hour,
			AmountRange:     AmountRange{MinPaise: 10000, MaxPaise: 200000}, // ₹100 - ₹2,000
			TransactionTypes: map[string]int{
				"transfer":   75,
				"deposit":    20,
				"withdrawal": 5,
			},
			ActiveHours:      []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21},
			BalanceThreshold: 150000, // ₹1500
		},
		PersonaInvestor: {
			Type:            PersonaInvestor,
			TransactionFreq: 48 * time.Hour,
			AmountRange:     AmountRange{MinPaise: 500000, MaxPaise: 10000000}, // ₹5,000 - ₹1,00,000
			TransactionTypes: map[string]int{
				"deposit":    50,
				"transfer":   45,
				"withdrawal": 5,
			},
			ActiveHours:      []int{10, 11, 14, 15, 16},
			BalanceThreshold: 10000000, // ₹1,00,000
		},
		PersonaCasual: {
			Type:            PersonaCasual,
			TransactionFreq: 12 * time.Hour,
			AmountRange:     AmountRange{MinPaise: 5000, MaxPaise: 100000}, // ₹50 - ₹1,000
			TransactionTypes: map[string]int{
				"transfer":   50,
				"deposit":    40,
				"withdrawal": 10,
			},
			ActiveHours:      []int{9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22},
			BalanceThreshold: 50000, // ₹500
		},
	}

	return personas[pType]
}

// AllPersonaTypes returns all available persona types
func AllPersonaTypes() []PersonaType {
	return []PersonaType{
		PersonaFrequentTrader,
		PersonaSaver,
		PersonaBillPayer,
		PersonaShopper,
		PersonaInvestor,
		PersonaCasual,
	}
}

// RandomAmount generates a random amount within the persona's range
func (p *Persona) RandomAmount() int64 {
	return p.AmountRange.MinPaise + rand.Int63n(p.AmountRange.MaxPaise-p.AmountRange.MinPaise+1) //nolint:gosec // G404: weak random acceptable for test data
}

// SelectTransactionType randomly selects a transaction type based on weights
func (p *Persona) SelectTransactionType() string {
	total := 0
	for _, weight := range p.TransactionTypes {
		total += weight
	}

	r := rand.Intn(total) //nolint:gosec // G404: weak random acceptable for test data
	cumulative := 0

	for txType, weight := range p.TransactionTypes {
		cumulative += weight
		if r < cumulative {
			return txType
		}
	}

	return "transfer" // fallback
}

// IsActiveHour checks if current hour is in active hours
func (p *Persona) IsActiveHour(hour int) bool {
	for _, h := range p.ActiveHours {
		if h == hour {
			return true
		}
	}
	return false
}
