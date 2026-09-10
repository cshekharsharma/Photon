package redis

// ConnectionConfig holds the configuration parameters required to establish a connection
// to a Redis server. It includes details such as server address, authentication credentials,
// target database index, and connection pool settings.
type ConnectionConfig struct {
	Address  string //  The Redis server address in "host:port" format (e.g., "localhost:6379").
	Username string // Username for authentication (used when Redis ACLs are enabled). Leave empty if not applicable.
	Password string // Password for authentication. Leave empty if no password is set.
	Database int    // Redis logical database number (default is 0).
	PoolSize int    // Maximum number of socket connections to maintain in the pool.
}
