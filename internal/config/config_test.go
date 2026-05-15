package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "driftwatch.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	raw := `
watch_paths:
  - /etc/app
  - /etc/nginx
interval: 1m
log_level: debug
webhook:
  url: https://hooks.example.com/alert
  timeout: 5s
  headers:
    Authorization: Bearer token123
`
	path := writeTempConfig(t, raw)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(cfg.WatchPaths) != 2 {
		t.Errorf("expected 2 watch paths, got %d", len(cfg.WatchPaths))
	}
	if cfg.Interval != time.Minute {
		t.Errorf("expected 1m interval, got %v", cfg.Interval)
	}
	if cfg.Webhook.URL != "https://hooks.example.com/alert" {
		t.Errorf("unexpected webhook URL: %s", cfg.Webhook.URL)
	}
}

func TestLoad_MissingWatchPaths(t *testing.T) {
	raw := `
webhook:
  url: https://hooks.example.com/alert
`
	path := writeTempConfig(t, raw)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing watch_paths")
	}
}

func TestLoad_MissingWebhookURL(t *testing.T) {
	raw := `
watch_paths:
  - /etc/app
`
	path := writeTempConfig(t, raw)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error for missing webhook.url")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/driftwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", cfg.Interval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got %s", cfg.LogLevel)
	}
}
