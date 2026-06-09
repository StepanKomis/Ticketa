package migrate

import "fmt"

// MigrateDown vrátí zpět naposledy aplikovanou migraci.
func (r *Runner) MigrateDown() error {
	applied, err := r.migrator.Applied()
	if err != nil {
		return fmt.Errorf("načítání čísla aplikované migrace: %w", err)
	}

	for i := len(r.migrations) - 1; i >= 0; i-- {
		m := r.migrations[i]
		if m.Number == applied {
			return r.migrator.Down(m)
		}
	}

	return nil
}
