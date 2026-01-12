package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/vnykmshr/nivo/shared/errors"
)

// NotificationChannel represents the delivery channel for a notification.
type NotificationChannel string

const (
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelPush  NotificationChannel = "push"
	NotificationChannelInApp NotificationChannel = "in_app"
)

// NotificationType represents the category of notification.
type NotificationType string

const (
	NotificationTypeWelcome          NotificationType = "welcome"
	NotificationTypeKYCStatus        NotificationType = "kyc_status"
	NotificationTypeWalletCreated    NotificationType = "wallet_created"
	NotificationTypeWalletActivated  NotificationType = "wallet_activated"
	NotificationTypeTransactionAlert NotificationType = "transaction_alert"
	NotificationTypeSecurityAlert    NotificationType = "security_alert"
	NotificationTypeOTP              NotificationType = "otp"
	NotificationTypeMarketing        NotificationType = "marketing"
	NotificationTypeSystemAlert      NotificationType = "system_alert"
)

// NotificationPriority represents the urgency of a notification.
type NotificationPriority string

const (
	NotificationPriorityCritical NotificationPriority = "critical"
	NotificationPriorityHigh     NotificationPriority = "high"
	NotificationPriorityNormal   NotificationPriority = "normal"
	NotificationPriorityLow      NotificationPriority = "low"
)

// SendNotificationRequest represents a request to send a notification.
type SendNotificationRequest struct {
	UserID        *string              `json:"user_id,omitempty"`
	Recipient     string               `json:"recipient"`
	Channel       NotificationChannel  `json:"channel"`
	Type          NotificationType     `json:"type"`
	Priority      NotificationPriority `json:"priority"`
	TemplateID    string               `json:"template_id"`
	Variables     map[string]any       `json:"variables,omitempty"`
	CorrelationID *string              `json:"correlation_id,omitempty"`
	SourceService string               `json:"source_service"`
	Metadata      map[string]any       `json:"metadata,omitempty"`
}

// SendNotificationResponse represents the response from sending a notification.
type SendNotificationResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	QueuedAt       time.Time `json:"queued_at"`
}

// NotificationClient handles communication with the notification service.
type NotificationClient struct {
	*BaseClient
	asyncTimeout time.Duration
}

// NewNotificationClient creates a new notification client.
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		BaseClient:   NewBaseClient(baseURL, DefaultTimeout),
		asyncTimeout: ShortTimeout,
	}
}

// SendNotification sends a notification via the notification service.
func (c *NotificationClient) SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, *errors.Error) {
	var result SendNotificationResponse
	if err := c.Post(ctx, "/v1/notifications/send", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SendNotificationAsync sends a notification asynchronously (fire and forget).
// It logs errors but does not block or return them.
func (c *NotificationClient) SendNotificationAsync(req *SendNotificationRequest, serviceName string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.asyncTimeout)
		defer cancel()

		_, err := c.SendNotification(ctx, req)
		if err != nil {
			fmt.Printf("[%s] Failed to send notification: %v\n", serviceName, err)
		}
	}()
}

// Helper functions to create common notifications

// SendWelcomeNotification sends a welcome notification to a new user.
func (c *NotificationClient) SendWelcomeNotification(ctx context.Context, userID, email, phone, fullName, templateID, sourceSvc string) *errors.Error {
	// Send email
	if email != "" {
		emailReq := &SendNotificationRequest{
			UserID:        &userID,
			Recipient:     email,
			Channel:       NotificationChannelEmail,
			Type:          NotificationTypeWelcome,
			Priority:      NotificationPriorityNormal,
			TemplateID:    templateID,
			Variables:     map[string]any{"full_name": fullName},
			CorrelationID: &userID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(emailReq, sourceSvc)
	}

	// Send SMS if phone provided
	if phone != "" {
		smsReq := &SendNotificationRequest{
			UserID:        &userID,
			Recipient:     phone,
			Channel:       NotificationChannelSMS,
			Type:          NotificationTypeWelcome,
			Priority:      NotificationPriorityNormal,
			TemplateID:    templateID,
			Variables:     map[string]any{"full_name": fullName},
			CorrelationID: &userID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(smsReq, sourceSvc)
	}

	return nil
}

// SendKYCStatusNotification sends a KYC status update notification.
func (c *NotificationClient) SendKYCStatusNotification(ctx context.Context, userID, email, phone, fullName, status, reason, templateID, sourceSvc string) *errors.Error {
	variables := map[string]any{
		"full_name": fullName,
		"status":    status,
	}
	if reason != "" {
		variables["reason"] = reason
	}

	correlationID := fmt.Sprintf("kyc-%s", userID)

	// Send email
	if email != "" {
		emailReq := &SendNotificationRequest{
			UserID:        &userID,
			Recipient:     email,
			Channel:       NotificationChannelEmail,
			Type:          NotificationTypeKYCStatus,
			Priority:      NotificationPriorityHigh,
			TemplateID:    templateID,
			Variables:     variables,
			CorrelationID: &correlationID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(emailReq, sourceSvc)
	}

	// Send SMS if phone provided
	if phone != "" {
		smsReq := &SendNotificationRequest{
			UserID:        &userID,
			Recipient:     phone,
			Channel:       NotificationChannelSMS,
			Type:          NotificationTypeKYCStatus,
			Priority:      NotificationPriorityHigh,
			TemplateID:    templateID,
			Variables:     variables,
			CorrelationID: &correlationID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(smsReq, sourceSvc)
	}

	return nil
}

// SendWalletNotification sends a wallet-related notification.
func (c *NotificationClient) SendWalletNotification(ctx context.Context, userID, email, walletID, walletType, currency, templateID, sourceSvc string, notifType NotificationType) *errors.Error {
	variables := map[string]any{
		"wallet_id":   walletID,
		"wallet_type": walletType,
		"currency":    currency,
	}

	correlationID := fmt.Sprintf("wallet-%s", walletID)

	req := &SendNotificationRequest{
		UserID:        &userID,
		Recipient:     email,
		Channel:       NotificationChannelEmail,
		Type:          notifType,
		Priority:      NotificationPriorityNormal,
		TemplateID:    templateID,
		Variables:     variables,
		CorrelationID: &correlationID,
		SourceService: sourceSvc,
	}

	c.SendNotificationAsync(req, sourceSvc)
	return nil
}

// SendTransactionNotification sends a transaction-related notification.
func (c *NotificationClient) SendTransactionNotification(ctx context.Context, userID, email, phone, transactionID, transactionType, amount, currency, templateID, sourceSvc string, priority NotificationPriority) *errors.Error {
	variables := map[string]any{
		"transaction_id":   transactionID,
		"transaction_type": transactionType,
		"amount":           amount,
		"currency":         currency,
	}

	correlationID := fmt.Sprintf("txn-%s", transactionID)

	// Send email
	if email != "" {
		emailReq := &SendNotificationRequest{
			UserID:        &userID,
			Recipient:     email,
			Channel:       NotificationChannelEmail,
			Type:          NotificationTypeTransactionAlert,
			Priority:      priority,
			TemplateID:    templateID,
			Variables:     variables,
			CorrelationID: &correlationID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(emailReq, sourceSvc)
	}

	// Send SMS if phone provided
	if phone != "" {
		smsReq := &SendNotificationRequest{
			UserID:        &userID,
			Recipient:     phone,
			Channel:       NotificationChannelSMS,
			Type:          NotificationTypeTransactionAlert,
			Priority:      priority,
			TemplateID:    templateID,
			Variables:     variables,
			CorrelationID: &correlationID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(smsReq, sourceSvc)
	}

	return nil
}
