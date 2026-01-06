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

	"github.com/vnykmshr/nivo/gateway/internal/handler"
	"github.com/vnykmshr/nivo/gateway/internal/proxy"
	"github.com/vnykmshr/nivo/gateway/internal/router"
	"github.com/vnykmshr/nivo/shared/config"
	"github.com/vnykmshr/nivo/shared/events"
	"github.com/vnykmshr/nivo/shared/logger"
)

const serviceName = "gateway"

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[%s] Failed to load configuration: %v", serviceName, err)
	}

	// Initialize logger
	appLogger := logger.NewDefault(serviceName)

	// Setup logging
	log.Printf("[%s] Starting API Gateway...", serviceName)
	log.Printf("[%s] Environment: %s", serviceName, cfg.Environment)
	log.Printf("[%s] Port: %d", serviceName, cfg.ServicePort)

	// Initialize service registry
	registry := proxy.NewServiceRegistry()
	log.Printf("[%s] Service registry initialized:", serviceName)
	for name, url := range registry.AllServices() {
		log.Printf("[%s]   - %s: %s", serviceName, name, url)
	}

	// Initialize gateway
	gateway := proxy.NewGateway(registry)
	log.Printf("[%s] Gateway proxy initialized", serviceName)

	// Initialize SSE broker
	broker := events.NewBroker()
	broker.Start()
	log.Printf("[%s] SSE event broker started", serviceName)

	// Initialize SSE handler
	sseHandler := handler.NewSSEHandler(broker, appLogger)
	log.Printf("[%s] SSE handler initialized", serviceName)

	// Initialize router
	apiRouter := router.NewRouter(gateway, sseHandler, appLogger)
	httpHandler := apiRouter.SetupRoutes()
	log.Printf("[%s] Routes configured", serviceName)

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
		log.Printf("[%s] Gateway listening on %s", serviceName, addr)
		log.Printf("[%s] Ready to accept requests at http://localhost%s/api/v1/", serviceName, addr)
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

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[%s] Server forced to shutdown: %v", serviceName, err)
	}

	log.Printf("[%s] Server shutdown complete", serviceName)
}
