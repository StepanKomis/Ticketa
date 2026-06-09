package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load načte YAML konfiguraci na path (pokud existuje) a překryje ji přes Defaults.
// Pokud soubor neexistuje, vrátí výchozí hodnoty bez chyby.
// Proměnná prostředí LOG_LEVEL přepisuje úroveň logování ze souboru.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if override := os.Getenv("LOG_LEVEL"); override != "" {
		cfg.Logging.Level = override
	}

	if len(cfg.TicketStatuses) < 3 {
		return nil, fmt.Errorf(
			"config: ticket_statuses must contain at least 3 entries (got %d); first=open, last=solved",
			len(cfg.TicketStatuses),
		)
	}

	return cfg, nil
}
