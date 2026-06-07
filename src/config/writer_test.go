package config

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestSave_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticketa.yaml")

	original := Defaults()
	original.Logging.Level = "debug"
	original.TicketStatuses = append(original.TicketStatuses, StatusConfig{
		Title: "Pozastaveno",
		Color: "#999999",
	})

	if err := Save(path, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}

	if loaded.Logging.Level != "debug" {
		t.Errorf("expected level 'debug', got %q", loaded.Logging.Level)
	}
	if len(loaded.TicketStatuses) != len(original.TicketStatuses) {
		t.Errorf("expected %d statuses, got %d", len(original.TicketStatuses), len(loaded.TicketStatuses))
	}
	if loaded.TicketStatuses[3].Title != "Pozastaveno" {
		t.Errorf("expected extra status 'Pozastaveno', got %q", loaded.TicketStatuses[3].Title)
	}
}

func TestSave_ConcurrentWritesDoNotCorrupt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ticketa.yaml")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cfg := Defaults()
			if err := Save(path, cfg); err != nil {
				t.Errorf("concurrent Save: %v", err)
			}
		}()
	}
	wg.Wait()

	if _, err := Load(path); err != nil {
		t.Fatalf("Load after concurrent saves: %v", err)
	}
}
