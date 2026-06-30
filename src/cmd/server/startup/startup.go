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
	"github.com/StepanKomis/Ticketa/src/internal/mailer"
	"github.com/StepanKomis/Ticketa/src/www"
	"github.com/StepanKomis/Ticketa/src/www/router"
)

// seedServerSettings uloží hodnoty z env proměnných do server_settings tabulky.
// Volá se při každém startu; from_env=true zabraňuje přepsání hodnot nastaveným
// přes admin UI (které mají from_env=false).
func seedServerSettings(ctx context.Context, queries *dbq.Queries) {
	type envSetting struct {
		key      string
		envName  string
		defVal   string
		required bool
	}
	settings := []envSetting{
		{key: "smtp_host", envName: "SMTP_HOST"},
		{key: "smtp_port", envName: "SMTP_PORT", defVal: "587"},
		{key: "smtp_user", envName: "SMTP_USER"},
		{key: "smtp_password", envName: "SMTP_PASSWORD"},
		{key: "smtp_from", envName: "SMTP_FROM"},
		{key: "pg_host", envName: "PG_HOST", defVal: "database"},
		{key: "pg_port", envName: "PG_PORT", defVal: "5432"},
		{key: "pg_user", envName: "PG_USER"},
		{key: "pg_database", envName: "PG_DATABASE", defVal: "ticketa"},
		{key: "pg_sslmode", envName: "PG_SSLMODE", defVal: "disable"},
	}
	for _, s := range settings {
		val := env.Get(s.envName, s.defVal)
		if val == "" {
			continue
		}
		stored := val
		_ = queries.UpsertServerSetting(ctx, dbq.UpsertServerSettingParams{
			Key:     s.key,
			Value:   stored,
			FromEnv: true,
		})
	}
}

// InitializeServer spustí celý server: připojí se k DB, aplikuje migrace, seeduje stavy
// tiketů, sestaví router a začne naslouchat na SERVER_PORT (výchozí 8080).
// Blokuje dokud server nespadne nebo nevrátí chybu.
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

	queries := dbq.New(db)

	l.Info("Seedování stavů tiketů z konfigurace...")
	if err := statuses.Seed(context.Background(), queries, cfgStore.Get().TicketStatuses); err != nil {
		return fmt.Errorf("seedování stavů tiketů: %w", err)
	}
	l.Info("Stavy tiketů seedovány.")

	l.Info("Seeding nastavení serveru z prostředí...")
	seedServerSettings(context.Background(), queries)

	m := mailer.New(l)
	if m == nil {
		m = mailer.NewFromDB(context.Background(), queries, l)
	}
	if m != nil {
		if err := m.Ping(); err != nil {
			l.Infof("SMTP: test připojení selhal: %s", err)
		} else {
			l.Info("SMTP: připojení ověřeno.")
		}
	}

	port := env.Get("SERVER_PORT", "8080")
	addr := ":" + port

	mux := router.NewRouter(www.StaticFiles, db, cfgStore, m)

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
