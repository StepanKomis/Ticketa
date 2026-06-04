package router

import (
	"io/fs"
	"net/http"

	"github.com/StepanKomis/Ticketa/src/www/router/handlers"
)

func NewRouter(staticFiles fs.FS) *http.ServeMux {
	sub, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("embed: cannot sub into static/: " + err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/", handlers.NewStaticHandler(sub))
	return mux
}
