package main

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/risk/internal/handler"
	"github.com/vnykmshr/nivo/services/risk/internal/repository"
	"github.com/vnykmshr/nivo/services/risk/internal/service"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	server.Run(server.ServiceConfig{
		Name: "risk",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repositories
			ruleRepo := repository.NewRiskRuleRepository(ctx.DB.DB)
			eventRepo := repository.NewRiskEventRepository(ctx.DB.DB)

			// Initialize services
			riskService := service.NewRiskService(ruleRepo, eventRepo)

			// Initialize router
			router := handler.NewRouter(riskService)

			return router.SetupRoutes(), nil
		},
	})
}
