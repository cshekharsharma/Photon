# Retry With Backoff

Use the backoff helper for transient operations such as downstream HTTP calls, queue publishes, or eventually consistent reads.

```go
package examples

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/coordination/backoff"
	"github.com/cshekharsharma/photon/core/logger"
)

var errPermanent = errors.New("permanent failure")

func RetryExample(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
	log := logger.Init(&logger.LoggerConfig{
		Name:     "retry",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelInfo,
	})

	retry := backoff.NewBackoff(
		backoff.WithMinDelay(100*time.Millisecond),
		backoff.WithMaxDelay(2*time.Second),
		backoff.WithMaxRetries(3),
		backoff.WithFactor(2),
		backoff.WithJitter(true),
		backoff.WithPerAttemptTimeout(500*time.Millisecond),
		backoff.WithLogger(log),
		backoff.WithRetryIf(func(err error) bool {
			return !errors.Is(err, errPermanent)
		}),
		backoff.WithMetrics(func(event backoff.RetryContext) {
			log.Info("retry_attempt=%d delay=%s err=%v", event.Attempt, event.DelayUsed, event.Error)
		}),
	)

	value, err := retry.Retry(ctx, func(ctx context.Context, args ...any) (any, error) {
		request := args[0].(*http.Request).Clone(ctx)
		resp, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= http.StatusInternalServerError {
			_ = resp.Body.Close()
			return nil, errors.New("transient downstream failure")
		}
		if resp.StatusCode >= http.StatusBadRequest {
			_ = resp.Body.Close()
			return nil, errPermanent
		}
		return resp, nil
	}, req)
	if err != nil {
		return nil, err
	}

	return value.(*http.Response), nil
}
```

