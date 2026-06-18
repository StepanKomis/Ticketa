package handlers_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

// ---------------------------------------------------------------------------
// Minimal SQL mock driver (mirrored from registration tests; registers under
// a different name so the two packages don't collide).
// ---------------------------------------------------------------------------

var (
	hdlDriverOnce sync.Once
	hdlDriver     = &hdlScriptDriver{ch: make(chan *hdlConnScript, 16)}
)

func init() {
	hdlDriverOnce.Do(func() {
		sql.Register("mock_handlers", hdlDriver)
	})
}

type hdlScriptDriver struct{ ch chan *hdlConnScript }

func (d *hdlScriptDriver) Open(_ string) (driver.Conn, error) {
	select {
	case s := <-d.ch:
		return &hdlConn{s: s}, nil
	default:
		return nil, fmt.Errorf("no script queued")
	}
}

type hdlConnScript struct {
	queries   []hdlQueryResult
	execs     []error
	commitErr error
	qi, ei    int
}

type hdlQueryResult struct {
	cols []string
	rows [][]driver.Value
	err  error
}

func newHandlerDB(t *testing.T, s *hdlConnScript) *sql.DB {
	t.Helper()
	hdlDriver.ch <- s
	db, err := sql.Open("mock_handlers", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	return db
}

type hdlConn struct{ s *hdlConnScript }

func (c *hdlConn) Prepare(_ string) (driver.Stmt, error) { return &hdlStmt{c: c}, nil }
func (c *hdlConn) Close() error                          { return nil }
func (c *hdlConn) Begin() (driver.Tx, error)             { return &hdlTx{c: c}, nil }

type hdlTx struct{ c *hdlConn }

func (t *hdlTx) Commit() error   { return t.c.s.commitErr }
func (t *hdlTx) Rollback() error { return nil }

type hdlStmt struct{ c *hdlConn }

func (s *hdlStmt) Close() error  { return nil }
func (s *hdlStmt) NumInput() int { return -1 }

func (s *hdlStmt) Exec(_ []driver.Value) (driver.Result, error) {
	sc := s.c.s
	if sc.ei >= len(sc.execs) {
		return nil, fmt.Errorf("unexpected Exec (index %d)", sc.ei)
	}
	err := sc.execs[sc.ei]
	sc.ei++
	return hdlResult{}, err
}

func (s *hdlStmt) Query(_ []driver.Value) (driver.Rows, error) {
	sc := s.c.s
	if sc.qi >= len(sc.queries) {
		return nil, fmt.Errorf("unexpected Query (index %d)", sc.qi)
	}
	r := sc.queries[sc.qi]
	sc.qi++
	if r.err != nil {
		return nil, r.err
	}
	return &hdlRows{cols: r.cols, data: r.rows}, nil
}

type hdlResult struct{}

func (hdlResult) LastInsertId() (int64, error) { return 0, nil }
func (hdlResult) RowsAffected() (int64, error) { return 1, nil }

type hdlRows struct {
	cols []string
	data [][]driver.Value
	i    int
}

func (r *hdlRows) Columns() []string { return r.cols }
func (r *hdlRows) Close() error      { return nil }
func (r *hdlRows) Next(dest []driver.Value) error {
	if r.i >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.i])
	r.i++
	return nil
}

func userQueryResult(id int64, email string) hdlQueryResult {
	return hdlQueryResult{
		cols: []string{
			"id", "email", "first_name", "last_name",
			"user_type", "provider", "is_active", "created_at", "last_login_at",
		},
		rows: [][]driver.Value{
			{id, email, "Jane", "Doe", "student", "local", true, time.Now(), nil},
		},
	}
}

// ---------------------------------------------------------------------------
// Test helper
// ---------------------------------------------------------------------------

func newTestHandler(t *testing.T, db *sql.DB) http.Handler {
	t.Helper()
	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()

	httpLogger, err := logs.NewLogger("http", cfg)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	t.Cleanup(func() { httpLogger.Close() })

	h, err := handlers.NewUserHandler(httpLogger, db, nil, cfg)
	if err != nil {
		t.Fatalf("NewUserHandler: %v", err)
	}
	return h
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestUserHandler_MethodNotAllowed(t *testing.T) {
	h := newTestHandler(t, nil)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/api/users", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, w.Code)
		}
	}
}

