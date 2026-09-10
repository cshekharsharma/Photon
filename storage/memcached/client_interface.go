package memcached

import (
	"github.com/bradfitz/gomemcache/memcache"
)

// MemcacheClientInterface defines the methods needed from a Memcached client.
// It abstracts the essential cache operations for easy mocking and testing.
type MemcachedClientInterface interface {
	// CRUD operations
	Get(key string) (*memcache.Item, error)
	GetMulti(keys []string) (map[string]*memcache.Item, error)
	GetAndTouch(key string, expiration int32) (*memcache.Item, error)
	Set(item *memcache.Item) error
	Add(item *memcache.Item) error
	Replace(item *memcache.Item) error
	Delete(key string) error
	DeleteAll() error
	FlushAll() error

	// Mutations
	Increment(key string, delta uint64) (uint64, error)
	Decrement(key string, delta uint64) (uint64, error)
	Append(item *memcache.Item) error
	Prepend(item *memcache.Item) error
	CompareAndSwap(item *memcache.Item) error

	// TTL Management
	Touch(key string, seconds int32) error

	// Health
	Ping() error

	// Connection
	Close() error
}
