package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	stdlog "log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/core/router"
	"github.com/cshekharsharma/photon/core/session"
	"github.com/cshekharsharma/photon/middleware"
	"github.com/cshekharsharma/photon/workers"
)

const (
	DefaultServerPort        int64         = 8080
	DefaultReadTimeout       time.Duration = 5 * time.Second
	DefaultReadHeaderTimeout time.Duration = 2 * time.Second
	DefaultWriteTimeout      time.Duration = 10 * time.Second
	DefaultIdleTimeout       time.Duration = 120 * time.Second
	DefaultShutdownTimeout   time.Duration = 10 * time.Second
	DefaultMaxHeaderBytes    int           = 1 << 20
)

type ServerConfig struct {
	ServerPort        int64
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	TLSConfig         *tls.Config
	TLSCertFile       string
	TLSKeyFile        string
	BaseContext       func(net.Listener) context.Context
	ConnContext       func(context.Context, net.Conn) context.Context
	ErrorLog          *stdlog.Logger

	SessionConfig  *session.Config
	SessionManager *session.Manager

	RouteProvider   string
	HttpRoutes      []*HttpRoute
	Middlewares     []func(http.Handler) http.Handler
	CorsOptions     *CorsOptions
	ApiMinifyConfig *ApiMinifyConfig

	WorkerSleepTimeout time.Duration
	BGWorkers          []*workers.WorkerConfig

	AccessLogger logger.Logger
	ErrorLogger  logger.Logger
	ServerLogger logger.Logger

	ShutdownHook func()
}

func (s *ServerConfig) Normalize() error {
	if s == nil {
		return errors.New("server config cannot be nil")
	}

	if s.RouteProvider == "" {
		s.RouteProvider = router.RouterCHI
	}
	if s.ServerPort == 0 {
		s.ServerPort = DefaultServerPort
	}
	if s.ReadTimeout == 0 {
		s.ReadTimeout = DefaultReadTimeout
	}
	if s.ReadHeaderTimeout == 0 {
		s.ReadHeaderTimeout = DefaultReadHeaderTimeout
	}
	if s.WriteTimeout == 0 {
		s.WriteTimeout = DefaultWriteTimeout
	}
	if s.IdleTimeout == 0 {
		s.IdleTimeout = DefaultIdleTimeout
	}
	if s.ShutdownTimeout == 0 {
		s.ShutdownTimeout = DefaultShutdownTimeout
	}
	if s.MaxHeaderBytes == 0 {
		s.MaxHeaderBytes = DefaultMaxHeaderBytes
	}
	if s.AccessLogger == nil {
		s.AccessLogger = logger.Init(defaultLoggerConfig("photon-access"))
	}
	if s.ErrorLogger == nil {
		s.ErrorLogger = logger.Init(defaultLoggerConfig("photon-error"))
	}
	if s.ServerLogger == nil {
		s.ServerLogger = logger.Init(defaultLoggerConfig("photon-server"))
	}

	return s.Validate()
}

func (s *ServerConfig) Validate() error {
	if s == nil {
		return errors.New("server config cannot be nil")
	}
	if s.RouteProvider != router.RouterCHI {
		return fmt.Errorf("unsupported route provider %q", s.RouteProvider)
	}
	if s.ServerPort < 1 || s.ServerPort > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", s.ServerPort)
	}
	if s.ReadTimeout < 0 {
		return errors.New("read timeout cannot be negative")
	}
	if s.ReadHeaderTimeout < 0 {
		return errors.New("read header timeout cannot be negative")
	}
	if s.WriteTimeout < 0 {
		return errors.New("write timeout cannot be negative")
	}
	if s.IdleTimeout < 0 {
		return errors.New("idle timeout cannot be negative")
	}
	if s.ShutdownTimeout < 0 {
		return errors.New("shutdown timeout cannot be negative")
	}
	if s.MaxHeaderBytes < 0 {
		return errors.New("max header bytes cannot be negative")
	}
	if (s.TLSCertFile == "") != (s.TLSKeyFile == "") {
		return errors.New("tls cert file and key file must be configured together")
	}
	if s.SessionConfig != nil && s.SessionManager != nil {
		return errors.New("session config and session manager cannot both be configured")
	}
	if s.AccessLogger == nil {
		return errors.New("access logger cannot be nil")
	}
	if s.ErrorLogger == nil {
		return errors.New("error logger cannot be nil")
	}
	if s.ServerLogger == nil {
		return errors.New("server logger cannot be nil")
	}
	if s.CorsOptions != nil {
		for _, origin := range s.CorsOptions.AllowedOrigins {
			if origin == "*" && s.CorsOptions.AllowCredentials {
				return errors.New("cors cannot allow credentials with wildcard origin")
			}
		}
	}
	for idx, route := range s.HttpRoutes {
		if route == nil {
			return fmt.Errorf("http route at index %d cannot be nil", idx)
		}
		if route.UrlRoute == "" {
			return fmt.Errorf("http route at index %d must have a url route", idx)
		}
		if route.HttpHandler == nil {
			return fmt.Errorf("http route %q must have a handler", route.UrlRoute)
		}
	}

	return nil
}

type HttpRoute struct {
	UrlRoute      string
	RequestMethod string
	HttpHandler   http.HandlerFunc
	Middlewares   []middleware.MiddlewareFn
	EnableMinify  bool
}

type CorsOptions struct {
	MaxAge           int
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type ApiMinifyConfig struct {
	IsEnabled bool
	KeyMap    *sync.Map
}

func defaultLoggerConfig(name string) *logger.LoggerConfig {
	return &logger.LoggerConfig{
		Name:     name,
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
		Level:    logger.LogLevelWarn,
	}
}
