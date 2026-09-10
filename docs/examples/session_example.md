# Session

Configure the session manager once, then attach its middleware to your router.

```go
package examples

import (
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/core/session"
)

func InitSession() (*session.Manager, error) {
	mgr, err := session.New(&session.Config{
		Store: session.StoreConfig{
			Type: session.StoreRedis,
			Options: session.StoreOptions{
				Address: "127.0.0.1:6379",
			},
		},
		Cookie: session.CookieConfig{
			Name:     "photon_session",
			Path:     "/",
			Secure:   session.Bool(true),
			HttpOnly: session.Bool(true),
			SameSite: http.SameSiteLaxMode,
			Persist:  session.Bool(true),
		},
		Encoding: session.EncodingJSON,
		Lifetime: 7 * 24 * time.Hour,
	})
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

// Use with your router: router.Use(mgr.Middleware())
func ExampleHandler(mgr *session.Manager, w http.ResponseWriter, r *http.Request) {
	_ = mgr.Put(r.Context(), "KeyName", "Value")
	_ = mgr.Renew(r.Context())
	_ = mgr.Destroy(r.Context())
}
```
