package models

import (
	"github.com/vnykmshr/nivo/shared/models"
)

// UPIDepositStatus represents the status of a UPI deposit.
type UPIDepositStatus string

const (
	UPIDepositStatusPending   UPIDepositStatus = "pending"
	UPIDepositStatusCompleted UPIDepositStatus = "completed"
	UPIDepositStatusFailed    UPIDepositStatus = "failed"
	UPIDepositStatusExpired   UPIDepositStatus = "expired"
)

// UPIDeposit represents a UPI deposit request.
type UPIDeposit struct {
	ID           string            `json:"id" db:"id"`
	WalletID     string            `json:"wallet_id" db:"wallet_id"`
	UserID       string            `json:"user_id" db:"user_id"`
	Amount       int64             `json:"amount" db:"amount"` // Amount in paise
	UPIReference string            `json:"upi_reference" db:"upi_reference"`
	Status       UPIDepositStatus  `json:"status" db:"status"`
	CreatedAt    models.Timestamp  `json:"created_at" db:"created_at"`
	ExpiresAt    models.Timestamp  `json:"expires_at" db:"expires_at"`
	CompletedAt  *models.Timestamp `json:"completed_at,omitempty" db:"completed_at"`
	FailedReason *string           `json:"failed_reason,omitempty" db:"failed_reason"`
}

// IsPending returns true if the deposit is pending.
func (d *UPIDeposit) IsPending() bool {
	return d.Status == UPIDepositStatusPending
}

// IsCompleted returns true if the deposit is completed.
func (d *UPIDeposit) IsCompleted() bool {
	return d.Status == UPIDepositStatusCompleted
}

// InitiateUPIDepositRequest represents a request to initiate a UPI deposit.
type InitiateUPIDepositRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0,lte=10000000"` // Max 1 lakh rupees (in paise)
}

// UPIDepositResponse represents the response for a UPI deposit initiation.
type UPIDepositResponse struct {
	Deposit   *UPIDeposit `json:"deposit"`
	UPIString string      `json:"upi_string"`
	QRCodeURL string      `json:"qr_code_url,omitempty"`
	ExpiresIn string      `json:"expires_in"`
	Message   string      `json:"message"`
}

// WalletUPIDetails represents UPI details for a wallet.
type WalletUPIDetails struct {
	WalletID  string `json:"wallet_id"`
	UPIVPA    string `json:"upi_vpa"`
	QRCodeURL string `json:"qr_code_url,omitempty"`
}
