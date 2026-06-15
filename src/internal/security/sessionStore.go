package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

const (
	TokenCookieName = "session_token"

	// SessionTTLSeconds určuje životnost session na serveru i Max-Age cookie —
	// obě expirace musí zůstat sjednocené.
	SessionTTLSeconds = int64(7 * 24 * 60 * 60)
)

type SessionStore struct {
	queries *db.Queries
}

func NewSessionStore(q *db.Queries) *SessionStore {
	return &SessionStore{queries: q}
}

// Create vytvoří novou session pro uživatele. Pokud již existuje (unikátní constraint
// na user_id), existující řádek se přegeneruje s novým tokenem a expirací.
func (s *SessionStore) Create(ctx context.Context, userID int64, r *http.Request) (db.Session, error) {
	token, err := generateToken()
	if err != nil {
		return db.Session{}, fmt.Errorf("generating token: %w", err)
	}

	ip := toInet(extractIP(r))

	session, err := s.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		Token:     token,
		Ip:        ip,
		UserAgent: r.UserAgent(),
		Column5:   SessionTTLSeconds,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return s.queries.RegenerateToken(ctx, db.RegenerateTokenParams{
				UserID:    userID,
				Token:     token,
				Ip:        ip,
				UserAgent: r.UserAgent(),
				Column5:   SessionTTLSeconds,
			})
		}
		return db.Session{}, fmt.Errorf("creating session: %w", err)
	}

	return session, nil
}

// GetByToken ověří token a atomicky aktualizuje last_seen_at.
// Vrátí sql.ErrNoRows pokud token chybí, vypršel nebo byl soft-smazán.
func (s *SessionStore) GetByToken(ctx context.Context, token string) (db.Session, error) {
	return s.queries.GetSessionByToken(ctx, token)
}

// Invalidate provede soft delete session — nastaví deleted=true v databázi.
// Token přestane být platný okamžitě; záznam zůstane pro účely auditu.
func (s *SessionStore) Invalidate(ctx context.Context, token string) error {
	return s.queries.SoftDeleteSession(ctx, token)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func extractIP(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return net.IPv4zero
	}
	return ip
}

func toInet(ip net.IP) pqtype.Inet {
	if v4 := ip.To4(); v4 != nil {
		return pqtype.Inet{
			IPNet: net.IPNet{IP: v4, Mask: net.CIDRMask(32, 32)},
			Valid: true,
		}
	}
	return pqtype.Inet{
		IPNet: net.IPNet{IP: ip.To16(), Mask: net.CIDRMask(128, 128)},
		Valid: true,
	}
}
