package service

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

// LedgerAccount represents a ledger account from the ledger service.
type LedgerAccount struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance"`
	Status   string `json:"status"`
}

// CreateLedgerAccountRequest represents the request to create a ledger account.
type CreateLedgerAccountRequest struct {
	Code     string            `json:"code"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Currency string            `json:"currency"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// LedgerClient handles communication with the ledger service.
type LedgerClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewLedgerClient creates a new ledger service client.
func NewLedgerClient(baseURL string) *LedgerClient {
	return &LedgerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateAccount creates a new ledger account.
// Uses internal endpoint for service-to-service communication (no auth required).
func (c *LedgerClient) CreateAccount(ctx context.Context, req *CreateLedgerAccountRequest) (*LedgerAccount, *errors.Error) {
	url := fmt.Sprintf("%s/internal/v1/accounts", c.baseURL)

	// Marshal request to JSON
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to marshal request: %v", err))
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to create HTTP request: %v", err))
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to call ledger service: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to read response body: %v", err))
	}

	// Check for error responses
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.Internal(fmt.Sprintf("ledger service returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	// Parse response envelope
	var envelope struct {
		Success bool           `json:"success"`
		Data    *LedgerAccount `json:"data"`
		Error   *string        `json:"error"`
	}

	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to parse ledger response: %v", err))
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, errors.Internal(fmt.Sprintf("ledger request failed: %s", errMsg))
	}

	return envelope.Data, nil
}

// GetAccount retrieves a ledger account by ID.
func (c *LedgerClient) GetAccount(ctx context.Context, accountID string) (*LedgerAccount, *errors.Error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s", c.baseURL, accountID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to create HTTP request: %v", err))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to call ledger service: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to read response body: %v", err))
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NotFound("ledger account not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Internal(fmt.Sprintf("ledger service returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	// Parse response envelope
	var envelope struct {
		Success bool           `json:"success"`
		Data    *LedgerAccount `json:"data"`
		Error   *string        `json:"error"`
	}

	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to parse ledger response: %v", err))
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, errors.Internal(fmt.Sprintf("ledger request failed: %s", errMsg))
	}

	return envelope.Data, nil
}

// GetAccountByCode retrieves a ledger account by its code.
// Uses internal endpoint for service-to-service communication.
// Returns nil (not an error) if the account doesn't exist - this supports idempotent wallet creation.
func (c *LedgerClient) GetAccountByCode(ctx context.Context, code string) (*LedgerAccount, *errors.Error) {
	url := fmt.Sprintf("%s/internal/v1/accounts/by-code/%s", c.baseURL, code)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to create HTTP request: %v", err))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to call ledger service: %v", err))
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to read response body: %v", err))
	}

	// Return nil if not found (this is not an error for idempotency)
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Internal(fmt.Sprintf("ledger service returned status %d: %s", resp.StatusCode, string(respBody)))
	}

	// Parse response envelope
	var envelope struct {
		Success bool           `json:"success"`
		Data    *LedgerAccount `json:"data"`
		Error   *string        `json:"error"`
	}

	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to parse ledger response: %v", err))
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, errors.Internal(fmt.Sprintf("ledger request failed: %s", errMsg))
	}

	return envelope.Data, nil
}
