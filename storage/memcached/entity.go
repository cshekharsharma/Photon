package memcached

import "time"

type ConnectionConfig struct {
	Addresses   []string
	Timeout     time.Duration
	MaxIdleConn int64
}
