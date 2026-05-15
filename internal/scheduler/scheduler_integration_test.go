package scheduler_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/scheduler"
)

// TestScheduler_MultipleTicksAccumulateCalls verifies that the scheduler fires
// on every tick, not just the first one.
func TestScheduler_MultipleTicksAccumulateCalls(t *testing.T) {
	paths := []string{"/etc/one.conf"}
	cfg := baseConfig(1, paths)

	var mu sync.Mutex
	var recorded []string

	checkFn := func(path string) error {
		mu.Lock()
		recorded = append(recorded, path)
		mu.Unlock()
		return nil
	}

	sched := scheduler.New(cfg, checkFn)
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	go sched.Start(ctx)
	<-ctx.Done()

	mu.Lock()
	count := len(recorded)
	mu.Unlock()

	if count < 2 {
		t.Errorf("expected at least 2 ticks, got %d", count)
	}
}

// TestScheduler_AllPathsCheckedEachTick verifies every path is visited per tick.
func TestScheduler_AllPathsCheckedEachTick(t *testing.T) {
	paths := []string{"/a", "/b", "/c"}
	cfg := &config.Config{
		WatchPaths:           paths,
		WebhookURL:           "http://example.com",
		CheckIntervalSeconds: 1,
	}

	seen := make(map[string]int)
	var mu sync.Mutex

	checkFn := func(path string) error {
		mu.Lock()
		seen[path]++
		mu.Unlock()
		return nil
	}

	sched := scheduler.New(cfg, checkFn)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	go sched.Start(ctx)
	<-ctx.Done()

	mu.Lock()
	defer mu.Unlock()
	for _, p := range paths {
		if seen[p] == 0 {
			t.Errorf("path %q was never checked", p)
		}
	}
	_ = fmt.Sprintf // keep import used by other test file
}
