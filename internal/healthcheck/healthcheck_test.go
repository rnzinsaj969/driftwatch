package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/healthcheck"
)

func TestHealthHandler_Returns200(t *testing.T) {
	h := healthcheck.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHealthHandler_ResponseIsJSON(t *testing.T) {
	h := healthcheck.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	var status healthcheck.Status
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if !status.Healthy {
		t.Error("expected healthy=true")
	}
}

func TestHealthHandler_StartsWithZeroChecks(t *testing.T) {
	h := healthcheck.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	var status healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&status)

	if status.Checks != 0 {
		t.Errorf("expected 0 checks, got %d", status.Checks)
	}
}

func TestHealthHandler_IncrementChecksReflectedInResponse(t *testing.T) {
	h := healthcheck.New()
	h.IncrementChecks()
	h.IncrementChecks()
	h.IncrementChecks()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	var status healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&status)

	if status.Checks != 3 {
		t.Errorf("expected 3 checks, got %d", status.Checks)
	}
}

func TestHealthHandler_StartedAtIsRecent(t *testing.T) {
	before := time.Now().UTC()
	h := healthcheck.New()
	after := time.Now().UTC()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rec, req)

	var status healthcheck.Status
	_ = json.NewDecoder(rec.Body).Decode(&status)

	if status.StartedAt.Before(before) || status.StartedAt.After(after) {
		t.Errorf("started_at %v not within expected range [%v, %v]", status.StartedAt, before, after)
	}
}
