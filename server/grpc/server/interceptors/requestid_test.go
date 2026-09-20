package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (s *testServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (s *testServerStream) SetTrailer(metadata.MD) {
}

func (s *testServerStream) Context() context.Context {
	return s.ctx
}

func (s *testServerStream) SendMsg(interface{}) error {
	return nil
}

func (s *testServerStream) RecvMsg(interface{}) error {
	return nil
}

func TestRequestIDInterceptor_GeneratesNewID(t *testing.T) {
	interceptor := RequestIDInterceptor()
	ctx := context.Background()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		rid := ctx.Value(requestIDKey)
		assert.NotNil(t, rid)
		assert.IsType(t, "", rid)
		assert.NotEmpty(t, rid)
		return rid, nil
	}

	rid, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.MissingRID"}, handler)
	assert.NoError(t, err)
	assert.NotEmpty(t, rid)
}

func TestRequestIDInterceptor_PreservesExistingID(t *testing.T) {
	existingRID := "test-id-123"
	md := metadata.Pairs(string(requestIDKey), existingRID)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := RequestIDInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		rid := ctx.Value(requestIDKey)
		assert.Equal(t, existingRID, rid)
		return rid, nil
	}

	rid, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.ExistingRID"}, handler)
	assert.NoError(t, err)
	assert.Equal(t, existingRID, rid)
}

func TestStreamRequestIDInterceptor_GeneratesNewID(t *testing.T) {
	interceptor := StreamRequestIDInterceptor()
	stream := &testServerStream{ctx: context.Background()}

	err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test.StreamMissingRID"}, func(_ interface{}, ss grpc.ServerStream) error {
		rid := ss.Context().Value(requestIDKey)
		assert.NotNil(t, rid)
		assert.IsType(t, "", rid)
		assert.NotEmpty(t, rid)
		return nil
	})

	assert.NoError(t, err)
}

func TestStreamRequestIDInterceptor_PreservesExistingID(t *testing.T) {
	existingRID := "stream-test-id-123"
	md := metadata.Pairs(string(requestIDKey), existingRID)
	stream := &testServerStream{ctx: metadata.NewIncomingContext(context.Background(), md)}

	interceptor := StreamRequestIDInterceptor()
	err := interceptor(nil, stream, &grpc.StreamServerInfo{FullMethod: "/test.StreamExistingRID"}, func(_ interface{}, ss grpc.ServerStream) error {
		assert.Equal(t, existingRID, ss.Context().Value(requestIDKey))
		return nil
	})

	assert.NoError(t, err)
}
