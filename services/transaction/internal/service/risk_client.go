package service

import (
	"context"

	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/errors"
)

// RiskClient handles communication with the Risk service.
type RiskClient struct {
	*clients.BaseClient
}

// NewRiskClient creates a new Risk service client.
func NewRiskClient(baseURL string) *RiskClient {
	return &RiskClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.ShortTimeout),
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
func (c *RiskClient) EvaluateTransaction(ctx context.Context, req *RiskEvaluationRequest) (*RiskEvaluationResult, *errors.Error) {
	var result RiskEvaluationResult
	if err := c.Post(ctx, "/api/v1/risk/evaluate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
