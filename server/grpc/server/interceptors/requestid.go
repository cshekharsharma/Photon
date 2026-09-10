package interceptors

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type grpcRequestIDKey string

const requestIDKey grpcRequestIDKey = "x-request-id"

// RequestIDInterceptor ensures that a request ID is present in the context metadata.
// If one is missing, it generates a new UUID and injects it.
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		rid := GetRequestID(md)
		if rid == "" {
			rid = uuid.New().String()
			md.Set(string(requestIDKey), rid)
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		// Save request ID in context for downstream use
		ctx = context.WithValue(ctx, requestIDKey, rid)
		return handler(ctx, req)
	}
}

// getRequestID extracts request ID from metadata.
func GetRequestID(md metadata.MD) string {
	ids := md.Get(string(requestIDKey))
	if len(ids) > 0 && ids[0] != "" {
		return strings.TrimSpace(ids[0])
	}
	return ""
}
