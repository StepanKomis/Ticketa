package psql

import (
	"database/sql"
	"fmt"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
	_ "github.com/lib/pq"
)

// getPsqlString sestaví připojovací řetězec pro Postgres z proměnných prostředí.
func getPsqlString() (string, error) {
	user, err := env.GetNeeded("PG_USER")
	if err != nil {
		return "", fmt.Errorf("proměnná prostředí PG_USER: %s", err.Error())
	}

	passwd, err := env.GetNeeded("PG_PASSWORD")
	if err != nil {
		return "", fmt.Errorf("proměnná prostředí PG_PASSWORD: %s", err.Error())
	}

	// Výchozí sslmode=disable odpovídá docker-compose nasazení s DB na interní síti.
	// Pro externí/managed databázi nastavte PG_SSLMODE=verify-full (případně require).
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		env.Get("PG_HOST", "database"), env.Get("PG_PORT", "5432"), user, passwd,
		env.Get("PG_DATABASE", "ticketa"), env.Get("PG_SSLMODE", "disable")), nil
}

// GetNewConnection vrátí nové připojení *sql.DB k Postgres instanci.
func GetNewConnection() (*sql.DB, error) {
	psqlInfo, err := getPsqlString()
	if err != nil {
		return nil, fmt.Errorf("chyba vytváření připojovacího řetězce: %s", err.Error())
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("chyba při připojování k Postgres: %s", err.Error())
	}

	return db, nil
}

// Init inicializuje první připojení k Postgres a ověří dosažitelnost databáze.
func Init() error {
	db, err := GetNewConnection()
	if err != nil {
		return fmt.Errorf("chyba inicializace prvního připojení k Postgres: %s", err.Error())
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("chyba ping databáze po prvním připojení: %s", err.Error())
	}

	db.Close()

	return nil
}
