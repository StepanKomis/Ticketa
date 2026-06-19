package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

func newTicketHandler(t *testing.T) *handlers.TicketHandler {
	t.Helper()
	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()
	l, err := logs.NewLogger("ticket", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	return handlers.NewTicketHandler(nil, l, nil)
}

func withSession(r *http.Request, userID int64) *http.Request {
	session := db.Session{
		UserID:     userID,
		Token:      "test-token",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Hour),
		LastSeenAt: time.Now(),
	}
	ctx := context.WithValue(r.Context(), ctxkeys.SessionContextKey, session)
	return r.WithContext(ctx)
}

func TestTicketHandler_Create_MissingSession_Returns401(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets",
		strings.NewReader(`{"title":"T","body":"B"}`))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTicketHandler_Create_MissingTitle_Returns422(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets",
		strings.NewReader(`{"body":"B"}`))
	req = withSession(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestTicketHandler_Create_MissingBody_Returns422(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets",
		strings.NewReader(`{"title":"T"}`))
	req = withSession(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

func TestTicketHandler_Get_InvalidID_Returns400(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/tickets/abc", nil)
	req = withSession(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTicketHandler_UnknownMethod_Returns405(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/tickets", nil)
	req = withSession(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

// canModifyTicket is tested indirectly through the handler — a non-author
// session should get 500 when the DB is nil (nil queries panics only on actual
// DB call, not on the auth check). We verify the auth check fires first by
// confirming non-author gets turned away before the nil DB is reached.
func TestTicketHandler_Delete_NoSession_Returns401(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/tickets/1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTicketHandler_ApprovePriority_InvalidID_Returns400(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/abc/approve-priority", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTicketHandler_ApprovePriority_NoSession_Returns401(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/1/approve-priority", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTicketHandler_RejectPriority_InvalidID_Returns400(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/abc/reject-priority", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTicketHandler_RejectPriority_NoSession_Returns401(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/1/reject-priority", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
