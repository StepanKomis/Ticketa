package middleware

import (
	"net/http"
	"strings"
)

// SecurityHeaders přidá obranné HTTP hlavičky ke každé odpovědi.
// Chrání před clickjackingem (X-Frame-Options), MIME sniffingem (X-Content-Type-Options)
// a zbytečným únikem referrer informací. API endpointy dostávají Cache-Control: no-store,
// aby prohlížeče ani proxy neukládaly citlivá data.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}

// BodyLimit omezí velikost těla požadavku na maxBytes bajtů.
// Přetečení způsobí chybu čtení v handleru (http.MaxBytesReader nastaví příznak),
// takže handlery dostávají 413 pokud tělo přečtou po překročení limitu.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
