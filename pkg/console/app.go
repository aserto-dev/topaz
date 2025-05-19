package console

import (
	"net/http"

	gorilla "github.com/gorilla/mux"

	console "github.com/aserto-dev/go-topaz-ui"
)

// uiFS is a http.FileSystem that always returns the same file ('console/index.html').
type uiFS struct {
	http.FileSystem
}

func (fs uiFS) Open(_ string) (http.File, error) {
	return fs.FileSystem.Open("console/index.html")
}

// publicFS is a http.FileSystem that replaces the `/public/` prefix with `console/`.
type publicFS struct {
	http.FileSystem
}

func (fs publicFS) Open(name string) (http.File, error) {
	return fs.FileSystem.Open("console" + name)
}

func RegisterAppHandlers(router *gorilla.Router) {
	// All paths that start with '/ui/' serve the same content ('console/index.html').
	router.PathPrefix("/ui/").Handler(http.FileServer(uiFS{http.FS(console.FS)}))

	// Redirect 'GET /' to the directory model page.
	router.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/directory/model", http.StatusSeeOther)
	})

	router.PathPrefix("/assets/").Handler(http.FileServer(publicFS{http.FS(console.FS)}))
}
