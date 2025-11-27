package router

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/transaction/internal/handler"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// SetupRoutes configures all routes for the transaction service using Go 1.22+ stdlib router.
func SetupRoutes(transactionHandler *handler.TransactionHandler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (public)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"transaction"}`))
	})

	// Metrics endpoint
	mux.Handle("GET /metrics", metrics.Handler())

	// Setup auth middleware
	authConfig := middleware.AuthConfig{
		JWTSecret: jwtSecret,
		SkipPaths: []string{"/health"},
	}
	authMiddleware := middleware.Auth(authConfig)

	// Rate limiting for money movement (prevent abuse)
	moneyRateLimit := middleware.RateLimit(middleware.StrictRateLimitConfig())

	// Permission middleware
	createTransferPerm := middleware.RequirePermission("transaction:transfer:create")
	createDepositPerm := middleware.RequirePermission("transaction:deposit:create")
	createWithdrawalPerm := middleware.RequirePermission("transaction:withdrawal:create")
	readTransactionPerm := middleware.RequirePermission("transaction:transaction:read")
	listTransactionsPerm := middleware.RequirePermission("transaction:transaction:list")
	reverseTransactionPerm := middleware.RequirePermission("transaction:transaction:reverse")

	// ========================================================================
	// Transaction Creation Endpoints (with strict rate limiting)
	// ========================================================================

	mux.Handle("POST /api/v1/transactions/transfer", moneyRateLimit(authMiddleware(createTransferPerm(http.HandlerFunc(transactionHandler.CreateTransfer)))))
	mux.Handle("POST /api/v1/transactions/deposit", moneyRateLimit(authMiddleware(createDepositPerm(http.HandlerFunc(transactionHandler.CreateDeposit)))))
	mux.Handle("POST /api/v1/transactions/deposit/upi", moneyRateLimit(authMiddleware(createDepositPerm(http.HandlerFunc(transactionHandler.InitiateUPIDeposit)))))
	mux.Handle("POST /api/v1/transactions/deposit/upi/complete", authMiddleware(http.HandlerFunc(transactionHandler.CompleteUPIDeposit))) // Webhook endpoint (no rate limit)
	mux.Handle("POST /api/v1/transactions/withdrawal", moneyRateLimit(authMiddleware(createWithdrawalPerm(http.HandlerFunc(transactionHandler.CreateWithdrawal)))))

	// ========================================================================
	// Transaction Retrieval Endpoints
	// ========================================================================

	mux.Handle("GET /api/v1/transactions/{id}", authMiddleware(readTransactionPerm(http.HandlerFunc(transactionHandler.GetTransaction))))
	mux.Handle("GET /api/v1/wallets/{walletId}/transactions", authMiddleware(listTransactionsPerm(http.HandlerFunc(transactionHandler.ListWalletTransactions))))

	// ========================================================================
	// Transaction Reversal Endpoint (Admin Operation - with strict rate limiting)
	// ========================================================================

	mux.Handle("POST /api/v1/transactions/{id}/reverse", moneyRateLimit(authMiddleware(reverseTransactionPerm(http.HandlerFunc(transactionHandler.ReverseTransaction)))))

	// ========================================================================
	// Internal Endpoints (no authentication - service-to-service)
	// ========================================================================

	// Process transfer (executes wallet transfer with limit checking)
	mux.HandleFunc("POST /internal/v1/transactions/{id}/process", transactionHandler.ProcessTransfer)

	// Apply middleware chain
	metricsCollector := metrics.NewCollector("transaction")
	handler := metricsCollector.Middleware("transaction")(mux)

	// Apply request ID
	handler = middleware.RequestID()(handler)

	return handler
}
