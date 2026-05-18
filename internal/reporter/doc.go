// Package reporter provides drift event recording and alert dispatching.
//
// The Reporter collects drift events detected by the watcher and forwards
// them to the alert package for webhook delivery. It maintains an in-memory
// log of recent events that can be queried by the health check endpoint.
//
// Usage:
//
//	cfg := config.DefaultConfig()
//	alerter := alert.New(cfg)
//	r := reporter.New(cfg, alerter)
//	r.Record(ctx, path, oldHash, newHash)
package reporter
