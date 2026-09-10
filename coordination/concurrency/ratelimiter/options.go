package ratelimiter

import "time"

// Options defines the configuration required to create a rate limiter instance.
// It encapsulates the type of rate limiter, the backend client, and the token
// bucket parameters such as maximum tokens and refill interval.
type Options struct {
	// Type determines the implementation of rate limiter to use (e.g., Redis).
	Type RateLimiterType

	// Client is the backend client instance (e.g., *redis.Client for Redis-based limiter).
	// It must match the expected type of the chosen rate limiter implementation.
	Client interface{}

	// MaxTokens specifies the maximum number of tokens in the token bucket.
	// This also defines the burst capacity for the limiter.
	MaxTokens int

	// Interval specifies the duration after which a new token is added to the bucket.
	// This controls the refill rate for tokens.
	Interval time.Duration
}
