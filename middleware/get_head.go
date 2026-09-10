package middleware

import "net/http"

func GetHead(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			r.Method = http.MethodGet
		}

		next.ServeHTTP(w, r)
	})
}
