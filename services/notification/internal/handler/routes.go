package handler

import (
	"net/http"

	"github.com/vnykmshr/nivo/shared/metrics"
	"github.com/vnykmshr/nivo/shared/middleware"
)

// Router handles HTTP routing for the notification service.
type Router struct {
	handler *NotificationHandler
	metrics *metrics.Collector
}

// NewRouter creates a new router.
func NewRouter(handler *NotificationHandler) *Router {
	return &Router{
		handler: handler,
		metrics: metrics.NewCollector("notification"),
	}
}

// SetupRoutes configures all routes for the notification service.
func (ro *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check (public)
	mux.HandleFunc("GET /health", ro.handler.Health)
	mux.HandleFunc("GET /ready", ro.handler.Health)

	// Metrics endpoint (for Prometheus)
	mux.Handle("GET /metrics", metrics.Handler())

	// Notification endpoints
	mux.HandleFunc("POST /v1/notifications/send", ro.handler.SendNotification)
	mux.HandleFunc("GET /v1/notifications/{id}", ro.handler.GetNotification)
	mux.HandleFunc("GET /v1/notifications", ro.handler.ListNotifications)

	// Template endpoints
	mux.HandleFunc("POST /v1/templates", ro.handler.CreateTemplate)
	mux.HandleFunc("GET /v1/templates/{id}", ro.handler.GetTemplate)
	mux.HandleFunc("GET /v1/templates", ro.handler.ListTemplates)
	mux.HandleFunc("PUT /v1/templates/{id}", ro.handler.UpdateTemplate)
	mux.HandleFunc("POST /v1/templates/{id}/preview", ro.handler.PreviewTemplate)

	// Admin endpoints (protected by RBAC in gateway)
	mux.HandleFunc("GET /admin/notifications/stats", ro.handler.GetStats)
	mux.HandleFunc("POST /admin/notifications/{id}/replay", ro.handler.ReplayNotification)

	// Apply middleware chain
	handler := ro.applyMiddleware(mux)
	return handler
}

// applyMiddleware applies the middleware chain to the handler.
func (ro *Router) applyMiddleware(handler http.Handler) http.Handler {
	// Apply metrics middleware (outermost layer)
	handler = ro.metrics.Middleware("notification")(handler)

	// Apply request ID generation/extraction
	handler = middleware.RequestID()(handler)
	return handler
}
