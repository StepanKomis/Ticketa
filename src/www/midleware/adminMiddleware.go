package middleware

import (
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// AdminMiddleware ověří, že přihlášený uživatel (načtený AuthMiddleware) má
// user_type = 'admin'. Musí být použit za AuthMiddleware, který do kontextu
// uloží objekt uživatele. Uživatelé bez role admin obdrží 403.
func AdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(ctxkeys.UserContextKey)
			if v == nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}
			user := v.(db.User)

			if user.UserType != db.UserTypeAdmin {
				handlers.WriteError(w, http.StatusForbidden, "přístup odepřen")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StaffOrAdminMiddleware povolí přístup uživatelům s rolí staff nebo admin.
// Školníci (maintainer) nemají správcovská práva.
func StaffOrAdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(ctxkeys.UserContextKey)
			if v == nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}
			user := v.(db.User)

			if user.UserType != db.UserTypeStaff && user.UserType != db.UserTypeAdmin {
				handlers.WriteError(w, http.StatusForbidden, "přístup odepřen")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
