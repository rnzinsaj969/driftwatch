package watcher_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/driftwatch/internal/config"
	"github.com/user/driftwatch/internal/watcher"
)

func tempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "drift-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func baseConfig(paths []string) *config.Config {
	return &config.Config{
		WatchPaths:          paths,
		WebhookURL:          "http://example.com/hook",
		PollIntervalSeconds: 1,
	}
}

func TestWatcher_NoDriftOnUnchangedFile(t *testing.T) {
	path := tempFile(t, "key: value\n")
	cfg := baseConfig([]string{path})

	w := watcher.New(cfg)
	if err := w.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	select {
	case evt := <-w.Events:
		t.Errorf("unexpected drift event: %+v", evt)
	case <-time.After(2500 * time.Millisecond):
		// no event expected — pass
	}
}

func TestWatcher_DetectsDrift(t *testing.T) {
	path := tempFile(t, "key: original\n")
	cfg := baseConfig([]string{path})

	w := watcher.New(cfg)
	if err := w.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer w.Stop()

	// Modify the file after watcher has started.
	time.Sleep(300 * time.Millisecond)
	if err := os.WriteFile(path, []byte("key: changed\n"), 0644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	select {
	case evt := <-w.Events:
		if evt.Path != path {
			t.Errorf("expected path %s, got %s", path, evt.Path)
		}
		if evt.OldHash == evt.NewHash {
			t.Error("expected old and new hashes to differ")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for drift event")
	}
}

func TestWatcher_StartFailsOnMissingFile(t *testing.T) {
	cfg := baseConfig([]string{"/nonexistent/path/config.yaml"})
	w := watcher.New(cfg)
	if err := w.Start(); err == nil {
		t.Error("expected error for missing file, got nil")
		w.Stop()
	}
}
