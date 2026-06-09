package router

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www"
	middleware "github.com/StepanKomis/Ticketa/src/www/midleware"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

func NewRouter(staticFiles fs.FS, sqlDB *sql.DB, cfgStore *config.Store) *http.ServeMux {
	cfg := cfgStore.Get()
	httpLogger, err := logs.NewLogger("http", cfg)
	if err != nil {
		panic("nepodařilo se vytvořit http logger pro router: " + err.Error())
	}

	queries := db.New(sqlDB)
	sessionStore := security.NewSessionStore(queries)

	auth := middleware.AuthMiddleware(sessionStore)
	admin := middleware.MaintainerMiddleware(sessionStore, queries)

	userHandler, err := handlers.NewUserHandler(httpLogger, sqlDB, sessionStore, cfg)
	if err != nil {
		httpLogger.Fatalf("nepodařilo se vytvořit user handler v routeru: %s", err)
	}

	ticketHandler := handlers.NewTicketHandler(queries, httpLogger)
	adminHandler := handlers.NewAdminHandler(queries, cfgStore, httpLogger)

	mux := http.NewServeMux()

	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		httpLogger.Fatalf("embed: nelze načíst podadresář static/: %s", err)
	}
	mux.Handle("/", handlers.NewStaticHandler(sub))

	// Docs — public, no auth
	docsHandler, err := handlers.NewDocsHandler(www.DocsFiles)
	if err != nil {
		httpLogger.Fatalf("docs: nelze vytvořit handler: %s", err)
	}
	mux.Handle("/docs/", docsHandler)
	mux.Handle("/docs", http.RedirectHandler("/docs/", http.StatusMovedPermanently))

	// Public routes
	mux.Handle("POST /api/register", userHandler)
	mux.Handle("POST /api/login", userHandler)

	// Authenticated routes (any active user)
	mux.Handle("POST /api/tickets", auth(ticketHandler))
	mux.Handle("GET /api/tickets", auth(ticketHandler))
	mux.Handle("GET /api/tickets/{id}", auth(ticketHandler))
	mux.Handle("PUT /api/tickets/{id}", auth(ticketHandler))
	mux.Handle("DELETE /api/tickets/{id}", auth(ticketHandler))

	// Admin routes (maintainer only)
	mux.Handle("GET /api/admin/config", admin(adminHandler))
	mux.Handle("PATCH /api/admin/config", admin(adminHandler))
	mux.Handle("GET /api/admin/ticket-statuses", admin(adminHandler))
	mux.Handle("POST /api/admin/ticket-statuses", admin(adminHandler))
	mux.Handle("PUT /api/admin/ticket-statuses/{id}", admin(adminHandler))
	mux.Handle("DELETE /api/admin/ticket-statuses/{id}", admin(adminHandler))
	mux.Handle("GET /api/admin/users", admin(adminHandler))
	mux.Handle("GET /api/admin/users/{id}", admin(adminHandler))
	mux.Handle("PATCH /api/admin/users/{id}", admin(adminHandler))

	return mux
}
