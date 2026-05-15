package alert_test

import (
	"fmt"
	"sync"

	"github.com/example/driftwatch/internal/alert"
)

// MockSender records calls to Send for use in other package tests.
type MockSender struct {
	mu       sync.Mutex
	Payloads []alert.Payload
	Err      error
}

// Send records the payload and returns the configured error.
func (m *MockSender) Send(p alert.Payload) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Payloads = append(m.Payloads, p)
	return m.Err
}

// CallCount returns the number of times Send was called.
func (m *MockSender) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Payloads)
}

// LastPayload returns the most recently recorded payload or an error if none.
func (m *MockSender) LastPayload() (alert.Payload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Payloads) == 0 {
		return alert.Payload{}, fmt.Errorf("mock: no payloads recorded")
	}
	return m.Payloads[len(m.Payloads)-1], nil
}

// Reset clears all recorded payloads and resets the configured error.
// This is useful when reusing a MockSender across multiple sub-tests.
func (m *MockSender) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Payloads = nil
	m.Err = nil
}
