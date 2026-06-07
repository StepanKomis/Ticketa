package config

import "testing"

func TestDefaults_MinimumStatuses(t *testing.T) {
	cfg := Defaults()
	if len(cfg.TicketStatuses) < 3 {
		t.Fatalf("expected at least 3 default statuses, got %d", len(cfg.TicketStatuses))
	}
}

func TestDefaults_FirstStatusIsOpen(t *testing.T) {
	cfg := Defaults()
	if cfg.TicketStatuses[0].Title != "Otevřeno" {
		t.Errorf("expected first status 'Otevřeno', got %q", cfg.TicketStatuses[0].Title)
	}
}

func TestDefaults_LastStatusIsSolved(t *testing.T) {
	cfg := Defaults()
	last := cfg.TicketStatuses[len(cfg.TicketStatuses)-1]
	if last.Title != "Vyřešeno" {
		t.Errorf("expected last status 'Vyřešeno', got %q", last.Title)
	}
}

func TestDefaults_LoggingLevel(t *testing.T) {
	cfg := Defaults()
	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level 'info', got %q", cfg.Logging.Level)
	}
}
