# Cache

Use the `caching.Cache` interface when application code should not depend on a specific cache backend. This example uses in-process FlashDB; the same request types work with the other cache providers.

```go
package examples

import (
	"fmt"

	"github.com/cshekharsharma/photon/caching"
)

func CacheExample() error {
	cache, err := caching.GetProvider(&caching.Options{
		Provider:    caching.ProviderFlashDB,
		Cluster:     "local",
		Hosts:       []string{"in-process"},
		ConnTimeout: 1,
		DefaultTTL:  300,
		Namespace:   "checkout",
		Collection:  "sessions",
	})
	if err != nil {
		return err
	}

	set := &caching.SetRequest{
		Key:   "session:123",
		Value: map[string]any{"user_id": "user-456"},
		TTL:   300,
	}
	set.SetNamespace("checkout")
	set.SetCollection("sessions")

	if _, err := cache.Set(set); err != nil {
		return err
	}

	get := &caching.GetRequest{Key: "session:123"}
	get.SetNamespace("checkout")
	get.SetCollection("sessions")

	value, err := cache.Get(get)
	if err != nil {
		return err
	}

	fmt.Println(value)
	return nil
}
```

