package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RiskClient handles communication with the Risk service.
type RiskClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewRiskClient creates a new Risk service client.
func NewRiskClient(baseURL string) *RiskClient {
	return &RiskClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// RiskEvaluationRequest represents a risk evaluation request.
type RiskEvaluationRequest struct {
	TransactionID   string `json:"transaction_id"`
	UserID          string `json:"user_id"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	TransactionType string `json:"transaction_type"`
	FromWalletID    string `json:"from_wallet_id,omitempty"`
	ToWalletID      string `json:"to_wallet_id,omitempty"`
}

// RiskEvaluationResult represents the risk evaluation result.
type RiskEvaluationResult struct {
	Allowed        bool     `json:"allowed"`
	Action         string   `json:"action"` // allow, block, flag
	RiskScore      int      `json:"risk_score"`
	Reason         string   `json:"reason"`
	TriggeredRules []string `json:"triggered_rules"`
	EventID        string   `json:"event_id"`
}

// EvaluateTransaction evaluates a transaction for risk.
func (c *RiskClient) EvaluateTransaction(ctx context.Context, req *RiskEvaluationRequest) (*RiskEvaluationResult, error) {
	url := fmt.Sprintf("%s/api/v1/risk/evaluate", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Risk service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("risk service returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response envelope
	var envelope struct {
		Success bool                  `json:"success"`
		Data    *RiskEvaluationResult `json:"data"`
		Error   *string               `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, fmt.Errorf("risk evaluation failed: %s", errMsg)
	}

	return envelope.Data, nil
}
