package caching

// CacheProvider is a string type that represents the identifier for different cache providers.
type CacheProvider string

const (
	// ProviderAerospike represents the cache provider identifier for Aerospike.
	// This constant is used to select the Aerospike-backed implementation of the Cache interface.
	ProviderAerospike CacheProvider = "aerospike"

	// ProviderRedis represents the cache provider identifier for Redis.
	// This constant is used to select the Redis-backed implementation of the Cache interface.
	ProviderRedis CacheProvider = "redis"

	// ProviderMemcached represents the cache provider identifier for Memcached.
	// This constant is used to select the Memcached-backed implementation of the Cache interface.
	ProviderMemcached CacheProvider = "memcached"

	// ProviderFlashDB represents the cache provider identifier for FlashDB.
	// This constant is used to select the FlashDB-backed implementation of the Cache interface.
	ProviderFlashDB CacheProvider = "flashdb"
)
