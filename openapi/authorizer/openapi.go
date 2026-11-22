package authorizer

import (
	_ "embed"
	"net/http"
	"sync"
	"text/template"
)

type Server struct {
	Scheme string
	Host   string
}

//go:embed authorizer.openapi.json
var staticString string

var buildTemplate = sync.OnceValues(func() (*template.Template, error) {
	return template.New("openapi.json").Parse(staticString)
})

// Static string value of the openapi.json file.
func Static() string {
	return staticString
}

// OpenAPIHandler, http handler to serve the OpenAPI specification file.
func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	template, err := buildTemplate()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}

	server := Server{
		Host:   r.Host,
		Scheme: scheme(r),
	}

	err = template.Execute(w, server)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}
}

func scheme(r *http.Request) string {
	scheme := r.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if r.TLS == nil {
			scheme = "http"
		} else {
			scheme = "https"
		}
	}

	return scheme
}
