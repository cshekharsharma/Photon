package server

import (
	"testing"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
)

func TestDefaultServerOptions_AllScenarios(t *testing.T) {
	t.Run("NilInput", func(t *testing.T) {
		opts := DefaultServerOptions(nil)

		assert.Equal(t, 9090, opts.Port)
		assert.Equal(t, 10*time.Second, opts.ShutdownTimeout)
		assert.NotNil(t, opts.ServerLogger)
		assert.Equal(t, 4*1024*1024, opts.MaxRecvMsgSize)
		assert.Equal(t, 4*1024*1024, opts.MaxSendMsgSize)
		assert.Equal(t, 2*time.Hour, opts.KeepaliveTime)
		assert.Equal(t, 20*time.Second, opts.KeepaliveTimeout)
	})

	t.Run("PartialInput", func(t *testing.T) {
		customLogger := logger.Init(&logger.LoggerConfig{
			Provider: logger.LoggerProviderZerolog,
			Name:     "custom",
			Type:     logger.LoggerTypeStdout,
		})

		input := &ServerOptions{
			Port:             8443,
			ShutdownTimeout:  3 * time.Second,
			ServerLogger:     customLogger,
			MaxRecvMsgSize:   1024,
			MaxSendMsgSize:   2048,
			KeepaliveTime:    time.Minute,
			KeepaliveTimeout: 5 * time.Second,
		}

		opts := DefaultServerOptions(input)

		assert.Equal(t, 8443, opts.Port)
		assert.Equal(t, 3*time.Second, opts.ShutdownTimeout)
		assert.Equal(t, customLogger, opts.ServerLogger)
		assert.Equal(t, 1024, opts.MaxRecvMsgSize)
		assert.Equal(t, 2048, opts.MaxSendMsgSize)
		assert.Equal(t, time.Minute, opts.KeepaliveTime)
		assert.Equal(t, 5*time.Second, opts.KeepaliveTimeout)
	})

	t.Run("WithZeroValues", func(t *testing.T) {
		opts := DefaultServerOptions(&ServerOptions{
			Port:               0,
			ShutdownTimeout:    0,
			ServerLogger:       nil,
			MaxRecvMsgSize:     0,
			MaxSendMsgSize:     0,
			KeepaliveTime:      0,
			KeepaliveTimeout:   0,
			RegisterFunc:       nil,
			UnaryInterceptors:  nil,
			StreamInterceptors: nil,
		})

		assert.Equal(t, 9090, opts.Port)
		assert.Equal(t, 10*time.Second, opts.ShutdownTimeout)
		assert.NotNil(t, opts.ServerLogger)
		assert.Equal(t, 4*1024*1024, opts.MaxRecvMsgSize)
		assert.Equal(t, 4*1024*1024, opts.MaxSendMsgSize)
		assert.Equal(t, 2*time.Hour, opts.KeepaliveTime)
		assert.Equal(t, 20*time.Second, opts.KeepaliveTimeout)
	})
}
