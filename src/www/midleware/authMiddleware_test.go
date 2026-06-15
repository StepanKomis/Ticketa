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
// Mocks
// ---------------------------------------------------------------------------

type mockSessionStore struct {
	session db.Session
	err     error
}

func (m *mockSessionStore) GetByToken(_ context.Context, _ string) (db.Session, error) {
	return m.session, m.err
}

type mockUserStore struct {
	user db.User
	err  error
}

func (m *mockUserStore) GetUserByID(_ context.Context, _ int32) (db.User, error) {
	return m.user, m.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// validHexToken je testovací token splňující formátové požadavky middleware
// (64znakový hex řetězec odpovídající 32 náhodným bajtům).
const validHexToken = "a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890"

func activeUser() db.User {
	return db.User{ID: 42, UserType: db.UserTypeStudent, IsActive: true}
}

func authMiddleware(sessions *mockSessionStore, users *mockUserStore) func(http.Handler) http.Handler {
	return middleware.AuthMiddleware(sessions, users)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAuthMiddleware_MissingCookie(t *testing.T) {
	handler := authMiddleware(&mockSessionStore{}, &mockUserStore{})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called when the session cookie is absent")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	handler := authMiddleware(&mockSessionStore{err: errors.New("session not found")}, &mockUserStore{})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called for an invalid token")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: "bogus-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ValidToken_PropagatesSession(t *testing.T) {
	now := time.Now()
	want := db.Session{
		Token:      validHexToken,
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Hour),
		LastSeenAt: now,
	}
	sessions := &mockSessionStore{session: want}
	users := &mockUserStore{user: activeUser()}

	var got db.Session
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(middleware.SessionContextKey)
		if v == nil {
			t.Fatal("session missing from request context")
		}
		got = v.(db.Session)
	})

	handler := authMiddleware(sessions, users)(next)
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

func TestAuthMiddleware_ValidToken_PropagatesUser(t *testing.T) {
	sessions := &mockSessionStore{session: db.Session{Token: validHexToken}}
	want := activeUser()
	users := &mockUserStore{user: want}

	var got db.User
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(middleware.UserContextKey)
		if v == nil {
			t.Fatal("user missing from request context")
		}
		got = v.(db.User)
	})

	handler := authMiddleware(sessions, users)(next)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: validHexToken})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if got.ID != want.ID {
		t.Errorf("user ID: got %d, want %d", got.ID, want.ID)
	}
}

func TestAuthMiddleware_InactiveUser_Gets401(t *testing.T) {
	sessions := &mockSessionStore{session: db.Session{Token: validHexToken}}
	users := &mockUserStore{user: db.User{ID: 1, IsActive: false}}

	handler := authMiddleware(sessions, users)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called for an inactive user")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: validHexToken})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_UserLookupFails_Gets401(t *testing.T) {
	sessions := &mockSessionStore{session: db.Session{Token: validHexToken}}
	users := &mockUserStore{err: errors.New("db error")}

	handler := authMiddleware(sessions, users)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("next handler must not be called when user lookup fails")
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: validHexToken})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
