package router

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

func NewRouter(staticFiles fs.FS, sqlDB *sql.DB) *http.ServeMux {
	httpLogger, err := logs.NewLogger("http")
	if err != nil {
		httpLogger.Fatalf("failed to create http logger for router: %s", err)
	}

	queries := db.New(sqlDB)
	store := security.NewSessionStore(queries)
	auth := handlers.AuthMiddleware(store)

	userHandler, err := handlers.NewUserHandler(httpLogger, sqlDB)
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
	mux.Handle("/api/users", userHandler)

	// Protected routes: wrap handlers with auth
	// mux.Handle("/api/...", auth(someHandler))
	_ = auth

	return mux
}
