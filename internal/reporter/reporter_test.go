package reporter_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/alert"
	"github.com/yourusername/driftwatch/internal/config"
	"github.com/yourusername/driftwatch/internal/reporter"
)

func baseConfig(webhookURL string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.WebhookURL = webhookURL
	return cfg
}

func TestReporter_RecordStoresEvent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := baseConfig(server.URL)
	a := alert.New(cfg)
	rep := reporter.New(cfg, a)

	event := reporter.DriftEvent{
		Path:       "/etc/app.conf",
		DetectedAt: time.Now(),
		OldHash:    "abc123",
		NewHash:    "def456",
	}

	if err := rep.Record(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := rep.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Path != event.Path {
		t.Errorf("expected path %q, got %q", event.Path, events[0].Path)
	}
}

func TestReporter_AlertPayloadContentsCorrect(t *testing.T) {
	var received alert.Payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := baseConfig(server.URL)
	a := alert.New(cfg)
	rep := reporter.New(cfg, a)

	now := time.Now().UTC().Truncate(time.Second)
	err := rep.Record(context.Background(), reporter.DriftEvent{
		Path:       "/etc/hosts",
		DetectedAt: now,
		OldHash:    "old",
		NewHash:    "new",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Path != "/etc/hosts" {
		t.Errorf("expected path /etc/hosts, got %q", received.Path)
	}
	if received.OldHash != "old" {
		t.Errorf("expected old hash 'old', got %q", received.OldHash)
	}
	if received.NewHash != "new" {
		t.Errorf("expected new hash 'new', got %q", received.NewHash)
	}
}

func TestReporter_ResetClearsEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := baseConfig(server.URL)
	a := alert.New(cfg)
	rep := reporter.New(cfg, a)

	_ = rep.Record(context.Background(), reporter.DriftEvent{Path: "/a", DetectedAt: time.Now()})
	_ = rep.Record(context.Background(), reporter.DriftEvent{Path: "/b", DetectedAt: time.Now()})

	rep.Reset()

	if len(rep.Events()) != 0 {
		t.Errorf("expected 0 events after reset, got %d", len(rep.Events()))
	}
}
