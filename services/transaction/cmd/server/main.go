package main

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/transaction/internal/handler"
	"github.com/vnykmshr/nivo/services/transaction/internal/repository"
	"github.com/vnykmshr/nivo/services/transaction/internal/router"
	"github.com/vnykmshr/nivo/services/transaction/internal/service"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	server.Run(server.ServiceConfig{
		Name: "transaction",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repository layer
			transactionRepo := repository.NewTransactionRepository(ctx.DB.DB)

			// Initialize external service clients
			riskClient := service.NewRiskClient(server.GetEnv("RISK_SERVICE_URL", "http://risk-service:8085"))
			walletClient := service.NewWalletClient(server.GetEnv("WALLET_SERVICE_URL", "http://wallet-service:8083"))
			ledgerClient := service.NewLedgerClient(server.GetEnv("LEDGER_SERVICE_URL", "http://ledger-service:8084"))

			// Initialize event publisher
			eventPublisher := events.NewPublisher(events.PublishConfig{
				GatewayURL:  server.GetEnv("GATEWAY_URL", "http://gateway:8000"),
				ServiceName: "transaction",
			})

			// Initialize service layer
			transactionService := service.NewTransactionService(transactionRepo, riskClient, walletClient, ledgerClient, eventPublisher)

			// Initialize handler layer
			transactionHandler := handler.NewTransactionHandler(transactionService, walletClient)

			// Setup routes
			jwtSecret := server.RequireEnv("JWT_SECRET")

			return router.SetupRoutes(transactionHandler, jwtSecret), nil
		},
	})
}
