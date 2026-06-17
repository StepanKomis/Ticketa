package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/StepanKomis/Ticketa/src/cmd/server/logs"
	"github.com/StepanKomis/Ticketa/src/config"
	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

func newTestAdminHandler(t *testing.T) (*handlers.AdminHandler, *config.Store) {
	t.Helper()

	cfg := config.Defaults()
	cfg.Logging.Dir = t.TempDir()

	cfgPath := filepath.Join(t.TempDir(), "ticketa.yaml")
	if err := config.Save(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	store := config.NewStore(cfg, cfgPath)

	l, err := logs.NewLogger("admin", cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })

	return handlers.NewAdminHandler(nil, store, l, nil), store
}

func TestAdminHandler_GetConfig(t *testing.T) {
	h, _ := newTestAdminHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/config", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var got config.Config
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.Logging.Level != "info" {
		t.Errorf("expected default level 'info', got %q", got.Logging.Level)
	}
}

func TestAdminHandler_PatchConfig_LogLevel(t *testing.T) {
	h, store := newTestAdminHandler(t)

	body := `{"logging": {"level": "debug"}}`
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/config", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if store.Get().Logging.Level != "debug" {
		t.Errorf("expected in-memory level 'debug', got %q", store.Get().Logging.Level)
	}

	loaded, err := config.Load(store.Path())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Logging.Level != "debug" {
		t.Errorf("expected persisted level 'debug', got %q", loaded.Logging.Level)
	}
}

func TestAdminHandler_PatchConfig_InvalidJSON(t *testing.T) {
	h, _ := newTestAdminHandler(t)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/config", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAdminHandler_UnknownRoute(t *testing.T) {
	h, _ := newTestAdminHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/unknown", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
