package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full driftwatch daemon configuration.
type Config struct {
	WatchPaths  []string      `yaml:"watch_paths"`
	Interval    time.Duration `yaml:"interval"`
	Webhook     WebhookConfig `yaml:"webhook"`
	LogLevel    string        `yaml:"log_level"`
}

// WebhookConfig holds settings for the alert webhook.
type WebhookConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
}

// DefaultConfig returns a Config populated with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Interval: 30 * time.Second,
		LogLevel: "info",
		Webhook: WebhookConfig{
			Timeout: 10 * time.Second,
		},
	}
}

// Load reads a YAML config file from path and returns a validated Config.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if len(c.WatchPaths) == 0 {
		return fmt.Errorf("watch_paths must contain at least one path")
	}
	for _, p := range c.WatchPaths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("watch_paths entry %q is not accessible: %w", p, err)
		}
	}
	if c.Webhook.URL == "" {
		return fmt.Errorf("webhook.url is required")
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be a positive duration")
	}
	return nil
}
