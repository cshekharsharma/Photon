# HTTP Service

A minimal HTTP service using Photon's server wrapper. `StartHttpServer` normalizes defaults, validates the config, applies middleware, and starts the listener.

```go
package examples

import (
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/core/router"
	"github.com/cshekharsharma/photon/middleware"
	server "github.com/cshekharsharma/photon/server/http"
)

func MainExample() {
	appLogger := logger.Init(&logger.LoggerConfig{
		Name:     "api",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelInfo,
	})

	server.StartHttpServer(&server.ServerConfig{
		ServerPort:    8080,
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  10 * time.Second,
		IdleTimeout:   120 * time.Second,
		RouteProvider: router.RouterCHI,
		HttpRoutes: []*server.HttpRoute{
			{
				UrlRoute:      "/health",
				RequestMethod: http.MethodGet,
				HttpHandler:   health,
				Middlewares: []middleware.MiddlewareFn{
					middleware.ToMiddlewareFn(requireAPIKey),
				},
			},
		},
		AccessLogger: appLogger,
		ErrorLogger:  appLogger,
		ServerLogger: appLogger,
	})
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
```

