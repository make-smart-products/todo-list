package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/make-smart-products/todo-list/internal/sim"
)

func TestServesBrowserClient(t *testing.T) {
	server := NewServer(sim.NewDefaultSimulation())
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for root page, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "Oil Worker") {
		t.Fatal("expected root page to include Oil Worker title")
	}
}

func TestResetRestoresScenario(t *testing.T) {
	server := NewServer(sim.NewDefaultSimulation())

	tickPayload, _ := json.Marshal(map[string]int{"hours": 3})
	server.ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/tick", bytes.NewReader(tickPayload)),
	)

	resetRecorder := httptest.NewRecorder()
	server.ServeHTTP(resetRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/reset", nil))

	if resetRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 from reset endpoint, got %d", resetRecorder.Code)
	}

	var snapshot sim.Snapshot
	if err := json.Unmarshal(resetRecorder.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("could not decode reset snapshot: %v", err)
	}

	if snapshot.TickHours != 0 {
		t.Fatalf("expected reset tick to return to zero, got %d", snapshot.TickHours)
	}
}
