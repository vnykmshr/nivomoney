// Package server provides common bootstrapping for all Nivo services.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
	"github.com/vnykmshr/nivo/shared/logger"
)

// HTTP server timeouts
const (
	ReadTimeout     = 15 * time.Second
	WriteTimeout    = 15 * time.Second
	IdleTimeout     = 60 * time.Second
	ShutdownTimeout = 30 * time.Second
)

// BootstrapContext contains initialized resources for service setup.
type BootstrapContext struct {
	Logger *logger.Logger
	Config *config.Config
	DB     *database.DB
}

// ServiceConfig defines how to bootstrap and run a service.
type ServiceConfig struct {
	// Name is the service name (used for logging and identification).
	Name string

	// SetupHandler is called after DB connection to initialize
	// repositories, services, and handlers. Returns the HTTP handler.
	SetupHandler func(ctx *BootstrapContext) (http.Handler, error)

	// Cleanup is called during graceful shutdown (optional).
	// Use this for closing additional resources like Redis connections.
	Cleanup func() error
}

// Run bootstraps and runs the service with common initialization.
func Run(cfg ServiceConfig) {
	// Initialize logger first
	appLogger := logger.NewFromEnv(cfg.Name)

	// Load configuration
	appConfig, err := config.Load()
	if err != nil {
		appLogger.Fatalf("Failed to load configuration: %v", err)
	}

	// Startup logging
	appLogger.Info("Starting " + cfg.Name + " service...")
	appLogger.WithField("environment", appConfig.Environment).Info("Environment configured")
	appLogger.WithField("port", appConfig.ServicePort).Info("Port configured")

	// Connect to database
	db, err := database.NewFromURL(appConfig.DatabaseURL)
	if err != nil {
		appLogger.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() { _ = db.Close() }()

	appLogger.Info("Connected to database successfully")

	// Run migrations
	if err := runMigrations(db, appLogger); err != nil {
		appLogger.Fatalf("Failed to run migrations: %v", err)
	}

	appLogger.Info("Database migrations completed")

	// Create bootstrap context for service-specific setup
	ctx := &BootstrapContext{
		Logger: appLogger,
		Config: appConfig,
		DB:     db,
	}

	// Call service-specific setup
	handler, err := cfg.SetupHandler(ctx)
	if err != nil {
		appLogger.Fatalf("Failed to setup service: %v", err)
	}

	// Create HTTP server
	addr := fmt.Sprintf(":%d", appConfig.ServicePort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.WithError(err).Warn("Server forced to shutdown")
	}

	// Run custom cleanup if provided
	if cfg.Cleanup != nil {
		if err := cfg.Cleanup(); err != nil {
			appLogger.WithError(err).Warn("Cleanup error during shutdown")
		}
	}

	appLogger.Info("Server stopped gracefully")
}

// runMigrations runs database migrations for the service.
func runMigrations(db *database.DB, log *logger.Logger) error {
	// Get migrations directory path
	migrationsDir := GetEnv("MIGRATIONS_DIR", "./migrations")

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

// GetEnv returns the environment variable value or a default value.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// RequireEnv returns the environment variable value or panics if not set.
// Use this for required configuration like JWT_SECRET.
func RequireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s environment variable is required and must not be empty", key))
	}
	return value
}
