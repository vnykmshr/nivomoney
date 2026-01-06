package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/middleware"
	"github.com/vnykmshr/nivo/shared/response"
)

// VerificationHandler handles verification-related HTTP requests.
type VerificationHandler struct {
	service *service.VerificationService
}

// NewVerificationHandler creates a new verification handler.
func NewVerificationHandler(svc *service.VerificationService) *VerificationHandler {
	return &VerificationHandler{service: svc}
}

// CreateVerification handles POST /api/v1/verifications
// Creates a new verification request for a sensitive operation.
func (h *VerificationHandler) CreateVerification(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req models.CreateVerificationRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.Error(w, errors.BadRequest("invalid request body"))
		return
	}

	// Validate operation type
	if !models.IsValidOperationType(req.OperationType) {
		response.Error(w, errors.BadRequest("invalid operation type"))
		return
	}

	// Create verification
	verification, createErr := h.service.CreateVerification(
		r.Context(),
		userID,
		req.OperationType,
		req.Metadata,
	)
	if createErr != nil {
		response.Error(w, createErr)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"data": map[string]interface{}{
			"verification": verification,
			"message":      "Verification created. Please check your admin portal for the OTP code.",
		},
	})
}

// GetPendingVerifications handles GET /api/v1/verifications/pending
// For User-Admin to see pending verifications with OTP codes.
func (h *VerificationHandler) GetPendingVerifications(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	// Get account type from context
	accountType, _ := middleware.GetAccountType(r.Context())

	// Only User-Admin can see OTP codes
	if accountType != string(models.AccountTypeUserAdmin) {
		response.Error(w, errors.Forbidden("only user-admin accounts can view verification codes"))
		return
	}

	verifications, err := h.service.GetPendingForUserAdmin(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"verifications": verifications,
			"count":         len(verifications),
		},
	})
}

// GetMyVerifications handles GET /api/v1/verifications/me
// For regular user to see their verification history (without OTP).
func (h *VerificationHandler) GetMyVerifications(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	// Get status filter from query
	status := r.URL.Query().Get("status") // "pending", "verified", "expired", "cancelled", "all"
	if status == "" {
		status = "all"
	}

	verifications, err := h.service.GetUserVerifications(r.Context(), userID, status)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Sanitize (remove OTP codes)
	sanitized := make([]*models.VerificationRequest, len(verifications))
	for i, v := range verifications {
		sanitized[i] = v.SanitizeForUser()
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"verifications": sanitized,
			"count":         len(sanitized),
		},
	})
}

// VerifyOTP handles POST /api/v1/verifications/{id}/verify
// Verifies the OTP and returns a verification token.
func (h *VerificationHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	// Get verification ID from path
	verificationID := r.PathValue("id")
	if verificationID == "" {
		response.Error(w, errors.BadRequest("verification ID required"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req models.VerifyOTPRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.Error(w, errors.BadRequest("invalid request body"))
		return
	}

	if req.OTP == "" {
		response.Error(w, errors.BadRequest("OTP is required"))
		return
	}

	// Verify OTP
	token, verifyErr := h.service.VerifyOTP(r.Context(), verificationID, userID, req.OTP)
	if verifyErr != nil {
		response.Error(w, verifyErr)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"verified": true,
			"token":    token,
			"message":  "Verification successful. Use the token to complete your operation.",
		},
	})
}

// CancelVerification handles DELETE /api/v1/verifications/{id}
// Cancels a pending verification request.
func (h *VerificationHandler) CancelVerification(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	// Get verification ID from path
	verificationID := r.PathValue("id")
	if verificationID == "" {
		response.Error(w, errors.BadRequest("verification ID required"))
		return
	}

	if err := h.service.CancelVerification(r.Context(), verificationID, userID); err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"cancelled": true,
			"message":   "Verification cancelled",
		},
	})
}

// GetVerification handles GET /api/v1/verifications/{id}
// Gets a specific verification request (sanitized for regular users).
func (h *VerificationHandler) GetVerification(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID == "" {
		response.Error(w, errors.Unauthorized("authentication required"))
		return
	}

	// Get verification ID from path
	verificationID := r.PathValue("id")
	if verificationID == "" {
		response.Error(w, errors.BadRequest("verification ID required"))
		return
	}

	verification, err := h.service.GetByID(r.Context(), verificationID)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Check ownership
	if verification.UserID != userID {
		// Check if this is a User-Admin accessing their paired user's verification
		accountType, _ := middleware.GetAccountType(r.Context())
		if accountType != string(models.AccountTypeUserAdmin) {
			response.Error(w, errors.Forbidden("not authorized to view this verification"))
			return
		}

		// Validate that this User-Admin is paired with the verification's owner
		canAccess, accessErr := h.service.CanUserAdminAccessVerification(r.Context(), userID, verification)
		if accessErr != nil || !canAccess {
			response.Error(w, errors.Forbidden("not authorized to view this verification"))
			return
		}

		// User-Admin can see full details including OTP for their paired user
		response.JSON(w, http.StatusOK, map[string]interface{}{
			"data": verification,
		})
		return
	}

	// Regular user sees sanitized version (no OTP)
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data": verification.SanitizeForUser(),
	})
}
