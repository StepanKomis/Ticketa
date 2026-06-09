package psql

import (
	"database/sql"
	"fmt"

	migrate "github.com/StepanKomis/Ticketa/src/database/migrations"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS migrations (
	number     INT         PRIMARY KEY,
	name       TEXT        NOT NULL,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

// PsqlMigrator implementuje migrate.Migrator pro PostgreSQL.
type PsqlMigrator struct {
	db *sql.DB
}

func NewMigrator() (*PsqlMigrator, error) {
	db, err := GetNewConnection()
	if err != nil {
		return nil, fmt.Errorf("migrátor: %w", err)
	}
	return &PsqlMigrator{db: db}, nil
}

// Init vytvoří tabulku migrací a zaznamená migraci 0 (samotné vytvoření tabulky).
func (p *PsqlMigrator) Init() error {
	_, err := p.db.Exec(createMigrationsTable)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	_, err = p.db.Exec(
		`INSERT INTO migrations (number, name) VALUES (0, 'create_migrations_table') ON CONFLICT (number) DO NOTHING`,
	)
	if err != nil {
		return fmt.Errorf("recording migration 0: %w", err)
	}

	return nil
}

// Applied vrátí nejvyšší zaznamenaný číslo migrace, nebo -1 pokud je tabulka prázdná.
func (p *PsqlMigrator) Applied() (int, error) {
	var n int
	err := p.db.QueryRow(`SELECT COALESCE(MAX(number), -1) FROM migrations`).Scan(&n)
	if err != nil {
		return -1, fmt.Errorf("querying applied migrations: %w", err)
	}
	return n, nil
}

// Up spustí Up funkci migrace a zaznamená ji do tabulky migrací.
func (p *PsqlMigrator) Up(m migrate.Migration) error {
	if err := m.Up(p.db); err != nil {
		return err
	}
	_, err := p.db.Exec(
		`INSERT INTO migrations (number, name) VALUES ($1, $2)`,
		m.Number, m.Name,
	)
	return err
}

// Down spustí Down funkci migrace a odstraní její záznam z tabulky migrací.
func (p *PsqlMigrator) Down(m migrate.Migration) error {
	if err := m.Down(p.db); err != nil {
		return err
	}
	_, err := p.db.Exec(`DELETE FROM migrations WHERE number = $1`, m.Number)
	return err
}
