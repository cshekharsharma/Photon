# Logger

Initialize named loggers once at startup, then retrieve them by name where needed.

```go
package examples

import "github.com/cshekharsharma/photon/core/logger"

func LoggerExample() {
	accessLogger := logger.Init(&logger.LoggerConfig{
		Name:     "access",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelInfo,
	})

	accessLogger.Info("service started")
	logger.Get("access").Debug("request completed")
}
```

