package middleware

import (
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// MaintainerMiddleware ověří, že přihlášený uživatel (načtený AuthMiddleware) má
// user_type = 'maintainer'. Musí být použit za AuthMiddleware, který do kontextu
// uloží objekt uživatele. Uživatelé bez role maintainer obdrží 403.
func MaintainerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(ctxkeys.UserContextKey)
			if v == nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}
			user := v.(db.User)

			if user.UserType != db.UserTypeMaintainer {
				handlers.WriteError(w, http.StatusForbidden, "přístup odepřen")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
