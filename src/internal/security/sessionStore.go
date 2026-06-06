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
	TokenCookieName   = "session_token"
	sessionTTLSeconds = int64(7 * 24 * 60 * 60)
)

type SessionStore struct {
	queries *db.Queries
}

func NewSessionStore(q *db.Queries) *SessionStore {
	return &SessionStore{queries: q}
}

// Create issues a new session for the user. If one already exists (unique constraint
// on user_id), the existing row is regenerated with a fresh token and expiry.
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
		Column5:   sessionTTLSeconds,
	})
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return s.queries.RegenerateToken(ctx, db.RegenerateTokenParams{
				UserID:    userID,
				Token:     token,
				Ip:        ip,
				UserAgent: r.UserAgent(),
				Column5:   sessionTTLSeconds,
			})
		}
		return db.Session{}, fmt.Errorf("creating session: %w", err)
	}

	return session, nil
}

// GetByToken validates the token and bumps last_seen_at atomically.
// Returns sql.ErrNoRows if the token is missing, expired, or soft-deleted.
func (s *SessionStore) GetByToken(ctx context.Context, token string) (db.Session, error) {
	return s.queries.GetSessionByToken(ctx, token)
}

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
