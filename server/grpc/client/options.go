package client

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"google.golang.org/grpc"
)

// ClientOptions holds the configuration required to initialize a GRPCClient.
type ClientOptions struct {
	Target              string         // gRPC server address
	TLSConfig           *tls.Config    // TLS configuration (optional)
	Insecure            bool           // Allow insecure connection
	MinConnectTimeout   time.Duration  // Minimum timeout for connection
	MaxRetries          int            // Retry attempts on failure
	EnableLoadBalancing bool           // Enable round robin DNS LB
	BackoffConfig       *BackoffConfig // Backoff configuration for retrying connections
	Logger              logger.Logger  // Logger for logging events during interceptors
	DialOptions         []grpc.DialOption
	ContextDialer       func(context.Context, string) (net.Conn, error)
}

// BackoffConfig holds the configuration for exponential backoff strategy.
// It is used to control the retry behavior of the gRPC client.
// The backoff strategy is used when the client fails to connect to the server.
type BackoffConfig struct {
	BaseDelay  time.Duration // BaseDelay is the amount of time to backoff after the first failure.
	Multiplier float64       // Multiplier is the factor for multiplying backoffs after failed retry. Should be greater than 1.
	Jitter     float64       // Jitter is the factor with which backoffs are randomized.
	MaxDelay   time.Duration // MaxDelay is the upper bound of backoff delay.
}
