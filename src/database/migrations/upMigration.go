package migrate

import "fmt"

// MigrateUp aplikuje všechny migrace s číslem vyšším než aktuálně aplikovaná.
func (r *Runner) MigrateUp() error {
	applied, err := r.migrator.Applied()
	if err != nil {
		return fmt.Errorf("načítání čísla aplikované migrace: %w", err)
	}

	for _, m := range r.migrations {
		if m.Number <= applied {
			continue
		}
		if err := r.migrator.Up(m); err != nil {
			return fmt.Errorf("migration %d (%s): %w", m.Number, m.Name, err)
		}
	}

	return nil
}
