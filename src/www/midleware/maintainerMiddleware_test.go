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

type mockUserStore struct {
	user db.User
	err  error
}

func (m *mockUserStore) GetUserByID(_ context.Context, _ int32) (db.User, error) {
	return m.user, m.err
}

func validSession() db.Session {
	now := time.Now()
	return db.Session{
		ID:         1,
		UserID:     42,
		Token:      "valid-token",
		CreatedAt:  now,
		ExpiresAt:  now.Add(time.Hour),
		LastSeenAt: now,
	}
}

func maintainerMiddleware(sessions *mockSessionStore, users *mockUserStore) http.Handler {
	mw := middleware.MaintainerMiddleware(sessions, users)
	return mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func TestMaintainerMiddleware_MissingCookie(t *testing.T) {
	h := maintainerMiddleware(&mockSessionStore{}, &mockUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_InvalidSession(t *testing.T) {
	sessions := &mockSessionStore{err: errors.New("session not found")}
	h := maintainerMiddleware(sessions, &mockUserStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: "bad-token"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_NonMaintainer_Gets403(t *testing.T) {
	sessions := &mockSessionStore{session: validSession()}
	users := &mockUserStore{user: db.User{UserType: db.UserTypeStudent}}
	h := maintainerMiddleware(sessions, users)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: "valid-token"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_Maintainer_PassesThrough(t *testing.T) {
	sessions := &mockSessionStore{session: validSession()}
	users := &mockUserStore{user: db.User{UserType: db.UserTypeMaintainer}}
	h := maintainerMiddleware(sessions, users)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: "valid-token"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_UserLookupFails_Gets401(t *testing.T) {
	sessions := &mockSessionStore{session: validSession()}
	users := &mockUserStore{err: errors.New("user not found")}
	h := maintainerMiddleware(sessions, users)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	req.AddCookie(&http.Cookie{Name: security.TokenCookieName, Value: "valid-token"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
