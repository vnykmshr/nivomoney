package metrics

import (
	"sync"
	"time"
)

// SimulationMetrics tracks simulation activity metrics.
type SimulationMetrics struct {
	mu sync.RWMutex

	// Operation counters.
	OperationsTotal     int64 `json:"operations_total"`
	OperationsDelayed   int64 `json:"operations_delayed"`
	OperationsFailed    int64 `json:"operations_failed"`
	OperationsSucceeded int64 `json:"operations_succeeded"`

	// Verification counters.
	VerificationsCreated      int64 `json:"verifications_created"`
	VerificationsAutoApproved int64 `json:"verifications_auto_approved"`

	// Persona activity.
	ActivePersonas        int   `json:"active_personas"`
	TransactionsGenerated int64 `json:"transactions_generated"`

	// User lifecycle.
	UsersCreated     int64 `json:"users_created"`
	UsersKYCVerified int64 `json:"users_kyc_verified"`
	UsersActivated   int64 `json:"users_activated"`

	// Timing metrics.
	AverageDelayMs float64   `json:"average_delay_ms"`
	StartedAt      time.Time `json:"started_at"`
	LastActivityAt time.Time `json:"last_activity_at"`

	// Current mode.
	CurrentMode string `json:"current_mode"`

	// Internal tracking.
	totalDelayMs int64
	delayCount   int64
}

// NewSimulationMetrics creates a new metrics tracker.
func NewSimulationMetrics() *SimulationMetrics {
	return &SimulationMetrics{
		StartedAt:   time.Now(),
		CurrentMode: "realistic",
	}
}

// RecordOperation records an operation execution.
func (m *SimulationMetrics) RecordOperation(delayed bool, failed bool, delayMs int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OperationsTotal++
	m.LastActivityAt = time.Now()

	if delayed && delayMs > 0 {
		m.OperationsDelayed++
		m.totalDelayMs += delayMs
		m.delayCount++
		m.AverageDelayMs = float64(m.totalDelayMs) / float64(m.delayCount)
	}

	if failed {
		m.OperationsFailed++
	} else {
		m.OperationsSucceeded++
	}
}

// RecordVerification records verification activity.
func (m *SimulationMetrics) RecordVerification(autoApproved bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.VerificationsCreated++
	if autoApproved {
		m.VerificationsAutoApproved++
	}
}

// RecordTransaction records a generated transaction.
func (m *SimulationMetrics) RecordTransaction() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TransactionsGenerated++
	m.LastActivityAt = time.Now()
}

// RecordUserCreated records a new user creation.
func (m *SimulationMetrics) RecordUserCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UsersCreated++
}

// RecordUserKYCVerified records a user KYC verification.
func (m *SimulationMetrics) RecordUserKYCVerified() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UsersKYCVerified++
}

// RecordUserActivated records a user activation.
func (m *SimulationMetrics) RecordUserActivated() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UsersActivated++
}

// SetActivePersonas updates the active persona count.
func (m *SimulationMetrics) SetActivePersonas(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ActivePersonas = count
}

// SetMode updates the current simulation mode.
func (m *SimulationMetrics) SetMode(mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CurrentMode = mode
}

// MetricsView is a JSON-safe representation of metrics (no mutex).
type MetricsView struct {
	OperationsTotal           int64     `json:"operations_total"`
	OperationsDelayed         int64     `json:"operations_delayed"`
	OperationsFailed          int64     `json:"operations_failed"`
	OperationsSucceeded       int64     `json:"operations_succeeded"`
	VerificationsCreated      int64     `json:"verifications_created"`
	VerificationsAutoApproved int64     `json:"verifications_auto_approved"`
	ActivePersonas            int       `json:"active_personas"`
	TransactionsGenerated     int64     `json:"transactions_generated"`
	UsersCreated              int64     `json:"users_created"`
	UsersKYCVerified          int64     `json:"users_kyc_verified"`
	UsersActivated            int64     `json:"users_activated"`
	AverageDelayMs            float64   `json:"average_delay_ms"`
	StartedAt                 time.Time `json:"started_at"`
	LastActivityAt            time.Time `json:"last_activity_at"`
	CurrentMode               string    `json:"current_mode"`
}

// GetSnapshot returns a JSON-safe view of current metrics.
func (m *SimulationMetrics) GetSnapshot() MetricsView {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return MetricsView{
		OperationsTotal:           m.OperationsTotal,
		OperationsDelayed:         m.OperationsDelayed,
		OperationsFailed:          m.OperationsFailed,
		OperationsSucceeded:       m.OperationsSucceeded,
		VerificationsCreated:      m.VerificationsCreated,
		VerificationsAutoApproved: m.VerificationsAutoApproved,
		ActivePersonas:            m.ActivePersonas,
		TransactionsGenerated:     m.TransactionsGenerated,
		UsersCreated:              m.UsersCreated,
		UsersKYCVerified:          m.UsersKYCVerified,
		UsersActivated:            m.UsersActivated,
		AverageDelayMs:            m.AverageDelayMs,
		StartedAt:                 m.StartedAt,
		LastActivityAt:            m.LastActivityAt,
		CurrentMode:               m.CurrentMode,
	}
}

// Reset resets all counters but preserves start time.
func (m *SimulationMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.OperationsTotal = 0
	m.OperationsDelayed = 0
	m.OperationsFailed = 0
	m.OperationsSucceeded = 0
	m.VerificationsCreated = 0
	m.VerificationsAutoApproved = 0
	m.TransactionsGenerated = 0
	m.UsersCreated = 0
	m.UsersKYCVerified = 0
	m.UsersActivated = 0
	m.AverageDelayMs = 0
	m.totalDelayMs = 0
	m.delayCount = 0
	m.StartedAt = time.Now()
}

// GetUptime returns how long the simulation has been running.
func (m *SimulationMetrics) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return time.Since(m.StartedAt)
}

// GetSuccessRate returns the operation success rate.
func (m *SimulationMetrics) GetSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.OperationsTotal == 0 {
		return 100.0
	}
	return float64(m.OperationsSucceeded) / float64(m.OperationsTotal) * 100
}
