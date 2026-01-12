package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vnykmshr/nivo/gateway/internal/handler"
	"github.com/vnykmshr/nivo/gateway/internal/proxy"
	"github.com/vnykmshr/nivo/gateway/internal/router"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/logger"
)

const serviceName = "gateway"

func main() {
	// Initialize logger first (before config load for early error logging)
	appLogger := logger.NewFromEnv(serviceName)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		appLogger.Fatalf("Failed to load configuration: %v", err)
	}

	// Startup logging
	appLogger.Info("Starting API Gateway...")
	appLogger.WithField("environment", cfg.Environment).Info("Environment configured")
	appLogger.WithField("port", cfg.ServicePort).Info("Port configured")

	// Initialize service registry
	registry := proxy.NewServiceRegistry()
	appLogger.Info("Service registry initialized")
	for name, url := range registry.AllServices() {
		appLogger.WithField("service", name).WithField("url", url).Debug("Registered service")
	}

	// Initialize gateway with logger
	gateway := proxy.NewGateway(registry, appLogger)
	appLogger.Info("Gateway proxy initialized")

	// Initialize SSE broker
	broker := events.NewBroker()
	broker.Start()
	appLogger.Info("SSE event broker started")

	// Initialize SSE handler
	sseHandler := handler.NewSSEHandler(broker, appLogger)
	appLogger.Info("SSE handler initialized")

	// Initialize router
	apiRouter := router.NewRouter(gateway, sseHandler, appLogger)
	httpHandler := apiRouter.SetupRoutes()
	appLogger.Info("Routes configured")

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.ServicePort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		appLogger.WithField("addr", addr).Info("Gateway listening")
		appLogger.Infof("Ready to accept requests at http://localhost%s/api/v1/", addr)
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

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Warn("Server forced to shutdown")
	}

	appLogger.Info("Server shutdown complete")
}
