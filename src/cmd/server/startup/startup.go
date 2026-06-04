package startup

import (
	"fmt"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	migrate "github.com/StepanKomis/Ticketa/src/database/migrations"
	psql "github.com/StepanKomis/Ticketa/src/database/postgres"
)

func InitializeServer(l *logs.Logger) error {
	l.Info("Starting server...")
	l.Info("Initializing Postgres connection...")

	err := psql.Init()
	if err != nil {
		return fmt.Errorf("Error initializing first Postgres connection: %s", err.Error())
	}

	l.Info("Postgres connection successful.")

	l.Info("Initializing migrations...")

	migrator, err := psql.NewMigrator()
	if err != nil {
		return fmt.Errorf("Error creating migrator: %s", err.Error())
	}

	if err := migrator.Init(); err != nil {
		return fmt.Errorf("Error initializing migrations: %s", err.Error())
	}

	runner := migrate.NewRunner(migrator, []migrate.Migration{})

	if err := runner.MigrateUp(); err != nil {
		return fmt.Errorf("Error running migrations: %s", err.Error())
	}

	l.Info("Migrations complete.")

	return nil
}
