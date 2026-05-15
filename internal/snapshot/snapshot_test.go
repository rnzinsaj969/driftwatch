package snapshot_test

import (
	"os"
	"testing"

	"github.com/yourusername/driftwatch/internal/snapshot"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "snap-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestHash_Deterministic(t *testing.T) {
	path := writeTempFile(t, "hello world")

	h1, err := snapshot.Hash(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	h2, err := snapshot.Hash(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h1 != h2 {
		t.Errorf("expected identical hashes, got %q and %q", h1, h2)
	}
}

func TestHash_MissingFile(t *testing.T) {
	_, err := snapshot.Hash("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestHasDrifted_NoDriftOnFirstCall(t *testing.T) {
	path := writeTempFile(t, "initial content")
	store := snapshot.New()

	drifted, err := store.HasDrifted(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if drifted {
		t.Error("expected no drift on first call, got drift")
	}
}

func TestHasDrifted_DetectsDrift(t *testing.T) {
	path := writeTempFile(t, "original")
	store := snapshot.New()

	// Establish baseline.
	if _, err := store.HasDrifted(path); err != nil {
		t.Fatalf("baseline error: %v", err)
	}

	// Mutate the file.
	if err := os.WriteFile(path, []byte("modified"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	drifted, err := store.HasDrifted(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !drifted {
		t.Error("expected drift to be detected after file modification")
	}
}

func TestHasDrifted_NoDriftOnUnchangedFile(t *testing.T) {
	path := writeTempFile(t, "stable content")
	store := snapshot.New()

	for i := 0; i < 3; i++ {
		drifted, err := store.HasDrifted(path)
		if err != nil {
			t.Fatalf("iteration %d error: %v", i, err)
		}
		if drifted {
			t.Errorf("iteration %d: unexpected drift on unchanged file", i)
		}
	}
}
