package redis

import "github.com/redis/go-redis/v9"

// IsNil reports whether the error represents a missing key.
func IsNil(err error) bool {
	return err == redis.Nil
}
