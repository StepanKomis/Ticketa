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

// PsqlMigrator implements migrate.Migrator for PostgreSQL.
type PsqlMigrator struct {
	db *sql.DB
}

func NewMigrator() (*PsqlMigrator, error) {
	db, err := GetNewConnection()
	if err != nil {
		return nil, fmt.Errorf("migrator: %w", err)
	}
	return &PsqlMigrator{db: db}, nil
}

// Init creates the migrations table and records migration 0 (the table creation itself).
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

// Applied returns the highest recorded migration number, or -1 if the table is empty.
func (p *PsqlMigrator) Applied() (int, error) {
	var n int
	err := p.db.QueryRow(`SELECT COALESCE(MAX(number), -1) FROM migrations`).Scan(&n)
	if err != nil {
		return -1, fmt.Errorf("querying applied migrations: %w", err)
	}
	return n, nil
}

// Up runs a migration's Up function then records it in the migrations table.
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

// Down runs a migration's Down function then removes its record from the migrations table.
func (p *PsqlMigrator) Down(m migrate.Migration) error {
	if err := m.Down(p.db); err != nil {
		return err
	}
	_, err := p.db.Exec(`DELETE FROM migrations WHERE number = $1`, m.Number)
	return err
}
