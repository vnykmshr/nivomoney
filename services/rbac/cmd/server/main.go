package main

import (
	"net/http"

	"github.com/vnykmshr/nivo/services/rbac/internal/handler"
	"github.com/vnykmshr/nivo/services/rbac/internal/repository"
	"github.com/vnykmshr/nivo/services/rbac/internal/service"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	server.Run(server.ServiceConfig{
		Name: "rbac",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repository layer
			rbacRepo := repository.NewRBACRepository(ctx.DB.DB)

			// Initialize service layer
			rbacService := service.NewRBACService(rbacRepo)

			// Initialize handler layer
			rbacHandler := handler.NewRBACHandler(rbacService)

			// Get JWT secret and setup routes
			jwtSecret := server.RequireEnv("JWT_SECRET")

			return handler.SetupRoutes(rbacHandler, jwtSecret), nil
		},
	})
}
