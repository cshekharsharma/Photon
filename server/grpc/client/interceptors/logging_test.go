package interceptors

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
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

func TestLoggingInterceptor_WithRealLogger(t *testing.T) {
	t.Run("LogsSuccessfulRequest", func(t *testing.T) {
		byteBuff := &bytes.Buffer{}
		log := getLogger("LogsSuccessfulRequest", byteBuff)

		interceptor := LoggingInterceptor(log)

		invoker := func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		err := interceptor(context.Background(), "/test.Service/Success", nil, nil, nil, invoker)
		assert.NoError(t, err)

		time.Sleep(10 * time.Millisecond)
		logOutput := byteBuff.String()
		assert.Contains(t, logOutput, `test.Service/Success`)
		assert.Contains(t, logOutput, `request completed`)
		assert.Contains(t, logOutput, `duration`)
		assert.NotContains(t, logOutput, `error`)
	})

	t.Run("LogsFailedRequest", func(t *testing.T) {
		byteBuff := &bytes.Buffer{}
		log := getLogger("LogsFailedRequest", byteBuff)
		interceptor := LoggingInterceptor(log)

		invoker := func(ctx context.Context, method string, req, reply interface{},
			cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			time.Sleep(5 * time.Millisecond)
			return errors.New("mock failure")
		}

		err := interceptor(context.Background(), "/test.Service/Fail", nil, nil, nil, invoker)
		assert.Error(t, err)

		logOutput := byteBuff.String()
		assert.Contains(t, logOutput, `test.Service/Fail`)
		assert.Contains(t, logOutput, `request completed`)
		assert.Contains(t, logOutput, `error`)
		assert.Contains(t, logOutput, `mock failure`)
	})
}
