package handlers

import (
	"io/fs"
	"net/http"
)

type StaticHandler struct {
	fs http.FileSystem
}

func NewStaticHandler(files fs.FS) *StaticHandler {
	return &StaticHandler{fs: http.FS(files)}
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	f, err := h.fs.Open(path)
	if err != nil {
		// Unknown path — let the SPA router handle it
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		http.FileServer(h.fs).ServeHTTP(w, r2)
		return
	}
	f.Close()

	http.FileServer(h.fs).ServeHTTP(w, r)
}
