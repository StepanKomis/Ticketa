package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

func newDocsFS() fstest.MapFS {
	return fstest.MapFS{
		"docs/index.html":    {Data: []byte("<html>swagger</html>")},
		"docs/openapi.yaml":  {Data: []byte("openapi: 3.0.3\n")},
	}
}

func TestDocsHandler_Root_ServesIndexHTML(t *testing.T) {
	h, err := handlers.NewDocsHandler(newDocsFS())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/docs/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "<html>swagger</html>" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestDocsHandler_NoDocs_PathServesIndexHTML(t *testing.T) {
	h, err := handlers.NewDocsHandler(newDocsFS())
	if err != nil {
		t.Fatal(err)
	}

	// /docs without trailing slash — after TrimPrefix results in empty string → "/"
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// FileServer will either serve 200 or redirect to /docs/ — either is acceptable
	if rr.Code != http.StatusOK && rr.Code/100 != 3 {
		t.Errorf("expected 200 or redirect, got %d", rr.Code)
	}
}

func TestDocsHandler_Spec_ServesYAML(t *testing.T) {
	h, err := handlers.NewDocsHandler(newDocsFS())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/docs/openapi.yaml", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if body := rr.Body.String(); body != "openapi: 3.0.3\n" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestDocsHandler_MissingFile_Returns404(t *testing.T) {
	h, err := handlers.NewDocsHandler(newDocsFS())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/docs/nonexistent.js", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
