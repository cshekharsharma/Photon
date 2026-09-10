package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

// RedisLimiter implements a token bucket rate limiter using Redis and Lua.
// It supports high-throughput, distributed rate limiting by storing token
// bucket state in Redis hashes and executing all logic atomically via Lua.
type RedisLimiter struct {
	client    redis.RedisInterface // Redis client instance
	maxTokens int                  // Maximum number of tokens in the bucket
	interval  time.Duration        // Time interval to refill one token
	script    RedisScript          // Precompiled Lua script used to apply rate limiting atomically
}

// NewRedisLimiter constructs a RedisLimiter with the specified token capacity and refill interval.
// The limiter stores token counts in Redis and uses Lua scripting to ensure atomic updates.
//
// Parameters:
//   - client: a valid go-redis client connected to the Redis server
//   - maxTokens: the maximum burst capacity (tokens) per key
//   - interval: the duration after which a new token is added to the bucket
//
// Returns:
//   - a *RedisLimiter instance that implements the Limiter interface
func NewRedisLimiter(client redis.RedisInterface, maxTokens int, interval time.Duration) *RedisLimiter {
	limiter := &RedisLimiter{
		client:    client,
		maxTokens: maxTokens,
		interval:  interval,
	}
	limiter.script = limiter.getScript()
	return limiter
}

// Allow determines whether a request for the given key should be allowed.
// It evaluates the current token count and performs refill logic based on elapsed time.
//
// Parameters:
//   - ctx: context for cancellation and timeout
//   - key: unique identifier for the rate limit bucket (e.g., user ID or IP)
//
// Returns:
//   - true if the request is allowed (i.e., a token was consumed)
//   - false if rate limit has been exceeded
//   - error if Redis interaction or script execution fails
func (r *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	result, err := r.script.Run(
		ctx,
		r.client.GetRawClient(),
		[]string{key}, r.maxTokens,
		int(r.interval.Milliseconds()),
		now,
	).Result()

	if err != nil {
		return false, fmt.Errorf("redis script error: %w", err)
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected script return type: %T", result)
	}

	return allowed == 1, nil
}

// getScript returns the precompiled Lua script used for token bucket logic.
// The script ensures atomic access to the Redis key and performs the following:
//
//   - Initializes bucket if missing
//   - Refills tokens based on elapsed time since last refill
//   - Consumes a token if available
//   - Updates token count and last refill timestamp in Redis
//   - Sets a TTL (time to live) to auto-expire unused buckets
//
// Returns:
//   - a *redis.Script that can be executed via script.Run()
func (r *RedisLimiter) getScript() RedisScript {
	return redisv9.NewScript(`
	local key = KEYS[1]
	local maxTokens = tonumber(ARGV[1])
	local refillInterval = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local bucket = redis.call("HMGET", key, "tokens", "last")
	local tokens = tonumber(bucket[1])
	local lastRefill = tonumber(bucket[2])

	if tokens == nil then
		tokens = maxTokens
		lastRefill = now
	end

	local delta = math.max(0, now - lastRefill)
	local refill = math.floor(delta / refillInterval)

	if refill > 0 then
		tokens = math.min(maxTokens, tokens + refill)
		lastRefill = now
	end

	local allowed = tokens > 0
	if allowed then
		tokens = tokens - 1
	end

	redis.call("HMSET", key, "tokens", tokens, "last", lastRefill)
	redis.call("PEXPIRE", key, refillInterval * 2)
	return allowed
	`)
}

// RedisScript defines the interface for executing Lua scripts in Redis.
type RedisScript interface {
	Run(ctx context.Context, client redisv9.Scripter, keys []string, args ...interface{}) *redisv9.Cmd
}
