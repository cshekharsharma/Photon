# Middleware

Photon middleware uses standard `func(http.Handler) http.Handler` composition, so it works with normal `net/http`, Chi, and the Photon HTTP server wrapper.

```go
package examples

import (
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/middleware"
)

func HandlerWithMiddleware() http.Handler {
	log := logger.Init(&logger.LoggerConfig{
		Name:     "http",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelInfo,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetRequestID(r.Context())
		w.Header().Set("X-App-Request-ID", requestID)
		w.WriteHeader(http.StatusOK)
	})

	return chain(
		handler,
		middleware.RequestId,
		middleware.SecurityHeaders,
		middleware.Timeout(2*time.Second),
		middleware.Recoverer(log),
	)
}

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
```

