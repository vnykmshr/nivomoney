//nolint:gosec // G404: math/rand acceptable for simulation timing
package behavior

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/vnykmshr/nivo/services/simulation/internal/config"
	"github.com/vnykmshr/nivo/services/simulation/internal/metrics"
)

// PendingVerification represents a pending verification from the database.
type PendingVerification struct {
	ID            string
	UserID        string
	OperationType string
	OTPCode       string
	ExpiresAt     time.Time
}

// AutoVerifier handles automatic verification for simulated users.
// It monitors pending verifications and auto-approves them after a delay.
type AutoVerifier struct {
	db            *sql.DB
	config        *config.SimulationConfig
	metrics       *metrics.SimulationMetrics
	mu            sync.RWMutex    // Protects simulatedUser map
	simulatedUser map[string]bool // Track which user IDs are simulated
	rngMu         sync.Mutex      // Protects rng
	rng           *rand.Rand
}

// NewAutoVerifier creates a new auto-verifier.
func NewAutoVerifier(
	db *sql.DB,
	cfg *config.SimulationConfig,
	met *metrics.SimulationMetrics,
) *AutoVerifier {
	return &AutoVerifier{
		db:            db,
		config:        cfg,
		metrics:       met,
		simulatedUser: make(map[string]bool),
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RegisterSimulatedUser marks a user ID as simulated.
func (v *AutoVerifier) RegisterSimulatedUser(userID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.simulatedUser[userID] = true
}

// UnregisterSimulatedUser removes a user ID from the simulated list.
func (v *AutoVerifier) UnregisterSimulatedUser(userID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.simulatedUser, userID)
}

// IsSimulatedUser checks if a user ID is registered as simulated.
func (v *AutoVerifier) IsSimulatedUser(userID string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.simulatedUser[userID]
}

// randFloat64 returns a random float64 in [0.0, 1.0) with thread safety.
func (v *AutoVerifier) randFloat64() float64 {
	v.rngMu.Lock()
	defer v.rngMu.Unlock()
	return v.rng.Float64()
}

// GetPendingVerifications retrieves pending verifications for simulated users.
func (v *AutoVerifier) GetPendingVerifications(ctx context.Context) ([]PendingVerification, error) {
	// Build user ID list under read lock
	v.mu.RLock()
	if len(v.simulatedUser) == 0 {
		v.mu.RUnlock()
		return nil, nil
	}

	userIDs := make([]string, 0, len(v.simulatedUser))
	for userID := range v.simulatedUser {
		userIDs = append(userIDs, userID)
	}
	v.mu.RUnlock()

	// Query pending verifications for simulated users
	query := `
		SELECT id, user_id, operation_type, otp_code, expires_at
		FROM verification_requests
		WHERE user_id = ANY($1)
		  AND status = 'pending'
		  AND expires_at > NOW()
		ORDER BY created_at ASC
	`

	rows, err := v.db.QueryContext(ctx, query, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var verifications []PendingVerification
	for rows.Next() {
		var ver PendingVerification
		if err := rows.Scan(&ver.ID, &ver.UserID, &ver.OperationType, &ver.OTPCode, &ver.ExpiresAt); err != nil {
			log.Printf("[simulation] Failed to scan verification: %v", err)
			continue
		}
		verifications = append(verifications, ver)
	}

	return verifications, rows.Err()
}

// AutoApproveVerification directly approves a verification in the database.
// This bypasses the normal OTP flow for simulated users.
func (v *AutoVerifier) AutoApproveVerification(ctx context.Context, verificationID string) error {
	// Apply configured delay before auto-approval
	delay := time.Duration(v.config.GetAutoVerificationDelayMs()) * time.Millisecond
	if v.config.IsRealistic() {
		// Add some randomness in realistic mode (±50%)
		jitter := float64(delay) * (0.5 + v.randFloat64())
		delay = time.Duration(jitter)
	}

	// Sleep before approving (simulates user reading OTP from User-Admin)
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return ctx.Err()
	}

	// Update verification status to 'verified'
	query := `
		UPDATE verification_requests
		SET status = 'verified',
		    verified_at = NOW()
		WHERE id = $1
		  AND status = 'pending'
		RETURNING id
	`

	var updatedID string
	err := v.db.QueryRowContext(ctx, query, verificationID).Scan(&updatedID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Verification was already processed or expired
			return nil
		}
		return err
	}

	// Record in metrics
	v.metrics.RecordVerification(true)

	log.Printf("[simulation] ✅ Auto-approved verification: %s", verificationID)
	return nil
}

// ProcessPendingVerifications processes all pending verifications for simulated users.
func (v *AutoVerifier) ProcessPendingVerifications(ctx context.Context) error {
	if !v.config.IsAutoVerificationEnabled() {
		return nil
	}

	verifications, err := v.GetPendingVerifications(ctx)
	if err != nil {
		return err
	}

	for _, ver := range verifications {
		if err := v.AutoApproveVerification(ctx, ver.ID); err != nil {
			log.Printf("[simulation] Failed to auto-approve verification %s: %v", ver.ID, err)
			continue
		}
	}

	return nil
}

// RunAutoVerificationLoop starts a background loop to process verifications.
func (v *AutoVerifier) RunAutoVerificationLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	log.Printf("[simulation] 🤖 Auto-verification loop started")

	for {
		select {
		case <-ctx.Done():
			log.Printf("[simulation] Auto-verification loop stopped")
			return
		case <-ticker.C:
			if err := v.ProcessPendingVerifications(ctx); err != nil {
				log.Printf("[simulation] Error processing verifications: %v", err)
			}
		}
	}
}
