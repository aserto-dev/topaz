package middleware

import "net/http"

// FieldsMask is http middlware that sets the Content-Type to "application/json+masked", which
// signals the marshaler not to emit unpopulated types, which is needed to
// serialize the masked result set.
// This happens if a fields.mask query parameter is present and set.
func FieldsMask(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := r.URL.Query()["fields.mask"]; ok && len(p) > 0 && p[0] != "" {
			r.Header.Set("Content-Type", "application/json+masked")
		}

		h.ServeHTTP(w, r)
	})
}
