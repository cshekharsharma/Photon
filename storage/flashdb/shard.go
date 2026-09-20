package flashdb

import (
	"sync"

	"github.com/cshekharsharma/photon/utils/stdlib"
)

// shard represents a single partition of the in-memory store.
// It manages a subset of the data, providing thread-safe access and eviction
// based on least-recently-used (LRU) policy. Each shard tracks its own
// capacity and current usage in bytes, and uses a mutex to synchronize
// access to its internal state.
type shard struct {
	mu          sync.RWMutex                  // Mutex for protecting data, lru, currentSize
	data        map[string]*record            // actual data storage
	lru         *stdlib.LRU[string, struct{}] // LRU for tracking recency
	capacity    uint64                        // max bytes allowed in this shard
	currentSize uint64                        // current bytes used in this shard
}

// newShard creates a new shard with the given ID and capacity.
func (sh *shard) tailKeyLocked() (string, bool) {
	return sh.lru.TailKey()
}

// removeLocked removes the least recently used item from the shard.
func (sh *shard) removeLocked(key string) {
	record := sh.data[key]
	delete(sh.data, key)

	sh.lru.Delete(key)

	if record != nil {
		sh.currentSize -= nonNegativeInt64ToUint64(record.size)
	}
}

// isFull tells if shard allocated memory is already full or not
func (sh *shard) isFull() bool {
	return sh.currentSize > sh.capacity
}

// isFullAfter tells if shard allocated memory will be full after adding newSize
func (sh *shard) isFullAfter(newSize uint64) bool {
	return sh.currentSize+newSize > sh.capacity
}
