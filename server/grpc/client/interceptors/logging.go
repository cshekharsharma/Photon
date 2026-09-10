package interceptors

import (
	"context"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"google.golang.org/grpc"
)

// LoggingInterceptor logs each outgoing gRPC client call with duration and status.
func LoggingInterceptor(logger logger.Logger) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {

		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		fields := map[string]interface{}{
			"method":   method,
			"duration": duration,
		}

		if err != nil {
			fields["error"] = err.Error()
		}

		logger.DebugWithFields(fields, "request completed")
		return err
	}
}
