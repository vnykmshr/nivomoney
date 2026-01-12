package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vnykmshr/nivo/services/notification/internal/handler"
	"github.com/vnykmshr/nivo/services/notification/internal/repository"
	"github.com/vnykmshr/nivo/services/notification/internal/service"
	"github.com/vnykmshr/nivo/shared/server"
)

func main() {
	// Track worker cancel function for cleanup
	var workerCancel context.CancelFunc

	server.Run(server.ServiceConfig{
		Name: "notification",
		SetupHandler: func(ctx *server.BootstrapContext) (http.Handler, error) {
			// Initialize repositories
			notifRepo := repository.NewNotificationRepository(ctx.DB.DB)
			templateRepo := repository.NewTemplateRepository(ctx.DB.DB)

			// Load simulation configuration
			simConfig := loadSimulationConfig()
			ctx.Logger.WithField("delay_ms", simConfig.DeliveryDelayMs).
				WithField("failure_rate", simConfig.FailureRatePercent).
				WithField("max_retries", simConfig.MaxRetryAttempts).
				Info("Simulation config loaded")

			// Initialize service
			notifService := service.NewNotificationService(notifRepo, templateRepo, simConfig)

			// Start background worker for processing queued notifications
			workerCtx, cancel := context.WithCancel(context.Background())
			workerCancel = cancel

			go func() {
				ctx.Logger.Info("Starting background worker for notification processing...")
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						if err := notifService.ProcessQueuedNotifications(workerCtx, 10); err != nil {
							ctx.Logger.WithError(err).Error("Worker error")
						}
					case <-workerCtx.Done():
						ctx.Logger.Info("Background worker stopped")
						return
					}
				}
			}()

			// Initialize handler and router
			notifHandler := handler.NewNotificationHandler(notifService)
			router := handler.NewRouter(notifHandler)

			return router.SetupRoutes(), nil
		},
		Cleanup: func() error {
			if workerCancel != nil {
				workerCancel()
			}
			return nil
		},
	})
}

// loadSimulationConfig loads simulation configuration from environment variables.
func loadSimulationConfig() service.SimulationConfig {
	config := service.DefaultSimulationConfig()

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
