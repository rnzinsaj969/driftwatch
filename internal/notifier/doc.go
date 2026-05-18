// Package notifier bridges the reporter and the alert sender.
//
// After each scheduler tick the watcher records drift events via the
// reporter. The notifier's Flush method reads those pending events,
// dispatches an alert payload for every event through the configured
// Sender, and then resets the reporter so the next tick starts clean.
//
// Usage:
//
//	n := notifier.New(rep, alertClient, logger)
//	if err := n.Flush(ctx); err != nil {
//		// handle send failure
//	}
package notifier
