package middleware

import (
	"context"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (db.User, error)
}

// MaintainerMiddleware ověří session cookie a zkontroluje, že přihlášený uživatel
// je aktivní a má user_type = 'maintainer'. Neaktivní uživatelé obdrží 401,
// uživatelé bez role maintainer 403.
func MaintainerMiddleware(sessions sessionGetter, users userGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(security.TokenCookieName)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			session, err := sessions.GetByToken(r.Context(), cookie.Value)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			user, err := users.GetUserByID(r.Context(), int32(session.UserID))
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			// Defense-in-depth: neaktivní účet nesmí projít, i kdyby jeho session
			// prošla validací (primární kontrola je v GetSessionByToken).
			if !user.IsActive {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			if user.UserType != db.UserTypeMaintainer {
				handlers.WriteError(w, http.StatusForbidden, "přístup odepřen")
				return
			}

			ctx := context.WithValue(r.Context(), ctxkeys.SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
