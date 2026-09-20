package interceptors

import (
	"context"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"google.golang.org/grpc"
)

// LoggingInterceptor logs each incoming gRPC request with duration and status.
// It uses the provided logger to log the information.
func LoggingInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		fields := map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration.String(),
		}
		if err != nil {
			fields["error"] = err.Error()
			log.ErrorWithFields(fields, "gRPC request failed")
		} else {
			log.InfoWithFields(fields, "gRPC request completed")
		}

		return resp, err
	}
}

// StreamLoggingInterceptor logs each incoming streaming gRPC request with duration and status.
func StreamLoggingInterceptor(log logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		fields := map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration.String(),
		}
		if err != nil {
			fields["error"] = err.Error()
			log.ErrorWithFields(fields, "gRPC stream failed")
		} else {
			log.InfoWithFields(fields, "gRPC stream completed")
		}

		return err
	}
}
