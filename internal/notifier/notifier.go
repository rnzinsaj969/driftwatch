// Package notifier orchestrates drift detection results and dispatches
// alerts through the configured alert sender.
package notifier

import (
	"context"
	"log"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/reporter"
)

// Sender is the interface used to dispatch an alert payload.
type Sender interface {
	Send(ctx context.Context, payload alert.Payload) error
}

// Notifier receives drift events from the reporter and sends alerts.
type Notifier struct {
	reporter *reporter.Reporter
	sender  Sender
	log     *log.Logger
}

// New creates a Notifier that reads events from r and dispatches via s.
func New(r *reporter.Reporter, s Sender, l *log.Logger) *Notifier {
	return &Notifier{
		reporter: r,
		sender:   s,
		log:      l,
	}
}

// Flush reads all pending events from the reporter, sends an alert for
// each one, then resets the reporter. It is safe to call concurrently.
func (n *Notifier) Flush(ctx context.Context) error {
	events := n.reporter.Events()
	if len(events) == 0 {
		return nil
	}

	for _, e := range events {
		payload := alert.Payload{
			Path:      e.Path,
			OldHash:   e.OldHash,
			NewHash:   e.NewHash,
			Timestamp: e.Timestamp,
			Message:   e.Message,
		}
		if err := n.sender.Send(ctx, payload); err != nil {
			n.log.Printf("notifier: failed to send alert for %s: %v", e.Path, err)
			return err
		}
		n.log.Printf("notifier: alert sent for %s", e.Path)
	}

	n.reporter.Reset()
	return nil
}
