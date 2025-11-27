package models

import (
	"github.com/vnykmshr/nivo/shared/models"
)

// UserStatus represents the current state of a user account.
type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"   // Awaiting KYC verification
	UserStatusActive    UserStatus = "active"    // KYC verified, account active
	UserStatusSuspended UserStatus = "suspended" // Temporarily disabled
	UserStatusClosed    UserStatus = "closed"    // Permanently closed
)

// KYCStatus represents the KYC verification status.
type KYCStatus string

const (
	KYCStatusPending  KYCStatus = "pending"  // KYC documents submitted
	KYCStatusVerified KYCStatus = "verified" // KYC approved
	KYCStatusRejected KYCStatus = "rejected" // KYC rejected
	KYCStatusExpired  KYCStatus = "expired"  // KYC documents expired
)

// User represents a Nivo user with India-specific identity fields.
type User struct {
	ID        string           `json:"id" db:"id"`
	Email     string           `json:"email" db:"email"`
	Phone     string           `json:"phone" db:"phone"` // Indian phone with +91
	FullName  string           `json:"full_name" db:"full_name"`
	Status    UserStatus       `json:"status" db:"status"`
	CreatedAt models.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt models.Timestamp `json:"updated_at" db:"updated_at"`

	// Password (hashed, never exposed in JSON)
	PasswordHash string `json:"-" db:"password_hash"`

	// KYC Information (India-specific)
	KYC KYCInfo `json:"kyc" db:"-"` // Embedded, stored separately
}

// KYCInfo contains India-specific KYC information.
type KYCInfo struct {
	UserID          string            `json:"user_id" db:"user_id"`
	Status          KYCStatus         `json:"status" db:"status"`
	PAN             string            `json:"pan" db:"pan"`                     // Permanent Account Number
	Aadhaar         string            `json:"-" db:"aadhaar"`                   // Never expose in API (PII)
	DateOfBirth     string            `json:"date_of_birth" db:"date_of_birth"` // YYYY-MM-DD
	Address         Address           `json:"address" db:"-"`                   // Stored as JSONB
	VerifiedAt      *models.Timestamp `json:"verified_at,omitempty" db:"verified_at"`
	RejectedAt      *models.Timestamp `json:"rejected_at,omitempty" db:"rejected_at"`
	RejectionReason string            `json:"rejection_reason,omitempty" db:"rejection_reason"`
	CreatedAt       models.Timestamp  `json:"created_at" db:"created_at"`
	UpdatedAt       models.Timestamp  `json:"updated_at" db:"updated_at"`
}

// Address represents an Indian address.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	PIN     string `json:"pin"`     // 6-digit Indian PIN code
	Country string `json:"country"` // Default: "IN"
}

// Session represents an active user session.
type Session struct {
	ID        string           `json:"id" db:"id"`
	UserID    string           `json:"user_id" db:"user_id"`
	Token     string           `json:"token" db:"token_hash"` // JWT token hash
	IPAddress string           `json:"ip_address" db:"ip_address"`
	UserAgent string           `json:"user_agent" db:"user_agent"`
	ExpiresAt models.Timestamp `json:"expires_at" db:"expires_at"`
	CreatedAt models.Timestamp `json:"created_at" db:"created_at"`
}

// CreateUserRequest represents the request to create a new user (registration).
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,indian_phone"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// LoginRequest represents the login credentials.
// Identifier can be either email or phone number.
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"` // Email or phone number
	Password   string `json:"password" validate:"required"`
}

// LoginResponse contains the authentication token.
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	User      *User  `json:"user"`
}

// UpdateKYCRequest represents KYC document submission.
type UpdateKYCRequest struct {
	PAN         string  `json:"pan" validate:"required,pan"`
	Aadhaar     string  `json:"aadhaar" validate:"required,aadhaar"`
	DateOfBirth string  `json:"date_of_birth" validate:"required"` // Format: YYYY-MM-DD
	Address     Address `json:"address" validate:"required"`
}

// UpdateProfileRequest represents the request to update user profile.
type UpdateProfileRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,indian_phone"`
}

// ChangePasswordRequest represents the request to change user password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
}

// Sanitize removes sensitive information from User before returning in API.
func (u *User) Sanitize() {
	u.PasswordHash = ""
	// Aadhaar is already excluded from JSON
}

// IsActive returns true if the user account is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsKYCVerified returns true if KYC is verified.
func (k *KYCInfo) IsKYCVerified() bool {
	return k.Status == KYCStatusVerified
}
