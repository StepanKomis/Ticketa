package psql

import (
	"database/sql"
	"fmt"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
	_ "github.com/lib/pq"
)

// returns sql connection string for postgrese instance, takes values from .env file
func getPsqlString() (string, error) {
	user, err := env.GetNeeded("PG_USER")
	if err != nil {
		return "", fmt.Errorf("Environment variable PG_USER: %s", err.Error())
	}

	passwd, err := env.GetNeeded("PG_PASSWORD")
	if err != nil {
		return "", fmt.Errorf("Environment variable PG_PASSWORD: %s", err.Error())
	}

	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		env.Get("PG_HOST", "database"), env.Get("PG_PORT", "5432"), user, passwd, env.Get("PG_DATABASE", "ticketa")), nil
}

// Returns *sql.DB connection to postgrese instance and error when something goes wrong
func GetNewConnection() (*sql.DB, error) {
	psqlInfo, err := getPsqlString()
	if err != nil {
		return nil, fmt.Errorf("Error creating database connection string: %s", err.Error())
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("Error when trying to connect to postgrese: %s", err.Error())
	}

	return db, nil
}

// Initializes first connectiont to postgrese and tests if the database is reachable
func Init() error {
	db, err := GetNewConnection()
	if err != nil {
		return fmt.Errorf("Error when trying to initialize first postgrese connection: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("Error pinging database after first connection: %s", err.Error())
	}

	db.Close()

	return nil
}
