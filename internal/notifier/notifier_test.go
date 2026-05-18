package notifier_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/config"
	"github.com/example/driftwatch/internal/notifier"
	"github.com/example/driftwatch/internal/reporter"
)

func baseConfig() *config.Config {
	return &config.Config{
		WebhookURL:    "http://example.com/hook",
		WatchPaths:    []string{"/etc/app/config.yaml"},
		CheckInterval: 10,
	}
}

type mockSender struct {
	called   int
	payloads []alert.Payload
	err      error
}

func (m *mockSender) Send(_ context.Context, p alert.Payload) error {
	m.called++
	m.payloads = append(m.payloads, p)
	return m.err
}

func testLogger() *log.Logger {
	return log.New(os.Stderr, "test: ", 0)
}

func TestFlush_NoEventsDoesNotCallSender(t *testing.T) {
	rep := reporter.New(baseConfig())
	sender := &mockSender{}
	n := notifier.New(rep, sender, testLogger())

	if err := n.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.called != 0 {
		t.Errorf("expected 0 calls, got %d", sender.called)
	}
}

func TestFlush_SendsAlertForEachEvent(t *testing.T) {
	rep := reporter.New(baseConfig())
	rep.Record("/etc/app/config.yaml", "aaa", "bbb", time.Now())
	rep.Record("/etc/app/other.yaml", "ccc", "ddd", time.Now())

	sender := &mockSender{}
	n := notifier.New(rep, sender, testLogger())

	if err := n.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.called != 2 {
		t.Errorf("expected 2 calls, got %d", sender.called)
	}
}

func TestFlush_ResetsReporterAfterSuccess(t *testing.T) {
	rep := reporter.New(baseConfig())
	rep.Record("/etc/app/config.yaml", "aaa", "bbb", time.Now())

	sender := &mockSender{}
	n := notifier.New(rep, sender, testLogger())

	_ = n.Flush(context.Background())

	if len(rep.Events()) != 0 {
		t.Errorf("expected reporter to be reset, got %d events", len(rep.Events()))
	}
}

func TestFlush_ReturnsErrorOnSendFailure(t *testing.T) {
	rep := reporter.New(baseConfig())
	rep.Record("/etc/app/config.yaml", "aaa", "bbb", time.Now())

	sender := &mockSender{err: errors.New("webhook down")}
	n := notifier.New(rep, sender, testLogger())

	if err := n.Flush(context.Background()); err == nil {
		t.Fatal("expected error but got nil")
	}
}
