package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// Recovery zachytí paniku v HTTP handleru, zapíše stack trace do logu a vrátí
// klientovi 500. Bez tohoto middleware by panika v handleru ukončila celý server.
func Recovery(logger *logs.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if v := recover(); v != nil {
					logger.Errorf("panika v HTTP handleru [%s %s]: %v\n%s",
						r.Method, r.URL.Path, v, debug.Stack())
					handlers.WriteError(w, http.StatusInternalServerError, "interní chyba serveru")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
