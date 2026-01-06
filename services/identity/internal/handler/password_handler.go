package handler

import (
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
	"github.com/vnykmshr/nivo/shared/response"
)

// PasswordHandler handles password-related HTTP requests.
type PasswordHandler struct {
	authService         *service.AuthService
	verificationService *service.VerificationService
}

// NewPasswordHandler creates a new password handler.
func NewPasswordHandler(
	authService *service.AuthService,
	verificationService *service.VerificationService,
) *PasswordHandler {
	return &PasswordHandler{
		authService:         authService,
		verificationService: verificationService,
	}
}

// ForgotPasswordRequest represents a forgot password request.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ForgotPassword handles POST /api/v1/auth/password/forgot
// Public endpoint - no authentication required.
// Creates a verification request for password reset.
func (h *PasswordHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	req, err := model.ParseInto[ForgotPasswordRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Look up user by email - only regular users can reset password
	user, lookupErr := h.authService.GetUserByEmail(r.Context(), req.Email, models.AccountTypeUser)
	if lookupErr != nil {
		// Don't reveal if email exists - prevents enumeration attacks
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"message":   "If an account exists with this email, a verification request has been created.",
			"next_step": "Check your admin portal for the verification code.",
		})
		return
	}

	// Create verification request for password reset
	verification, verifyErr := h.verificationService.CreateVerification(
		r.Context(),
		user.ID,
		models.OpPasswordChange, // Using password_change operation type
		map[string]interface{}{
			"email":   user.Email,
			"request": "password_reset",
		},
	)
	if verifyErr != nil {
		// Rate limited
		if verifyErr.Code == errors.ErrCodeRateLimit {
			response.Error(w, verifyErr)
			return
		}
		// For other errors, return generic message
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"message":   "If an account exists with this email, a verification request has been created.",
			"next_step": "Check your admin portal for the verification code.",
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"verification_id": verification.ID,
		"message":         "Verification request created. Login to your admin portal to get the OTP code.",
		"next_step":       "Enter the OTP from your admin portal to reset your password.",
		"expires_in":      "10 minutes",
	})
}

// ResetPasswordRequest represents a password reset request.
type ResetPasswordRequest struct {
	VerificationToken string `json:"verification_token" validate:"required"`
	NewPassword       string `json:"new_password" validate:"required,min:8"`
}

// ResetPassword handles POST /api/v1/auth/password/reset
// Public endpoint - uses verification token for authorization.
func (h *PasswordHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	req, err := model.ParseInto[ResetPasswordRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Validate password strength
	if strengthErr := validatePasswordStrength(req.NewPassword); strengthErr != nil {
		response.Error(w, strengthErr)
		return
	}

	// Process password reset using verification token
	resetErr := h.authService.ResetPasswordWithToken(r.Context(), req.VerificationToken, req.NewPassword)
	if resetErr != nil {
		response.Error(w, resetErr)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Password has been reset successfully. Please login with your new password.",
	})
}

// InitiatePasswordChangeRequest represents a request to initiate password change.
type InitiatePasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
}

// InitiatePasswordChange handles POST /api/v1/auth/password/change/initiate
// Protected endpoint - requires authentication.
// Creates a verification request for password change.
func (h *PasswordHandler) InitiatePasswordChange(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	req, err := model.ParseInto[InitiatePasswordChangeRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Verify current password first
	if verifyErr := h.authService.VerifyCurrentPassword(r.Context(), userID, req.CurrentPassword); verifyErr != nil {
		response.Error(w, verifyErr)
		return
	}

	// Create verification request
	verification, verifyErr := h.verificationService.CreateVerification(
		r.Context(),
		userID,
		models.OpPasswordChange,
		nil,
	)
	if verifyErr != nil {
		response.Error(w, verifyErr)
		return
	}

	response.JSON(w, http.StatusAccepted, map[string]interface{}{
		"verification_required": true,
		"verification_id":       verification.ID,
		"message":               "Please verify this action via your admin portal to complete the password change.",
		"expires_in":            "10 minutes",
	})
}

// CompletePasswordChangeRequest represents a request to complete password change.
type CompletePasswordChangeRequest struct {
	VerificationToken string `json:"verification_token" validate:"required"`
	NewPassword       string `json:"new_password" validate:"required,min:8"`
}

// CompletePasswordChange handles POST /api/v1/auth/password/change/complete
// Protected endpoint - requires authentication + verification token.
func (h *PasswordHandler) CompletePasswordChange(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	req, err := model.ParseInto[CompletePasswordChangeRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Validate password strength
	if strengthErr := validatePasswordStrength(req.NewPassword); strengthErr != nil {
		response.Error(w, strengthErr)
		return
	}

	// Complete password change with verification token
	changeErr := h.authService.ChangePasswordWithToken(r.Context(), userID, req.VerificationToken, req.NewPassword)
	if changeErr != nil {
		response.Error(w, changeErr)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Password changed successfully.",
	})
}

// validatePasswordStrength checks password meets security requirements.
func validatePasswordStrength(password string) *errors.Error {
	// Min length check
	if len(password) < 8 {
		return errors.BadRequest("password must be at least 8 characters")
	}

	// Max length check - bcrypt has a 72-byte limit
	if len(password) > 72 {
		return errors.BadRequest("password must be at most 72 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasNumber = true
		case strings.ContainsRune(specialChars, c):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.BadRequest("password must contain uppercase, lowercase, number, and special character")
	}

	return nil
}
