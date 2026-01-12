package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vnykmshr/nivo/services/identity/internal/handler"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/cache"
	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/logger"
)

const serviceName = "identity"

func main() {
	// Initialize logger first
	appLogger := logger.NewFromEnv(serviceName)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatalf("Failed to load configuration: %v", err)
	}

	// Startup logging
	appLogger.Info("Starting Identity Service...")
	appLogger.WithField("environment", cfg.Environment).Info("Environment configured")
	appLogger.WithField("port", cfg.ServicePort).Info("Port configured")

	// Connect to database
	db, err := database.NewFromURL(cfg.DatabaseURL)
	if err != nil {
		appLogger.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() { _ = db.Close() }()

	appLogger.Info("Connected to database successfully")

	// Run database migrations
	if err := runMigrations(db, cfg, appLogger); err != nil {
		appLogger.Fatalf("Failed to run migrations: %v", err)
	}

	appLogger.Info("Database migrations completed")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	userAdminRepo := repository.NewUserAdminRepository(db)
	kycRepo := repository.NewKYCRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	verificationRepo := repository.NewVerificationRepository(db)

	// Initialize RBAC client
	rbacURL := getEnvOrDefault("RBAC_SERVICE_URL", "http://rbac-service:8082")
	rbacClient := service.NewRBACClient(rbacURL)

	// Initialize event publisher
	gatewayURL := getEnvOrDefault("GATEWAY_URL", "http://gateway:8000")
	eventPublisher := events.NewPublisher(events.PublishConfig{
		GatewayURL:  gatewayURL,
		ServiceName: "identity",
	})
	appLogger.WithField("gateway", gatewayURL).Info("Event publisher initialized")

	// Initialize notification client
	notificationURL := getEnvOrDefault("NOTIFICATION_SERVICE_URL", "http://notification-service:8087")
	notificationClient := clients.NewNotificationClient(notificationURL)
	appLogger.WithField("url", notificationURL).Info("Notification client initialized")

	// Initialize wallet client
	walletURL := getEnvOrDefault("WALLET_SERVICE_URL", "http://wallet-service:8083")
	walletClient := service.NewWalletClient(walletURL)
	appLogger.WithField("url", walletURL).Info("Wallet client initialized")

	// Initialize Redis cache (optional - graceful degradation if unavailable)
	var sessionCache cache.Cache
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		redisCfg := cache.DefaultRedisConfig(redisURL)
		redisCache, err := cache.NewRedisCache(redisCfg)
		if err != nil {
			appLogger.WithError(err).Warn("Redis connection failed, running without cache")
		} else {
			sessionCache = redisCache
			appLogger.Info("Redis cache initialized successfully")
			defer func() { _ = redisCache.Close() }()
		}
	} else {
		appLogger.Info("REDIS_URL not set, running without session cache")
	}

	// Initialize services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		appLogger.Fatal("JWT_SECRET environment variable is required and must not be empty")
	}
	jwtExpiry := 24 * time.Hour // 24 hours
	authService := service.NewAuthService(userRepo, userAdminRepo, kycRepo, sessionRepo, rbacClient, walletClient, notificationClient, jwtSecret, jwtExpiry, eventPublisher)

	// Enable session caching if Redis is available
	if sessionCache != nil {
		authService.SetCache(sessionCache)
	}

	verificationService := service.NewVerificationService(verificationRepo, userAdminRepo)

	// Initialize router
	router := handler.NewRouter(authService, verificationService)
	httpHandler := router.SetupRoutes()

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.ServicePort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.WithField("addr", addr).Info("Server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-quit
	appLogger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Warn("Server forced to shutdown")
	}

	appLogger.Info("Server stopped gracefully")
}

// runMigrations runs database migrations for the Identity Service.
func runMigrations(db *database.DB, cfg *config.Config, log *logger.Logger) error {
	// Get migrations directory path
	migrationsDir := getEnvOrDefault("MIGRATIONS_DIR", "./migrations")

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.WithField("dir", migrationsDir).Info("Migrations directory not found, skipping migrations")
		return nil
	}

	// Run migrations
	migrator := database.NewMigrator(db.DB, migrationsDir)
	if err := migrator.Up(); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// getEnvOrDefault returns the environment variable value or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
