package middleware

import (
	"context"
	"encoding/hex"
	"net/http"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// SessionContextKey je klíč kontextu pod kterým autentizační middleware ukládá
// validovanou session. Handlery ji čtou přes r.Context().Value(middleware.SessionContextKey).
const SessionContextKey = ctxkeys.SessionContextKey

// UserContextKey je klíč kontextu pod kterým autentizační middleware ukládá
// načteného uživatele. Handlery ho čtou přes r.Context().Value(middleware.UserContextKey).
const UserContextKey = ctxkeys.UserContextKey

// sessionGetter abstrahuje vyhledávání v session store, aby byl AuthMiddleware testovatelný
// bez živé databáze. *security.SessionStore toto rozhraní implementuje.
type sessionGetter interface {
	GetByToken(ctx context.Context, token string) (db.Session, error)
}

// userGetter abstrahuje načítání uživatele z DB. *db.Queries toto rozhraní implementuje.
type userGetter interface {
	GetUserByID(ctx context.Context, id int32) (db.User, error)
}

// AuthMiddleware ověří session cookie, načte uživatele z DB a vloží obojí do kontextu.
// Vrátí 401 pokud cookie chybí, token má nesprávný formát, session neexistuje/vypršela
// nebo účet je neaktivní.
func AuthMiddleware(store sessionGetter, users userGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(security.TokenCookieName)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			// Token musí být přesně 64znakový hex řetězec (32 náhodných bajtů → hex).
			// Odmítnutí abnormálních hodnot před DB dotazem zabraňuje zbytečné zátěži.
			if len(cookie.Value) != 64 {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}
			if _, err := hex.DecodeString(cookie.Value); err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			session, err := store.GetByToken(r.Context(), cookie.Value)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			user, err := users.GetUserByID(r.Context(), int32(session.UserID))
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}
			if !user.IsActive {
				handlers.WriteError(w, http.StatusUnauthorized, "nepřihlášen")
				return
			}

			ctx := context.WithValue(r.Context(), SessionContextKey, session)
			ctx = context.WithValue(ctx, UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
