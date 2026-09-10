package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClientInterface defines a unified interface over the go-redis Client.
type RedisClientInterface interface {
	// Basic Key operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
	MSet(ctx context.Context, values ...interface{}) *redis.StatusCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Persist(ctx context.Context, key string) *redis.BoolCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd

	// Increment/Decrement
	Incr(ctx context.Context, key string) *redis.IntCmd
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	DecrBy(ctx context.Context, key string, value int64) *redis.IntCmd

	// String operations
	Append(ctx context.Context, key, value string) *redis.IntCmd
	GetSet(ctx context.Context, key string, value interface{}) *redis.StringCmd
	StrLen(ctx context.Context, key string) *redis.IntCmd

	// Hash operations
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	HExists(ctx context.Context, key, field string) *redis.BoolCmd
	HLen(ctx context.Context, key string) *redis.IntCmd
	HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd
	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	HIncrBy(ctx context.Context, key, field string, incr int64) *redis.IntCmd

	// List operations
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	LPop(ctx context.Context, key string) *redis.StringCmd
	RPop(ctx context.Context, key string) *redis.StringCmd
	LRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd
	LRem(ctx context.Context, key string, count int64, value interface{}) *redis.IntCmd
	LTrim(ctx context.Context, key string, start, stop int64) *redis.StatusCmd
	LLen(ctx context.Context, key string) *redis.IntCmd

	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd

	// Sorted Set operations
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd
	ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd

	// Pub/Sub
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub

	// Scripting
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd

	// Locking/Concurrency (Optional helpers)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd

	// Ping
	Ping(ctx context.Context) *redis.StatusCmd

	// Connection/Client operations
	Close() error
}
