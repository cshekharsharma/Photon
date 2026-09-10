package caching

// CacheRequest defines the interface for all cache request types.
// It ensures that each request contains a namespace and collection,
// which are required to interact with specific cache stores (e.g., Aerospike).
type CacheRequest interface {
	GetNamespace() string
	GetCollection() string
	SetNamespace(namespace string)
	SetCollection(collection string)
}

// cacheRequest is an embedded struct used in all concrete cache request types.
// It provides shared fields and implementations for namespace and collection access.
type cacheRequest struct {
	Namespace  string
	Collection string
}

// GetNamespace returns the namespace for the cache request.
func (c *cacheRequest) GetNamespace() string {
	return c.Namespace
}

// GetCollection returns the collection for the cache request.
func (c *cacheRequest) GetCollection() string {
	return c.Collection
}

// SetNamespace sets the namespace for the cache request.
func (c *cacheRequest) SetNamespace(namespace string) {
	c.Namespace = namespace
}

// SetCollection sets the collection for the cache request.
func (c *cacheRequest) SetCollection(collection string) {
	c.Collection = collection
}

// ExistsRequest is used to check the existence of a specific key in the cache.
type ExistsRequest struct {
	cacheRequest
	Key string // The cache key to check
}

// GetRequest is used to retrieve a specific record or fields from the cache.
type GetRequest struct {
	cacheRequest
	Key    string   // The cache key to retrieve
	Fields []string // Optional list of fields (bins) to fetch
}

// SetRequest is used to insert or update a key in the cache.
type SetRequest struct {
	cacheRequest
	Key    string         // The key to set
	Value  any            // The value to set (optional if using Fields)
	Fields map[string]any // Field-level values for bin-based stores like Aerospike
	TTL    int64          // Time-to-live in seconds
}

// DeleteRequest is used to delete a specific key from the cache.
type DeleteRequest struct {
	cacheRequest
	Key string // The key to delete
}

// MultiGetRequest is used to fetch multiple keys in a single request.
type MultiGetRequest struct {
	cacheRequest
	Keys []string // List of keys to retrieve
}

// MultiSetRequest is used to set multiple keys in a single operation.
type MultiSetRequest struct {
	cacheRequest
	FieldsMap map[string]map[string]any // Map of keys to their field maps
	ValueMap  map[string]any            // Optional values for simple stores
	TTL       int64                     // TTL applied to all keys
}

// MultiDeleteRequest is used to delete multiple keys from the cache.
type MultiDeleteRequest struct {
	cacheRequest
	Keys []string // Keys to delete
}

// IncrementRequest is used to increment integer values of fields in a record.
type IncrementRequest struct {
	cacheRequest
	Key    string           // Key of the record to increment
	Fields map[string]int64 // Fields and their increment values
	Value  int64            // Optional single field increment value
}

// DecrementRequest is used to decrement integer values of fields in a record.
type DecrementRequest struct {
	cacheRequest
	Key    string           // Key of the record to decrement
	Fields map[string]int64 // Fields and their decrement values
	Value  int64            // Optional single field decrement value
}

// AppendRequest is used to append string values to fields in a record.
type AppendRequest struct {
	cacheRequest
	Key    string            // Key of the record to append to
	Fields map[string]string // Fields and the values to append
	Value  string            // Optional single field append value
}

// GetTTLRequest is used to fetch the TTL (time-to-live) of a specific key.
type GetTTLRequest struct {
	cacheRequest
	Key string // The key for which TTL is queried
}

// SetTTLRequest is used to update the TTL of a specific key.
type SetTTLRequest struct {
	cacheRequest
	Key string // The key to update
	TTL int64  // New TTL in seconds
}
