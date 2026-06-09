package migrate

import "sort"

// Migration představuje jednu jednotku změny schématu. Up a Down obdrží nativní
// připojení backendu (např. *sql.DB pro Postgres) jako typ any.
type Migration struct {
	Number int
	Name   string
	Up     func(db any) error
	Down   func(db any) error
}

// Migrator je rozhraní, které musí implementovat každý databázový backend.
type Migrator interface {
	// Init inicializuje infrastrukturu pro sledování migrací a zaznamená migraci 0.
	Init() error
	// Applied vrátí nejvyšší zaznamenané číslo migrace.
	Applied() (int, error)
	// Up aplikuje migraci a zaznamená ji.
	Up(m Migration) error
	// Down vrátí zpět migraci a odstraní její záznam.
	Down(m Migration) error
}

// Runner spouští seřazené migrace vůči Migrator backendu.
type Runner struct {
	migrator   Migrator
	migrations []Migration
}

// NewRunner vytvoří Runner s migracemi seřazenými vzestupně podle Number.
func NewRunner(migrator Migrator, migrations []Migration) *Runner {
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Number < sorted[j].Number
	})
	return &Runner{migrator: migrator, migrations: sorted}
}
