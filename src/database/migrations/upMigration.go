package migrate

import "fmt"

// MigrateUp applies all migrations with Number greater than the currently applied one.
func (r *Runner) MigrateUp() error {
	applied, err := r.migrator.Applied()
	if err != nil {
		return fmt.Errorf("fetching applied migration number: %w", err)
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
