// Package main is the entry point for the driftwatch daemon.
// It wires together configuration, scheduling, snapshotting, and alerting
// to monitor infrastructure config files for unexpected drift.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/driftwatch/internal/alert"
	"github.com/yourusername/driftwatch/internal/config"
	"github.com/yourusername/driftwatch/internal/reporter"
	"github.com/yourusername/driftwatch/internal/scheduler"
	"github.com/yourusername/driftwatch/internal/snapshot"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to driftwatch config file")
	flag.Parse()

	// Load configuration from the specified file.
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("driftwatch starting — watching %d path(s) every %s",
		len(cfg.WatchPaths), cfg.Interval)

	// Build core dependencies.
	alerter := alert.New(cfg)
	snap := snapshot.New()
	rep := reporter.New(cfg, snap, alerter)

	// Build the scheduler with the reporter's check function.
	sched := scheduler.New(cfg, rep.Check)

	// Handle OS signals for graceful shutdown.
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Start the scheduler; it blocks until the context is cancelled.
	if err := sched.Start(ctx); err != nil {
		log.Fatalf("scheduler exited with error: %v", err)
	}

	log.Println("driftwatch stopped")
}
