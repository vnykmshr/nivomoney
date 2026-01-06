package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// Router sets up HTTP routes for the Identity Service.
type Router struct {
	authHandler         *AuthHandler
	verificationHandler *VerificationHandler
	passwordHandler     *PasswordHandler
	authMiddleware      *AuthMiddleware
	userAdminValidation *UserAdminValidation
	metrics             *metrics.Collector
}

// NewRouter creates a new router with all handlers and middleware.
func NewRouter(authService *service.AuthService, verificationService *service.VerificationService) *Router {
	return &Router{
		authHandler:         NewAuthHandler(authService),
		verificationHandler: NewVerificationHandler(verificationService),
		passwordHandler:     NewPasswordHandler(authService, verificationService),
		authMiddleware:      NewAuthMiddleware(authService),
		userAdminValidation: NewUserAdminValidation(authService),
		metrics:             metrics.NewCollector("identity"),
	}
}

// SetupRoutes configures all HTTP routes for the Identity Service.
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Rate limiting for auth endpoints (prevent brute force)
	authRateLimit := middleware.RateLimit(middleware.DefaultRateLimitConfig())
	strictRateLimit := middleware.RateLimit(middleware.StrictRateLimitConfig())

	// Public routes (no authentication required) - with rate limiting
	mux.Handle("POST /api/v1/auth/register", authRateLimit(http.HandlerFunc(r.authHandler.Register)))
	mux.Handle("POST /api/v1/auth/login", authRateLimit(http.HandlerFunc(r.authHandler.Login)))

	// ========================================================================
	// Password Reset Routes (public - no auth required)
	// ========================================================================

	// Initiate password reset (creates verification request)
	mux.Handle("POST /api/v1/auth/password/forgot",
		strictRateLimit(http.HandlerFunc(r.passwordHandler.ForgotPassword)))

	// Complete password reset using verification token
	mux.Handle("POST /api/v1/auth/password/reset",
		strictRateLimit(http.HandlerFunc(r.passwordHandler.ResetPassword)))

	// Protected routes (authentication required)
	mux.Handle("POST /api/v1/auth/logout",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Logout)))

	mux.Handle("POST /api/v1/auth/logout-all",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.LogoutAll)))

	mux.Handle("GET /api/v1/auth/me",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetProfile)))

	// Alias for frontend compatibility (uses /users/me instead of /auth/me)
	mux.Handle("GET /api/v1/users/me",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetProfile)))

	mux.Handle("PUT /api/v1/users/me",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.UpdateProfile)))

	mux.Handle("PUT /api/v1/users/me/password",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.ChangePassword)))

	// ========================================================================
	// Password Change Routes (protected - requires authentication + verification)
	// ========================================================================

	// Initiate password change (creates verification request)
	mux.Handle("POST /api/v1/auth/password/change/initiate",
		r.authMiddleware.Authenticate(
			http.HandlerFunc(r.passwordHandler.InitiatePasswordChange)))

	// Complete password change using verification token
	mux.Handle("POST /api/v1/auth/password/change/complete",
		r.authMiddleware.Authenticate(
			http.HandlerFunc(r.passwordHandler.CompletePasswordChange)))

	// User lookup (rate limited to prevent phone number enumeration)
	mux.Handle("GET /api/v1/users/lookup",
		strictRateLimit(
			r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.LookupUser))))

	mux.Handle("GET /api/v1/auth/kyc",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetKYC)))

	mux.Handle("PUT /api/v1/auth/kyc",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.UpdateKYC)))

	// Admin routes (authentication + permission required) - with strict rate limiting
	kycVerifyPermission := r.authMiddleware.RequirePermission("identity:kyc:verify")
	kycRejectPermission := r.authMiddleware.RequirePermission("identity:kyc:reject")
	kycListPermission := r.authMiddleware.RequirePermission("identity:kyc:list")
	userSuspendPermission := r.authMiddleware.RequirePermission("identity:user:suspend")
	userUnsuspendPermission := r.authMiddleware.RequirePermission("identity:user:unsuspend")

	mux.Handle("GET /api/v1/admin/kyc/pending",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycListPermission(http.HandlerFunc(r.authHandler.ListPendingKYCs)))))

	mux.Handle("GET /api/v1/admin/stats",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycListPermission(http.HandlerFunc(r.authHandler.GetAdminStats)))))

	mux.Handle("GET /api/v1/admin/users/search",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycListPermission(http.HandlerFunc(r.authHandler.SearchUsers)))))

	mux.Handle("GET /api/v1/admin/users/{id}",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycListPermission(http.HandlerFunc(r.authHandler.GetUserDetails)))))

	mux.Handle("POST /api/v1/admin/kyc/verify",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycVerifyPermission(http.HandlerFunc(r.authHandler.VerifyKYC)))))

	mux.Handle("POST /api/v1/admin/kyc/reject",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycRejectPermission(http.HandlerFunc(r.authHandler.RejectKYC)))))

	mux.Handle("POST /api/v1/admin/users/{id}/suspend",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				userSuspendPermission(http.HandlerFunc(r.authHandler.SuspendUser)))))

	mux.Handle("POST /api/v1/admin/users/{id}/unsuspend",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				userUnsuspendPermission(http.HandlerFunc(r.authHandler.UnsuspendUser)))))

	// ========================================================================
	// Verification Routes (OTP-based verification for sensitive operations)
	// ========================================================================

	// Create verification request (any authenticated user)
	mux.Handle("POST /api/v1/verifications",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				http.HandlerFunc(r.verificationHandler.CreateVerification))))

	// Get pending verifications with OTP codes (User-Admin only)
	mux.Handle("GET /api/v1/verifications/pending",
		r.authMiddleware.Authenticate(
			http.HandlerFunc(r.verificationHandler.GetPendingVerifications)))

	// Get my verification history (any authenticated user, sanitized)
	mux.Handle("GET /api/v1/verifications/me",
		r.authMiddleware.Authenticate(
			http.HandlerFunc(r.verificationHandler.GetMyVerifications)))

	// Get specific verification by ID
	mux.Handle("GET /api/v1/verifications/{id}",
		r.authMiddleware.Authenticate(
			http.HandlerFunc(r.verificationHandler.GetVerification)))

	// Verify OTP and get verification token
	mux.Handle("POST /api/v1/verifications/{id}/verify",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				http.HandlerFunc(r.verificationHandler.VerifyOTP))))

	// Cancel a pending verification
	mux.Handle("DELETE /api/v1/verifications/{id}",
		r.authMiddleware.Authenticate(
			http.HandlerFunc(r.verificationHandler.CancelVerification)))

	// ========================================================================
	// User-Admin Routes (for User-Admin accounts only)
	// ========================================================================

	// Get paired user profile (User-Admin only)
	// The LoadPairedUserID middleware adds the paired user ID to context
	mux.Handle("GET /api/v1/user-admin/paired-user",
		r.authMiddleware.Authenticate(
			r.authMiddleware.RequireRole("user_admin")(
				r.userAdminValidation.LoadPairedUserID(
					http.HandlerFunc(r.authHandler.GetPairedUserProfile)))))

	// Health check endpoint
	mux.HandleFunc("GET /health", healthCheck)

	// Metrics endpoint
	mux.Handle("GET /metrics", metrics.Handler())

	// Apply middleware chain
	handler := r.applyMiddleware(mux)
	return handler
}

// applyMiddleware applies the middleware chain to the handler.
func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	// Apply metrics (outermost layer)
	handler = r.metrics.Middleware("identity")(handler)

	// Apply request ID generation/extraction
	handler = middleware.RequestID()(handler)

	return handler
}

// healthCheck is a simple health check endpoint.
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy","service":"identity"}`))
}
