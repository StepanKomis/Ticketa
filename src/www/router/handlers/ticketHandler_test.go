package handlers_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

// ---------------------------------------------------------------------------
// Claim a Patch (přiřazení tiketů údržbářům) — DB-backed testy přes mock
// ovladač definovaný v userHandler_test.go (hdlScriptDriver/newHandlerDB),
// sdílený v rámci package handlers_test.
// ---------------------------------------------------------------------------

func newTicketHandlerWithDB(t *testing.T, sqlDB *sql.DB) *handlers.TicketHandler {
	t.Helper()
	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()
	l, err := logs.NewLogger("ticket", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	return handlers.NewTicketHandler(db.New(sqlDB), l, nil)
}

// getTicketRow odpovídá scan pořadí GetTicketRow (tickets.sql.go) — 17 sloupců.
func getTicketRow(id int64, authorID int32, assignedTo *int32, statusID *int32, priority string) hdlQueryResult {
	var assignedVal, statusVal driver.Value
	if assignedTo != nil {
		assignedVal = int64(*assignedTo)
	}
	if statusID != nil {
		statusVal = int64(*statusID)
	}
	return hdlQueryResult{
		cols: []string{
			"id", "title", "body", "created_at", "author_id", "status_id", "priority",
			"assigned_to", "location", "category", "updated_at", "requested_priority",
			"priority_approved_by", "author_name", "assignee_name", "vote_count", "user_has_voted",
		},
		rows: [][]driver.Value{
			{id, "Titulek", "Tělo", time.Now(), int64(authorID), statusVal, priority,
				assignedVal, "Místo", "Kategorie", time.Now(), nil,
				nil, "Autor", "Řešitel", int64(0), false},
		},
	}
}

// metaTicketRow odpovídá scan pořadí Ticket (UpdateTicketMeta RETURNING *) — 13 sloupců.
func metaTicketRow(id int64, authorID int32, assignedTo *int32, statusID *int32, priority string) hdlQueryResult {
	var assignedVal, statusVal driver.Value
	if assignedTo != nil {
		assignedVal = int64(*assignedTo)
	}
	if statusID != nil {
		statusVal = int64(*statusID)
	}
	return hdlQueryResult{
		cols: []string{
			"id", "title", "body", "created_at", "author_id", "status_id", "priority",
			"assigned_to", "location", "category", "updated_at", "requested_priority", "priority_approved_by",
		},
		rows: [][]driver.Value{
			{id, "Titulek", "Tělo", time.Now(), int64(authorID), statusVal, priority,
				assignedVal, "Místo", "Kategorie", time.Now(), nil, nil},
		},
	}
}

func userRow(id int32, userType string) hdlQueryResult {
	return hdlQueryResult{
		cols: []string{
			"id", "email", "first_name", "last_name", "user_type",
			"provider", "is_active", "created_at", "last_login_at", "requested_role", "approved_by",
		},
		rows: [][]driver.Value{
			{int64(id), "u@example.com", "Jana", "Nováková", userType, "local", true, time.Now(), nil, nil, nil},
		},
	}
}

func i32(v int32) *int32 { return &v }

func TestTicketHandler_Claim_InvalidID_Returns400(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/abc/claim", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestTicketHandler_Claim_NoSession_Returns401(t *testing.T) {
	h := newTicketHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/1/claim", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestTicketHandler_Claim_NonMaintainer_Returns403(t *testing.T) {
	h := newTicketHandler(t) // role check happens before any DB call — nil queries je v pořádku
	req := httptest.NewRequest(http.MethodPost, "/api/tickets/1/claim", nil)
	req = withSession(req, 7)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeStaff})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestTicketHandler_Claim_AlreadyAssigned_Returns409(t *testing.T) {
	sqlDB := newHandlerDB(t, &hdlConnScript{
		queries: []hdlQueryResult{
			getTicketRow(1, 99, i32(5), nil, "medium"), // už přiřazeno uživateli 5
		},
	})
	h := newTicketHandlerWithDB(t, sqlDB)

	req := httptest.NewRequest(http.MethodPost, "/api/tickets/1/claim", nil)
	req = withSession(req, 7)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeMaintainer})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTicketHandler_Claim_Success_Returns200(t *testing.T) {
	sqlDB := newHandlerDB(t, &hdlConnScript{
		queries: []hdlQueryResult{
			getTicketRow(1, 99, nil, nil, "medium"),      // nepřiřazeno
			metaTicketRow(1, 99, i32(7), nil, "medium"), // po UpdateTicketMeta přiřazeno mně (7)
		},
	})
	h := newTicketHandlerWithDB(t, sqlDB)

	req := httptest.NewRequest(http.MethodPost, "/api/tickets/1/claim", nil)
	req = withSession(req, 7)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeMaintainer})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"AssignedTo":7`) {
		t.Errorf("expected AssignedTo:7 in response, got %s", rr.Body.String())
	}
}

func TestTicketHandler_Patch_MaintainerNotAssignee_Returns403(t *testing.T) {
	sqlDB := newHandlerDB(t, &hdlConnScript{
		queries: []hdlQueryResult{
			getTicketRow(1, 99, i32(5), nil, "medium"), // přiřazeno jinému uživateli (5), ne mně (7)
		},
	})
	h := newTicketHandlerWithDB(t, sqlDB)

	req := httptest.NewRequest(http.MethodPatch, "/api/tickets/1", strings.NewReader(`{"status_id":2}`))
	req = withSession(req, 7)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeMaintainer})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTicketHandler_Patch_MaintainerExtraField_Returns403(t *testing.T) {
	sqlDB := newHandlerDB(t, &hdlConnScript{
		queries: []hdlQueryResult{
			getTicketRow(1, 99, i32(7), nil, "medium"), // přiřazeno mně (7)
		},
	})
	h := newTicketHandlerWithDB(t, sqlDB)

	// Údržbář smí měnit jen stav — ne assigned_to.
	req := httptest.NewRequest(http.MethodPatch, "/api/tickets/1", strings.NewReader(`{"status_id":2,"assigned_to":5}`))
	req = withSession(req, 7)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeMaintainer})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestTicketHandler_Patch_MaintainerOwnStatus_Returns200(t *testing.T) {
	sqlDB := newHandlerDB(t, &hdlConnScript{
		queries: []hdlQueryResult{
			getTicketRow(1, 99, i32(7), i32(1), "medium"),  // přiřazeno mně (7), stav 1
			metaTicketRow(1, 99, i32(7), i32(2), "medium"), // po patchu stav 2, přiřazení nezměněno
		},
	})
	h := newTicketHandlerWithDB(t, sqlDB)

	req := httptest.NewRequest(http.MethodPatch, "/api/tickets/1", strings.NewReader(`{"status_id":2}`))
	req = withSession(req, 7)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeMaintainer})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `"AssignedTo":7`) {
		t.Errorf("expected AssignedTo zůstává 7 (regrese NULL-overwrite bugu), got %s", rr.Body.String())
	}
}

func TestTicketHandler_Patch_InvalidAssigneeRole_Returns422(t *testing.T) {
	sqlDB := newHandlerDB(t, &hdlConnScript{
		queries: []hdlQueryResult{
			getTicketRow(1, 99, nil, nil, "medium"),
			userRow(42, "student"), // cíl přiřazení je student — nepovolené
		},
	})
	h := newTicketHandlerWithDB(t, sqlDB)

	req := httptest.NewRequest(http.MethodPatch, "/api/tickets/1", strings.NewReader(`{"assigned_to":42}`))
	req = withSession(req, 3)
	req = withUser(req, db.User{ID: 3, UserType: db.UserTypeStaff})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d: %s", rr.Code, rr.Body.String())
	}
}
