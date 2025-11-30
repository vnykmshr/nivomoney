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

	"github.com/vnykmshr/nivo/services/transaction/internal/handler"
	"github.com/vnykmshr/nivo/services/transaction/internal/repository"
	"github.com/vnykmshr/nivo/services/transaction/internal/router"
	"github.com/vnykmshr/nivo/services/transaction/internal/service"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/events"
)

const (
	serviceName = "transaction"
	apiVersion  = "v1"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] Failed to load configuration: %v", serviceName, err)
	}

	// Setup logging
	log.Printf("[%s] Starting Transaction Service...", serviceName)
	log.Printf("[%s] Environment: %s", serviceName, cfg.Environment)
	log.Printf("[%s] Port: %d", serviceName, cfg.ServicePort)

	// Connect to database
	db, err := database.NewFromURL(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[%s] Failed to connect to database: %v", serviceName, err)
	}
	defer func() { _ = db.Close() }()

	log.Printf("[%s] Connected to database successfully", serviceName)

	// Run migrations
	if err := runMigrations(db, cfg); err != nil {
		log.Fatalf("[%s] Failed to run migrations: %v", serviceName, err)
	}

	log.Printf("[%s] Database migrations completed", serviceName)

	// Initialize repository layer
	transactionRepo := repository.NewTransactionRepository(db.DB)

	// Initialize Risk Service client
	riskServiceURL := getEnvOrDefault("RISK_SERVICE_URL", "http://risk-service:8085")
	riskClient := service.NewRiskClient(riskServiceURL)
	log.Printf("[%s] Risk Service URL: %s", serviceName, riskServiceURL)

	// Initialize Wallet Service client
	walletServiceURL := getEnvOrDefault("WALLET_SERVICE_URL", "http://wallet-service:8083")
	walletClient := service.NewWalletClient(walletServiceURL)
	log.Printf("[%s] Wallet Service URL: %s", serviceName, walletServiceURL)

	// Initialize Ledger Service client
	ledgerServiceURL := getEnvOrDefault("LEDGER_SERVICE_URL", "http://ledger-service:8084")
	ledgerClient := service.NewLedgerClient(ledgerServiceURL)
	log.Printf("[%s] Ledger Service URL: %s", serviceName, ledgerServiceURL)

	// Initialize event publisher
	gatewayURL := getEnvOrDefault("GATEWAY_URL", "http://gateway:8000")
	eventPublisher := events.NewPublisher(events.PublishConfig{
		GatewayURL:  gatewayURL,
		ServiceName: serviceName,
	})
	log.Printf("[%s] Event publisher initialized (Gateway: %s)", serviceName, gatewayURL)

	// Initialize service layer
	transactionService := service.NewTransactionService(transactionRepo, riskClient, walletClient, ledgerClient, eventPublisher)

	// Initialize handler layer
	transactionHandler := handler.NewTransactionHandler(transactionService)

	// Get JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("[%s] JWT_SECRET environment variable is required and must not be empty", serviceName)
	}

	// Setup routes
	httpHandler := router.SetupRoutes(transactionHandler, jwtSecret)

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

// runMigrations runs database migrations for the Transaction Service.
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
