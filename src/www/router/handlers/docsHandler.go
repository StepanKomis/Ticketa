package handlers

import (
	"io/fs"
	"net/http"
	"strings"
)

type DocsHandler struct {
	fs http.FileSystem
}

func NewDocsHandler(files fs.FS) (*DocsHandler, error) {
	sub, err := fs.Sub(files, "docs")
	if err != nil {
		return nil, err
	}
	return &DocsHandler{fs: http.FS(sub)}, nil
}

func (h *DocsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r2 := r.Clone(r.Context())
	r2.URL.Path = strings.TrimPrefix(r.URL.Path, "/docs")
	if r2.URL.Path == "" {
		r2.URL.Path = "/"
	}
	http.FileServer(h.fs).ServeHTTP(w, r2)
}
