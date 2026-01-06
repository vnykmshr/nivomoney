//nolint:gosec // G404: math/rand acceptable for simulation behavior randomization
package behavior

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/vnykmshr/nivo/services/simulation/internal/config"
	"github.com/vnykmshr/nivo/shared/errors"
)

// BehaviorInjector adds realistic delays and failures to operations.
type BehaviorInjector struct {
	config *config.SimulationConfig
	mu     sync.Mutex // Protects rng
	rng    *rand.Rand
}

// NewBehaviorInjector creates a new behavior injector.
func NewBehaviorInjector(cfg *config.SimulationConfig) *BehaviorInjector {
	return &BehaviorInjector{
		config: cfg,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// randIntn returns a random int in [0, n) with thread safety.
func (b *BehaviorInjector) randIntn(n int) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rng.Intn(n)
}

// randFloat64 returns a random float64 in [0.0, 1.0) with thread safety.
func (b *BehaviorInjector) randFloat64() float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.rng.Float64()
}

// ApplyDelay applies configured delay for an operation.
// Returns error if context is cancelled during delay.
func (b *BehaviorInjector) ApplyDelay(ctx context.Context, operation string) error {
	if b.config.IsDemo() {
		// Minimal delay in demo mode
		delay := time.Duration(b.config.GetMinDelayMs()) * time.Millisecond
		return b.sleep(ctx, delay)
	}

	// Get operation-specific delay range
	minMs := b.config.GetMinDelayMs()
	maxMs := b.getMaxDelay(operation)

	// Ensure valid range
	if maxMs <= minMs {
		maxMs = minMs + 100
	}

	// Random delay within range
	delayMs := minMs + b.randIntn(maxMs-minMs+1)
	delay := time.Duration(delayMs) * time.Millisecond

	return b.sleep(ctx, delay)
}

// getMaxDelay returns the maximum delay for an operation.
func (b *BehaviorInjector) getMaxDelay(operation string) int {
	switch operation {
	case "transfer":
		return b.config.GetTransferDelayMs()
	case "kyc_review":
		return b.config.GetKYCReviewDelayMs()
	case "verification":
		return b.config.GetVerificationDelayMs()
	default:
		return b.config.GetMaxDelayMs()
	}
}

// ShouldFail determines if an operation should fail based on configured rates.
func (b *BehaviorInjector) ShouldFail(operation string) bool {
	if !b.config.IsFailuresEnabled() || b.config.IsDemo() {
		return false
	}

	rate := b.getFailureRate(operation)
	return b.randFloat64() < rate
}

// getFailureRate returns the failure rate for an operation.
func (b *BehaviorInjector) getFailureRate(operation string) float64 {
	switch operation {
	case "transfer":
		return b.config.GetTransferFailureRate()
	case "kyc_review":
		return b.config.GetKYCRejectRate()
	default:
		return b.config.GetFailureRate()
	}
}

// GetFailureError returns an appropriate error for a failed operation.
func (b *BehaviorInjector) GetFailureError(operation string) *errors.Error {
	messages := b.getFailureMessages(operation)
	msg := messages[b.randIntn(len(messages))]
	return errors.Unavailable(msg)
}

// getFailureMessages returns realistic error messages for an operation.
func (b *BehaviorInjector) getFailureMessages(operation string) []string {
	switch operation {
	case "transfer":
		return []string{
			"insufficient funds in source account",
			"recipient account temporarily unavailable",
			"transaction limit exceeded for today",
			"network timeout during processing",
			"beneficiary verification pending",
		}
	case "kyc_review":
		return []string{
			"document quality insufficient",
			"information mismatch detected",
			"additional verification required",
			"document expired",
		}
	case "deposit":
		return []string{
			"payment gateway timeout",
			"bank server unavailable",
			"daily deposit limit reached",
		}
	case "withdrawal":
		return []string{
			"withdrawal processing delayed",
			"bank validation pending",
			"daily withdrawal limit reached",
		}
	default:
		return []string{
			"operation temporarily unavailable",
			"please try again later",
			"service experiencing high load",
		}
	}
}

// sleep performs a context-aware sleep.
func (b *BehaviorInjector) sleep(ctx context.Context, duration time.Duration) error {
	select {
	case <-time.After(duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetDelayDuration returns a random delay duration for metrics tracking.
func (b *BehaviorInjector) GetDelayDuration(operation string) time.Duration {
	if b.config.IsDemo() {
		return time.Duration(b.config.GetMinDelayMs()) * time.Millisecond
	}

	minMs := b.config.GetMinDelayMs()
	maxMs := b.getMaxDelay(operation)

	if maxMs <= minMs {
		maxMs = minMs + 100
	}

	delayMs := minMs + b.randIntn(maxMs-minMs+1)
	return time.Duration(delayMs) * time.Millisecond
}
