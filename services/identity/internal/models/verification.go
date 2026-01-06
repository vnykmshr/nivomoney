package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
	sharedModels "github.com/vnykmshr/nivo/shared/models"
)

// VerificationStatus represents the status of a verification request.
type VerificationStatus string

const (
	VerificationStatusPending   VerificationStatus = "pending"
	VerificationStatusVerified  VerificationStatus = "verified"
	VerificationStatusExpired   VerificationStatus = "expired"
	VerificationStatusCancelled VerificationStatus = "cancelled"
)

// OperationType represents the type of operation requiring verification.
type OperationType string

const (
	OpPasswordChange    OperationType = "password_change"
	OpEmailChange       OperationType = "email_change"
	OpPhoneChange       OperationType = "phone_change"
	OpHighValueTransfer OperationType = "high_value_transfer"
	OpBeneficiaryAdd    OperationType = "beneficiary_add"
	Op2FAEnable         OperationType = "2fa_enable"
	Op2FADisable        OperationType = "2fa_disable"
)

// HighValueThreshold is the amount (in paisa) above which transfers require verification.
const HighValueThreshold int64 = 1000000 // ₹10,000.00

// VerificationRequest represents a verification request with OTP code.
type VerificationRequest struct {
	ID            string                  `json:"id" db:"id"`
	UserID        string                  `json:"user_id" db:"user_id"`
	OperationType OperationType           `json:"operation_type" db:"operation_type"`
	OTPCode       string                  `json:"otp_code,omitempty" db:"otp_code"` // Only visible to User-Admin
	Status        VerificationStatus      `json:"status" db:"status"`
	Metadata      VerificationMeta        `json:"metadata,omitempty" db:"metadata"`
	ExpiresAt     sharedModels.Timestamp  `json:"expires_at" db:"expires_at"`
	CreatedAt     sharedModels.Timestamp  `json:"created_at" db:"created_at"`
	VerifiedAt    *sharedModels.Timestamp `json:"verified_at,omitempty" db:"verified_at"`
	AttemptCount  int                     `json:"attempt_count" db:"attempt_count"`
	LastAttemptAt *sharedModels.Timestamp `json:"last_attempt_at,omitempty" db:"last_attempt_at"`
}

// VerificationMeta stores operation-specific context data.
type VerificationMeta map[string]interface{}

// Scan implements sql.Scanner for VerificationMeta.
func (m *VerificationMeta) Scan(value interface{}) error {
	if value == nil {
		*m = make(VerificationMeta)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		*m = make(VerificationMeta)
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer for VerificationMeta.
func (m VerificationMeta) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// SanitizeForUser removes OTP code (for regular user view).
func (v *VerificationRequest) SanitizeForUser() *VerificationRequest {
	return &VerificationRequest{
		ID:            v.ID,
		UserID:        v.UserID,
		OperationType: v.OperationType,
		Status:        v.Status,
		Metadata:      v.Metadata,
		ExpiresAt:     v.ExpiresAt,
		CreatedAt:     v.CreatedAt,
		VerifiedAt:    v.VerifiedAt,
		AttemptCount:  v.AttemptCount,
		// OTPCode intentionally omitted
	}
}

// IsExpired checks if the verification request has expired.
func (v *VerificationRequest) IsExpired() bool {
	return time.Now().After(v.ExpiresAt.Time)
}

// IsPending checks if the verification is still pending.
func (v *VerificationRequest) IsPending() bool {
	return v.Status == VerificationStatusPending
}

// VerificationToken represents a short-lived token for completing verified operations.
type VerificationToken struct {
	Token         string           `json:"token"`
	OperationType OperationType    `json:"operation_type"`
	Metadata      VerificationMeta `json:"metadata,omitempty"`
	ExpiresAt     time.Time        `json:"expires_at"`
}

// VerificationClaims contains the JWT claims for a verification token.
type VerificationClaims struct {
	VerificationID string           `json:"verification_id"`
	UserID         string           `json:"user_id"`
	OperationType  OperationType    `json:"operation_type"`
	Metadata       VerificationMeta `json:"metadata,omitempty"`
	jwt.RegisteredClaims
}

// CreateVerificationRequest represents the request to create a verification.
type CreateVerificationRequest struct {
	OperationType OperationType          `json:"operation_type" validate:"required"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// VerifyOTPRequest represents the request to verify an OTP.
type VerifyOTPRequest struct {
	OTP string `json:"otp" validate:"required,len=6,numeric"`
}

// ValidOperationTypes returns the list of valid operation types.
func ValidOperationTypes() map[OperationType]bool {
	return map[OperationType]bool{
		OpPasswordChange:    true,
		OpEmailChange:       true,
		OpPhoneChange:       true,
		OpHighValueTransfer: true,
		OpBeneficiaryAdd:    true,
		Op2FAEnable:         true,
		Op2FADisable:        true,
	}
}

// IsValidOperationType checks if the operation type is valid.
func IsValidOperationType(op OperationType) bool {
	return ValidOperationTypes()[op]
}
