package handlers_test

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	db "github.com/StepanKomis/Ticketa/src/database/postgres/queries"
	"github.com/StepanKomis/Ticketa/src/internal/ctxkeys"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

func withUser(r *http.Request, u db.User) *http.Request {
	ctx := context.WithValue(r.Context(), ctxkeys.UserContextKey, u)
	return r.WithContext(ctx)
}

func activityRowNoActor(id int64, eventType string, createdAt time.Time) hdlQueryResult {
	return hdlQueryResult{
		cols: []string{"id", "event_type", "actor_id", "target_type", "target_id", "payload", "created_at"},
		rows: [][]driver.Value{
			{id, eventType, nil, nil, nil, nil, createdAt},
		},
	}
}

func activityCountRow(n int64) hdlQueryResult {
	return hdlQueryResult{cols: []string{"count"}, rows: [][]driver.Value{{n}}}
}

func TestActivityHandler_ListForUser_Self_Returns200(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{
			activityRowNoActor(1, "tiket_vytvoren", time.Now()),
			activityCountRow(1),
		},
	}
	sqlDB := newHandlerDB(t, script)
	h := handlers.NewActivityHandler(db.New(sqlDB), mustTestLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/api/users/7/activity", nil)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeStudent})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var got struct {
		Items  []map[string]any `json:"items"`
		Total  int64            `json:"total"`
		Limit  int              `json:"limit"`
		Offset int              `json:"offset"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(got.Items) != 1 || got.Total != 1 {
		t.Errorf("expected 1 item and total=1, got items=%d total=%d", len(got.Items), got.Total)
	}
}

func TestActivityHandler_ListForUser_OtherUserNonAdmin_Returns403(t *testing.T) {
	h := handlers.NewActivityHandler(nil, mustTestLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/api/users/9/activity", nil)
	req = withUser(req, db.User{ID: 7, UserType: db.UserTypeStudent})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestActivityHandler_ListForUser_OtherUserAdmin_Returns200(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{
			activityRowNoActor(1, "uzivatel_schvalen", time.Now()),
			activityCountRow(1),
		},
	}
	sqlDB := newHandlerDB(t, script)
	h := handlers.NewActivityHandler(db.New(sqlDB), mustTestLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/api/users/9/activity", nil)
	req = withUser(req, db.User{ID: 1, UserType: db.UserTypeAdmin})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestActivityHandler_ListGlobal_DefaultAndClampedPagination(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"defaults", "", 50, 0},
		{"custom", "?limit=10&offset=5", 10, 5},
		{"clamped", "?limit=9999", 200, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			script := &hdlConnScript{
				queries: []hdlQueryResult{
					activityRowNoActor(1, "tiket_vytvoren", time.Now()),
					activityCountRow(1),
				},
			}
			sqlDB := newHandlerDB(t, script)
			h := handlers.NewActivityHandler(db.New(sqlDB), mustTestLogger(t))

			req := httptest.NewRequest(http.MethodGet, "/api/activity"+c.query, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
			}
			var got struct {
				Limit  int `json:"limit"`
				Offset int `json:"offset"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if got.Limit != c.wantLimit || got.Offset != c.wantOffset {
				t.Errorf("limit/offset = %d/%d, want %d/%d", got.Limit, got.Offset, c.wantLimit, c.wantOffset)
			}
		})
	}
}

func TestActivityHandler_ListGlobal_EventTypeFilterDoesNotError(t *testing.T) {
	script := &hdlConnScript{
		queries: []hdlQueryResult{
			activityRowNoActor(1, "komentar_vytvoren", time.Now()),
			activityCountRow(1),
		},
	}
	sqlDB := newHandlerDB(t, script)
	h := handlers.NewActivityHandler(db.New(sqlDB), mustTestLogger(t))

	req := httptest.NewRequest(http.MethodGet, "/api/activity?event_type=komentar_vytvoren", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func mustTestLogger(t *testing.T) *logs.Logger {
	t.Helper()
	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()
	l, err := logs.NewLogger("activity", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	return l
}
