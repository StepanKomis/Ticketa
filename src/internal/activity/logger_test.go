package activity

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
)

// ---------------------------------------------------------------------------
// Minimal mock SQL driver — supports only Exec, which is all CreateActivityLog needs.
// ---------------------------------------------------------------------------

var (
	mockDriverOnce sync.Once
	mockDriver     = &actScriptDriver{ch: make(chan *actConnScript, 16)}
)

func init() {
	mockDriverOnce.Do(func() {
		sql.Register("mock_activity", mockDriver)
	})
}

type actScriptDriver struct{ ch chan *actConnScript }

func (d *actScriptDriver) Open(_ string) (driver.Conn, error) {
	select {
	case s := <-d.ch:
		return &actConn{s: s}, nil
	default:
		return nil, fmt.Errorf("no script queued")
	}
}

type actConnScript struct {
	execErr  error
	lastArgs []driver.Value
}

func newActivityDB(t *testing.T, s *actConnScript) *sql.DB {
	t.Helper()
	mockDriver.ch <- s
	sqlDB, err := sql.Open("mock_activity", "")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	return sqlDB
}

type actConn struct{ s *actConnScript }

func (c *actConn) Prepare(_ string) (driver.Stmt, error) { return &actStmt{c: c}, nil }
func (c *actConn) Close() error                          { return nil }
func (c *actConn) Begin() (driver.Tx, error)              { return nil, fmt.Errorf("transactions not supported by mock") }

type actStmt struct{ c *actConn }

func (s *actStmt) Close() error  { return nil }
func (s *actStmt) NumInput() int { return -1 }

func (s *actStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.c.s.lastArgs = args
	if s.c.s.execErr != nil {
		return nil, s.c.s.execErr
	}
	return actResult{}, nil
}

func (s *actStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return nil, fmt.Errorf("query not supported by mock")
}

type actResult struct{}

func (actResult) LastInsertId() (int64, error) { return 0, nil }
func (actResult) RowsAffected() (int64, error) { return 1, nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestActivityLogger(t *testing.T, script *actConnScript, file *bytes.Buffer) *ActivityLogger {
	t.Helper()
	sqlDB := newActivityDB(t, script)
	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()
	l, err := logs.NewLogger("activity-test", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	return &ActivityLogger{queries: db.New(sqlDB), file: file, logger: l}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestActivityLogger_WritesDBAndFile(t *testing.T) {
	script := &actConnScript{}
	var fileBuf bytes.Buffer
	a := newTestActivityLogger(t, script, &fileBuf)

	a.LogTiketVytvoren(context.Background(), 7, 42, "Projektor nefunguje")

	if len(script.lastArgs) == 0 {
		t.Fatal("expected CreateActivityLog exec to be called")
	}
	if got := script.lastArgs[0]; got != string(EventTiketVytvoren) {
		t.Errorf("event_type = %v, want %v", got, EventTiketVytvoren)
	}

	line := strings.TrimSpace(fileBuf.String())
	var decoded map[string]any
	if err := json.Unmarshal([]byte(line), &decoded); err != nil {
		t.Fatalf("file log line is not valid JSON: %v (line: %q)", err, line)
	}
	if decoded["event_type"] != string(EventTiketVytvoren) {
		t.Errorf("file log event_type = %v, want %v", decoded["event_type"], EventTiketVytvoren)
	}
}

func TestActivityLogger_DBFailure_StillWritesFileAndDoesNotPanic(t *testing.T) {
	script := &actConnScript{execErr: fmt.Errorf("boom")}
	var fileBuf bytes.Buffer
	a := newTestActivityLogger(t, script, &fileBuf)

	a.LogTiketSmazan(context.Background(), 1, 99, "Rozbitá židle")

	if fileBuf.Len() == 0 {
		t.Error("expected file write to still happen after a DB failure")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, fmt.Errorf("disk full") }

func TestActivityLogger_FileFailure_StillWritesDBAndDoesNotPanic(t *testing.T) {
	script := &actConnScript{}
	a := newTestActivityLogger(t, script, nil)
	a.file = failingWriter{}

	a.LogKomentarVytvoren(context.Background(), 3, 10, 42)

	if len(script.lastArgs) == 0 {
		t.Error("expected DB write to still happen after a file failure")
	}
}

func TestActivityLogger_NilLogger_IsNoOp(t *testing.T) {
	var a *ActivityLogger
	a.LogTiketVytvoren(context.Background(), 1, 1, "x")
	a.LogUzivatelSchvalen(context.Background(), 1, 2, "a@b.cz")
}

func TestActivityLogger_AllEventTypes(t *testing.T) {
	cases := []struct {
		name string
		call func(a *ActivityLogger)
		want EventType
	}{
		{"created", func(a *ActivityLogger) { a.LogTiketVytvoren(context.Background(), 1, 1, "t") }, EventTiketVytvoren},
		{"updated", func(a *ActivityLogger) { a.LogTiketAktualizovan(context.Background(), 1, 1, []string{"title"}) }, EventTiketAktualizovan},
		{"status_changed", func(a *ActivityLogger) { a.LogStavZmenen(context.Background(), 1, 1, "open", "resolved") }, EventTiketStavZmenen},
		{"assigned", func(a *ActivityLogger) {
			newID := int32(5)
			a.LogTiketPrirazen(context.Background(), 1, 1, nil, &newID)
		}, EventTiketPrirazen},
		{"deleted", func(a *ActivityLogger) { a.LogTiketSmazan(context.Background(), 1, 1, "t") }, EventTiketSmazan},
		{"comment_created", func(a *ActivityLogger) { a.LogKomentarVytvoren(context.Background(), 1, 2, 1) }, EventKomentarVytvoren},
		{"comment_updated", func(a *ActivityLogger) { a.LogKomentarAktualizovan(context.Background(), 1, 2, 1) }, EventKomentarAktualizovan},
		{"comment_deleted", func(a *ActivityLogger) { a.LogKomentarSmazan(context.Background(), 1, 2, 1) }, EventKomentarSmazan},
		{"user_registered", func(a *ActivityLogger) { a.LogUzivatelRegistrovan(context.Background(), 9, "a@b.cz") }, EventUzivatelRegistrovan},
		{"user_approved", func(a *ActivityLogger) { a.LogUzivatelSchvalen(context.Background(), 1, 9, "a@b.cz") }, EventUzivatelSchvalen},
		{"user_rejected", func(a *ActivityLogger) { a.LogUzivatelZamitnuv(context.Background(), 1, 9, "a@b.cz") }, EventUzivatelZamitnuv},
		{"user_deactivated", func(a *ActivityLogger) { a.LogUzivatelDeaktivovan(context.Background(), 1, 9, "a@b.cz") }, EventUzivatelDeaktivovan},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			script := &actConnScript{}
			var fileBuf bytes.Buffer
			a := newTestActivityLogger(t, script, &fileBuf)

			c.call(a)

			if len(script.lastArgs) == 0 {
				t.Fatal("expected CreateActivityLog exec to be called")
			}
			if got := script.lastArgs[0]; got != string(c.want) {
				t.Errorf("event_type = %v, want %v", got, c.want)
			}
		})
	}
}
