package startup

import (
	"fmt"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
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

	return nil
}
