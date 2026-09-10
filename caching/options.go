package caching

// Options holds the configuration settings required to initialize a Cache provider.
type Options struct {
	Provider    CacheProvider // Provider specifies the cache provider to use (e.g., "aerospike").
	Cluster     string        // Cluster defines the target cluster identifier or connection alias for the cache provider.
	Hosts       []string      // Hosts is a list of host addresses for the cache provider (e.g., Aerospike nodes).
	ConnTimeout int64         // ConnTimeout is the connection timeout duration in seconds.
	DefaultTTL  int64         // DefaultTTL is the default time-to-live for cache entries, in seconds.
	Namespace   string        // Namespace specifies the namespace within the cache system (used in Aerospike).
	Collection  string        // Collection denotes the logical grouping of records, similar to a set (used in Aerospike).
	Username    string        // Username  denotes the authentication username required to login (Optional)
	Password    string        // Password denotes the authentication password required to login (Optional)
}
