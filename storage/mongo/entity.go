package mongo

import (
	"fmt"
	"strings"
	"time"
)

type ConnectionConfig struct {
	Hosts              []string
	Username           string
	Password           string
	ConnectionTimeout  time.Duration
	ConnectionPoolsize int64
}

func (c *ConnectionConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("mongo: connection config is required")
	}
	if len(c.Hosts) == 0 {
		return fmt.Errorf("mongo: at least one host is required")
	}
	for _, host := range c.Hosts {
		if strings.TrimSpace(host) == "" {
			return fmt.Errorf("mongo: host cannot be empty")
		}
	}
	if c.ConnectionTimeout < 0 {
		return fmt.Errorf("mongo: connection timeout cannot be negative")
	}
	if c.ConnectionPoolsize < 0 {
		return fmt.Errorf("mongo: connection pool size cannot be negative")
	}
	return nil
}

func (c *ConnectionConfig) normalized() *ConnectionConfig {
	if c == nil {
		return nil
	}
	normalized := *c
	normalized.Hosts = append([]string(nil), c.Hosts...)
	for i := range normalized.Hosts {
		normalized.Hosts[i] = strings.TrimSpace(normalized.Hosts[i])
	}
	if normalized.ConnectionTimeout == 0 {
		normalized.ConnectionTimeout = DefaultConnectionTimeout
	}
	if normalized.ConnectionPoolsize == 0 {
		normalized.ConnectionPoolsize = int64(DefaultConnectionPoolSize)
	}
	return &normalized
}
