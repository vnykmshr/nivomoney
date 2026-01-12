package main

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/wallet/internal/handler"
	"github.com/vnykmshr/nivo/services/wallet/internal/repository"
	"github.com/vnykmshr/nivo/services/wallet/internal/router"
	"github.com/vnykmshr/nivo/services/wallet/internal/service"
	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	server.Run(server.ServiceConfig{
		Name: "wallet",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repository layer
			walletRepo := repository.NewWalletRepository(ctx.DB.DB)
			beneficiaryRepo := repository.NewBeneficiaryRepository(ctx.DB.DB)
			upiDepositRepo := repository.NewUPIDepositRepository(ctx.DB.DB)
			virtualCardRepo := repository.NewVirtualCardRepository(ctx.DB.DB)

			// Initialize event publisher
			eventPublisher := events.NewPublisher(events.PublishConfig{
				GatewayURL:  server.GetEnv("GATEWAY_URL", "http://gateway:8000"),
				ServiceName: "wallet",
			})

			// Initialize external service clients
			ledgerClient := service.NewLedgerClient(server.GetEnv("LEDGER_SERVICE_URL", "http://ledger-service:8081"))
			notificationClient := clients.NewNotificationClient(server.GetEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8087"))
			identityClient := service.NewIdentityClient(server.GetEnv("IDENTITY_SERVICE_URL", "http://identity-service:8080"))

			// Initialize service layer
			walletService := service.NewWalletService(walletRepo, eventPublisher, ledgerClient, notificationClient, identityClient)
			beneficiaryService := service.NewBeneficiaryService(beneficiaryRepo, walletRepo, identityClient, eventPublisher)
			upiDepositService := service.NewUPIDepositService(upiDepositRepo, walletRepo, eventPublisher)
			virtualCardService := service.NewVirtualCardService(virtualCardRepo, walletRepo)

			// Initialize handler layer
			walletHandler := handler.NewWalletHandler(walletService)
			beneficiaryHandler := handler.NewBeneficiaryHandler(beneficiaryService)
			upiDepositHandler := handler.NewUPIDepositHandler(upiDepositService)
			virtualCardHandler := handler.NewVirtualCardHandler(virtualCardService)

			// Setup routes
			jwtSecret := server.RequireEnv("JWT_SECRET")

			return router.SetupRoutes(walletHandler, beneficiaryHandler, upiDepositHandler, virtualCardHandler, jwtSecret), nil
		},
	})
}
