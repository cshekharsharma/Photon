package interceptors

import (
	"bytes"
	"context"
	"testing"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getLogger(name string, buff *bytes.Buffer) logger.Logger {
	return logger.Init(&logger.LoggerConfig{
		Provider: logger.LoggerProviderZerolog,
		Name:     name,
		Level:    logger.LogLevelDebug,
		Type:     logger.LoggerTypeStdout,
		Writer:   buff,
	})
}

func TestRecoverInterceptor_NoPanic(t *testing.T) {
	buff := &bytes.Buffer{}
	interceptor := RecoverInterceptor(getLogger("TestRecoverInterceptor_NoPanic", buff))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.SafeMethod",
	}, handler)

	assert.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestRecoverInterceptor_WithPanic(t *testing.T) {
	buff := &bytes.Buffer{}
	interceptor := RecoverInterceptor(getLogger("TestRecoverInterceptor_WithPanic", buff))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("something broke")
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.PanicMethod",
	}, handler)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "internal server error")
}

func TestStreamRecoverInterceptor_NoPanic(t *testing.T) {
	buff := &bytes.Buffer{}
	interceptor := StreamRecoverInterceptor(getLogger("TestStreamRecoverInterceptor_NoPanic", buff))

	err := interceptor(nil, &testServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{
		FullMethod: "/test.StreamSafeMethod",
	}, func(_ interface{}, _ grpc.ServerStream) error {
		return nil
	})

	assert.NoError(t, err)
}

func TestStreamRecoverInterceptor_WithPanic(t *testing.T) {
	buff := &bytes.Buffer{}
	interceptor := StreamRecoverInterceptor(getLogger("TestStreamRecoverInterceptor_WithPanic", buff))

	err := interceptor(nil, &testServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{
		FullMethod: "/test.StreamPanicMethod",
	}, func(_ interface{}, _ grpc.ServerStream) error {
		panic("stream broke")
	})

	assert.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "internal server error")
}
