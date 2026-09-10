# Auth Middleware

Compose normal `net/http` middleware with Photon's session manager.

```go
package examples

import (
	"net/http"

	"github.com/cshekharsharma/photon/core/session"
)

func AuthMiddleware(mgr *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return mgr.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			exists, err := mgr.Exists(r.Context(), "user_id")
			if err != nil || !exists {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}
```
