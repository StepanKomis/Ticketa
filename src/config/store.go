package config

import (
	"fmt"
	"sync"
)

// Store holds the live config in memory and keeps it in sync with the YAML file.
// All methods are safe for concurrent use.
type Store struct {
	mu   sync.RWMutex
	cfg  *Config
	path string
}

// NewStore wraps cfg (already loaded) and the file path it came from.
func NewStore(cfg *Config, path string) *Store {
	return &Store{cfg: cfg, path: path}
}

// Get returns a copy of the current config. Callers must not mutate the returned value.
func (s *Store) Get() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := *s.cfg
	return &cp
}

// Update calls fn with a copy of the current config. If fn returns nil, the
// result is written to disk and replaces the in-memory config atomically.
// If fn or the file write fails, the in-memory config is unchanged.
func (s *Store) Update(fn func(*Config) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	draft := *s.cfg
	if err := fn(&draft); err != nil {
		return fmt.Errorf("config update: %w", err)
	}

	if len(draft.TicketStatuses) < 3 {
		return fmt.Errorf("config update: ticket_statuses must contain at least 3 entries")
	}

	if err := Save(s.path, &draft); err != nil {
		return fmt.Errorf("config update: persisting to disk: %w", err)
	}

	s.cfg = &draft
	return nil
}
