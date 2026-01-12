package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/shared/clients"
)

// WalletClient handles communication with the Wallet service.
type WalletClient struct {
	*clients.BaseClient
}

// NewWalletClient creates a new wallet service client.
func NewWalletClient(baseURL string) *WalletClient {
	return &WalletClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.DefaultTimeout),
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

// CreateDefaultWallet creates a default INR wallet for a user.
func (c *WalletClient) CreateDefaultWallet(ctx context.Context, userID string) (*WalletResponse, error) {
	req := CreateWalletRequest{
		UserID:   userID,
		Type:     "default",
		Currency: "INR",
	}

	var result WalletResponse
	if err := c.Post(ctx, "/api/v1/wallets", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListUserWallets retrieves all wallets for a user.
func (c *WalletClient) ListUserWallets(ctx context.Context, userID string) ([]WalletResponse, error) {
	var result []WalletResponse
	path := fmt.Sprintf("/api/v1/wallets?user_id=%s", userID)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ActivateWallet activates a wallet by ID (called after KYC approval).
func (c *WalletClient) ActivateWallet(ctx context.Context, walletID string) error {
	path := fmt.Sprintf("/api/v1/wallets/%s/activate", walletID)
	if err := c.Post(ctx, path, nil, nil); err != nil {
		return err
	}
	return nil
}
