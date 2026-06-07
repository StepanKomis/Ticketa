package startup

import (
	"fmt"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	migrate "github.com/StepanKomis/Ticketa/src/database/migrations"
	psql "github.com/StepanKomis/Ticketa/src/database/postgres"
	psqlmigrations "github.com/StepanKomis/Ticketa/src/database/postgres/migrations"
	"github.com/StepanKomis/Ticketa/src/www"
	"github.com/StepanKomis/Ticketa/src/www/router"
)

func InitializeServer(l *logs.Logger, cfg *config.Config) error {
	l.Info("Starting server...")
	l.Info("Initializing Postgres connection...")

	// Initializes Postgres connection
	err := psql.Init()
	if err != nil {
		return fmt.Errorf("Error initializing first Postgres connection: %s", err.Error())
	}

	db, err := psql.GetNewConnection()
	if err != nil {
		return fmt.Errorf(
			"Error during creation of new database connection whileinitializing the server: %s",
			err,
		)
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

	runner := migrate.NewRunner(migrator, psqlmigrations.All)

	if err := runner.MigrateUp(); err != nil {
		return fmt.Errorf("Error running migrations: %s", err.Error())
	}

	l.Info("Migrations complete.")

	port := env.Get("SERVER_PORT", "8080")
	addr := ":" + port

	mux := router.NewRouter(www.StaticFiles, db, cfg)

	l.Infof("Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("HTTP server error: %s", err.Error())
	}

	return nil
}
