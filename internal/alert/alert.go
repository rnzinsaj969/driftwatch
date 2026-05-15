package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Payload represents the webhook alert body sent on drift detection.
type Payload struct {
	Timestamp time.Time `json:"timestamp"`
	FilePath  string    `json:"file_path"`
	Message   string    `json:"message"`
	Checksum  string    `json:"checksum"`
}

// Sender sends drift alerts to a configured webhook URL.
type Sender struct {
	WebhookURL string
	Client     *http.Client
}

// New creates a new Sender with the given webhook URL.
func New(webhookURL string) *Sender {
	return &Sender{
		WebhookURL: webhookURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send marshals the payload and posts it to the webhook URL.
func (s *Sender) Send(p Payload) error {
	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now().UTC()
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("alert: marshal payload: %w", err)
	}

	resp, err := s.Client.Post(s.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("alert: post webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: unexpected status from webhook: %d", resp.StatusCode)
	}

	return nil
}
