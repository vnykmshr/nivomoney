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

// NewWalletClient creates a new Wallet service client.
func NewWalletClient(baseURL string) *WalletClient {
	return &WalletClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// WalletBalance represents a wallet balance.
type WalletBalance struct {
	WalletID         string `json:"wallet_id"`
	Balance          int64  `json:"balance"`
	AvailableBalance int64  `json:"available_balance"`
	HeldAmount       int64  `json:"held_amount"`
}

// CheckAndReserveLimitRequest represents a limit check and reservation request.
type CheckAndReserveLimitRequest struct {
	WalletID string `json:"wallet_id"`
	Amount   int64  `json:"amount"`
}

// TransferRequest represents an internal wallet transfer request.
type TransferRequest struct {
	SourceWalletID      string `json:"source_wallet_id"`
	DestinationWalletID string `json:"destination_wallet_id"`
	Amount              int64  `json:"amount"`
	TransactionID       string `json:"transaction_id"`
	Description         string `json:"description"`
}

// GetBalance retrieves the balance of a wallet.
func (c *WalletClient) GetBalance(ctx context.Context, walletID string) (*WalletBalance, error) {
	url := fmt.Sprintf("%s/api/v1/wallets/%s/balance", c.baseURL, walletID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Wallet service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wallet service returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response envelope
	var envelope struct {
		Success bool           `json:"success"`
		Data    *WalletBalance `json:"data"`
		Error   *string        `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !envelope.Success || envelope.Data == nil {
		errMsg := "unknown error"
		if envelope.Error != nil {
			errMsg = *envelope.Error
		}
		return nil, fmt.Errorf("get balance failed: %s", errMsg)
	}

	return envelope.Data, nil
}

// CheckAndReserveLimit checks if transfer is within limits and reserves the amount.
// This is called as part of the transfer processing flow.
func (c *WalletClient) CheckAndReserveLimit(ctx context.Context, walletID string, amount int64) error {
	url := fmt.Sprintf("%s/internal/v1/wallets/%s/limits/reserve", c.baseURL, walletID)

	req := CheckAndReserveLimitRequest{
		WalletID: walletID,
		Amount:   amount,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call Wallet service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)

		// Parse error response to get user-friendly message
		var envelope struct {
			Success bool    `json:"success"`
			Error   *string `json:"error"`
		}
		if json.Unmarshal(respBody, &envelope) == nil && envelope.Error != nil {
			return fmt.Errorf("%s", *envelope.Error)
		}

		return fmt.Errorf("limit check failed: %s", string(respBody))
	}

	return nil
}

// ExecuteTransfer executes a wallet-to-wallet transfer (internal endpoint).
// This updates wallet balances and creates holds as needed.
func (c *WalletClient) ExecuteTransfer(ctx context.Context, req *TransferRequest) error {
	url := fmt.Sprintf("%s/internal/v1/wallets/transfer", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call Wallet service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("transfer execution failed: %s", string(respBody))
	}

	return nil
}

// DepositRequest represents an internal deposit request.
type DepositRequest struct {
	WalletID      string `json:"wallet_id"`
	Amount        int64  `json:"amount"`
	TransactionID string `json:"transaction_id"`
	Description   string `json:"description"`
}

// CreditDeposit credits a deposit to a wallet (internal endpoint).
// This directly updates the wallet balance for successful deposits.
func (c *WalletClient) CreditDeposit(ctx context.Context, req *DepositRequest) error {
	url := fmt.Sprintf("%s/internal/v1/wallets/deposit", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call Wallet service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deposit credit failed: %s", string(respBody))
	}

	return nil
}

// WalletInfo represents wallet details including ownership.
type WalletInfo struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

// GetWalletInfo retrieves wallet information including owner (internal endpoint).
func (c *WalletClient) GetWalletInfo(ctx context.Context, walletID string) (*WalletInfo, error) {
	url := fmt.Sprintf("%s/internal/v1/wallets/%s/info", c.baseURL, walletID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Wallet service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("wallet not found")
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get wallet info: %s", string(respBody))
	}

	var envelope struct {
		Success bool        `json:"success"`
		Data    *WalletInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !envelope.Success || envelope.Data == nil {
		return nil, fmt.Errorf("failed to get wallet info")
	}

	return envelope.Data, nil
}

// VerifyWalletOwnership checks if a wallet belongs to the specified user.
func (c *WalletClient) VerifyWalletOwnership(ctx context.Context, walletID, userID string) error {
	info, err := c.GetWalletInfo(ctx, walletID)
	if err != nil {
		return err
	}

	if info.UserID != userID {
		return fmt.Errorf("wallet does not belong to user")
	}

	return nil
}
