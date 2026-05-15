package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/alert"
)

func TestSend_Success(t *testing.T) {
	var received alert.Payload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := alert.New(ts.URL)
	p := alert.Payload{
		Timestamp: time.Now().UTC(),
		FilePath:  "/etc/app/config.yaml",
		Message:   "drift detected",
		Checksum:  "abc123",
	}

	if err := s.Send(p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received.FilePath != p.FilePath {
		t.Errorf("file_path: want %q, got %q", p.FilePath, received.FilePath)
	}
	if received.Checksum != p.Checksum {
		t.Errorf("checksum: want %q, got %q", p.Checksum, received.Checksum)
	}
}

func TestSend_NonSuccessStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s := alert.New(ts.URL)
	err := s.Send(alert.Payload{FilePath: "/tmp/test", Message: "drift"})
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	s := alert.New("http://127.0.0.1:0/webhook")
	err := s.Send(alert.Payload{FilePath: "/tmp/test", Message: "drift"})
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
}

func TestSend_TimestampDefaultsToNow(t *testing.T) {
	var received alert.Payload
	before := time.Now().UTC()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	s := alert.New(ts.URL)
	// zero-value Timestamp — should be filled in by Send
	_ = s.Send(alert.Payload{FilePath: "/tmp/x", Message: "test"})

	if received.Timestamp.Before(before) {
		t.Errorf("expected timestamp >= %v, got %v", before, received.Timestamp)
	}
}
