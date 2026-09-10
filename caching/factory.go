package caching

import (
	"fmt"

	"github.com/cshekharsharma/photon/storage/aerospike"
	"github.com/cshekharsharma/photon/storage/memcached"
	"github.com/cshekharsharma/photon/storage/redis"
)

// GetProvider returns a cache instance based on the provider specified in the options.
// Currently, it supports Aerospike as the cache provider.
//
// Parameters:
//   - opts: A pointer to Options struct containing provider type, cluster name, namespace, etc.
//
// Returns:
//   - Cache: An implementation of the Cache interface corresponding to the selected provider.
//   - error: If validation fails or the provider is unsupported.
//
// Example:
//
//	cache, err := GetProvider(&Options{
//	  Provider: ProviderAerospike,
//	  Cluster:  "my-aerospike-cluster",
//	  Namespace: "ns1",
//	  Collection: "users",
//	})
func GetProvider(opts *Options) (Cache, error) {
	if err := validate(opts); err != nil {
		return nil, err
	}

	switch opts.Provider {
	case ProviderAerospike:
		aerocache, err := NewAerospikeCache(opts, &aerospike.AerospikeConnector{})
		return aerocache, err

	case ProviderRedis:
		redis, err := NewRedisCache(opts, &redis.RedisConnector{})
		return redis, err

	case ProviderMemcached:
		memcache, err := NewMemcachedCache(opts, &memcached.MemcachedConnector{})
		return memcache, err

	case ProviderFlashDB:
		flashcache, err := NewFlashDBCache(*opts)
		return flashcache, err

	default:
		return nil, fmt.Errorf("unsupported cache provider: %s", opts.Provider)
	}
}

// validate performs basic validation on the provided Options.
//
// Parameters:
//   - opts: A pointer to Options struct that needs to be validated.
//
// Returns:
//   - error: Returns an error if mandatory fields like Provider or Cluster are missing, nil otherwise.
func validate(opts *Options) error {
	if opts.Provider == "" {
		return fmt.Errorf("Options.Provider is not set")
	}

	if len(opts.Hosts) == 0 {
		return fmt.Errorf("Options.Hosts is not set")
	}

	if opts.ConnTimeout == 0 {
		return fmt.Errorf("Options.ConnTimeout is not set")
	}

	if opts.DefaultTTL == 0 {
		return fmt.Errorf("Options.DefaultTTL is not set")
	}

	if opts.Cluster == "" {
		return fmt.Errorf("Options.Cluster is not set")
	}

	return nil
}
