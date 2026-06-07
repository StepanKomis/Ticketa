package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ticketa-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_MissingFile_ReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("expected default log level 'info', got %q", cfg.Logging.Level)
	}
	if len(cfg.TicketStatuses) < 3 {
		t.Errorf("expected at least 3 default statuses, got %d", len(cfg.TicketStatuses))
	}
}

func TestLoad_ValidFile_OverridesDefaults(t *testing.T) {
	yaml := `
logging:
  level: debug
  dir: /tmp/logs
ticket_statuses:
  - title: "Nový"
    color: "#aabbcc"
  - title: "Zpracovává se"
    color: "#ddeeff"
  - title: "Uzavřeno"
    color: "#112233"
`
	path := writeTempConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected level 'debug', got %q", cfg.Logging.Level)
	}
	if cfg.Logging.Dir != "/tmp/logs" {
		t.Errorf("expected dir '/tmp/logs', got %q", cfg.Logging.Dir)
	}
	if cfg.TicketStatuses[0].Title != "Nový" {
		t.Errorf("expected first status 'Nový', got %q", cfg.TicketStatuses[0].Title)
	}
}

func TestLoad_LogLevelEnvOverridesYAML(t *testing.T) {
	t.Setenv("LOG_LEVEL", "warn")
	yaml := `
logging:
  level: debug
ticket_statuses:
  - title: "A"
    color: "#111111"
  - title: "B"
    color: "#222222"
  - title: "C"
    color: "#333333"
`
	path := writeTempConfig(t, yaml)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("expected env override 'warn', got %q", cfg.Logging.Level)
	}
}

func TestLoad_FewerThanThreeStatuses_ReturnsError(t *testing.T) {
	yaml := `
ticket_statuses:
  - title: "Jediný"
    color: "#ffffff"
  - title: "Druhý"
    color: "#000000"
`
	path := writeTempConfig(t, yaml)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for fewer than 3 statuses, got nil")
	}
}
