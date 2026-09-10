package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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
