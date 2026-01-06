package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vnykmshr/nivo/services/simulation/internal/config"
	"github.com/vnykmshr/nivo/services/simulation/internal/metrics"
	"github.com/vnykmshr/nivo/services/simulation/internal/service"
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

// StartSimulation starts the simulation engine
func (h *SimulationHandler) StartSimulation(w http.ResponseWriter, r *http.Request) {
	if h.engine.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "simulation already running",
		})
		return
	}

	h.engine.Start(r.Context())

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "simulation started",
	})
}

// StopSimulation stops the simulation engine
func (h *SimulationHandler) StopSimulation(w http.ResponseWriter, r *http.Request) {
	if !h.engine.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "simulation not running",
		})
		return
	}

	h.engine.Stop()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "simulation stopped",
	})
}

// GetConfig handles GET /api/v1/simulation/config
// Returns current simulation configuration.
func (h *SimulationHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.config.GetView()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to read request body",
		})
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req UpdateConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Validate input values
	if req.FailureRate != nil && (*req.FailureRate < 0 || *req.FailureRate > 1.0) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "failure_rate must be between 0.0 and 1.0",
		})
		return
	}
	if req.MinDelayMs != nil && *req.MinDelayMs < 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "min_delay_ms must be non-negative",
		})
		return
	}
	if req.MaxDelayMs != nil && *req.MaxDelayMs < 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "max_delay_ms must be non-negative",
		})
		return
	}
	if req.Mode != nil && *req.Mode != "realistic" && *req.Mode != "demo" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "mode must be 'realistic' or 'demo'",
		})
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
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to read request body",
		})
		return
	}
	defer func() { _ = r.Body.Close() }()

	var req SetModeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if req.Mode != "realistic" && req.Mode != "demo" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "mode must be 'realistic' or 'demo'",
		})
		return
	}

	h.config.SetMode(config.SimulationMode(req.Mode))
	h.metrics.SetMode(req.Mode)

	cfg := h.config.GetView()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "mode switched to " + req.Mode,
		"config":  cfg,
	})
}

// GetMetrics handles GET /api/v1/simulation/metrics
// Returns simulation metrics.
func (h *SimulationHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot := h.metrics.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics":      snapshot,
		"uptime":       h.metrics.GetUptime().String(),
		"success_rate": h.metrics.GetSuccessRate(),
	})
}

// ResetMetrics handles POST /api/v1/simulation/metrics/reset
// Resets simulation metrics.
func (h *SimulationHandler) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	h.metrics.Reset()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "metrics reset",
	})
}
