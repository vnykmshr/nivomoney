package config

import (
	"sync"
)

// SimulationMode defines the operation mode.
type SimulationMode string

const (
	// ModeRealistic enables delays and failures for realistic behavior.
	ModeRealistic SimulationMode = "realistic"
	// ModeDemo uses fast paths with no failures for presentations.
	ModeDemo SimulationMode = "demo"
)

// SimulationConfig holds runtime configuration for simulation behavior.
type SimulationConfig struct {
	mu sync.RWMutex

	// Mode determines overall behavior profile.
	Mode SimulationMode `json:"mode"`

	// Delays configuration.
	Delays DelayConfig `json:"delays"`

	// Failures configuration.
	Failures FailureConfig `json:"failures"`

	// AutoVerification configuration for simulated users.
	AutoVerification AutoVerificationConfig `json:"auto_verification"`

	// Persona activity configuration.
	Personas PersonaConfig `json:"personas"`
}

// DelayConfig configures operation delays in milliseconds.
type DelayConfig struct {
	// MinDelayMs is the minimum delay for any operation.
	MinDelayMs int `json:"min_delay_ms"`
	// MaxDelayMs is the maximum general delay.
	MaxDelayMs int `json:"max_delay_ms"`

	// Operation-specific delays.
	TransferDelayMs     int `json:"transfer_delay_ms"`
	KYCReviewDelayMs    int `json:"kyc_review_delay_ms"`
	VerificationDelayMs int `json:"verification_delay_ms"`
}

// FailureConfig configures failure injection.
type FailureConfig struct {
	// Enabled controls whether failures are injected.
	Enabled bool `json:"enabled"`
	// FailureRate is the default failure probability (0.0-1.0).
	FailureRate float64 `json:"failure_rate"`

	// Operation-specific failure rates.
	TransferFailureRate float64 `json:"transfer_failure_rate"`
	KYCRejectRate       float64 `json:"kyc_reject_rate"`
}

// AutoVerificationConfig configures automatic verification for simulated users.
type AutoVerificationConfig struct {
	// Enabled controls whether simulated users bypass verification.
	Enabled bool `json:"enabled"`
	// DelayMs is the delay before auto-approving verification.
	DelayMs int `json:"delay_ms"`
}

// PersonaConfig configures persona activity.
type PersonaConfig struct {
	// Enabled controls whether personas are active.
	Enabled bool `json:"enabled"`
	// IntervalSeconds is the activity check interval.
	IntervalSeconds int `json:"interval_seconds"`
	// TransactionsPerHour is the target transaction rate.
	TransactionsPerHour int `json:"transactions_per_hour"`
}

// NewDefaultConfig returns realistic mode configuration.
func NewDefaultConfig() *SimulationConfig {
	return &SimulationConfig{
		Mode: ModeRealistic,
		Delays: DelayConfig{
			MinDelayMs:          500,
			MaxDelayMs:          3000,
			TransferDelayMs:     3000,
			KYCReviewDelayMs:    10000,
			VerificationDelayMs: 1000,
		},
		Failures: FailureConfig{
			Enabled:             true,
			FailureRate:         0.05, // 5%
			TransferFailureRate: 0.03, // 3%
			KYCRejectRate:       0.10, // 10%
		},
		AutoVerification: AutoVerificationConfig{
			Enabled: true,
			DelayMs: 500,
		},
		Personas: PersonaConfig{
			Enabled:             true,
			IntervalSeconds:     60,
			TransactionsPerHour: 10,
		},
	}
}

// NewDemoConfig returns demo mode configuration with minimal delays and no failures.
func NewDemoConfig() *SimulationConfig {
	return &SimulationConfig{
		Mode: ModeDemo,
		Delays: DelayConfig{
			MinDelayMs:          100,
			MaxDelayMs:          500,
			TransferDelayMs:     300,
			KYCReviewDelayMs:    500,
			VerificationDelayMs: 100,
		},
		Failures: FailureConfig{
			Enabled:             false,
			FailureRate:         0,
			TransferFailureRate: 0,
			KYCRejectRate:       0,
		},
		AutoVerification: AutoVerificationConfig{
			Enabled: true,
			DelayMs: 100,
		},
		Personas: PersonaConfig{
			Enabled:             true,
			IntervalSeconds:     30,
			TransactionsPerHour: 20,
		},
	}
}

