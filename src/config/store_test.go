package config

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ticketa.yaml")
	cfg := Defaults()
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	return NewStore(cfg, path)
}

func TestStore_GetReturnsCopy(t *testing.T) {
	s := newTestStore(t)
	a := s.Get()
	b := s.Get()
	if a == b {
		t.Error("Get should return distinct copies, not the same pointer")
	}
}

func TestStore_Update_PersistsToFile(t *testing.T) {
	s := newTestStore(t)

	err := s.Update(func(c *Config) error {
		c.Logging.Level = "debug"
		return nil
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	loaded, err := Load(s.path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Logging.Level != "debug" {
		t.Errorf("expected persisted level 'debug', got %q", loaded.Logging.Level)
	}
}

func TestStore_Update_FailureLeavesPreviousConfig(t *testing.T) {
	s := newTestStore(t)
	original := s.Get().Logging.Level

	err := s.Update(func(c *Config) error {
		c.Logging.Level = "error"
		return errors.New("intentional failure")
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if s.Get().Logging.Level != original {
		t.Errorf("config should be unchanged after failed update, got %q", s.Get().Logging.Level)
	}
}

func TestStore_Update_RejectsFewerThanThreeStatuses(t *testing.T) {
	s := newTestStore(t)

	err := s.Update(func(c *Config) error {
		c.TicketStatuses = c.TicketStatuses[:2]
		return nil
	})
	if err == nil {
		t.Fatal("expected error for fewer than 3 statuses, got nil")
	}
}

func TestStore_ConcurrentGetAndUpdate(t *testing.T) {
	s := newTestStore(t)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = s.Get()
		}()
		go func() {
			defer wg.Done()
			_ = s.Update(func(c *Config) error {
				c.Logging.Level = "info"
				return nil
			})
		}()
	}
	wg.Wait()
}
