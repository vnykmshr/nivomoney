package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vnykmshr/nivo/services/identity/internal/handler"
	"github.com/vnykmshr/nivo/services/identity/internal/repository"
	"github.com/vnykmshr/nivo/services/identity/internal/service"
	"github.com/vnykmshr/nivo/shared/clients"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/events"
)

const (
	serviceName = "identity"
	apiVersion  = "v1"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] Failed to load configuration: %v", serviceName, err)
	}

	// Setup logging
	log.Printf("[%s] Starting Identity Service...", serviceName)
	log.Printf("[%s] Environment: %s", serviceName, cfg.Environment)
	log.Printf("[%s] Port: %d", serviceName, cfg.ServicePort)

	// Connect to database
	db, err := database.NewFromURL(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[%s] Failed to connect to database: %v", serviceName, err)
	}
	defer func() { _ = db.Close() }()

	log.Printf("[%s] Connected to database successfully", serviceName)

	// Run database migrations
	if err := runMigrations(db, cfg); err != nil {
		log.Fatalf("[%s] Failed to run migrations: %v", serviceName, err)
	}

	log.Printf("[%s] Database migrations completed", serviceName)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	kycRepo := repository.NewKYCRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// Initialize RBAC client
	rbacURL := getEnvOrDefault("RBAC_SERVICE_URL", "http://rbac-service:8082")
	rbacClient := service.NewRBACClient(rbacURL)

	// Initialize event publisher
	gatewayURL := getEnvOrDefault("GATEWAY_URL", "http://gateway:8000")
	eventPublisher := events.NewPublisher(events.PublishConfig{
		GatewayURL:  gatewayURL,
		ServiceName: "identity",
	})
	log.Printf("[%s] Event publisher initialized (Gateway: %s)", serviceName, gatewayURL)

	// Initialize notification client
	notificationURL := getEnvOrDefault("NOTIFICATION_SERVICE_URL", "http://notification-service:8087")
	notificationClient := clients.NewNotificationClient(notificationURL)
	log.Printf("[%s] Notification client initialized (Service: %s)", serviceName, notificationURL)

	// Initialize services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("[%s] JWT_SECRET environment variable is required and must not be empty", serviceName)
	}
	jwtExpiry := 24 * time.Hour // 24 hours
	authService := service.NewAuthService(userRepo, kycRepo, sessionRepo, rbacClient, notificationClient, jwtSecret, jwtExpiry, eventPublisher)

	// Initialize router
	router := handler.NewRouter(authService)
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
		log.Printf("[%s] Server listening on %s", serviceName, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[%s] Server failed to start: %v", serviceName, err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-quit
	log.Printf("[%s] Shutting down server...", serviceName)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] Server forced to shutdown: %v", serviceName, err)
	}

	log.Printf("[%s] Server stopped gracefully", serviceName)
}

// runMigrations runs database migrations for the Identity Service.
func runMigrations(db *database.DB, cfg *config.Config) error {
	// Get migrations directory path
	migrationsDir := getEnvOrDefault("MIGRATIONS_DIR", "./migrations")

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Printf("[%s] Migrations directory not found: %s (skipping migrations)", serviceName, migrationsDir)
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
