package middleware

import (
	"context"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// SessionContextKey je klíč kontextu pod kterým autentizační middleware ukládá
// validovanou session. Handlery ji čtou přes r.Context().Value(middleware.SessionContextKey).
const SessionContextKey = ctxkeys.SessionContextKey

// sessionGetter abstrahuje vyhledávání v session store, aby byl AuthMiddleware testovatelný
// bez živé databáze. *security.SessionStore toto rozhraní implementuje.
type sessionGetter interface {
	GetByToken(ctx context.Context, token string) (db.Session, error)
}

// TODO: ověřit délku a znakovou sadu tokenu (např. musí být 64znakový hex řetězec)
// před dotazem do databáze, aby se zabránilo abnormálně velkým nebo poškozeným hodnotám cookie.
func AuthMiddleware(store sessionGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(security.TokenCookieName)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			session, err := store.GetByToken(r.Context(), cookie.Value)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
