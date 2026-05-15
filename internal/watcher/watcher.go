package watcher

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/user/driftwatch/internal/config"
)

// FileState holds the last known hash of a watched file.
type FileState struct {
	Path    string
	Hash    string
	ModTime time.Time
}

// DriftEvent is emitted when a file's content has changed unexpectedly.
type DriftEvent struct {
	Path     string
	OldHash  string
	NewHash  string
	Detected time.Time
}

// Watcher monitors a set of files for content drift.
type Watcher struct {
	cfg      *config.Config
	states   map[string]FileState
	mu       sync.RWMutex
	Events   chan DriftEvent
	stopCh   chan struct{}
}

// New creates a new Watcher from the provided config.
func New(cfg *config.Config) *Watcher {
	return &Watcher{
		cfg:    cfg,
		states: make(map[string]FileState),
		Events: make(chan DriftEvent, 16),
		stopCh: make(chan struct{}),
	}
}

// Start initialises baseline hashes and begins polling.
func (w *Watcher) Start() error {
	for _, path := range w.cfg.WatchPaths {
		hash, modTime, err := hashFile(path)
		if err != nil {
			return fmt.Errorf("watcher: baseline hash for %s: %w", path, err)
		}
		w.states[path] = FileState{Path: path, Hash: hash, ModTime: modTime}
	}

	interval := time.Duration(w.cfg.PollIntervalSeconds) * time.Second
	go w.poll(interval)
	return nil
}

// Stop signals the polling goroutine to exit.
func (w *Watcher) Stop() {
	close(w.stopCh)
}

func (w *Watcher) poll(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			w.checkAll()
		case <-w.stopCh:
			return
		}
	}
}

func (w *Watcher) checkAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for path, state := range w.states {
		newHash, modTime, err := hashFile(path)
		if err != nil {
			continue
		}
		if newHash != state.Hash {
			w.Events <- DriftEvent{
				Path:     path,
				OldHash:  state.Hash,
				NewHash:  newHash,
				Detected: time.Now(),
			}
			w.states[path] = FileState{Path: path, Hash: newHash, ModTime: modTime}
		}
	}
}

func hashFile(path string) (string, time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", time.Time{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", time.Time{}, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", time.Time{}, err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), info.ModTime(), nil
}
