package directory

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
	"sync"
)

//go:embed directory.openapi.json
var directory embed.FS

var (
	loadOnce sync.Once
	errLoad  error
	etag     string
	buf      []byte
)

func loadOpenAPI() ([]byte, error) {
	loadOnce.Do(func() {
		buf, errLoad = fs.ReadFile(directory, "directory.openapi.json")
		sum := sha256.Sum256(buf)
		etag = `"` + hex.EncodeToString(sum[:]) + `"`
	})

	return buf, errLoad
}

func OpenAPIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if match := r.Header.Get("If-None-Match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		data, err := loadOpenAPI()
		if err != nil {
			http.Error(w, "failed to load OpenAPI spec", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