func TestUserHandler_Post_InvalidJSON(t *testing.T) {
	h := newTestHandler(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_Post_PasswordTooShort(t *testing.T) {
	h := newTestHandler(t, nil)

	body := `{"email":"jane@example.com","password":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short password, got %d", w.Code)
	}
}

func TestUserHandler_Post_NoSpecialCharacter(t *testing.T) {
	h := newTestHandler(t, nil)

	body := `{"email":"jane@example.com","password":"Secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing special char, got %d", w.Code)
	}
}

func TestUserHandler_Post_DBError(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{{err: fmt.Errorf("duplicate email")}},
	}
	db := newHandlerDB(t, script)
	h := newTestHandler(t, db)

	body := `{"email":"jane@example.com","password":"Secret1!","user_type":"student"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on DB error, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// PATCH /api/me/password
// ---------------------------------------------------------------------------

func patchPasswordReq(t *testing.T, sqlDB *sql.DB, sessionUserID int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := newTestHandler(t, sqlDB)
	req := httptest.NewRequest(http.MethodPatch, "/api/me/password", strings.NewReader(body))

	if sessionUserID != 0 {
		ctx := context.WithValue(req.Context(), ctxkeys.SessionContextKey, db.Session{UserID: sessionUserID})
		req = req.WithContext(ctx)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestPatchMyPassword_NoSession(t *testing.T) {
	w := patchPasswordReq(t, nil, 0, `{"current_password":"Old1!","new_password":"New1@"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPatchMyPassword_InvalidJSON(t *testing.T) {
	w := patchPasswordReq(t, nil, 1, "not json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPatchMyPassword_WeakNewPassword(t *testing.T) {
	w := patchPasswordReq(t, nil, 1, `{"current_password":"Old1!","new_password":"weak"}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestPatchMyPassword_NoLocalLogin(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{{err: fmt.Errorf("not found")}},
	}
	db := newHandlerDB(t, script)
	w := patchPasswordReq(t, db, 1, `{"current_password":"Old1!","new_password":"NewPass1@"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPatchMyPassword_WrongCurrentPassword(t *testing.T) {
	// Bcrypt hash pro "CorrectPass1!" vygenerovaný s cost=4 pro rychlost testů
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1!"), 4)
	script := &hdlConnScript{
		queries: []hdlQueryResult{{
			cols: []string{"id", "password_hash", "must_change_pw", "pw_changed_at"},
			rows: [][]driver.Value{{int64(1), string(hash), false, nil}},
		}},
	}
	db := newHandlerDB(t, script)
	w := patchPasswordReq(t, db, 1, `{"current_password":"WrongPass1!","new_password":"NewPass1@"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPatchMyPassword_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("OldPass1!"), 4)
	script := &hdlConnScript{
		queries: []hdlQueryResult{{
			cols: []string{"id", "password_hash", "must_change_pw", "pw_changed_at"},
			rows: [][]driver.Value{{int64(1), string(hash), false, nil}},
		}},
		execs: []error{nil}, // UpdateLocalLoginPassword
	}
	db := newHandlerDB(t, script)
	w := patchPasswordReq(t, db, 1, `{"current_password":"OldPass1!","new_password":"NewPass1@"}`)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// PATCH /api/me/email
// ---------------------------------------------------------------------------

func fullUserRow(id int32, email string) hdlQueryResult {
	return hdlQueryResult{
		cols: []string{
			"id", "email", "first_name", "last_name",
			"user_type", "provider", "is_active", "created_at", "last_login_at",
			"requested_role", "approved_by",
		},
		rows: [][]driver.Value{
			{id, email, "Jane", "Doe", "student", "local", true, time.Now(), nil, nil, nil},
		},
	}
}

func patchEmailReq(t *testing.T, sqlDB *sql.DB, sessionUserID int64, body string) *httptest.ResponseRecorder {
	t.Helper()
	h := newTestHandler(t, sqlDB)
	req := httptest.NewRequest(http.MethodPatch, "/api/me/email", strings.NewReader(body))

	if sessionUserID != 0 {
		ctx := context.WithValue(req.Context(), ctxkeys.SessionContextKey, db.Session{UserID: sessionUserID})
		req = req.WithContext(ctx)
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestPatchMyEmail_NoSession(t *testing.T) {
	w := patchEmailReq(t, nil, 0, `{"current_password":"Old1!","new_email":"new@example.com"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPatchMyEmail_InvalidJSON(t *testing.T) {
	w := patchEmailReq(t, nil, 1, "not json")
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPatchMyEmail_EmptyNewEmail(t *testing.T) {
	w := patchEmailReq(t, nil, 1, `{"current_password":"Old1!","new_email":"   "}`)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestPatchMyEmail_NoLocalLogin(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{{err: fmt.Errorf("not found")}},
	}
	db := newHandlerDB(t, script)
	w := patchEmailReq(t, db, 1, `{"current_password":"Old1!","new_email":"new@example.com"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPatchMyEmail_WrongCurrentPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1!"), 4)
	script := &hdlConnScript{
		queries: []hdlQueryResult{{
			cols: []string{"id", "password_hash", "must_change_pw", "pw_changed_at"},
			rows: [][]driver.Value{{int64(1), string(hash), false, nil}},
		}},
	}
	db := newHandlerDB(t, script)
	w := patchEmailReq(t, db, 1, `{"current_password":"WrongPass1!","new_email":"new@example.com"}`)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestPatchMyEmail_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1!"), 4)
	script := &hdlConnScript{
		queries: []hdlQueryResult{
			{
				cols: []string{"id", "password_hash", "must_change_pw", "pw_changed_at"},
				rows: [][]driver.Value{{int64(1), string(hash), false, nil}},
			},
			fullUserRow(1, "new@example.com"),
		},
	}
	db := newHandlerDB(t, script)
	w := patchEmailReq(t, db, 1, `{"current_password":"CorrectPass1!","new_email":"NEW@Example.com"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", w.Code, w.Body.String())
	}
	var got struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %q", got.Email)
	}
}

func TestPatchMyEmail_DuplicateEmail(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1!"), 4)
	script := &hdlConnScript{
		queries: []hdlQueryResult{
			{
				cols: []string{"id", "password_hash", "must_change_pw", "pw_changed_at"},
				rows: [][]driver.Value{{int64(1), string(hash), false, nil}},
			},
			{err: &pq.Error{Code: "23505"}},
		},
	}
	db := newHandlerDB(t, script)
	w := patchEmailReq(t, db, 1, `{"current_password":"CorrectPass1!","new_email":"taken@example.com"}`)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestUserHandler_Post_Success(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{userQueryResult(7, "jane@example.com")},
		execs:   []error{nil},
	}
	db := newHandlerDB(t, script)
	h := newTestHandler(t, db)

	body := `{"email":"jane@example.com","password":"Secret1!","first_name":"Jane","last_name":"Doe","user_type":"student"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
