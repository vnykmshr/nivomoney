package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/ledger/internal/service"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// Router sets up HTTP routes for the Ledger Service.
type Router struct {
	ledgerHandler *LedgerHandler
	jwtSecret     string
	metrics       *metrics.Collector
}

// NewRouter creates a new router with all handlers.
func NewRouter(ledgerService *service.LedgerService, jwtSecret string) *Router {
	return &Router{
		ledgerHandler: NewLedgerHandler(ledgerService),
		jwtSecret:     jwtSecret,
		metrics:       metrics.NewCollector("ledger"),
	}
}

// SetupRoutes configures all HTTP routes for the Ledger Service.
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (public)
	mux.HandleFunc("GET /health", healthCheck)

	// Metrics endpoint
	mux.Handle("GET /metrics", metrics.Handler())

	// Setup auth middleware
	authConfig := middleware.AuthConfig{
		JWTSecret: r.jwtSecret,
		SkipPaths: []string{"/health"},
	}
	authMiddleware := middleware.Auth(authConfig)

	// Permission middleware for different operations
	accountantPermission := middleware.RequireAnyPermission("ledger:account:create", "ledger:account:update")
	viewLedgerPermission := middleware.RequireAnyPermission("ledger:account:read", "ledger:entry:read")

	// Account endpoints (protected)
	mux.Handle("POST /api/v1/accounts",
		authMiddleware(accountantPermission(http.HandlerFunc(r.ledgerHandler.CreateAccount))))

	mux.Handle("GET /api/v1/accounts",
		authMiddleware(viewLedgerPermission(http.HandlerFunc(r.ledgerHandler.ListAccounts))))

	mux.Handle("GET /api/v1/accounts/{id}/balance",
		authMiddleware(viewLedgerPermission(http.HandlerFunc(r.ledgerHandler.GetAccountBalance))))

	mux.Handle("GET /api/v1/accounts/{id}",
		authMiddleware(viewLedgerPermission(http.HandlerFunc(r.ledgerHandler.GetAccount))))

	mux.Handle("PUT /api/v1/accounts/{id}",
		authMiddleware(accountantPermission(http.HandlerFunc(r.ledgerHandler.UpdateAccount))))

	// Journal entry endpoints (protected)
	mux.Handle("POST /api/v1/journal-entries",
		authMiddleware(middleware.RequirePermission("ledger:entry:create")(http.HandlerFunc(r.ledgerHandler.CreateJournalEntry))))

	mux.Handle("GET /api/v1/journal-entries/{id}",
		authMiddleware(viewLedgerPermission(http.HandlerFunc(r.ledgerHandler.GetJournalEntry))))

	mux.Handle("GET /api/v1/journal-entries",
		authMiddleware(viewLedgerPermission(http.HandlerFunc(r.ledgerHandler.ListJournalEntries))))

	mux.Handle("POST /api/v1/journal-entries/{id}/post",
		authMiddleware(middleware.RequirePermission("ledger:entry:post")(http.HandlerFunc(r.ledgerHandler.PostJournalEntry))))

	mux.Handle("POST /api/v1/journal-entries/{id}/void",
		authMiddleware(middleware.RequirePermission("ledger:entry:void")(http.HandlerFunc(r.ledgerHandler.VoidJournalEntry))))

	mux.Handle("POST /api/v1/journal-entries/{id}/reverse",
		authMiddleware(middleware.RequirePermission("ledger:entry:reverse")(http.HandlerFunc(r.ledgerHandler.ReverseJournalEntry))))

	// ========================================================================
	// Internal Endpoints (No Authentication - Service-to-Service Only)
	// ========================================================================

	// Internal endpoints for wallet service
	mux.HandleFunc("POST /internal/v1/accounts", r.ledgerHandler.CreateAccountInternal)
	mux.HandleFunc("GET /internal/v1/accounts/by-code/{code}", r.ledgerHandler.GetAccountByCode)

	// Apply middleware chain
	handler := r.applyMiddleware(mux)
	return handler
}

// applyMiddleware applies the middleware chain to the handler.
func (r *Router) applyMiddleware(handler http.Handler) http.Handler {
	// Apply metrics (outermost layer)
	handler = r.metrics.Middleware("ledger")(handler)

	// Apply request ID generation/extraction
	handler = middleware.RequestID()(handler)

	return handler
}

// healthCheck is a simple health check endpoint.
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy","service":"ledger"}`))
}
