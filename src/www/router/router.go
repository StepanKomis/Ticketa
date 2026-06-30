package router

import (
	"database/sql"
	"io/fs"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/internal/activity"
	"github.com/StepanKomis/Ticketa/src/internal/mailer"
	"github.com/StepanKomis/Ticketa/src/internal/notifications"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www"
	middleware "github.com/StepanKomis/Ticketa/src/www/midleware"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

func NewRouter(staticFiles fs.FS, sqlDB *sql.DB, cfgStore *config.Store, m *mailer.Mailer) *http.ServeMux {
	cfg := cfgStore.Get()
	httpLogger, err := logs.NewLogger("http", cfg)
	if err != nil {
		panic("nepodařilo se vytvořit http logger pro router: " + err.Error())
	}

	queries := db.New(sqlDB)
	sessionStore := security.NewSessionStore(queries)

	auth := middleware.AuthMiddleware(sessionStore, queries)
	mustChangePw := middleware.MustChangePwMiddleware(queries)
	// authEnforced = auth + blokování uživatelů s must_change_pw = TRUE.
	// Whitelistované routes (/api/me/password, /api/logout) používají jen auth.
	authEnforced := func(h http.Handler) http.Handler { return auth(mustChangePw(h)) }
	admin := func(h http.Handler) http.Handler { return authEnforced(middleware.AdminMiddleware()(h)) }
	staffAdmin := func(h http.Handler) http.Handler { return authEnforced(middleware.StaffOrAdminMiddleware()(h)) }

	activityLogger := activity.NewActivityLogger(queries, cfg, httpLogger)
	notifier := notifications.NewNotifier(queries, httpLogger, m)

	userHandler, err := handlers.NewUserHandler(httpLogger, sqlDB, sessionStore, cfg, activityLogger)
	if err != nil {
		httpLogger.Fatalf("nepodařilo se vytvořit user handler v routeru: %s", err)
	}

	ticketHandler := handlers.NewTicketHandler(queries, httpLogger, activityLogger, notifier)
	commentHandler := handlers.NewCommentHandler(queries, httpLogger, activityLogger)
	adminHandler := handlers.NewAdminHandler(queries, cfgStore, httpLogger, activityLogger, notifier)
	activityHandler := handlers.NewActivityHandler(queries, httpLogger)
	notificationHandler := handlers.NewNotificationHandler(queries, httpLogger)
	notificationPreferencesHandler := handlers.NewNotificationPreferencesHandler(queries, httpLogger)
	serverSettingsHandler := handlers.NewServerSettingsHandler(queries, m, httpLogger)

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
	mux.Handle("GET /api/setup-status", userHandler)
	mux.Handle("POST /api/register", userHandler)
	mux.Handle("POST /api/login", userHandler)
	mux.Handle("POST /api/auth/invite/accept", userHandler)
	mux.Handle("POST /api/setup/smtp/test", serverSettingsHandler)
	mux.Handle("POST /api/setup/db/test", serverSettingsHandler)

	// User routes (authenticated)
	// /api/me/password, /api/me/email a /api/logout jsou whitelistovány — uživatel
	// s must_change_pw = TRUE musí mít přístup ke změně hesla/e-mailu a k odhlášení
	// bez blokování middlewarem.
	mux.Handle("PATCH /api/me/password", auth(userHandler))
	mux.Handle("PATCH /api/me/email", auth(userHandler))
	mux.Handle("POST /api/logout", auth(userHandler))
	mux.Handle("GET /api/me", authEnforced(userHandler))
	mux.Handle("PATCH /api/me", authEnforced(userHandler))

	// Authenticated routes (any active user)
	mux.Handle("POST /api/tickets", authEnforced(ticketHandler))
	mux.Handle("GET /api/tickets", authEnforced(ticketHandler))
	mux.Handle("GET /api/tickets/{id}", authEnforced(ticketHandler))
	mux.Handle("PUT /api/tickets/{id}", authEnforced(ticketHandler))
	mux.Handle("PATCH /api/tickets/{id}", authEnforced(ticketHandler))
	mux.Handle("DELETE /api/tickets/{id}", authEnforced(ticketHandler))
	mux.Handle("POST /api/tickets/{id}/vote", authEnforced(ticketHandler))
	mux.Handle("DELETE /api/tickets/{id}/vote", authEnforced(ticketHandler))
	mux.Handle("POST /api/tickets/{id}/approve-priority", staffAdmin(ticketHandler))
	mux.Handle("POST /api/tickets/{id}/reject-priority", staffAdmin(ticketHandler))
	// claim: jen authEnforced — role (maintainer) se ověřuje uvnitř handleru,
	// staffAdmin() by tu maintainera naopak zablokoval.
	mux.Handle("POST /api/tickets/{id}/claim", authEnforced(ticketHandler))

	// Comment routes (any active user; delete also allowed for staff/maintainer)
	mux.Handle("POST /api/tickets/{id}/comments", authEnforced(commentHandler))
	mux.Handle("GET /api/tickets/{id}/comments", authEnforced(commentHandler))
	mux.Handle("PUT /api/comments/{id}", authEnforced(commentHandler))
	mux.Handle("DELETE /api/comments/{id}", authEnforced(commentHandler))

	// Veřejný (auth-only) přístup ke stavům tiketů — potřebují i staff uživatelé
	mux.Handle("GET /api/ticket-statuses", authEnforced(adminHandler))
	// History tiketu
	mux.Handle("GET /api/tickets/{id}/history", authEnforced(ticketHandler))

	// Notifications
	mux.Handle("GET /api/notifications", authEnforced(notificationHandler))
	mux.Handle("POST /api/notifications/mark-viewed", authEnforced(notificationHandler))
	mux.Handle("GET /api/notifications/preferences", authEnforced(notificationPreferencesHandler))
	mux.Handle("PUT /api/notifications/preferences", authEnforced(notificationPreferencesHandler))

	// Activity log
	mux.Handle("GET /api/activity", admin(activityHandler))
	mux.Handle("GET /api/users/{id}/activity", authEnforced(activityHandler))

	// Setup wizard completion (admin only, only before wizard_completed)
	mux.Handle("POST /api/setup/complete", admin(serverSettingsHandler))

	// Server settings (admin only)
	mux.Handle("GET /api/admin/server-settings", admin(serverSettingsHandler))
	mux.Handle("PATCH /api/admin/server-settings/smtp", admin(serverSettingsHandler))
	mux.Handle("POST /api/admin/server-settings/smtp/test", admin(serverSettingsHandler))

	// Admin routes (maintainer only)
	mux.Handle("GET /api/admin/config", admin(adminHandler))
	mux.Handle("PATCH /api/admin/config", admin(adminHandler))
	mux.Handle("GET /api/admin/ticket-statuses", admin(adminHandler))
	mux.Handle("POST /api/admin/ticket-statuses", admin(adminHandler))
	mux.Handle("PUT /api/admin/ticket-statuses/{id}", admin(adminHandler))
	mux.Handle("DELETE /api/admin/ticket-statuses/{id}", admin(adminHandler))
	mux.Handle("GET /api/admin/users", staffAdmin(adminHandler))
	mux.Handle("GET /api/admin/users/{id}", admin(adminHandler))
	mux.Handle("PATCH /api/admin/users/{id}", admin(adminHandler))
	mux.Handle("POST /api/admin/users/{id}/approve", staffAdmin(adminHandler))
	mux.Handle("POST /api/admin/users/{id}/reject", staffAdmin(adminHandler))
	mux.Handle("POST /api/admin/invitations", admin(adminHandler))

	return mux
}
