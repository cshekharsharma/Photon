# Rate Limiter

Use the Redis-backed limiter when rate-limit state must be shared across service instances.

```go
package examples

import (
	"net/http"
	"os"
	"time"

	"github.com/cshekharsharma/photon/coordination/concurrency/ratelimiter"
	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/middleware"
	"github.com/cshekharsharma/photon/storage/redis"
)

func RateLimitedHandler(next http.Handler) (http.Handler, error) {
	redis.SetConnectionConfig("rate-limit", &redis.ConnectionConfig{
		Address:  "localhost:6379",
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		Database: 0,
		PoolSize: 20,
	})

	redisClient, err := redis.Connect(&redis.RedisConnector{}, "rate-limit")
	if err != nil {
		return nil, err
	}

	limiter, err := ratelimiter.NewRateLimiter(&ratelimiter.Options{
		Type:      ratelimiter.RateLimiterTypeRedis,
		Client:    redisClient,
		MaxTokens: 100,
		Interval: 1 * time.Minute,
	})
	if err != nil {
		return nil, err
	}

	log := logger.Init(&logger.LoggerConfig{
		Name:     "rate-limiter",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelWarn,
	})

	return middleware.RateLimitMiddleware(limiter, log, 100*time.Millisecond)(next), nil
}
```

Photon's HTTP middleware keys by client IP. For user/API-key limits, create a custom middleware around `ratelimiter.Limiter`.

