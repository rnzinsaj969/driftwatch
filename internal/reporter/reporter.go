// Package reporter provides functionality for aggregating and reporting
// drift detection results across multiple watched paths.
package reporter

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/yourusername/driftwatch/internal/alert"
	"github.com/yourusername/driftwatch/internal/config"
)

// DriftEvent represents a single drift detection occurrence.
type DriftEvent struct {
	Path      string
	DetectedAt time.Time
	OldHash   string
	NewHash   string
}

// Reporter collects drift events and dispatches alerts.
type Reporter struct {
	cfg    *config.Config
	alerter *alert.Alerter
	mu     sync.Mutex
	events []DriftEvent
}

// New creates a new Reporter with the given config and alerter.
func New(cfg *config.Config, alerter *alert.Alerter) *Reporter {
	return &Reporter{
		cfg:     cfg,
		alerter: alerter,
		events:  make([]DriftEvent, 0),
	}
}

// Record stores a drift event and immediately dispatches an alert.
func (r *Reporter) Record(ctx context.Context, event DriftEvent) error {
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()

	log.Printf("[reporter] drift detected: path=%s old=%s new=%s",
		event.Path, event.OldHash, event.NewHash)

	return r.alerter.Send(ctx, alert.Payload{
		Path:      event.Path,
		OldHash:   event.OldHash,
		NewHash:   event.NewHash,
		Timestamp: event.DetectedAt,
	})
}

// Events returns a copy of all recorded drift events.
func (r *Reporter) Events() []DriftEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]DriftEvent, len(r.events))
	copy(out, r.events)
	return out
}

// Reset clears all recorded events.
func (r *Reporter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = r.events[:0]
}
