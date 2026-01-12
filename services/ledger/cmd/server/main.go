package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vnykmshr/nivo/services/ledger/internal/handler"
	"github.com/vnykmshr/nivo/services/ledger/internal/repository"
	"github.com/vnykmshr/nivo/services/ledger/internal/service"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/logger"
)

const serviceName = "ledger"

func main() {
	// Initialize logger first
	appLogger := logger.NewFromEnv(serviceName)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatalf("Failed to load configuration: %v", err)
	}

	// Startup logging
	appLogger.Info("Starting Ledger Service...")
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
	if err := runMigrations(db, appLogger); err != nil {
		appLogger.Fatalf("Failed to run migrations: %v", err)
	}

	appLogger.Info("Database migrations completed")

	// Initialize repositories
	accountRepo := repository.NewAccountRepository(db)
	journalRepo := repository.NewJournalEntryRepository(db)

	// Initialize services
	ledgerService := service.NewLedgerService(accountRepo, journalRepo)

	// Get JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		appLogger.Fatal("JWT_SECRET environment variable is required and must not be empty")
	}

	// Initialize router
	router := handler.NewRouter(ledgerService, jwtSecret)
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

// runMigrations runs database migrations for the Ledger Service.
func runMigrations(db *database.DB, log *logger.Logger) error {
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
