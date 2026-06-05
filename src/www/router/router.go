package router

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

func NewRouter(staticFiles fs.FS, db *sql.DB) *http.ServeMux {
	// Creating the HTTP logger
	httpLogger, err := logs.NewLogger("http")
  if err != nil {
  	httpLogger.Fatalf("failed to create http logger for router: %s", err)
  }
	
	// User handler initialization
	userHandler, err := handlers.NewUserHandler(httpLogger, db)
	if err != nil {
		httpLogger.Fatalf("Failed to create user handler in router: %s", err)
	}

	mux := http.NewServeMux()
	
	// Static route for the react front-end
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		httpLogger.Fatalf("embed: cannot sub into static/: %s", err)
	}
	mux.Handle("/", handlers.NewStaticHandler(sub))

	// API routes
	mux.Handle("/api/users", userHandler)
	return mux
}
