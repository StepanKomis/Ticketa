package migrate

import "fmt"

// MigrateDown rolls back the most recently applied migration.
func (r *Runner) MigrateDown() error {
	applied, err := r.migrator.Applied()
	if err != nil {
		return fmt.Errorf("fetching applied migration number: %w", err)
	}

	for i := len(r.migrations) - 1; i >= 0; i-- {
		m := r.migrations[i]
		if m.Number == applied {
			return r.migrator.Down(m)
		}
	}

	return nil
}
