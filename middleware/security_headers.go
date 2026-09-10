package middleware

import (
	"net/http"
)

// Add security http headers to the response object, to avoid the
// common security loopholes and threats. In future more sophisticated
// security hooks can also be added here.
func SecurityHeaders(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
