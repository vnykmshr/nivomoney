package models

import (
	"encoding/json"

	"github.com/vnykmshr/nivo/shared/models"
)

// WalletType represents the type of wallet.
type WalletType string

const (
	WalletTypeDefault WalletType = "default" // Default wallet (one per user per currency)
)

// WalletStatus represents the status of a wallet.
type WalletStatus string

const (
	WalletStatusActive   WalletStatus = "active"   // Wallet is active and can be used
	WalletStatusFrozen   WalletStatus = "frozen"   // Wallet is frozen (compliance/security)
	WalletStatusClosed   WalletStatus = "closed"   // Wallet is permanently closed
	WalletStatusInactive WalletStatus = "inactive" // Wallet is inactive (KYC pending, etc.)
)

// Wallet represents a user's wallet in the neobank.
type Wallet struct {
	ID               string            `json:"id" db:"id"`
	UserID           string            `json:"user_id" db:"user_id"`                     // Owner of the wallet
	Type             WalletType        `json:"type" db:"type"`                           // Wallet type
	Currency         models.Currency   `json:"currency" db:"currency"`                   // Wallet currency
	Balance          int64             `json:"balance" db:"balance"`                     // Current balance in smallest unit (paise)
	AvailableBalance int64             `json:"available_balance" db:"available_balance"` // Balance minus holds/freezes
	Status           WalletStatus      `json:"status" db:"status"`
	LedgerAccountID  string            `json:"ledger_account_id" db:"ledger_account_id"` // Link to Ledger Service account
	Metadata         map[string]string `json:"metadata,omitempty" db:"metadata"`         // JSONB metadata
	CreatedAt        models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt        models.Timestamp  `json:"updated_at" db:"updated_at"`
	ClosedAt         *models.Timestamp `json:"closed_at,omitempty" db:"closed_at"`
	ClosedReason     *string           `json:"closed_reason,omitempty" db:"closed_reason"`
}

// IsActive returns true if the wallet is active.
func (w *Wallet) IsActive() bool {
	return w.Status == WalletStatusActive
}

// CanTransact returns true if the wallet can be used for transactions.
func (w *Wallet) CanTransact() bool {
	return w.Status == WalletStatusActive && w.AvailableBalance > 0
}

// CreateWalletRequest represents a request to create a new wallet.
type CreateWalletRequest struct {
	UserID          string          `json:"user_id" validate:"required,uuid"`
	Type            WalletType      `json:"type" validate:"required"`
	Currency        models.Currency `json:"currency" validate:"required,len:3"`
	LedgerAccountID string          `json:"ledger_account_id,omitempty" validate:"omitempty,uuid"` // Optional - auto-created if not provided
	MetadataRaw     json.RawMessage `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *CreateWalletRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// UpdateWalletStatusRequest represents a request to update wallet status.
type UpdateWalletStatusRequest struct {
	Status WalletStatus `json:"status" validate:"required"`
	Reason string       `json:"reason,omitempty" validate:"omitempty,min:10,max:500"`
}

// FreezeWalletRequest represents a request to freeze a wallet.
type FreezeWalletRequest struct {
	Reason string `json:"reason" validate:"required,min:10,max:500"`
}

// CloseWalletRequest represents a request to close a wallet.
type CloseWalletRequest struct {
	Reason string `json:"reason" validate:"required,min:10,max:500"`
}

// WalletBalance represents a wallet's balance information.
type WalletBalance struct {
	WalletID         string `json:"wallet_id"`
	Balance          int64  `json:"balance"`
	AvailableBalance int64  `json:"available_balance"`
	HeldAmount       int64  `json:"held_amount"` // Balance - AvailableBalance
}

// WalletLimits represents transfer limits for a wallet.
type WalletLimits struct {
	ID             string           `json:"id" db:"id"`
	WalletID       string           `json:"wallet_id" db:"wallet_id"`
	DailyLimit     int64            `json:"daily_limit" db:"daily_limit"`           // In smallest unit (paise)
	DailySpent     int64            `json:"daily_spent" db:"daily_spent"`           // Amount spent today
	DailyResetAt   models.Timestamp `json:"daily_reset_at" db:"daily_reset_at"`     // When daily limit resets
	MonthlyLimit   int64            `json:"monthly_limit" db:"monthly_limit"`       // In smallest unit (paise)
	MonthlySpent   int64            `json:"monthly_spent" db:"monthly_spent"`       // Amount spent this month
	MonthlyResetAt models.Timestamp `json:"monthly_reset_at" db:"monthly_reset_at"` // When monthly limit resets
	CreatedAt      models.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      models.Timestamp `json:"updated_at" db:"updated_at"`
}

// DailyRemaining returns the remaining daily transfer limit.
func (wl *WalletLimits) DailyRemaining() int64 {
	remaining := wl.DailyLimit - wl.DailySpent
	if remaining < 0 {
		return 0
	}
	return remaining
}

// MonthlyRemaining returns the remaining monthly transfer limit.
func (wl *WalletLimits) MonthlyRemaining() int64 {
	remaining := wl.MonthlyLimit - wl.MonthlySpent
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CanTransfer checks if a transfer amount is within limits.
func (wl *WalletLimits) CanTransfer(amount int64) bool {
	return amount <= wl.DailyRemaining() && amount <= wl.MonthlyRemaining()
}

// UpdateLimitsRequest represents a request to update wallet transfer limits.
// Note: Authentication is handled via JWT - no additional password required.
type UpdateLimitsRequest struct {
	DailyLimit   int64 `json:"daily_limit" validate:"required,gt=0"`
	MonthlyLimit int64 `json:"monthly_limit" validate:"required,gt=0"`
}

// ProcessTransferRequest represents an internal request to process a wallet transfer.
// This is called by the transaction service to execute approved transfers.
type ProcessTransferRequest struct {
	SourceWalletID      string `json:"source_wallet_id" validate:"required,uuid"`
	DestinationWalletID string `json:"destination_wallet_id" validate:"required,uuid"`
	Amount              int64  `json:"amount" validate:"required,gt=0"`
	TransactionID       string `json:"transaction_id" validate:"required,uuid"`
}

// ProcessDepositRequest represents an internal request to process a deposit.
// This is called by the transaction service to credit deposits to wallets.
type ProcessDepositRequest struct {
	WalletID      string `json:"wallet_id" validate:"required,uuid"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	TransactionID string `json:"transaction_id" validate:"required,uuid"`
	Description   string `json:"description,omitempty"`
}
