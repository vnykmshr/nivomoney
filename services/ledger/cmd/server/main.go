package main

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/ledger/internal/handler"
	"github.com/vnykmshr/nivo/services/ledger/internal/repository"
	"github.com/vnykmshr/nivo/services/ledger/internal/service"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	server.Run(server.ServiceConfig{
		Name: "ledger",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repositories
			accountRepo := repository.NewAccountRepository(ctx.DB)
			journalRepo := repository.NewJournalEntryRepository(ctx.DB)

			// Initialize services
			ledgerService := service.NewLedgerService(accountRepo, journalRepo)

			// Get JWT secret and setup router
			jwtSecret := server.RequireEnv("JWT_SECRET")
			router := handler.NewRouter(ledgerService, jwtSecret)

			return router.SetupRoutes(), nil
		},
	})
}
