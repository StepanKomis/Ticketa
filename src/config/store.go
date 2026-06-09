package config

import (
	"fmt"
	"sync"
)

// Store udržuje živou konfiguraci v paměti a synchronizuje ji s YAML souborem.
// Všechny metody jsou bezpečné pro souběžné použití.
type Store struct {
	mu   sync.RWMutex
	cfg  *Config
	path string
}

// NewStore obalí cfg (již načtenou) a cestu k souboru ze kterého pochází.
func NewStore(cfg *Config, path string) *Store {
	return &Store{cfg: cfg, path: path}
}

// Path vrátí cestu k souboru, do kterého Store persistuje.
func (s *Store) Path() string { return s.path }

// Get vrátí kopii aktuální konfigurace. Volající nesmí vrácenou hodnotu měnit.
func (s *Store) Get() *Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cp := *s.cfg
	return &cp
}

// Update zavolá fn s kopií aktuální konfigurace. Pokud fn vrátí nil, výsledek
// je zapsán na disk a atomicky nahradí in-memory konfiguraci.
// Pokud fn nebo zápis souboru selžou, in-memory konfigurace zůstane nezměněna.
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
