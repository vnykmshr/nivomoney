package main

import (
	"net/http"
	"os"
	"time"

	"github.com/vnykmshr/nivo/services/identity/internal/handler"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/cache"
	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	// Track Redis cache for cleanup
	var redisCache *cache.RedisCache

	server.Run(server.ServiceConfig{
		Name: "identity",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repositories
			userRepo := repository.NewUserRepository(ctx.DB)
			userAdminRepo := repository.NewUserAdminRepository(ctx.DB)
			kycRepo := repository.NewKYCRepository(ctx.DB)
			sessionRepo := repository.NewSessionRepository(ctx.DB)
			verificationRepo := repository.NewVerificationRepository(ctx.DB)

			// Initialize external service clients
			rbacClient := service.NewRBACClient(server.GetEnv("RBAC_SERVICE_URL", "http://rbac-service:8082"))
			walletClient := service.NewWalletClient(server.GetEnv("WALLET_SERVICE_URL", "http://wallet-service:8083"))
			notificationClient := clients.NewNotificationClient(server.GetEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8087"))

			// Initialize event publisher
			eventPublisher := events.NewPublisher(events.PublishConfig{
				GatewayURL:  server.GetEnv("GATEWAY_URL", "http://gateway:8000"),
				ServiceName: "identity",
			})

			// Initialize Redis cache (optional - graceful degradation if unavailable)
			var sessionCache cache.Cache
			redisURL := os.Getenv("REDIS_URL")
			if redisURL != "" {
				redisCfg := cache.DefaultRedisConfig(redisURL)
				var err error
				redisCache, err = cache.NewRedisCache(redisCfg)
				if err != nil {
					ctx.Logger.WithError(err).Warn("Redis connection failed, running without cache")
				} else {
					sessionCache = redisCache
					ctx.Logger.Info("Redis cache initialized successfully")
				}
			} else {
				ctx.Logger.Info("REDIS_URL not set, running without session cache")
			}

			// Initialize services
			jwtSecret := server.RequireEnv("JWT_SECRET")
			jwtExpiry := 24 * time.Hour
			authService := service.NewAuthService(userRepo, userAdminRepo, kycRepo, sessionRepo, rbacClient, walletClient, notificationClient, jwtSecret, jwtExpiry, eventPublisher)

			// Enable session caching if Redis is available
			if sessionCache != nil {
				authService.SetCache(sessionCache)
			}

			verificationService := service.NewVerificationService(verificationRepo, userAdminRepo)

			// Initialize router
			router := handler.NewRouter(authService, verificationService)

			return router.SetupRoutes(), nil
		},
		Cleanup: func() error {
			if redisCache != nil {
				return redisCache.Close()
			}
			return nil
		},
	})
}
