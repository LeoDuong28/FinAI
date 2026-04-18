package middleware

import (
	"net/http"
	"strings"
)

// ContentType enforces that write requests have an acceptable Content-Type.
func ContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			ct := r.Header.Get("Content-Type")
			if ct == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnsupportedMediaType)
				w.Write([]byte(`{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type header required"}}`))
				return
			}

			ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
			allowed := ct == "application/json" ||
				ct == "application/x-www-form-urlencoded" ||
				ct == "multipart/form-data"

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnsupportedMediaType)
				w.Write([]byte(`{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Unsupported Content-Type"}}`))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
