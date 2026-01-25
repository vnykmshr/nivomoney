package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/vnykmshr/nivo/services/simulation/internal/config"
	"github.com/vnykmshr/nivo/services/simulation/internal/metrics"
	"github.com/vnykmshr/nivo/services/simulation/internal/service"
	"github.com/vnykmshr/nivo/shared/errors"
	"github.com/vnykmshr/nivo/shared/response"
)

// SimulationHandler handles HTTP requests for simulation control
type SimulationHandler struct {
	engine  *service.SimulationEngine
	config  *config.SimulationConfig
	metrics *metrics.SimulationMetrics
}

// NewSimulationHandler creates a new simulation handler
func NewSimulationHandler(
	engine *service.SimulationEngine,
	cfg *config.SimulationConfig,
	met *metrics.SimulationMetrics,
) *SimulationHandler {
	return &SimulationHandler{
		engine:  engine,
		config:  cfg,
		metrics: met,
	}
}

// StatusResponse represents the status response
type StatusResponse struct {
	Running bool   `json:"running"`
	Message string `json:"message"`
}

// GetStatus returns the current simulation status
func (h *SimulationHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := StatusResponse{
		Running: h.engine.IsRunning(),
	}

	if status.Running {
		status.Message = "Simulation is running"
	} else {
		status.Message = "Simulation is stopped"
	}

	response.OK(w, status)
}

// StartSimulation starts the simulation engine
func (h *SimulationHandler) StartSimulation(w http.ResponseWriter, r *http.Request) {
	if h.engine.IsRunning() {
		response.Error(w, errors.Conflict("simulation already running"))
		return
	}

	// Use background context for the simulation engine lifecycle.
	// The HTTP request context would cancel when the response is sent,
	// but we want the simulation to run independently.
	h.engine.Start(context.Background())

	response.OK(w, map[string]string{
		"message": "simulation started",
	})
}

// StopSimulation stops the simulation engine
func (h *SimulationHandler) StopSimulation(w http.ResponseWriter, r *http.Request) {
	if !h.engine.IsRunning() {
		response.Error(w, errors.Conflict("simulation not running"))
		return
	}

	h.engine.Stop()

	response.OK(w, map[string]string{
		"message": "simulation stopped",
	})
}

// GetConfig handles GET /api/v1/simulation/config
// Returns current simulation configuration.
func (h *SimulationHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.config.GetView()

	response.OK(w, map[string]interface{}{
		"config": cfg,
	})
}

// UpdateConfigRequest represents a configuration update request.
type UpdateConfigRequest struct {
	Mode            *string  `json:"mode,omitempty"`
	FailureRate     *float64 `json:"failure_rate,omitempty"`
	MinDelayMs      *int     `json:"min_delay_ms,omitempty"`
	MaxDelayMs      *int     `json:"max_delay_ms,omitempty"`
	FailuresEnabled *bool    `json:"failures_enabled,omitempty"`
}

// UpdateConfig handles PUT /api/v1/simulation/config
// Updates simulation configuration.
func (h *SimulationHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req UpdateConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.Error(w, errors.BadRequest("invalid request body"))
		return
	}

	// Validate input values
	if req.FailureRate != nil && (*req.FailureRate < 0 || *req.FailureRate > 1.0) {
		response.Error(w, errors.BadRequest("failure_rate must be between 0.0 and 1.0"))
		return
	}
	if req.MinDelayMs != nil && *req.MinDelayMs < 0 {
		response.Error(w, errors.BadRequest("min_delay_ms must be non-negative"))
		return
	}
	if req.MaxDelayMs != nil && *req.MaxDelayMs < 0 {
		response.Error(w, errors.BadRequest("max_delay_ms must be non-negative"))
		return
	}
	if req.Mode != nil && *req.Mode != "realistic" && *req.Mode != "demo" {
		response.Error(w, errors.BadRequest("mode must be 'realistic' or 'demo'"))
		return
	}

	// Handle mode change separately (SetMode acquires its own lock)
	if req.Mode != nil {
		h.config.SetMode(config.SimulationMode(*req.Mode))
		h.metrics.SetMode(*req.Mode)
	}

	// Update other fields
	h.config.Update(func(cfg *config.SimulationConfig) {
		if req.FailureRate != nil {
			cfg.Failures.FailureRate = *req.FailureRate
		}
		if req.MinDelayMs != nil {
			cfg.Delays.MinDelayMs = *req.MinDelayMs
		}
		if req.MaxDelayMs != nil {
			cfg.Delays.MaxDelayMs = *req.MaxDelayMs
		}
		if req.FailuresEnabled != nil {
			cfg.Failures.Enabled = *req.FailuresEnabled
		}
	})

	cfg := h.config.GetView()
	response.OK(w, map[string]interface{}{
		"message": "configuration updated",
		"config":  cfg,
	})
}

// SetModeRequest represents a mode switch request.
type SetModeRequest struct {
	Mode string `json:"mode"`
}

// SetMode handles POST /api/v1/simulation/mode
// Switches between realistic and demo modes.
func (h *SimulationHandler) SetMode(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, errors.BadRequest("failed to read request body"))
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req SetModeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		response.Error(w, errors.BadRequest("invalid request body"))
		return
	}

	if req.Mode != "realistic" && req.Mode != "demo" {
		response.Error(w, errors.BadRequest("mode must be 'realistic' or 'demo'"))
		return
	}

	h.config.SetMode(config.SimulationMode(req.Mode))
	h.metrics.SetMode(req.Mode)

	cfg := h.config.GetView()
	response.OK(w, map[string]interface{}{
		"message": "mode switched to " + req.Mode,
		"config":  cfg,
	})
}

// GetMetrics handles GET /api/v1/simulation/metrics
// Returns simulation metrics.
func (h *SimulationHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot := h.metrics.GetSnapshot()

	response.OK(w, map[string]interface{}{
		"metrics":      snapshot,
		"uptime":       h.metrics.GetUptime().String(),
		"success_rate": h.metrics.GetSuccessRate(),
	})
}

// ResetMetrics handles POST /api/v1/simulation/metrics/reset
// Resets simulation metrics.
func (h *SimulationHandler) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	h.metrics.Reset()

	response.OK(w, map[string]string{
		"message": "metrics reset",
	})
}
