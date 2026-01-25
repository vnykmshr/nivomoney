package router

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/vnykmshr/nivo/gateway/internal/handler"
	"github.com/vnykmshr/nivo/gateway/internal/middleware"
	"github.com/vnykmshr/nivo/gateway/internal/proxy"
	"github.com/vnykmshr/nivo/shared/logger"
	"github.com/vnykmshr/nivo/shared/metrics"
	sharedMiddleware "github.com/vnykmshr/nivo/shared/middleware"
)

// Router configures HTTP routes for the API Gateway.
type Router struct {
	gateway    *proxy.Gateway
	sseHandler *handler.SSEHandler
	validator  *middleware.JWTValidator
	logger     *logger.Logger
	metrics    *metrics.Collector
}

// NewRouter creates a new router with all handlers and middleware.
func NewRouter(gateway *proxy.Gateway, sseHandler *handler.SSEHandler, log *logger.Logger) *Router {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	return &Router{
		gateway:    gateway,
		sseHandler: sseHandler,
		validator:  middleware.NewJWTValidator(jwtSecret),
		logger:     log,
		metrics:    metrics.NewCollector("gateway"),
	}
}

// SetupRoutes configures all HTTP routes for the gateway.
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (gateway-level)
	mux.HandleFunc("GET /health", r.healthCheck)
	mux.HandleFunc("GET /api/health", r.healthCheck)

	// Metrics endpoint (Prometheus)
	mux.Handle("GET /metrics", metrics.Handler())

	// Public routes (no authentication required)
	// Authentication endpoints - these should go directly to identity service
	// Support both canonical paths (/api/v1/identity/auth/*) and alias paths (/api/v1/auth/*)
	mux.HandleFunc("POST /api/v1/identity/auth/register", r.gateway.ProxyRequest)
	mux.HandleFunc("POST /api/v1/identity/auth/login", r.gateway.ProxyRequest)
	mux.HandleFunc("POST /api/v1/auth/register", r.gateway.ProxyRequest)
	mux.HandleFunc("POST /api/v1/auth/login", r.gateway.ProxyRequest)

	// Password reset endpoints (public - no auth required)
	mux.HandleFunc("POST /api/v1/auth/password/forgot", r.gateway.ProxyRequest)
	mux.HandleFunc("POST /api/v1/auth/password/reset", r.gateway.ProxyRequest)

	// SSE endpoints (authentication optional, can subscribe to public events)
	mux.HandleFunc("GET /api/v1/events", r.sseHandler.HandleEvents)
	mux.HandleFunc("GET /api/v1/events/stats", r.sseHandler.HandleStats)
	mux.HandleFunc("POST /api/v1/events/broadcast", r.sseHandler.HandleBroadcast)

	// Protected routes (authentication required)
	// All other API routes require authentication
	authenticatedHandler := r.validator.Authenticate(http.HandlerFunc(r.gateway.ProxyRequest))
	mux.Handle("/api/v1/", authenticatedHandler)

	// Apply middleware chain
	handler := r.applyMiddleware(mux)

	return handler
}

// applyMiddleware applies the middleware chain to the handler.
func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	// Apply metrics (outermost layer - captures everything)
	handler = r.metrics.Middleware("gateway")(handler)

	// Apply CORS
	corsConfig := r.getCORSConfig()
	handler = sharedMiddleware.CORS(corsConfig)(handler)

	// Apply CSRF protection (production only)
	// In development, frontend doesn't send CSRF tokens
	if os.Getenv("ENVIRONMENT") == "production" {
		csrfConfig := sharedMiddleware.CSRFConfig{
			SkipPaths: []string{
				"/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/auth/refresh",
				"/api/v1/auth/password/forgot",
				"/api/v1/auth/password/reset",
				"/api/v1/identity/auth/login",
				"/api/v1/identity/auth/register",
				"/api/v1/events",
				"/health",
				"/metrics",
				// Internal service endpoints (auth-protected, no browser CSRF risk)
				"/api/v1/identity/auth/kyc",
				// Transaction endpoints (service-to-service, JWT authenticated)
				"/api/v1/transaction/transactions/deposit",
				"/api/v1/transaction/transactions/transfer",
				"/api/v1/transaction/transactions/withdrawal",
			},
			CookiePath:     "/",
			CookieSecure:   true,
			CookieSameSite: http.SameSiteLaxMode,
		}
		handler = sharedMiddleware.CSRF(csrfConfig)(handler)
	}

	// Apply request ID generation
	handler = sharedMiddleware.RequestID()(handler)

	// Apply logging
	handler = sharedMiddleware.Logging(r.logger)(handler)

	// Apply panic recovery
	handler = sharedMiddleware.Recovery(r.logger)(handler)

	// Apply rate limiting (gateway-wide)
	handler = sharedMiddleware.RateLimit(sharedMiddleware.DefaultRateLimitConfig())(handler)

	return handler
}

// getCORSConfig returns CORS configuration based on environment.
// In development, allows localhost origins. In production, uses CORS_ORIGINS env var.
func (r *Router) getCORSConfig() sharedMiddleware.CORSConfig {
	config := sharedMiddleware.DefaultCORSConfig()

	// Check for explicit CORS_ORIGINS environment variable
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		config.AllowedOrigins = splitAndTrim(origins, ",")
	} else if os.Getenv("ENVIRONMENT") != "production" {
		// Development defaults: allow localhost origins
		config.AllowedOrigins = []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:3002",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:3002",
		}
	}

	// Enable credentials for authenticated requests
	config.AllowCredentials = true

	return config
}

// splitAndTrim splits a string by separator and trims whitespace from each part.
func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, part := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// healthCheck is the gateway health check endpoint.
func (r *Router) healthCheck(w http.ResponseWriter, req *http.Request) {
	health := map[string]interface{}{
		"status":  "healthy",
		"service": "gateway",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		r.logger.WithError(err).Error("failed to encode health check response")
	}
}
