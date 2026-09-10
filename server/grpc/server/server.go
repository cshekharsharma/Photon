package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/server/grpc/server/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const networkTCP = "tcp"

var grpcServeFn = func(server *grpc.Server, lis net.Listener) error {
	return server.Serve(lis)
}

type stoppableServer interface {
	GracefulStop()
	Stop()
}

// StartGRPCServer starts a gRPC server using the provided ServerOptions. It supports TLS or
// insecure mode, health check service, reflection for debugging, chained interceptors,
// and graceful shutdown on SIGINT or SIGTERM. If RegisterFunc is provided, it is used to
// register the application-specific services. Any missing options are populated with defaults
// via DefaultServerOptions.
func StartGRPCServer(opts *ServerOptions) error {
	opts = DefaultServerOptions(opts)

	listenAddr := fmt.Sprintf(":%d", opts.Port)
	lis, err := net.Listen(networkTCP, listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	var grpcOpts []grpc.ServerOption

	// set TLS or insecure credentials
	if opts.TLSConfig != nil {
		grpcOpts = append(grpcOpts, grpc.Creds(credentials.NewTLS(opts.TLSConfig)))
	} else {
		grpcOpts = append(grpcOpts, grpc.Creds(insecure.NewCredentials()))
	}

	// append required interceptors (both in-build & user-defined)
	unaryInterceptors := make([]grpc.UnaryServerInterceptor, 0)

	if len(opts.UnaryInterceptors) > 0 {
		unaryInterceptors = append(unaryInterceptors, opts.UnaryInterceptors...)
	}

	unaryInterceptors = append(unaryInterceptors,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggingInterceptor(opts.ServerLogger),
		interceptors.RecoverInterceptor(opts.ServerLogger),
	)
	grpcOpts = append(grpcOpts, grpc.ChainUnaryInterceptor(unaryInterceptors...))

	if len(opts.StreamInterceptors) > 0 {
		grpcOpts = append(grpcOpts, grpc.ChainStreamInterceptor(opts.StreamInterceptors...))
	}

	// Set message size limits
	grpcOpts = append(grpcOpts,
		grpc.MaxRecvMsgSize(opts.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(opts.MaxSendMsgSize),
	)

	grpcServer := grpc.NewServer(grpcOpts...)

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	if opts.RegisterFunc != nil {
		opts.RegisterFunc(grpcServer)
	}

	if opts.EnableReflection {
		reflection.Register(grpcServer)
	}

	errChan := make(chan error, 1)
	go func() {
		opts.ServerLogger.Info("Starting gRPC server on %s", listenAddr)
		errChan <- grpcServeFn(grpcServer, lis)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)

	select {
	case sig := <-sigChan:
		opts.ServerLogger.Info("Received shutdown signal: %v", sig)
		if opts.ShutdownHook != nil {
			opts.ShutdownHook()
		}
		stopGracefully(grpcServer, opts.ShutdownTimeout, opts.ServerLogger)
		return nil
	case err := <-errChan:
		return fmt.Errorf("gRPC server error: %w", err)
	}
}

// stopGracefully attempts a graceful shutdown of the provided gRPC server
// within the specified timeout duration. If the server fails to stop within
// the allotted time, it forcefully stops it. Shutdown events are logged using
// the provided logger. This function is intended to be called internally during
// the termination flow of StartGRPCServer.
func stopGracefully(s stoppableServer, timeout time.Duration, log logger.Logger) {
	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Info("gRPC server shut down gracefully")
	case <-time.After(timeout):
		log.Warn("Graceful shutdown timed out, forcing stop")
		s.Stop()
	}
}
