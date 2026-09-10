package flashdb

import "time"

// Options for FlashDB.
// These options control the behavior and limits of the in-memory store.
type Options struct {
	Shards         uint8         // default 64
	MaxStorageSize uint64        // REQUIRED: per-store budget in bytes (split evenly per shard)
	DefaultTTL     time.Duration // default TTL
	JanitorEvery   time.Duration // default 1m (0 disables)
	LFUDepth       int           // default 4
	LFUWidth       uint64        // default 1<<16
	LFUAgingEvery  uint64        // default 200_000
}
