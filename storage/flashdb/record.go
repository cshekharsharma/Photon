package flashdb

import "time"

// record represents a single entry in the in-memory store.
// It contains the actual value as a map, the cached size in bytes,
// and an optional expiry time (zero value means no TTL).
type record struct {
	value  map[string]any // actual value
	size   int64          // size of cached value (bytes)
	expiry time.Time      // zero => no TTL
}

// isExpired checks if the record has expired based on the current time.
func (r *record) isExpired(now time.Time) bool {
	return !r.expiry.IsZero() && now.After(r.expiry)
}
