package ratelimiter

import (
	"fmt"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"
)

func TestNewRateLimiter_NilOptions(t *testing.T) {
	_, err := NewRateLimiter(nil)
	if err == nil || err.Error() != "options cannot be nil" {
		t.Errorf("expected error for nil options, got: %v", err)
	}
}

func TestNewRateLimiter_InvalidClientType(t *testing.T) {
	opts := &Options{
		Type:      RateLimiterTypeRedis,
		Client:    "not-a-redis-client",
		MaxTokens: 10,
		Interval:  time.Second,
	}
	_, err := NewRateLimiter(opts)
	if err == nil || err.Error() != "invalid client type for Redis rate limiter" {
		t.Errorf("expected error for invalid client type, got: %v", err)
	}
}

func TestNewRateLimiter_UnsupportedType(t *testing.T) {
	opts := &Options{
		Type:      RateLimiterType("unknown"),
		Client:    nil,
		MaxTokens: 10,
		Interval:  time.Second,
	}
	_, err := NewRateLimiter(opts)
	expectedErr := fmt.Sprintf("unsupported rate limiter type: %s", opts.Type)
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected unsupported type error, got: %v", err)
	}
}

func TestNewRateLimiter_ValidRedisClient(t *testing.T) {
	redis.SetConnectionConfig("testserver", &redis.ConnectionConfig{Address: "localhost:6379"})
	client, err := redis.Connect(&redis.RedisConnector{}, "testserver")
	if err != nil {
		t.Fatalf("failed to connect to Redis: %v", err)
	}

	opts := &Options{
		Type:      RateLimiterTypeRedis,
		Client:    client,
		MaxTokens: 5,
		Interval:  time.Second,
	}
	limiter, err := NewRateLimiter(opts)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if limiter == nil {
		t.Errorf("expected limiter instance, got nil")
	}
}
