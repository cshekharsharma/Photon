package server

import (
	"crypto/tls"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"google.golang.org/grpc"
)

// ServerOptions defines the configuration for launching a gRPC server.
// It supports various tuning parameters such as interceptors, TLS credentials,
// reflection toggles, logging, shutdown behavior, and message size limits.
// This struct is intended to be passed into StartGRPCServer() to configure
// the server at startup.
type ServerOptions struct {
	Port                        int                            // Port on which the gRPC server will listen for incoming connections.
	Insecure                    bool                           // If true, the server will not use TLS.
	TLSConfig                   *tls.Config                    // TLSConfig is used to configure the server for secure connections.
	RegisterFunc                func(*grpc.Server)             // RegisterFunc is a callback to register application-specific services with the gRPC server.
	UnaryInterceptors           []grpc.UnaryServerInterceptor  // UnaryInterceptors are gRPC interceptors for unary RPCs.
	StreamInterceptors          []grpc.StreamServerInterceptor // StreamInterceptors are gRPC interceptors for streaming RPCs.
	EnableReflection            bool                           // If true, enables reflection for the gRPC server, useful for debugging and introspection.
	Environment                 string                         // Environment name, e.g. "production".
	AllowReflectionInProduction bool                           // Must be true to enable reflection when Environment is production.
	ShutdownTimeout             time.Duration                  // Duration to wait for ongoing RPCs to finish before shutting down the server.
	ServerLogger                logger.Logger                  // Logger for logging server-specific information, such as startup and shutdown events.
	ShutdownHook                func()                         // ShutdownHook is a callback function that will be called when the server is shutting down.
	MaxRecvMsgSize              int                            // Maximum size of a message that the server can receive. Default is 4MB.
	MaxSendMsgSize              int                            // Maximum size of a message that the server can send. Default is 4MB.
	KeepaliveTime               time.Duration                  // Duration for which the server will wait before sending a keepalive ping to the client.
	KeepaliveTimeout            time.Duration                  // Duration for which the server will wait for a keepalive ping ack from the client.
}

const (
	DefaultGRPCPort             = 9090
	DefaultGRPCShutdownTimeout  = 10 * time.Second
	DefaultGRPCMaxRecvMsgSize   = 4 * 1024 * 1024
	DefaultGRPCMaxSendMsgSize   = 4 * 1024 * 1024
	DefaultGRPCKeepaliveTime    = 2 * time.Hour
	DefaultGRPCKeepaliveTimeout = 20 * time.Second
	ProductionEnvironment       = "production"
)

// DefaultServerOptions returns a new ServerOptions struct with default values filled in
// for any fields not explicitly set by the caller. This function ensures that the gRPC server
// has safe and reasonable production defaults, such as sensible port, timeout durations,
// logger instances, and message size limits.
func DefaultServerOptions(opts *ServerOptions) *ServerOptions {
	if opts == nil {
		opts = &ServerOptions{}
	}

	defaultLogger := logger.Init(&logger.LoggerConfig{
		Provider: logger.LoggerProviderZerolog,
		Level:    logger.LogLevelDebug,
		Name:     "gRPCServer",
		Type:     logger.LoggerTypeStdout,
	})

	if opts.Port == 0 {
		opts.Port = DefaultGRPCPort
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = DefaultGRPCShutdownTimeout
	}
	if opts.ServerLogger == nil {
		opts.ServerLogger = defaultLogger
	}
	if opts.MaxRecvMsgSize == 0 {
		opts.MaxRecvMsgSize = DefaultGRPCMaxRecvMsgSize
	}
	if opts.MaxSendMsgSize == 0 {
		opts.MaxSendMsgSize = DefaultGRPCMaxSendMsgSize
	}
	if opts.KeepaliveTime == 0 {
		opts.KeepaliveTime = DefaultGRPCKeepaliveTime
	}
	if opts.KeepaliveTimeout == 0 {
		opts.KeepaliveTimeout = DefaultGRPCKeepaliveTimeout
	}
	return opts
}
