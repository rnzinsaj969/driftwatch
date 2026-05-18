package reporter

import "time"

// EventKind describes the category of a drift event.
type EventKind string

const (
	// KindDrift indicates that a monitored file's content has changed unexpectedly.
	KindDrift EventKind = "drift"

	// KindError indicates that a file could not be read or hashed during a check.
	KindError EventKind = "error"
)

// Event represents a single drift or error occurrence captured by the Reporter.
type Event struct {
	// Kind classifies the event as a drift detection or a read/hash error.
	Kind EventKind `json:"kind"`

	// Path is the filesystem path of the monitored file that triggered the event.
	Path string `json:"path"`

	// OldHash is the SHA-256 hex digest recorded before the change was detected.
	// Empty string when Kind is KindError.
	OldHash string `json:"old_hash,omitempty"`

	// NewHash is the SHA-256 hex digest recorded at the time of detection.
	// Empty string when Kind is KindError.
	NewHash string `json:"new_hash,omitempty"`

	// Message holds a human-readable description, used primarily for KindError events.
	Message string `json:"message,omitempty"`

	// DetectedAt is the UTC timestamp when the event was recorded.
	DetectedAt time.Time `json:"detected_at"`
}

// newDriftEvent constructs an Event of KindDrift.
func newDriftEvent(path, oldHash, newHash string) Event {
	return Event{
		Kind:       KindDrift,
		Path:       path,
		OldHash:    oldHash,
		NewHash:    newHash,
		DetectedAt: time.Now().UTC(),
	}
}

// newErrorEvent constructs an Event of KindError.
func newErrorEvent(path, message string) Event {
	return Event{
		Kind:       KindError,
		Path:       path,
		Message:    message,
		DetectedAt: time.Now().UTC(),
	}
}
