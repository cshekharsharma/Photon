package aerospike

// ConnectionConfig struct for initialising aerospike connection
type ConnectionConfig struct {
	Hosts             []string
	ConnectionTimeout int64
	DefaultTTL        uint32
}
