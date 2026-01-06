package router

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/wallet/internal/handler"
	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// SetupRoutes configures all routes for the wallet service using Go 1.22+ stdlib router.
func SetupRoutes(walletHandler *handler.WalletHandler, beneficiaryHandler *handler.BeneficiaryHandler, upiHandler *handler.UPIDepositHandler, cardHandler *handler.VirtualCardHandler, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint (public)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","service":"wallet"}`))
	})

	// Metrics endpoint
	mux.Handle("GET /metrics", metrics.Handler())

	// Setup auth middleware
	authConfig := middleware.AuthConfig{
		JWTSecret: jwtSecret,
		SkipPaths: []string{"/health"},
	}
	authMiddleware := middleware.Auth(authConfig)

	// Permission middleware
	createWalletPerm := middleware.RequirePermission("wallet:wallet:create")
	readWalletPerm := middleware.RequirePermission("wallet:wallet:read")
	manageWalletPerm := middleware.RequireAnyPermission("wallet:wallet:activate", "wallet:wallet:freeze", "wallet:wallet:close")

	// ========================================================================
	// Wallet Management Endpoints
	// ========================================================================

	// Wallet CRUD operations
	mux.Handle("POST /api/v1/wallets", authMiddleware(createWalletPerm(http.HandlerFunc(walletHandler.CreateWallet))))
	mux.Handle("GET /api/v1/wallets/{id}", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.GetWallet))))
	mux.Handle("GET /api/v1/wallets/{id}/balance", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.GetWalletBalance))))

	// Wallet limits endpoints (users can read and update their own limits)
	mux.Handle("GET /api/v1/wallets/{id}/limits", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.GetWalletLimits))))
	mux.Handle("PUT /api/v1/wallets/{id}/limits", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.UpdateWalletLimits))))

	// Wallet status management (admin/support operations)
	mux.Handle("POST /api/v1/wallets/{id}/activate", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.ActivateWallet))))
	mux.Handle("POST /api/v1/wallets/{id}/freeze", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.FreezeWallet))))
	mux.Handle("POST /api/v1/wallets/{id}/unfreeze", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.UnfreezeWallet))))
	mux.Handle("POST /api/v1/wallets/{id}/close", authMiddleware(manageWalletPerm(http.HandlerFunc(walletHandler.CloseWallet))))

	// User wallets listing
	mux.Handle("GET /api/v1/users/{userId}/wallets", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.ListUserWallets))))

	// List wallets for authenticated user (convenience endpoint)
	mux.Handle("GET /api/v1/wallets", authMiddleware(readWalletPerm(http.HandlerFunc(walletHandler.ListMyWallets))))

	// ========================================================================
	// UPI Deposit Endpoints
	// ========================================================================

	// UPI deposit operations
	mux.Handle("POST /api/v1/wallets/{id}/deposit/upi", authMiddleware(readWalletPerm(http.HandlerFunc(upiHandler.InitiateDeposit))))
	mux.Handle("GET /api/v1/wallets/{id}/upi", authMiddleware(readWalletPerm(http.HandlerFunc(upiHandler.GetWalletUPIDetails))))
	mux.Handle("GET /api/v1/deposits/upi", authMiddleware(readWalletPerm(http.HandlerFunc(upiHandler.ListDeposits))))
	mux.Handle("GET /api/v1/deposits/upi/{id}", authMiddleware(readWalletPerm(http.HandlerFunc(upiHandler.GetDeposit))))

	// ========================================================================
	// Internal Endpoints (no authentication - service-to-service)
	// ========================================================================

	// Process wallet transfer (called by transaction service)
	mux.HandleFunc("POST /internal/v1/wallets/transfer", walletHandler.ProcessTransfer)
	mux.HandleFunc("POST /internal/v1/wallets/deposit", walletHandler.ProcessDeposit)
	mux.HandleFunc("GET /internal/v1/wallets/{id}/info", walletHandler.GetWalletInfo)

	// ========================================================================
	// Beneficiary Management Endpoints
	// ========================================================================

	// Rate limiting for beneficiary endpoints (prevent abuse/enumeration)
	beneficiaryRateLimit := middleware.RateLimit(middleware.DefaultRateLimitConfig())

	// Permission middleware for beneficiaries
	manageBeneficiaryPerm := middleware.RequirePermission("wallet:beneficiary:manage")

	// Beneficiary CRUD operations (with rate limiting to prevent abuse)
	mux.Handle("POST /api/v1/beneficiaries",
		beneficiaryRateLimit(authMiddleware(manageBeneficiaryPerm(http.HandlerFunc(beneficiaryHandler.AddBeneficiary)))))
	mux.Handle("GET /api/v1/beneficiaries",
		authMiddleware(manageBeneficiaryPerm(http.HandlerFunc(beneficiaryHandler.ListBeneficiaries))))
	mux.Handle("GET /api/v1/beneficiaries/{id}",
		authMiddleware(manageBeneficiaryPerm(http.HandlerFunc(beneficiaryHandler.GetBeneficiary))))
	mux.Handle("PUT /api/v1/beneficiaries/{id}",
		beneficiaryRateLimit(authMiddleware(manageBeneficiaryPerm(http.HandlerFunc(beneficiaryHandler.UpdateBeneficiary)))))
	mux.Handle("DELETE /api/v1/beneficiaries/{id}",
		beneficiaryRateLimit(authMiddleware(manageBeneficiaryPerm(http.HandlerFunc(beneficiaryHandler.DeleteBeneficiary)))))

	// ========================================================================
	// Virtual Card Management Endpoints
	// ========================================================================

	// Permission middleware for cards
	manageCardPerm := middleware.RequirePermission("wallet:card:manage")

	// Card CRUD operations (with rate limiting)
	mux.Handle("POST /api/v1/wallets/{walletId}/cards",
		beneficiaryRateLimit(authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.CreateCard)))))
	mux.Handle("GET /api/v1/wallets/{walletId}/cards",
		authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.ListCards))))
	mux.Handle("GET /api/v1/cards/{id}",
		authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.GetCard))))

	// Card control operations
	mux.Handle("POST /api/v1/cards/{id}/freeze",
		authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.FreezeCard))))
	mux.Handle("POST /api/v1/cards/{id}/unfreeze",
		authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.UnfreezeCard))))
	mux.Handle("DELETE /api/v1/cards/{id}",
		authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.CancelCard))))

	// Card limits management
	mux.Handle("PATCH /api/v1/cards/{id}/limits",
		authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.UpdateCardLimits))))

	// Card details reveal (requires additional security in production)
	mux.Handle("GET /api/v1/cards/{id}/reveal",
		beneficiaryRateLimit(authMiddleware(manageCardPerm(http.HandlerFunc(cardHandler.RevealCardDetails)))))

	// Apply middleware chain
	metricsCollector := metrics.NewCollector("wallet")
	handler := metricsCollector.Middleware("wallet")(mux)

	// Apply request ID
	handler = middleware.RequestID()(handler)

	return handler
}
