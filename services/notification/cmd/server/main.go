package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vnykmshr/nivo/services/notification/internal/handler"
	"github.com/vnykmshr/nivo/services/notification/internal/repository"
	"github.com/vnykmshr/nivo/services/notification/internal/service"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/database"
)

const serviceName = "notification"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] Failed to load configuration: %v", serviceName, err)
	}

	// Setup logging
	log.Printf("[%s] Starting Notification Service...", serviceName)
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
	notifRepo := repository.NewNotificationRepository(db.DB)
	templateRepo := repository.NewTemplateRepository(db.DB)

	// Load simulation configuration
	simConfig := loadSimulationConfig()
	log.Printf("[%s] Simulation config: delay=%dms, final_delay=%dms, failure_rate=%.1f%%, max_retries=%d",
		serviceName, simConfig.DeliveryDelayMs, simConfig.FinalDelayMs,
		simConfig.FailureRatePercent, simConfig.MaxRetryAttempts)

	// Initialize service
	notifService := service.NewNotificationService(notifRepo, templateRepo, simConfig)

	// Initialize handler and router
	notifHandler := handler.NewNotificationHandler(notifService)
	router := handler.NewRouter(notifHandler)
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

	// Start background worker for processing queued notifications
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	go func() {
		log.Printf("[%s] Starting background worker for notification processing...", serviceName)
		ticker := time.NewTicker(5 * time.Second) // Process every 5 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Process queued notifications in batches
				if err := notifService.ProcessQueuedNotifications(workerCtx, 10); err != nil {
					log.Printf("[%s] Worker error: %v", serviceName, err)
				}
			case <-workerCtx.Done():
				log.Printf("[%s] Background worker stopped", serviceName)
				return
			}
		}
	}()

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

	// Stop background worker
	workerCancel()

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] Server forced to shutdown: %v", serviceName, err)
	}

	log.Printf("[%s] Server stopped gracefully", serviceName)
}

// runMigrations runs database migrations for the Notification Service.
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

// loadSimulationConfig loads simulation configuration from environment variables.
func loadSimulationConfig() service.SimulationConfig {
	config := service.DefaultSimulationConfig()

	// Override with environment variables if present
	if val := os.Getenv("SIM_DELIVERY_DELAY_MS"); val != "" {
		if delay, err := strconv.Atoi(val); err == nil {
			config.DeliveryDelayMs = delay
		}
	}

	if val := os.Getenv("SIM_FINAL_DELAY_MS"); val != "" {
		if delay, err := strconv.Atoi(val); err == nil {
			config.FinalDelayMs = delay
		}
	}

	if val := os.Getenv("SIM_FAILURE_RATE_PERCENT"); val != "" {
		if rate, err := strconv.ParseFloat(val, 64); err == nil {
			config.FailureRatePercent = rate
		}
	}

	if val := os.Getenv("SIM_MAX_RETRY_ATTEMPTS"); val != "" {
		if retries, err := strconv.Atoi(val); err == nil {
			config.MaxRetryAttempts = retries
		}
	}

	if val := os.Getenv("SIM_RETRY_DELAY_MS"); val != "" {
		if delay, err := strconv.Atoi(val); err == nil {
			config.RetryDelayMs = delay
		}
	}

	return config
}

// getEnvOrDefault returns the environment variable value or a default value.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
