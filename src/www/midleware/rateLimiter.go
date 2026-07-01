package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

type rateBucket struct {
	mu   sync.Mutex
	hits []time.Time
}

// RateLimiter vrátí middleware implementující sliding-window rate limiting keyed by client IP.
// max = maximální počet požadavků, window = délka okna (např. time.Minute).
// Po překročení vrátí 429 Too Many Requests s hlavičkou Retry-After.
func RateLimiter(max int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	ips := make(map[string]*rateBucket)

	// Background goroutine odstraňuje záznamy IP adres bez aktivit v posledním okně,
	// aby se předešlo neomezenému růstu paměti při provozu s mnoha IP adresami.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			mu.Lock()
			for ip, b := range ips {
				b.mu.Lock()
				cutoff := now.Add(-window)
				active := false
				for _, t := range b.hits {
					if t.After(cutoff) {
						active = true
						break
					}
				}
				if !active {
					delete(ips, ip)
				}
				b.mu.Unlock()
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			mu.Lock()
			b, ok := ips[ip]
			if !ok {
				b = &rateBucket{}
				ips[ip] = b
			}
			mu.Unlock()

			b.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-window)
			valid := b.hits[:0]
			for _, t := range b.hits {
				if t.After(cutoff) {
					valid = append(valid, t)
				}
			}
			b.hits = valid
			allowed := len(b.hits) < max
			if allowed {
				b.hits = append(b.hits, now)
			}
			b.mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", "60")
				handlers.WriteError(w, http.StatusTooManyRequests, "příliš mnoho požadavků, zkuste to za chvíli")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP vrátí IP adresu klienta. Preferuje X-Real-IP nastavenou reverzní proxy
// (nginx, Traefik) před RemoteAddr — za proxy je RemoteAddr vždy IP proxy serveru.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
