package service

import (
	"context"
	"fmt"

	"github.com/vnykmshr/nivo/shared/clients"
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
	*clients.BaseClient
}

// NewLedgerClient creates a new ledger service client.
func NewLedgerClient(baseURL string) *LedgerClient {
	return &LedgerClient{
		BaseClient: clients.NewBaseClient(baseURL, clients.DefaultTimeout),
	}
}

// CreateAccount creates a new ledger account.
// Uses internal endpoint for service-to-service communication (no auth required).
func (c *LedgerClient) CreateAccount(ctx context.Context, req *CreateLedgerAccountRequest) (*LedgerAccount, *errors.Error) {
	var result LedgerAccount
	if err := c.Post(ctx, "/internal/v1/accounts", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAccount retrieves a ledger account by ID.
func (c *LedgerClient) GetAccount(ctx context.Context, accountID string) (*LedgerAccount, *errors.Error) {
	var result LedgerAccount
	path := fmt.Sprintf("/api/v1/accounts/%s", accountID)
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAccountByCode retrieves a ledger account by its code.
// Uses internal endpoint for service-to-service communication.
// Returns nil (not an error) if the account doesn't exist - this supports idempotent wallet creation.
func (c *LedgerClient) GetAccountByCode(ctx context.Context, code string) (*LedgerAccount, *errors.Error) {
	var result LedgerAccount
	path := fmt.Sprintf("/internal/v1/accounts/by-code/%s", code)
	if err := c.Get(ctx, path, &result); err != nil {
		// Return nil if not found (this is not an error for idempotency)
		if err.HTTPStatusCode() == 404 {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}
