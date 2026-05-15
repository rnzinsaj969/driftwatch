package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/scheduler"
)

func baseConfig(interval int, paths []string) *config.Config {
	cfg := config.DefaultConfig()
	cfg.CheckIntervalSeconds = interval
	cfg.WatchPaths = paths
	cfg.WebhookURL = "http://example.com/hook"
	return cfg
}

func TestScheduler_InvokeCheckFnForEachPath(t *testing.T) {
	paths := []string{"/etc/a.conf", "/etc/b.conf"}
	cfg := baseConfig(1, paths)

	var callCount atomic.Int64
	calledPaths := make(chan string, 10)

	checkFn := func(path string) error {
		callCount.Add(1)
		calledPaths <- path
		return nil
	}

	sched := scheduler.New(cfg, checkFn)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	go sched.Start(ctx)

	<-ctx.Done()

	if callCount.Load() < int64(len(paths)) {
		t.Errorf("expected at least %d calls, got %d", len(paths), callCount.Load())
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	cfg := baseConfig(1, []string{"/etc/test.conf"})

	var callCount atomic.Int64
	checkFn := func(path string) error {
		callCount.Add(1)
		return nil
	}

	sched := scheduler.New(cfg, checkFn)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		sched.Start(ctx)
		close(done)
	}()

	time.Sleep(1200 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// scheduler exited cleanly
	case <-time.After(500 * time.Millisecond):
		t.Error("scheduler did not stop after context cancellation")
	}

	snap := callCount.Load()
	time.Sleep(600 * time.Millisecond)
	if callCount.Load() != snap {
		t.Error("scheduler continued running after context was cancelled")
	}
}

func TestScheduler_CheckFnErrorDoesNotPanic(t *testing.T) {
	cfg := baseConfig(1, []string{"/nonexistent/path.conf"})

	checkFn := func(path string) error {
		return fmt.Errorf("simulated error for %s", path)
	}

	sched := scheduler.New(cfg, checkFn)
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	// Should not panic
	go sched.Start(ctx)
	<-ctx.Done()
}
