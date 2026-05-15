// Package snapshot provides utilities for computing and comparing
// file content hashes to detect configuration drift.
package snapshot

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
)

// Store holds a thread-safe map of file paths to their last known SHA-256 hash.
type Store struct {
	mu     sync.RWMutex
	hashes map[string]string
}

// New returns an initialised, empty Store.
func New() *Store {
	return &Store{
		hashes: make(map[string]string),
	}
}

// Hash computes the SHA-256 hex digest of the file at path.
func Hash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("snapshot: open %q: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("snapshot: hash %q: %w", path, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Record stores the current hash for path, overwriting any previous value.
func (s *Store) Record(path, hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hashes[path] = hash
}

// Get returns the stored hash for path and whether it was found.
func (s *Store) Get(path string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.hashes[path]
	return h, ok
}

// HasDrifted returns true when the file at path has a different hash from the
// one previously recorded. If no baseline exists the hash is recorded and
// HasDrifted returns false.
func (s *Store) HasDrifted(path string) (bool, error) {
	current, err := Hash(path)
	if err != nil {
		return false, err
	}

	previous, known := s.Get(path)
	if !known {
		s.Record(path, current)
		return false, nil
	}

	if current != previous {
		s.Record(path, current)
		return true, nil
	}
	return false, nil
}
