package middleware

import (
	"context"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

type localLoginGetter interface {
	GetLocalLoginByUserID(ctx context.Context, id int32) (db.LocalLogin, error)
}

// MustChangePwMiddleware blokuje lokální uživatele, u nichž admin nastavil
// must_change_pw = TRUE. Vrátí 403 se statusem "must_change_password" pro
// všechna volání kromě endpointů na whitelistu (změna hesla, odhlášení).
// Frontend na tento status přesměruje uživatele na /settings/password.
func MustChangePwMiddleware(queries localLoginGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value(ctxkeys.UserContextKey).(db.User)
			if !ok || user.Provider != "local" {
				next.ServeHTTP(w, r)
				return
			}

			ll, err := queries.GetLocalLoginByUserID(r.Context(), user.ID)
			if err != nil || !ll.MustChangePw {
				next.ServeHTTP(w, r)
				return
			}

			handlers.WriteErrorWithStatus(w, http.StatusForbidden, "must_change_password", "před pokračováním je nutné změnit heslo")
		})
	}
}
