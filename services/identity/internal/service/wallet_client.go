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

// WalletClient handles communication with the Wallet service.
type WalletClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewWalletClient creates a new wallet service client.
func NewWalletClient(baseURL string) *WalletClient {
	return &WalletClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateWalletRequest represents the request to create a wallet.
type CreateWalletRequest struct {
	UserID   string `json:"user_id"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
}

// WalletResponse represents a wallet from the wallet service.
type WalletResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Type             string `json:"type"`
	Currency         string `json:"currency"`
	Balance          int64  `json:"balance"`
	AvailableBalance int64  `json:"available_balance"`
	Status           string `json:"status"`
	LedgerAccountID  string `json:"ledger_account_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// APIResponse represents a standard API response from the wallet service.
type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *APIError       `json:"error,omitempty"`
}

// APIError represents an error from the wallet service.
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CreateDefaultWallet creates a default INR wallet for a user.
func (c *WalletClient) CreateDefaultWallet(ctx context.Context, userID string) (*WalletResponse, error) {
	req := CreateWalletRequest{
		UserID:   userID,
		Type:     "default",
		Currency: "INR",
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/wallets", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("wallet service error: %s - %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("wallet service returned error without details")
	}

	var wallet WalletResponse
	if err := json.Unmarshal(apiResp.Data, &wallet); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %w", err)
	}

	return &wallet, nil
}

// ListUserWalletsRequest represents the request to list user wallets.
type ListUserWalletsRequest struct {
	UserID string `json:"user_id"`
	Status string `json:"status,omitempty"`
}

// ListUserWallets retrieves all wallets for a user.
func (c *WalletClient) ListUserWallets(ctx context.Context, userID string) ([]WalletResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/wallets?user_id=%s", c.baseURL, userID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return nil, fmt.Errorf("wallet service error: %s - %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return nil, fmt.Errorf("wallet service returned error without details")
	}

	var wallets []WalletResponse
	if err := json.Unmarshal(apiResp.Data, &wallets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallets data: %w", err)
	}

	return wallets, nil
}

// ActivateWallet activates a wallet by ID (called after KYC approval).
func (c *WalletClient) ActivateWallet(ctx context.Context, walletID string) error {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/wallets/%s/activate", c.baseURL, walletID), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error != nil {
			return fmt.Errorf("wallet service error: %s - %s", apiResp.Error.Code, apiResp.Error.Message)
		}
		return fmt.Errorf("wallet service returned error without details")
	}

	return nil
}
