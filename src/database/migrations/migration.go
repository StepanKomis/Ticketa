package migrate

import "sort"

// Migration represents a single schema change unit. Up and Down receive the
// backend's native connection (e.g. *sql.DB for Postgres) cast as any.
type Migration struct {
	Number int
	Name   string
	Up     func(db any) error
	Down   func(db any) error
}

// Migrator is the interface any database backend must satisfy.
type Migrator interface {
	// Init bootstraps migration tracking infrastructure and records migration 0.
	Init() error
	// Applied returns the highest migration number that has been recorded.
	Applied() (int, error)
	// Up applies a migration and records it.
	Up(m Migration) error
	// Down rolls back a migration and removes its record.
	Down(m Migration) error
}

// Runner executes ordered migrations against a Migrator backend.
type Runner struct {
	migrator   Migrator
	migrations []Migration
}

// NewRunner creates a Runner with migrations sorted ascending by Number.
func NewRunner(migrator Migrator, migrations []Migration) *Runner {
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Number < sorted[j].Number
	})
	return &Runner{migrator: migrator, migrations: sorted}
}
