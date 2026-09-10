package mongo

import "time"

const (

	// DefaultConnectionTimeout specifies the default timeout duration in seconds.
	DefaultConnectionTimeout time.Duration = 10 * time.Second

	// DefaultConnectionPoolSize defines the default size for the connection pool
	DefaultConnectionPoolSize uint64 = 100
)
