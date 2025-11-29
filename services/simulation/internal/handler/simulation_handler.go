package handler

import (
	"encoding/json"
	"net/http"

	"github.com/vnykmshr/nivo/services/simulation/internal/service"
)

// SimulationHandler handles HTTP requests for simulation control
type SimulationHandler struct {
	engine *service.SimulationEngine
}

// NewSimulationHandler creates a new simulation handler
func NewSimulationHandler(engine *service.SimulationEngine) *SimulationHandler {
	return &SimulationHandler{
		engine: engine,
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
