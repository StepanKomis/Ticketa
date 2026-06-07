package middleware

import (
	"context"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// SessionContextKey is the context key under which auth middleware stores the
// validated session. Handlers retrieve it via r.Context().Value(middleware.SessionContextKey).
const SessionContextKey = ctxkeys.SessionContextKey

// sessionGetter abstracts session-store lookups to make AuthMiddleware testable
// without a live database. *security.SessionStore satisfies this interface.
type sessionGetter interface {
	GetByToken(ctx context.Context, token string) (db.Session, error)
}

// TODO: validate token length and character set (e.g. must be a 64-char hex string)
// before querying the database to guard against abnormally large or malformed cookie values.
func AuthMiddleware(store sessionGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(security.TokenCookieName)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			session, err := store.GetByToken(r.Context(), cookie.Value)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
