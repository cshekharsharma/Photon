package middleware

import (
	"net/http"
	"strings"

	"github.com/cshekharsharma/photon/utils/rest"
)

// Heartbeat endpoint middleware useful to setting up a path like
// `/ping` that load balancers or uptime testing external services
// can make a request before hitting any routes. It's also convenient
// to place this above ACL middlewares as well.
func Heartbeat(endpoint string) func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if (r.Method == http.MethodGet || r.Method == http.MethodHead) && strings.EqualFold(r.URL.Path, endpoint) {
				w.Header().Set(rest.HeaderContentType, rest.ContentTypePlainText)
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(".")); err != nil {
					return
				}
				return
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	return f
}
