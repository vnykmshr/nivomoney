package models

import (
	"github.com/vnykmshr/nivo/shared/models"
)

// CardStatus represents the status of a virtual card.
type CardStatus string

const (
	CardStatusActive    CardStatus = "active"
	CardStatusFrozen    CardStatus = "frozen"
	CardStatusExpired   CardStatus = "expired"
	CardStatusCancelled CardStatus = "cancelled"
)

// CardType represents the type of card.
type CardType string

const (
	CardTypeVirtual  CardType = "virtual"
	CardTypePhysical CardType = "physical"
)

// VirtualCard represents a virtual debit/credit card linked to a wallet.
type VirtualCard struct {
	ID                  string            `json:"id" db:"id"`
	WalletID            string            `json:"wallet_id" db:"wallet_id"`
	UserID              string            `json:"user_id" db:"user_id"`
	CardNumber          string            `json:"card_number" db:"card_number"`
	CardHolderName      string            `json:"card_holder_name" db:"card_holder_name"`
	ExpiryMonth         int               `json:"expiry_month" db:"expiry_month"`
	ExpiryYear          int               `json:"expiry_year" db:"expiry_year"`
	CVV                 string            `json:"-" db:"cvv"` // Never expose in JSON
	CardType            CardType          `json:"card_type" db:"card_type"`
	Status              CardStatus        `json:"status" db:"status"`
	DailyLimit          int64             `json:"daily_limit" db:"daily_limit"`
	MonthlyLimit        int64             `json:"monthly_limit" db:"monthly_limit"`
	PerTransactionLimit int64             `json:"per_transaction_limit" db:"per_transaction_limit"`
	DailySpent          int64             `json:"daily_spent" db:"daily_spent"`
	MonthlySpent        int64             `json:"monthly_spent" db:"monthly_spent"`
	LastUsedAt          *models.Timestamp `json:"last_used_at,omitempty" db:"last_used_at"`
	FrozenAt            *models.Timestamp `json:"frozen_at,omitempty" db:"frozen_at"`
	FrozenReason        *string           `json:"frozen_reason,omitempty" db:"frozen_reason"`
	CancelledAt         *models.Timestamp `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CancelledReason     *string           `json:"cancelled_reason,omitempty" db:"cancelled_reason"`
	CreatedAt           models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt           models.Timestamp  `json:"updated_at" db:"updated_at"`
}

// IsActive returns true if the card is active.
func (c *VirtualCard) IsActive() bool {
	return c.Status == CardStatusActive
}

// IsFrozen returns true if the card is frozen.
func (c *VirtualCard) IsFrozen() bool {
	return c.Status == CardStatusFrozen
}

// MaskedCardNumber returns the card number with middle digits masked.
func (c *VirtualCard) MaskedCardNumber() string {
	if len(c.CardNumber) < 16 {
		return c.CardNumber
	}
	return c.CardNumber[:4] + " **** **** " + c.CardNumber[12:]
}

// VirtualCardResponse represents the response for virtual card endpoints.
// Masks sensitive data for API responses.
type VirtualCardResponse struct {
	ID                  string            `json:"id"`
	WalletID            string            `json:"wallet_id"`
	UserID              string            `json:"user_id"`
	CardNumberMasked    string            `json:"card_number_masked"`
	CardHolderName      string            `json:"card_holder_name"`
	ExpiryMonth         int               `json:"expiry_month"`
	ExpiryYear          int               `json:"expiry_year"`
	CardType            CardType          `json:"card_type"`
	Status              CardStatus        `json:"status"`
	DailyLimit          int64             `json:"daily_limit"`
	MonthlyLimit        int64             `json:"monthly_limit"`
	PerTransactionLimit int64             `json:"per_transaction_limit"`
	DailySpent          int64             `json:"daily_spent"`
	MonthlySpent        int64             `json:"monthly_spent"`
	LastUsedAt          *models.Timestamp `json:"last_used_at,omitempty"`
	CreatedAt           models.Timestamp  `json:"created_at"`
}

// ToResponse converts a VirtualCard to VirtualCardResponse with masked data.
func (c *VirtualCard) ToResponse() *VirtualCardResponse {
	return &VirtualCardResponse{
		ID:                  c.ID,
		WalletID:            c.WalletID,
		UserID:              c.UserID,
		CardNumberMasked:    c.MaskedCardNumber(),
		CardHolderName:      c.CardHolderName,
		ExpiryMonth:         c.ExpiryMonth,
		ExpiryYear:          c.ExpiryYear,
		CardType:            c.CardType,
		Status:              c.Status,
		DailyLimit:          c.DailyLimit,
		MonthlyLimit:        c.MonthlyLimit,
		PerTransactionLimit: c.PerTransactionLimit,
		DailySpent:          c.DailySpent,
		MonthlySpent:        c.MonthlySpent,
		LastUsedAt:          c.LastUsedAt,
		CreatedAt:           c.CreatedAt,
	}
}

// CreateVirtualCardRequest represents a request to create a virtual card.
type CreateVirtualCardRequest struct {
	CardHolderName string `json:"card_holder_name" validate:"required,min=3,max=100"`
}

// RevealCardDetailsResponse represents the response with full card details.
// Only returned on explicit reveal request with additional authentication.
type RevealCardDetailsResponse struct {
	CardNumber  string `json:"card_number"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	CVV         string `json:"cvv"`
}

// FreezeCardRequest represents a request to freeze a card.
type FreezeCardRequest struct {
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}

// UnfreezeCardRequest represents a request to unfreeze a card.
type UnfreezeCardRequest struct {
	// No additional fields required
}

// UpdateCardLimitsRequest represents a request to update card limits.
type UpdateCardLimitsRequest struct {
	DailyLimit          *int64 `json:"daily_limit,omitempty" validate:"omitempty,gte=0"`
	MonthlyLimit        *int64 `json:"monthly_limit,omitempty" validate:"omitempty,gte=0"`
	PerTransactionLimit *int64 `json:"per_transaction_limit,omitempty" validate:"omitempty,gte=0"`
}

// CancelCardRequest represents a request to cancel a card.
type CancelCardRequest struct {
	Reason string `json:"reason" validate:"required,min=3,max=500"`
}
