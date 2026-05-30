package startup

import (
	"fmt"
	"log"

	psql "github.com/StepanKomis/Ticketa/src/database/postgres"
)

func InitializeServer() error {
	log.Printf("Starting server...")
	log.Printf("Initializing Postgres connection...")

	err := psql.Init()
	if err != nil {
		return fmt.Errorf("Error initializing first Postgres connection: %s", err.Error())
	}

	log.Printf("Postgres connection successful.")

	return nil
}
