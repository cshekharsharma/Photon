package ratelimiter

// RateLimiterType represents the type of backend implementation to use
// for rate limiting. This allows selecting among supported strategies,
// such as Redis-based limiters.
type RateLimiterType string

const (
	// RateLimiterTypeRedis specifies that the rate limiter should use a
	// Redis backend for token bucket storage and atomic operations.
	RateLimiterTypeRedis RateLimiterType = "redis"
)
