// Package healthcheck provides a simple HTTP health endpoint for driftwatch.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status holds the current health state of the daemon.
type Status struct {
	Healthy   bool      `json:"healthy"`
	StartedAt time.Time `json:"started_at"`
	Checks    int64     `json:"checks_performed"`
}

// Handler is an HTTP handler that reports daemon health.
type Handler struct {
	startedAt time.Time
	checks    atomic.Int64
}

// New creates a new Handler.
func New() *Handler {
	return &Handler{
		startedAt: time.Now().UTC(),
	}
}

// IncrementChecks records that one drift check cycle has completed.
func (h *Handler) IncrementChecks() {
	h.checks.Add(1)
}

// ServeHTTP writes a JSON health response.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := Status{
		Healthy:   true,
		StartedAt: h.startedAt,
		Checks:    h.checks.Load(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}
