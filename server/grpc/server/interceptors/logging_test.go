package interceptors

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestLoggingInterceptor_Success(t *testing.T) {
	buff := &bytes.Buffer{}
	interceptor := LoggingInterceptor(getLogger("TestLoggingInterceptor_Success", buff))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "logged", nil
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Logged",
	}, handler)

	assert.NoError(t, err)
	assert.Equal(t, "logged", resp)
}

func TestLoggingInterceptor_Error(t *testing.T) {
	buff := &bytes.Buffer{}
	interceptor := LoggingInterceptor(getLogger("TestLoggingInterceptor_Error", buff))

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		time.Sleep(5 * time.Millisecond)
		return nil, errors.New("test error")
	}

	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
		FullMethod: "/test.Fail",
	}, handler)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "test error")
}
