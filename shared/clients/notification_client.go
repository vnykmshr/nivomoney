package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	NotificationTypeWelcome           NotificationType = "welcome"
	NotificationTypeKYCStatus         NotificationType = "kyc_status"
	NotificationTypeWalletCreated     NotificationType = "wallet_created"
	NotificationTypeWalletActivated   NotificationType = "wallet_activated"
	NotificationTypeTransactionAlert  NotificationType = "transaction_alert"
	NotificationTypeSecurityAlert     NotificationType = "security_alert"
	NotificationTypeOTP               NotificationType = "otp"
	NotificationTypeMarketing         NotificationType = "marketing"
	NotificationTypeSystemAlert       NotificationType = "system_alert"
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
	UserID        *string                    `json:"user_id,omitempty"`
	Recipient     string                     `json:"recipient"`
	Channel       NotificationChannel        `json:"channel"`
	Type          NotificationType           `json:"type"`
	Priority      NotificationPriority       `json:"priority"`
	TemplateID    string                     `json:"template_id"`
	Variables     map[string]interface{}     `json:"variables,omitempty"`
	CorrelationID *string                    `json:"correlation_id,omitempty"`
	SourceService string                     `json:"source_service"`
	Metadata      map[string]interface{}     `json:"metadata,omitempty"`
}

// SendNotificationResponse represents the response from sending a notification.
type SendNotificationResponse struct {
	NotificationID string    `json:"notification_id"`
	Status         string    `json:"status"`
	QueuedAt       time.Time `json:"queued_at"`
}

// NotificationClient handles communication with the notification service.
type NotificationClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewNotificationClient creates a new notification client.
func NewNotificationClient(baseURL string) *NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		timeout: 5 * time.Second,
	}
}

// SendNotification sends a notification via the notification service.
func (c *NotificationClient) SendNotification(ctx context.Context, req *SendNotificationRequest) (*SendNotificationResponse, *errors.Error) {
	// Create request body
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Internal("failed to marshal notification request")
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/notifications/send", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, errors.Internal("failed to create notification request")
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to send notification: %v", err))
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Internal("failed to read notification response")
	}

	// Check status code
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Internal(fmt.Sprintf("notification service returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	// Parse response
	var apiResp struct {
		Success bool                      `json:"success"`
		Data    *SendNotificationResponse `json:"data"`
		Error   *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, errors.Internal("failed to parse notification response")
	}

	if !apiResp.Success || apiResp.Data == nil {
		errMsg := "unknown error"
		if apiResp.Error != nil {
			errMsg = apiResp.Error.Message
		}
		return nil, errors.Internal(fmt.Sprintf("notification service error: %s", errMsg))
	}

	return apiResp.Data, nil
}

// SendNotificationAsync sends a notification asynchronously (fire and forget).
// It logs errors but does not block or return them.
func (c *NotificationClient) SendNotificationAsync(req *SendNotificationRequest, serviceName string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
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
			Variables:     map[string]interface{}{"full_name": fullName},
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
			Variables:     map[string]interface{}{"full_name": fullName},
			CorrelationID: &userID,
			SourceService: sourceSvc,
		}
		c.SendNotificationAsync(smsReq, sourceSvc)
	}

	return nil
}

// SendKYCStatusNotification sends a KYC status update notification.
func (c *NotificationClient) SendKYCStatusNotification(ctx context.Context, userID, email, phone, fullName, status, reason, templateID, sourceSvc string) *errors.Error {
	variables := map[string]interface{}{
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
	variables := map[string]interface{}{
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
	variables := map[string]interface{}{
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
