package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/security"
	middleware "github.com/StepanKomis/Ticketa/src/www/midleware"
)

// ---------------------------------------------------------------------------
// Mock
// ---------------------------------------------------------------------------

type mockSessionStore struct {
	session db.Session
	err     error
}

func (m *mockSessionStore) GetByToken(_ context.Context, _ string) (db.Session, error) {
	return m.session, m.err
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	store := &mockSessionStore{}
	handler := middleware.AuthMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler must not be called when the session cookie is absent")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	store := &mockSessionStore{err: errors.New("session not found")}
	handler := middleware.AuthMiddleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler must not be called for an invalid token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: "bogus-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// validHexToken je testovací token splňující formátové požadavky middleware
// (64znakový hex řetězec odpovídající 32 náhodným bajtům).
const validHexToken = "a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890"

func TestAuthMiddleware_ValidToken_PropagatesSession(t *testing.T) {
	now := time.Now()
	want := db.Session{
		Token:      validHexToken,
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Hour),
		LastSeenAt: now,
	}
	store := &mockSessionStore{session: want}

	var got db.Session
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(middleware.SessionContextKey)
		if v == nil {
			t.Fatal("session missing from request context")
		}
		got = v.(db.Session)
	})

	handler := middleware.AuthMiddleware(store)(next)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: validHexToken})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if got.Token != want.Token {
		t.Errorf("session token: got %q, want %q", got.Token, want.Token)
	}
}
