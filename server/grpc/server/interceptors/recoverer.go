package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/cshekharsharma/photon/core/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoverInterceptor returns a unary interceptor that recovers from panics
// and converts them into gRPC internal errors with stack trace logging.
func RecoverInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorWithFields(map[string]interface{}{
					"panic":  r,
					"stack":  string(debug.Stack()),
					"method": info.FullMethod,
				}, "panic recovered in unary gRPC call")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamRecoverInterceptor recovers from panics in streaming RPC handlers.
func StreamRecoverInterceptor(log logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				log.ErrorWithFields(map[string]interface{}{
					"panic":  r,
					"stack":  string(debug.Stack()),
					"method": info.FullMethod,
				}, "panic recovered in streaming gRPC call")
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}
