package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/make-smart-products/todo-list/internal/sim"
)

type Server struct {
	mu  sync.Mutex
	sim *sim.Simulation
	mux *http.ServeMux
}

type tickRequest struct {
	Hours int `json:"hours"`
}

type maintenanceRequest struct {
	StationID string `json:"station_id"`
}

type loadRequest struct {
	StationID  string  `json:"station_id"`
	LoadFactor float64 `json:"load_factor"`
}

type reserveRequest struct {
	StationID string `json:"station_id"`
}

type bypassRequest struct {
	Share float64 `json:"share"`
}

func NewServer(simulation *sim.Simulation) *Server {
	server := &Server{
		sim: simulation,
		mux: http.NewServeMux(),
	}

	server.routes()
	return server
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/state", s.handleState)
	s.mux.HandleFunc("POST /api/v1/tick", s.handleTick)
	s.mux.HandleFunc("POST /api/v1/stations/maintenance", s.handleMaintenance)
	s.mux.HandleFunc("POST /api/v1/stations/load", s.handleLoad)
	s.mux.HandleFunc("POST /api/v1/stations/prepare_reserve", s.handlePrepareReserve)
	s.mux.HandleFunc("POST /api/v1/network/bypass_share", s.handleBypassShare)
}

func (s *Server) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleState(writer http.ResponseWriter, _ *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	writeJSON(writer, http.StatusOK, s.sim.Snapshot())
}

func (s *Server) handleTick(writer http.ResponseWriter, request *http.Request) {
	var payload tickRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	writeJSON(writer, http.StatusOK, s.sim.Advance(payload.Hours))
}

func (s *Server) handleMaintenance(writer http.ResponseWriter, request *http.Request) {
	var payload maintenanceRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.sim.PerformMaintenance(payload.StationID)
	if err != nil {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(writer, http.StatusOK, snapshot)
}

func (s *Server) handleLoad(writer http.ResponseWriter, request *http.Request) {
	var payload loadRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.sim.SetStationLoad(payload.StationID, payload.LoadFactor)
	if err != nil {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(writer, http.StatusOK, snapshot)
}

func (s *Server) handlePrepareReserve(writer http.ResponseWriter, request *http.Request) {
	var payload reserveRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot, err := s.sim.PrepareReserve(payload.StationID)
	if err != nil {
		writeJSON(writer, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(writer, http.StatusOK, snapshot)
}

func (s *Server) handleBypassShare(writer http.ResponseWriter, request *http.Request) {
	var payload bypassRequest
	if err := decodeJSON(request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	writeJSON(writer, http.StatusOK, s.sim.SetBypassShare(payload.Share))
}

func decodeJSON(request *http.Request, destination any) error {
	defer request.Body.Close()

	if request.Body == nil {
		return errors.New("request body is required")
	}

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}

	return nil
}

func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}
