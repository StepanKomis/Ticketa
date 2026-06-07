package middleware

import (
	"context"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (db.User, error)
}

// MaintainerMiddleware validates the session cookie and then checks that the
// authenticated user has user_type = 'maintainer'. Non-maintainers receive 403.
func MaintainerMiddleware(sessions sessionGetter, users userGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(security.TokenCookieName)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			session, err := sessions.GetByToken(r.Context(), cookie.Value)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			user, err := users.GetUserByID(r.Context(), int32(session.UserID))
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			if user.UserType != db.UserTypeMaintainer {
				handlers.WriteError(w, http.StatusForbidden, "forbidden")
				return
			}

			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
