package middleware

import (
	"net/http"
	"path"
)

func CleanPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" {
			r.URL.Path = path.Clean(r.URL.Path)
		}

		next.ServeHTTP(w, r)
	})
}
