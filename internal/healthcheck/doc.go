// Package healthcheck exposes a lightweight HTTP handler that reports the
// operational status of the driftwatch daemon.
//
// The handler responds to GET requests with a JSON payload containing:
//   - healthy: always true while the process is running
//   - started_at: the UTC timestamp when the daemon started
//   - checks_performed: the cumulative number of drift check cycles completed
//
// Usage:
//
//	h := healthcheck.New()
//	http.Handle("/health", h)
//
// Call h.IncrementChecks() after each completed scheduler tick to keep
// the counter accurate.
package healthcheck
