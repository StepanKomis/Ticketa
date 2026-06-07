package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads the YAML config at path (if present) on top of Defaults.
// If the file does not exist the defaults are returned without error.
// The LOG_LEVEL env var overrides the logging level from the file.
func Load(path string) (*Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
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
