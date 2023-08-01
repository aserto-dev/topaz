package rapidoc

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path"
)

// Opts configures the RapiDoc middlewares.
type Opts struct {
	// BasePath for the UI path, defaults to: / .
	BasePath string
	// Path combines with BasePath for the full UI path, defaults to: docs.
	Path string
	// SpecURL the url to find the spec for.
	SpecURL string
	// RapiDocURL for the js that generates the rapidoc site, defaults to: https://cdn.jsdelivr.net/npm/rapidoc/bundles/rapidoc.standalone.js .
	RapiDocURL string
	// Title for the documentation site, default to: API documentation.
	Title string
}

// EnsureDefaults in case some options are missing.
func (r *Opts) EnsureDefaults() {
	if r.BasePath == "" {
		r.BasePath = "/"
	}
	if r.Path == "" {
		r.Path = "docs"
	}
	if r.SpecURL == "" {
		r.SpecURL = "/openapi.json"
	}
	if r.RapiDocURL == "" {
		r.RapiDocURL = rapidocLatest
	}
	if r.Title == "" {
		r.Title = "API documentation"
	}
}

// Handler creates a middleware to serve a documentation site for a swagger spec.
// This allows for altering the spec before starting the http listener.
func Handler(opts *Opts, next http.Handler) http.Handler {
	opts.EnsureDefaults()

	pth := path.Join(opts.BasePath, opts.Path)
	tmpl := template.Must(template.New("rapidoc").Parse(rapidocTemplate))

	buf := bytes.NewBuffer(nil)
	_ = tmpl.Execute(buf, opts)
	b := buf.Bytes()

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path == pth {
			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusOK)

			_, _ = rw.Write(b)
			return
		}

		if next == nil {
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(rw, "%q not found", pth)
			return
		}
		next.ServeHTTP(rw, r)
	})
}

const (
	rapidocLatest   = "https://unpkg.com/rapidoc/dist/rapidoc-min.js"
	rapidocTemplate = `<!doctype html>
<html>
<head>
  <title>{{ .Title }}</title>
  <meta charset="utf-8"> <!-- Important: rapi-doc uses utf8 characters -->
  <script type="module" src="{{ .RapiDocURL }}"></script>
</head>
<body>
  <rapi-doc
	spec-url="{{ .SpecURL }}"
	theme="dark"
	bg-color = "#161719"
  	text-color = "#bbb"
	header-color = "#444444"
	primary-color = "#8a959f"
	regular-font = "-apple-system, BlinkMacSystemFont, 'Roboto', 'Oxygen','Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue', sans-serif"
	mono-font = "'Ubuntu Mono', source-code-pro, Menlo, Monaco, Consolas, 'Courier New', monospace"
	layout = "column"
	render-style = "read"
	schema-style = "tree"
	show-header = "true"
	allow-spec-url-load = "false"
	allow-spec-file-load = "false"
	allow-server-selection = "false"
  >
  <div slot="logo" style="display: flex; align-items: center; justify-content: center;">
  	<img src = "https://www.aserto.com/images/Aserto-logo-color-120px.png" style="width:120px; margin-right: 40px"> <span style="color:#fff"></span>
  </div>
  </rapi-doc>
</body>
</html>
`
)
