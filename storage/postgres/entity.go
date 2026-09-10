package postgres

import (
	"fmt"
	"strconv"
	"time"
)

// PgBouncerMode defines how PgBouncer pools backend connections.
// Only transaction/statement pooling is incompatible with server-side prepares.
//
// Values:
//   - "" (empty): not using PgBouncer / unknown
//   - "session"
//   - "transaction"
//   - "statement"
type PgBouncerMode string

const (
	PgBouncerNone        PgBouncerMode = ""
	PgBouncerSession     PgBouncerMode = "session"
	PgBouncerTransaction PgBouncerMode = "transaction"
	PgBouncerStatement   PgBouncerMode = "statement"
)

// ConnectionConfig holds settings for connecting to Postgres.
type ConnectionConfig struct {
	Host     string
	Port     string
	UserName string
	Password string
	DbName   string

	// TLS/SSL: "disable", "require", "verify-ca", "verify-full"
	SSLMode string

	// Pool sizing (per process)
	MaxOpenConn int
	MaxIdleConn int

	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Defaults applied if zero.
	ConnectTimeout time.Duration
	PingTimeout    time.Duration

	// DefaultQueryTimeout is used by DBContext when it needs to add a timeout
	// to context-less calls.
	DefaultQueryTimeout time.Duration

	// PgBouncer config — if transaction/statement pooling, we will disable
	// server-side prepared statements at the driver level.
	PgBouncer PgBouncerMode

	// RuntimeParams are sent to Postgres at connection startup:
	// application_name, search_path, statement_timeout, timezone, etc.
	RuntimeParams map[string]string
}

func (c *ConnectionConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("postgres: connection config is required")
	}
	if c.Host == "" || c.Port == "" || c.UserName == "" || c.DbName == "" {
		return fmt.Errorf("postgres: missing required connection fields (host/port/user/dbname)")
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("postgres: invalid port: %q", c.Port)
	}
	switch c.SSLMode {
	case "", "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("postgres: invalid sslmode: %q", c.SSLMode)
	}
	switch c.PgBouncer {
	case PgBouncerNone, PgBouncerSession, PgBouncerTransaction, PgBouncerStatement:
	default:
		return fmt.Errorf("postgres: invalid pgbouncer mode: %q", c.PgBouncer)
	}
	if c.MaxOpenConn < 0 {
		return fmt.Errorf("postgres: MaxOpenConn cannot be negative")
	}
	if c.MaxIdleConn < 0 {
		return fmt.Errorf("postgres: MaxIdleConn cannot be negative")
	}
	if c.MaxOpenConn > 0 && c.MaxIdleConn > c.MaxOpenConn {
		return fmt.Errorf("postgres: MaxIdleConn cannot exceed MaxOpenConn")
	}
	if c.ConnMaxLifetime < 0 {
		return fmt.Errorf("postgres: ConnMaxLifetime cannot be negative")
	}
	if c.ConnMaxIdleTime < 0 {
		return fmt.Errorf("postgres: ConnMaxIdleTime cannot be negative")
	}
	if c.ConnectTimeout < 0 {
		return fmt.Errorf("postgres: ConnectTimeout cannot be negative")
	}
	if c.PingTimeout < 0 {
		return fmt.Errorf("postgres: PingTimeout cannot be negative")
	}
	if c.DefaultQueryTimeout < 0 {
		return fmt.Errorf("postgres: DefaultQueryTimeout cannot be negative")
	}
	return nil
}

func (c *ConnectionConfig) normalized() (*ConnectionConfig, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	normalized := *c
	if c.RuntimeParams != nil {
		normalized.RuntimeParams = make(map[string]string, len(c.RuntimeParams))
		for k, v := range c.RuntimeParams {
			normalized.RuntimeParams[k] = v
		}
	}
	applyDefaults(&normalized)
	return &normalized, nil
}
