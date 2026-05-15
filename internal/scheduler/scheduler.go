// Package scheduler provides a periodic tick-based runner that drives
// the drift-check loop for each watched path.
package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/example/driftwatch/internal/config"
)

// CheckFunc is called on every tick for a given file path.
type CheckFunc func(path string) error

// Scheduler drives periodic drift checks.
type Scheduler struct {
	cfg      *config.Config
	checkFn  CheckFunc
	ticker   *time.Ticker
	done     chan struct{}
}

// New creates a Scheduler that will call checkFn for every watched path
// at the interval defined in cfg.
func New(cfg *config.Config, checkFn CheckFunc) *Scheduler {
	return &Scheduler{
		cfg:     cfg,
		checkFn: checkFn,
		done:    make(chan struct{}),
	}
}

// Start begins the scheduling loop and blocks until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	interval := time.Duration(s.cfg.CheckIntervalSeconds) * time.Second
	s.ticker = time.NewTicker(interval)
	defer s.ticker.Stop()

	log.Printf("scheduler: starting with interval %s for %d path(s)",
		interval, len(s.cfg.WatchPaths))

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler: context cancelled, stopping")
			return
		case <-s.ticker.C:
			s.runChecks()
		}
	}
}

// runChecks iterates over all watched paths and invokes the check function.
// Errors are logged but do not halt checks for remaining paths.
func (s *Scheduler) runChecks() {
	var errCount int
	for _, path := range s.cfg.WatchPaths {
		if err := s.checkFn(path); err != nil {
			log.Printf("scheduler: check error for %q: %v", path, err)
			errCount++
		}
	}
	if errCount > 0 {
		log.Printf("scheduler: %d/%d check(s) failed in this cycle",
			errCount, len(s.cfg.WatchPaths))
	}
}