// ConfigView is a JSON-safe representation of the configuration (no mutex).
type ConfigView struct {
	Mode             SimulationMode         `json:"mode"`
	Delays           DelayConfig            `json:"delays"`
	Failures         FailureConfig          `json:"failures"`
	AutoVerification AutoVerificationConfig `json:"auto_verification"`
	Personas         PersonaConfig          `json:"personas"`
}

// GetView returns a JSON-safe view of the current configuration.
func (c *SimulationConfig) GetView() ConfigView {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ConfigView{
		Mode:             c.Mode,
		Delays:           c.Delays,
		Failures:         c.Failures,
		AutoVerification: c.AutoVerification,
		Personas:         c.Personas,
	}
}

// Get returns a thread-safe copy of the current configuration.
// NOTE: This copies the mutex - use GetView() for JSON serialization.
func (c *SimulationConfig) Get() SimulationConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return SimulationConfig{
		Mode:             c.Mode,
		Delays:           c.Delays,
		Failures:         c.Failures,
		AutoVerification: c.AutoVerification,
		Personas:         c.Personas,
	}
}

// Update applies changes to the configuration in a thread-safe manner.
func (c *SimulationConfig) Update(update func(*SimulationConfig)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	update(c)
}

// SetMode switches between realistic and demo modes.
func (c *SimulationConfig) SetMode(mode SimulationMode) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if mode == ModeDemo {
		demo := NewDemoConfig()
		c.Mode = demo.Mode
		c.Delays = demo.Delays
		c.Failures = demo.Failures
		c.AutoVerification = demo.AutoVerification
	} else {
		realistic := NewDefaultConfig()
		c.Mode = realistic.Mode
		c.Delays = realistic.Delays
		c.Failures = realistic.Failures
		c.AutoVerification = realistic.AutoVerification
	}
}

// IsRealistic returns true if running in realistic mode.
func (c *SimulationConfig) IsRealistic() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Mode == ModeRealistic
}

// IsDemo returns true if running in demo mode.
func (c *SimulationConfig) IsDemo() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Mode == ModeDemo
}

// Delay getters - thread-safe accessors for delay configuration.

// GetMinDelayMs returns the minimum delay in milliseconds.
func (c *SimulationConfig) GetMinDelayMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Delays.MinDelayMs
}

// GetMaxDelayMs returns the maximum delay in milliseconds.
func (c *SimulationConfig) GetMaxDelayMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Delays.MaxDelayMs
}

// GetTransferDelayMs returns the transfer delay in milliseconds.
func (c *SimulationConfig) GetTransferDelayMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Delays.TransferDelayMs
}

// GetKYCReviewDelayMs returns the KYC review delay in milliseconds.
func (c *SimulationConfig) GetKYCReviewDelayMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Delays.KYCReviewDelayMs
}

// GetVerificationDelayMs returns the verification delay in milliseconds.
func (c *SimulationConfig) GetVerificationDelayMs() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Delays.VerificationDelayMs
}

// Failure getters - thread-safe accessors for failure configuration.

// IsFailuresEnabled returns true if failure injection is enabled.
func (c *SimulationConfig) IsFailuresEnabled() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Failures.Enabled
}

// GetFailureRate returns the general failure rate.
func (c *SimulationConfig) GetFailureRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Failures.FailureRate
}

// GetTransferFailureRate returns the transfer failure rate.
func (c *SimulationConfig) GetTransferFailureRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Failures.TransferFailureRate
}

// GetKYCRejectRate returns the KYC rejection rate.
func (c *SimulationConfig) GetKYCRejectRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Failures.KYCRejectRate
}
