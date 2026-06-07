package router

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	middleware "github.com/StepanKomis/Ticketa/src/www/midleware"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

func NewRouter(staticFiles fs.FS, sqlDB *sql.DB, cfg *config.Config) *http.ServeMux {
	httpLogger, err := logs.NewLogger("http", cfg)
	if err != nil {
		panic("failed to create http logger for router: " + err.Error())
	}

	queries := db.New(sqlDB)
	store := security.NewSessionStore(queries)
	auth := middleware.AuthMiddleware(store)

	userHandler, err := handlers.NewUserHandler(httpLogger, sqlDB, store, cfg)
	if err != nil {
		httpLogger.Fatalf("Failed to create user handler in router: %s", err)
	}

	mux := http.NewServeMux()

	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		httpLogger.Fatalf("embed: cannot sub into static/: %s", err)
	}
	mux.Handle("/", handlers.NewStaticHandler(sub))

	// Public routes
	mux.Handle("POST /api/register", userHandler)
	mux.Handle("POST /api/login", userHandler)

	// Protected routes: wrap handlers with auth
	// mux.Handle("/api/...", auth(someHandler))
	_ = auth

	return mux
}
