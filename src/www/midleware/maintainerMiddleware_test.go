package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	middleware "github.com/StepanKomis/Ticketa/src/www/midleware"
)

func maintainerHandler() http.Handler {
	mw := middleware.MaintainerMiddleware()
	return mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

func requestWithUser(user db.User) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	ctx := context.WithValue(req.Context(), ctxkeys.UserContextKey, user)
	return req.WithContext(ctx)
}

func TestMaintainerMiddleware_NoUserInContext_Gets401(t *testing.T) {
	h := maintainerHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_NonMaintainer_Gets403(t *testing.T) {
	h := maintainerHandler()
	req := requestWithUser(db.User{UserType: db.UserTypeStudent, IsActive: true})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_Staff_Gets403(t *testing.T) {
	h := maintainerHandler()
	req := requestWithUser(db.User{UserType: db.UserTypeStaff, IsActive: true})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestMaintainerMiddleware_Maintainer_PassesThrough(t *testing.T) {
	h := maintainerHandler()
	req := requestWithUser(db.User{UserType: db.UserTypeMaintainer, IsActive: true})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
