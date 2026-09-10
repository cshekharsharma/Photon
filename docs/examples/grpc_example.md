# gRPC

Photon wraps gRPC startup and clients with defaults for interceptors, health checks, retries, backoff, TLS, and graceful shutdown.

```go
package examples

import (
	"context"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	grpcclient "github.com/cshekharsharma/photon/server/grpc/client"
	grpcserver "github.com/cshekharsharma/photon/server/grpc/server"
	"google.golang.org/grpc"
)

func StartGRPCServer() error {
	log := logger.Init(&logger.LoggerConfig{
		Name:     "grpc-server",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelInfo,
	})

	return grpcserver.StartGRPCServer(&grpcserver.ServerOptions{
		Port:             9090,
		Insecure:         true,
		EnableReflection: true,
		ShutdownTimeout:  10 * time.Second,
		ServerLogger:     log,
		RegisterFunc: func(server *grpc.Server) {
			// Register generated protobuf services here.
			// examplepb.RegisterCheckoutServiceServer(server, checkoutService)
		},
	})
}

func NewGRPCClient(ctx context.Context) (grpcclient.Client, error) {
	log := logger.Init(&logger.LoggerConfig{
		Name:     "grpc-client",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelInfo,
	})

	client, err := grpcclient.NewGRPCClient(&grpcclient.ClientOptions{
		Target:            "localhost:9090",
		Insecure:          true,
		MinConnectTimeout: 500 * time.Millisecond,
		MaxRetries:        3,
		BackoffConfig: &grpcclient.BackoffConfig{
			BaseDelay:  100 * time.Millisecond,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   2 * time.Second,
		},
		Logger: log,
	})
	if err != nil {
		return nil, err
	}

	healthCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := client.HealthCheck(healthCtx); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
```

Use `TLSConfig` instead of `Insecure` outside local development.

