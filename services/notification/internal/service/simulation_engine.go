package service

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/vnykmshr/nivo/services/notification/internal/models"
	"github.com/vnykmshr/nivo/shared/errors"
)

// SimulationConfig holds configuration for the notification simulation engine.
type SimulationConfig struct {
	DeliveryDelayMs      int     // Delay in milliseconds before marking as sent
	FinalDelayMs         int     // Delay in milliseconds before final status (delivered/failed)
	FailureRatePercent   float64 // Percentage of notifications that should fail (0-100)
	MaxRetryAttempts     int     // Maximum retry attempts for failed notifications
	RetryDelayMs         int     // Base delay between retries (exponential backoff)
	CriticalPriorityOnly bool    // Process only critical priority (for testing)
}

// DefaultSimulationConfig returns sensible defaults for simulation.
func DefaultSimulationConfig() SimulationConfig {
	return SimulationConfig{
		DeliveryDelayMs:    1000, // 1 second to mark as sent
		FinalDelayMs:       2000, // 2 seconds to mark as delivered/failed
		FailureRatePercent: 10.0, // 10% failure rate
		MaxRetryAttempts:   3,    // Retry up to 3 times
		RetryDelayMs:       2000, // 2 seconds base delay
	}
}

// SimulationEngine simulates notification delivery with realistic behavior.
type SimulationEngine struct {
	config SimulationConfig
	repo   NotificationRepositoryInterface
	rand   *rand.Rand
}

// NewSimulationEngine creates a new simulation engine.
func NewSimulationEngine(config SimulationConfig, repo NotificationRepositoryInterface) *SimulationEngine {
	return &SimulationEngine{
		config: config,
		repo:   repo,
		//nolint:gosec // Using math/rand for simulation randomness, not cryptographic security
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ProcessNotification simulates processing a single notification.
// This is the core simulation logic that mimics real-world delivery.
func (e *SimulationEngine) ProcessNotification(ctx context.Context, notif *models.Notification) *errors.Error {
	log.Printf("[simulation] Processing notification %s (type=%s, channel=%s, priority=%s)",
		notif.ID, notif.Type, notif.Channel, notif.Priority)

	// Step 1: Simulate network delay before sending
	time.Sleep(time.Duration(e.config.DeliveryDelayMs) * time.Millisecond)

	// Update status to 'sent'
	if err := e.repo.UpdateStatus(ctx, notif.ID, models.StatusSent, nil); err != nil {
		log.Printf("[simulation] Failed to update notification %s to sent: %v", notif.ID, err)
		return err
	}

	log.Printf("[simulation] Notification %s marked as sent", notif.ID)

	// Step 2: Simulate processing delay before final status
	time.Sleep(time.Duration(e.config.FinalDelayMs) * time.Millisecond)

	// Step 3: Determine if delivery should fail (random simulation)
	shouldFail := e.shouldSimulateFailure()

	if shouldFail {
		// Simulate failure
		failureReason := e.generateFailureReason(notif.Channel)
		if err := e.repo.UpdateStatus(ctx, notif.ID, models.StatusFailed, &failureReason); err != nil {
			log.Printf("[simulation] Failed to update notification %s to failed: %v", notif.ID, err)
			return err
		}

		log.Printf("[simulation] Notification %s marked as failed: %s", notif.ID, failureReason)

		// Check if retry is needed
		if notif.RetryCount < e.config.MaxRetryAttempts {
			log.Printf("[simulation] Notification %s will be retried (attempt %d/%d)",
				notif.ID, notif.RetryCount+1, e.config.MaxRetryAttempts)
		} else {
			log.Printf("[simulation] Notification %s exceeded max retries", notif.ID)
		}

		return nil
	}

	// Simulate successful delivery
	if err := e.repo.UpdateStatus(ctx, notif.ID, models.StatusDelivered, nil); err != nil {
		log.Printf("[simulation] Failed to update notification %s to delivered: %v", notif.ID, err)
		return err
	}

	log.Printf("[simulation] Notification %s marked as delivered successfully", notif.ID)
	return nil
}

// shouldSimulateFailure determines if the current notification should fail.
func (e *SimulationEngine) shouldSimulateFailure() bool {
	if e.config.FailureRatePercent <= 0 {
		return false
	}
	if e.config.FailureRatePercent >= 100 {
		return true
	}

	randomValue := e.rand.Float64() * 100 // 0-100
	return randomValue < e.config.FailureRatePercent
}

// generateFailureReason generates a realistic failure reason based on channel.
func (e *SimulationEngine) generateFailureReason(channel models.NotificationChannel) string {
	reasons := map[models.NotificationChannel][]string{
		models.ChannelSMS: {
			"Invalid phone number",
			"Phone number not reachable",
			"Network timeout",
			"Provider error",
			"Number blocked by recipient",
		},
		models.ChannelEmail: {
			"Invalid email address",
			"Mailbox full",
			"Email rejected by recipient server",
			"SMTP timeout",
			"Email marked as spam",
		},
		models.ChannelPush: {
			"Device token invalid",
			"Device not registered",
			"Push service unavailable",
			"Token expired",
			"App not installed",
		},
		models.ChannelInApp: {
			"User session expired",
			"Storage quota exceeded",
			"Database write error",
		},
	}

	channelReasons, ok := reasons[channel]
	if !ok {
		return "Unknown delivery error"
	}

	// Pick random reason
	return channelReasons[e.rand.Intn(len(channelReasons))]
}

// CalculateRetryDelay calculates the delay before next retry using exponential backoff.
func (e *SimulationEngine) CalculateRetryDelay(retryCount int) time.Duration {
	// Exponential backoff: base * 2^retryCount
	baseDelay := time.Duration(e.config.RetryDelayMs) * time.Millisecond
	multiplier := 1 << retryCount // 2^retryCount
	return baseDelay * time.Duration(multiplier)
}

// GetProcessingDelay returns the total processing delay for a notification.
func (e *SimulationEngine) GetProcessingDelay() time.Duration {
	return time.Duration(e.config.DeliveryDelayMs+e.config.FinalDelayMs) * time.Millisecond
}

// NotificationRepositoryInterface defines the repository methods needed by the simulation engine.
type NotificationRepositoryInterface interface {
	UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, failureReason *string) *errors.Error
	IncrementRetryCount(ctx context.Context, id string) *errors.Error
	GetQueuedNotifications(ctx context.Context, limit int) ([]*models.Notification, *errors.Error)
}
