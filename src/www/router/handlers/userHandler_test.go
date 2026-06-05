package handlers_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
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
	t.Setenv("LOG_DIR", t.TempDir())

	httpLogger, err := logs.NewLogger("http")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	t.Cleanup(func() { httpLogger.Close() })

	h, err := handlers.NewUserHandler(httpLogger, db)
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

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_Post_PasswordTooShort(t *testing.T) {
	h := newTestHandler(t, nil)

	body := `{"email":"jane@example.com","password":"x"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short password, got %d", w.Code)
	}
}

func TestUserHandler_Post_NoSpecialCharacter(t *testing.T) {
	h := newTestHandler(t, nil)

	body := `{"email":"jane@example.com","password":"Secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
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

	body := `{"email":"jane@example.com","password":"Secret1!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on DB error, got %d", w.Code)
	}
}

func TestUserHandler_Post_Success(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{userQueryResult(7, "jane@example.com")},
		execs:   []error{nil},
	}
	db := newHandlerDB(t, script)
	h := newTestHandler(t, db)

	body := `{"email":"jane@example.com","password":"Secret1!","first_name":"Jane","last_name":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
