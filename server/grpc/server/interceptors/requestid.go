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
		return handler(contextWithRequestID(ctx), req)
	}
}

type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextServerStream) Context() context.Context {
	return s.ctx
}

// StreamRequestIDInterceptor ensures that a request ID is present for streaming RPCs.
func StreamRequestIDInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := contextWithRequestID(ss.Context())
		return handler(srv, &contextServerStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

func contextWithRequestID(ctx context.Context) context.Context {
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

	return context.WithValue(ctx, requestIDKey, rid)
}

// getRequestID extracts request ID from metadata.
func GetRequestID(md metadata.MD) string {
	ids := md.Get(string(requestIDKey))
	if len(ids) > 0 && ids[0] != "" {
		return strings.TrimSpace(ids[0])
	}
	return ""
}
