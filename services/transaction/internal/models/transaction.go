package models

import (
	"encoding/json"

	"github.com/vnykmshr/nivo/shared/models"
)

// TransactionType represents the type of transaction.
type TransactionType string

const (
	TransactionTypeTransfer   TransactionType = "transfer"   // Transfer between wallets
	TransactionTypeDeposit    TransactionType = "deposit"    // Deposit to wallet
	TransactionTypeWithdrawal TransactionType = "withdrawal" // Withdrawal from wallet
	TransactionTypeReversal   TransactionType = "reversal"   // Reversal of a transaction
	TransactionTypeFee        TransactionType = "fee"        // Fee charge
	TransactionTypeRefund     TransactionType = "refund"     // Refund
)

// TransactionStatus represents the status of a transaction.
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"    // Transaction initiated
	TransactionStatusProcessing TransactionStatus = "processing" // Transaction being processed
	TransactionStatusCompleted  TransactionStatus = "completed"  // Transaction completed successfully
	TransactionStatusFailed     TransactionStatus = "failed"     // Transaction failed
	TransactionStatusReversed   TransactionStatus = "reversed"   // Transaction reversed
	TransactionStatusCancelled  TransactionStatus = "cancelled"  // Transaction cancelled
)

// Transaction represents a financial transaction in the neobank.
type Transaction struct {
	ID                  string            `json:"id" db:"id"`
	Type                TransactionType   `json:"type" db:"type"`
	Status              TransactionStatus `json:"status" db:"status"`
	SourceWalletID      *string           `json:"source_wallet_id,omitempty" db:"source_wallet_id"`
	DestinationWalletID *string           `json:"destination_wallet_id,omitempty" db:"destination_wallet_id"`
	Amount              int64             `json:"amount" db:"amount"` // In smallest unit (paise)
	Currency            models.Currency   `json:"currency" db:"currency"`
	Description         string            `json:"description" db:"description"`
	Reference           *string           `json:"reference,omitempty" db:"reference"` // External reference
	LedgerEntryID       *string           `json:"ledger_entry_id,omitempty" db:"ledger_entry_id"`
	ParentTransactionID *string           `json:"parent_transaction_id,omitempty" db:"parent_transaction_id"` // For reversals/refunds
	Metadata            map[string]string `json:"metadata,omitempty" db:"metadata"`
	FailureReason       *string           `json:"failure_reason,omitempty" db:"failure_reason"`
	ProcessedAt         *models.Timestamp `json:"processed_at,omitempty" db:"processed_at"`
	CompletedAt         *models.Timestamp `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt           models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt           models.Timestamp  `json:"updated_at" db:"updated_at"`
}

// IsCompleted returns true if the transaction is completed.
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// IsFailed returns true if the transaction failed.
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}

// IsPending returns true if the transaction is pending.
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending || t.Status == TransactionStatusProcessing
}

// CreateTransferRequest represents a request to create a transfer transaction.
type CreateTransferRequest struct {
	SourceWalletID      string          `json:"source_wallet_id" validate:"required,uuid"`
	DestinationWalletID string          `json:"destination_wallet_id" validate:"required,uuid"`
	Amount              int64           `json:"amount" validate:"required,gt=0"`
	Currency            models.Currency `json:"currency" validate:"required,len=3"`
	Description         string          `json:"description" validate:"required,min=3,max=500"`
	Reference           string          `json:"reference,omitempty" validate:"omitempty,max=100"`
	MetadataRaw         json.RawMessage `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *CreateTransferRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// CreateDepositRequest represents a request to create a deposit transaction.
type CreateDepositRequest struct {
	WalletID    string          `json:"wallet_id" validate:"required,uuid"`
	Amount      int64           `json:"amount" validate:"required,gt=0"`
	Currency    models.Currency `json:"currency" validate:"required,len=3"`
	Description string          `json:"description" validate:"required,min=3,max=500"`
	Reference   string          `json:"reference,omitempty" validate:"omitempty,max=100"`
	MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *CreateDepositRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// CreateWithdrawalRequest represents a request to create a withdrawal transaction.
type CreateWithdrawalRequest struct {
	WalletID    string          `json:"wallet_id" validate:"required,uuid"`
	Amount      int64           `json:"amount" validate:"required,gt=0"`
	Currency    models.Currency `json:"currency" validate:"required,len=3"`
	Description string          `json:"description" validate:"required,min=3,max=500"`
	Reference   string          `json:"reference,omitempty" validate:"omitempty,max=100"`
	MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

// GetMetadata parses and returns the metadata map.
func (r *CreateWithdrawalRequest) GetMetadata() (map[string]string, error) {
	if len(r.MetadataRaw) == 0 {
		return make(map[string]string), nil
	}

	var metadata map[string]string
	if err := json.Unmarshal(r.MetadataRaw, &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// ReverseTransactionRequest represents a request to reverse a transaction.
type ReverseTransactionRequest struct {
	Reason string `json:"reason" validate:"required,min=10,max=500"`
}

// TransactionFilter represents filters for listing transactions.
type TransactionFilter struct {
	WalletID  *string
	Status    *TransactionStatus
	Type      *TransactionType
	StartDate *models.Timestamp
	EndDate   *models.Timestamp
	Limit     int
	Offset    int
}
