package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/errors"
)

// WalletClient handles communication with the Wallet service.
type WalletClient struct {
	*clients.BaseClient
}

// NewWalletClient creates a new Wallet service client.
func NewWalletClient(baseURL string) *WalletClient {
	return &WalletClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.DefaultTimeout),
	}
}

// NewWalletClientWithSecret creates a Wallet client with internal service authentication.
func NewWalletClientWithSecret(baseURL, internalSecret string) *WalletClient {
	return &WalletClient{
		BaseClient: clients.NewInternalClient(baseURL, clients.DefaultTimeout, internalSecret),
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

// DepositRequest represents an internal deposit request.
type DepositRequest struct {
	WalletID      string `json:"wallet_id"`
	Amount        int64  `json:"amount"`
	TransactionID string `json:"transaction_id"`
	Description   string `json:"description"`
}

// WalletInfo represents wallet details including ownership.
type WalletInfo struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	Status          string `json:"status"`
	LedgerAccountID string `json:"ledger_account_id"`
}

// GetBalance retrieves the balance of a wallet.
func (c *WalletClient) GetBalance(ctx context.Context, walletID string) (*WalletBalance, *errors.Error) {
	var result WalletBalance
	path := fmt.Sprintf("/api/v1/wallets/%s/balance", walletID)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CheckAndReserveLimit checks if transfer is within limits and reserves the amount.
// This is called as part of the transfer processing flow.
func (c *WalletClient) CheckAndReserveLimit(ctx context.Context, walletID string, amount int64) *errors.Error {
	req := CheckAndReserveLimitRequest{
		WalletID: walletID,
		Amount:   amount,
	}
	path := fmt.Sprintf("/internal/v1/wallets/%s/limits/reserve", walletID)
	return c.Post(ctx, path, req, nil)
}

// ExecuteTransfer executes a wallet-to-wallet transfer (internal endpoint).
// This updates wallet balances and creates holds as needed.
func (c *WalletClient) ExecuteTransfer(ctx context.Context, req *TransferRequest) *errors.Error {
	return c.Post(ctx, "/internal/v1/wallets/transfer", req, nil)
}

// CreditDeposit credits a deposit to a wallet (internal endpoint).
// This directly updates the wallet balance for successful deposits.
func (c *WalletClient) CreditDeposit(ctx context.Context, req *DepositRequest) *errors.Error {
	return c.Post(ctx, "/internal/v1/wallets/deposit", req, nil)
}

// GetWalletInfo retrieves wallet information including owner (internal endpoint).
func (c *WalletClient) GetWalletInfo(ctx context.Context, walletID string) (*WalletInfo, *errors.Error) {
	var result WalletInfo
	path := fmt.Sprintf("/internal/v1/wallets/%s/info", walletID)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifyWalletOwnership checks if a wallet belongs to the specified user.
func (c *WalletClient) VerifyWalletOwnership(ctx context.Context, walletID, userID string) *errors.Error {
	info, err := c.GetWalletInfo(ctx, walletID)
	if err != nil {
		return err
	}

	if info.UserID != userID {
		return errors.Forbidden("wallet does not belong to user")
	}

	return nil
}
