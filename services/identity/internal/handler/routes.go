package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// Router sets up HTTP routes for the Identity Service.
type Router struct {
	authHandler    *AuthHandler
	authMiddleware *AuthMiddleware
	metrics        *metrics.Collector
}

// NewRouter creates a new router with all handlers and middleware.
func NewRouter(authService *service.AuthService) *Router {
	return &Router{
		authHandler:    NewAuthHandler(authService),
		authMiddleware: NewAuthMiddleware(authService),
		metrics:        metrics.NewCollector("identity"),
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

	// Protected routes (authentication required)
	mux.Handle("POST /api/v1/auth/logout",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.Logout)))

	mux.Handle("POST /api/v1/auth/logout-all",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.LogoutAll)))

	mux.Handle("GET /api/v1/auth/me",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetProfile)))

	mux.Handle("PUT /api/v1/users/me",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.UpdateProfile)))

	mux.Handle("PUT /api/v1/users/me/password",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.ChangePassword)))

	mux.Handle("GET /api/v1/auth/kyc",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.GetKYC)))

	mux.Handle("PUT /api/v1/auth/kyc",
		r.authMiddleware.Authenticate(http.HandlerFunc(r.authHandler.UpdateKYC)))

	// Admin routes (authentication + permission required) - with strict rate limiting
	kycVerifyPermission := r.authMiddleware.RequirePermission("identity:kyc:verify")
	kycRejectPermission := r.authMiddleware.RequirePermission("identity:kyc:reject")
	kycListPermission := r.authMiddleware.RequirePermission("identity:kyc:list")

	mux.Handle("GET /api/v1/admin/kyc/pending",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycListPermission(http.HandlerFunc(r.authHandler.ListPendingKYCs)))))

	mux.Handle("GET /api/v1/admin/stats",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycListPermission(http.HandlerFunc(r.authHandler.GetAdminStats)))))

	mux.Handle("POST /api/v1/admin/kyc/verify",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycVerifyPermission(http.HandlerFunc(r.authHandler.VerifyKYC)))))

	mux.Handle("POST /api/v1/admin/kyc/reject",
		strictRateLimit(
			r.authMiddleware.Authenticate(
				kycRejectPermission(http.HandlerFunc(r.authHandler.RejectKYC)))))

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
