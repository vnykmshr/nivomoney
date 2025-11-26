package handler

import (
	"io"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/vnykmshr/gopantic/pkg/model"
	"github.com/vnykmshr/nivo/services/identity/internal/models"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max:255"`
	Phone    string `json:"phone" validate:"required,indian_phone"`
	FullName string `json:"full_name" validate:"required,min:2,max:100"`
	Password string `json:"password" validate:"required,min:8,max:72"`
}

// Register handles user registration.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request using gopantic
	req, err := model.ParseInto[RegisterRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Create user request
	createReq := &models.CreateUserRequest{
		Email:    req.Email,
		Phone:    normalizeIndianPhone(req.Phone),
		FullName: req.FullName,
		Password: req.Password,
	}

	// Register user
	user, svcErr := h.authService.Register(r.Context(), createReq)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.Created(w, user)
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

// Login handles user authentication.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request using gopantic
	req, err := model.ParseInto[LoginRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Extract IP address and user agent
	ipAddress := extractIPAddress(r)
	userAgent := r.UserAgent()

	// Login request
	loginReq := &models.LoginRequest{
		Identifier: normalizeIndianPhone(req.Identifier),
		Password:   req.Password,
	}

	// Authenticate user
	loginResp, svcErr := h.authService.Login(r.Context(), loginReq, ipAddress, userAgent)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, loginResp)
}

// Logout handles session termination.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	token := extractBearerToken(r)
	if token == "" {
		response.Error(w, errors.Unauthorized("missing authorization token"))
		return
	}

	// Logout user
	if svcErr := h.authService.Logout(r.Context(), token); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.NoContent(w)
}

// LogoutAll handles termination of all sessions for a user.
// POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	// Extract user from context (set by auth middleware)
	user := getUserFromContext(r.Context())
	if user == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Logout all sessions
	if svcErr := h.authService.LogoutAll(r.Context(), user.ID); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.NoContent(w)
}

// GetProfile retrieves the current user's profile.
// GET /api/v1/auth/me
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user from context (set by auth middleware)
	user := getUserFromContext(r.Context())
	if user == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Get fresh user data with KYC
	freshUser, svcErr := h.authService.GetUserByID(r.Context(), user.ID)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, freshUser)
}

// UpdateKYCRequest represents a KYC update request.
type UpdateKYCRequest struct {
	PAN         string         `json:"pan" validate:"required,pan"`
	Aadhaar     string         `json:"aadhaar" validate:"required,aadhaar"`
	DateOfBirth string         `json:"date_of_birth" validate:"required,date:2006-01-02"`
	Address     AddressRequest `json:"address" validate:"required"`
}

// AddressRequest represents an address in a request.
type AddressRequest struct {
	Street  string `json:"street" validate:"required,min:5,max:200"`
	City    string `json:"city" validate:"required,min:2,max:100"`
	State   string `json:"state" validate:"required,min:2,max:100"`
	PIN     string `json:"pin" validate:"required,pincode"`
	Country string `json:"country" validate:"required,len:2"`
}

// UpdateKYC handles KYC information submission/update.
// PUT /api/v1/auth/kyc
func (h *AuthHandler) UpdateKYC(w http.ResponseWriter, r *http.Request) {
	// Extract user from context (set by auth middleware)
	user := getUserFromContext(r.Context())
	if user == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request using gopantic
	req, err := model.ParseInto[UpdateKYCRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Create KYC update request
	kycReq := &models.UpdateKYCRequest{
		PAN:         req.PAN,
		Aadhaar:     req.Aadhaar,
		DateOfBirth: req.DateOfBirth,
		Address: models.Address{
			Street:  req.Address.Street,
			City:    req.Address.City,
			State:   req.Address.State,
			PIN:     req.Address.PIN,
			Country: req.Address.Country,
		},
	}

	// Update KYC
	kyc, svcErr := h.authService.UpdateKYC(r.Context(), user.ID, kycReq)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, kyc)
}

// GetKYC retrieves the current user's KYC information.
// GET /api/v1/auth/kyc
func (h *AuthHandler) GetKYC(w http.ResponseWriter, r *http.Request) {
	// Extract user from context (set by auth middleware)
	user := getUserFromContext(r.Context())
	if user == nil {
		response.Error(w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Return KYC from user context (already loaded)
	response.OK(w, user.KYC)
}

// VerifyKYCRequest represents a KYC verification request (admin only).
type VerifyKYCRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

// VerifyKYC approves a user's KYC (admin operation).
// POST /api/v1/admin/kyc/verify
func (h *AuthHandler) VerifyKYC(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request using gopantic
	req, err := model.ParseInto[VerifyKYCRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Verify KYC
	if svcErr := h.authService.VerifyKYC(r.Context(), req.UserID); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.NoContent(w)
}

// RejectKYCRequest represents a KYC rejection request (admin only).
type RejectKYCRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Reason string `json:"reason" validate:"required,min:10,max:500"`
}

// RejectKYC rejects a user's KYC (admin operation).
// POST /api/v1/admin/kyc/reject
func (h *AuthHandler) RejectKYC(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}

	// Parse and validate request using gopantic
	req, err := model.ParseInto[RejectKYCRequest](body)
	if err != nil {
		response.Error(w, errors.Validation(err.Error()))
		return
	}

	// Reject KYC
	if svcErr := h.authService.RejectKYC(r.Context(), req.UserID, req.Reason); svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.NoContent(w)
}

// ListPendingKYCs retrieves all pending KYC submissions (admin operation).
// GET /api/v1/admin/kyc/pending
func (h *AuthHandler) ListPendingKYCs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get pending KYCs
	kycList, svcErr := h.authService.ListPendingKYCs(r.Context(), limit, offset)
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, kycList)
}

// AdminStatsResponse represents admin dashboard statistics.
type AdminStatsResponse struct {
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	PendingKYC        int `json:"pending_kyc"`
	TotalWallets      int `json:"total_wallets"`
	TotalTransactions int `json:"total_transactions"`
}

// GetAdminStats retrieves statistics for admin dashboard (admin operation).
// GET /api/v1/admin/stats
func (h *AuthHandler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	// Get stats from service
	stats, svcErr := h.authService.GetAdminStats(r.Context())
	if svcErr != nil {
		response.Error(w, svcErr)
		return
	}

	response.OK(w, stats)
}

// extractBearerToken extracts the JWT token from the Authorization header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// extractIPAddress extracts the client IP address from the request.
func extractIPAddress(r *http.Request) string {
	// Try X-Forwarded-For header first (for proxied requests)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	// Try X-Real-IP header
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr (strip port if present)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails, return as-is (might be IPv6 without port)
		return r.RemoteAddr
	}
	return host
}

// normalizeIndianPhone normalizes Indian phone numbers by adding +91 prefix
// if the input is a 10-digit number starting with 6-9.
// Otherwise, returns the input as-is (could be email or already formatted phone).
func normalizeIndianPhone(input string) string {
	// Pattern: exactly 10 digits starting with 6, 7, 8, or 9
	tenDigitPattern := regexp.MustCompile(`^[6-9][0-9]{9}$`)

	if tenDigitPattern.MatchString(input) {
		return "+91" + input
	}

	return input
}
