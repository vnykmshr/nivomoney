package models

import (
	"github.com/vnykmshr/nivo/shared/models"
)

// Beneficiary represents a saved recipient for quick transfers.
type Beneficiary struct {
	ID                  string            `json:"id" db:"id"`
	OwnerUserID         string            `json:"owner_user_id" db:"owner_user_id"`                 // User who saved this beneficiary
	BeneficiaryUserID   string            `json:"beneficiary_user_id" db:"beneficiary_user_id"`     // User being saved
	BeneficiaryWalletID string            `json:"beneficiary_wallet_id" db:"beneficiary_wallet_id"` // Default wallet for transfers
	Nickname            string            `json:"nickname" db:"nickname"`                           // Friendly name
	BeneficiaryPhone    string            `json:"beneficiary_phone" db:"beneficiary_phone"`         // Phone for display
	Metadata            map[string]string `json:"metadata,omitempty" db:"metadata"`                 // JSONB metadata
	CreatedAt           models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt           models.Timestamp  `json:"updated_at" db:"updated_at"`
}

// AddBeneficiaryRequest represents a request to add a new beneficiary.
type AddBeneficiaryRequest struct {
	Phone    string `json:"phone" validate:"required,e164"`             // Phone number to add (e.g., "+919876543210")
	Nickname string `json:"nickname" validate:"required,min=1,max=100"` // Friendly name
}

// UpdateBeneficiaryRequest represents a request to update a beneficiary's nickname.
type UpdateBeneficiaryRequest struct {
	Nickname string `json:"nickname" validate:"required,min=1,max=100"` // New nickname
}

// BeneficiaryResponse represents a beneficiary in API responses.
type BeneficiaryResponse struct {
	ID        string           `json:"id"`
	Nickname  string           `json:"nickname"`
	Phone     string           `json:"phone"`
	WalletID  string           `json:"wallet_id"`
	CreatedAt models.Timestamp `json:"created_at"`
}

// ToBeneficiaryResponse converts a Beneficiary to a BeneficiaryResponse.
func ToBeneficiaryResponse(b *Beneficiary) *BeneficiaryResponse {
	return &BeneficiaryResponse{
		ID:        b.ID,
		Nickname:  b.Nickname,
		Phone:     b.BeneficiaryPhone,
		WalletID:  b.BeneficiaryWalletID,
		CreatedAt: b.CreatedAt,
	}
}
