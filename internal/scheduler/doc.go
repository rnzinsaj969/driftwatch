// Package scheduler implements a tick-based loop that periodically triggers
// drift checks across all configured watch paths.
//
// Usage:
//
//	sched := scheduler.New(cfg, func(path string) error {
//		// perform drift check for path
//		return nil
//	})
//	sched.Start(ctx) // blocks until ctx is cancelled
//
// The check interval is read from config.Config.CheckIntervalSeconds.
// Each tick causes checkFn to be called once per watched path.
// Errors returned by checkFn are logged but do not stop the scheduler.
package scheduler
