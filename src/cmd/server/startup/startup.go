package startup

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/env"
	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/config/statuses"
	migrate "github.com/StepanKomis/Ticketa/src/database/migrations"
	psql "github.com/StepanKomis/Ticketa/src/database/postgres"
	psqlmigrations "github.com/StepanKomis/Ticketa/src/database/postgres/migrations"
	dbq "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/www"
	"github.com/StepanKomis/Ticketa/src/www/router"
)

func InitializeServer(l *logs.Logger, cfgStore *config.Store) error {
	l.Info("Spouštění serveru...")
	l.Info("Inicializace připojení k Postgres...")

	err := psql.Init()
	if err != nil {
		return fmt.Errorf("chyba inicializace prvního připojení k Postgres: %s", err.Error())
	}

	db, err := psql.GetNewConnection()
	if err != nil {
		return fmt.Errorf("chyba vytváření databázového připojení při inicializaci serveru: %s", err)
	}

	l.Info("Připojení k Postgres úspěšné.")

	l.Info("Inicializace migrací...")

	migrator, err := psql.NewMigrator()
	if err != nil {
		return fmt.Errorf("chyba vytváření migratoru: %s", err.Error())
	}

	if err := migrator.Init(); err != nil {
		return fmt.Errorf("chyba inicializace migrací: %s", err.Error())
	}

	runner := migrate.NewRunner(migrator, psqlmigrations.All)

	if err := runner.MigrateUp(); err != nil {
		return fmt.Errorf("chyba spouštění migrací: %s", err.Error())
	}

	l.Info("Migrace dokončeny.")

	l.Info("Seedování stavů tiketů z konfigurace...")
	if err := statuses.Seed(context.Background(), dbq.New(db), cfgStore.Get().TicketStatuses); err != nil {
		return fmt.Errorf("seedování stavů tiketů: %w", err)
	}
	l.Info("Stavy tiketů seedovány.")

	port := env.Get("SERVER_PORT", "8080")
	addr := ":" + port

	mux := router.NewRouter(www.StaticFiles, db, cfgStore)

	// Timeouty brání vyčerpání zdrojů pomalými či nedokončenými požadavky (Slowloris).
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	l.Infof("Nasloucháno na %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("chyba HTTP serveru: %s", err.Error())
	}

	return nil
}
