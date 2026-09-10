package mocks

import (
	"context"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redisv9.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redisv9.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redisv9.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.StringCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redisv9.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) MGet(ctx context.Context, keys ...string) *redisv9.SliceCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redisv9.SliceCmd)
}

func (m *MockRedisClient) MSet(ctx context.Context, values ...interface{}) *redisv9.StatusCmd {
	args := m.Called(ctx, values)
	return args.Get(0).(*redisv9.StatusCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redisv9.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redisv9.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redisv9.BoolCmd)
}

func (m *MockRedisClient) TTL(ctx context.Context, key string) *redisv9.DurationCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.DurationCmd)
}

func (m *MockRedisClient) Persist(ctx context.Context, key string) *redisv9.BoolCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.BoolCmd)
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redisv9.StringSliceCmd {
	args := m.Called(ctx, pattern)
	return args.Get(0).(*redisv9.StringSliceCmd)
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redisv9.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) IncrBy(ctx context.Context, key string, value int64) *redisv9.IntCmd {
	args := m.Called(ctx, key, value)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) Decr(ctx context.Context, key string) *redisv9.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) DecrBy(ctx context.Context, key string, value int64) *redisv9.IntCmd {
	args := m.Called(ctx, key, value)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) Append(ctx context.Context, key, value string) *redisv9.IntCmd {
	args := m.Called(ctx, key, value)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) GetSet(ctx context.Context, key string, value interface{}) *redisv9.StringCmd {
	args := m.Called(ctx, key, value)
	return args.Get(0).(*redisv9.StringCmd)
}

func (m *MockRedisClient) StrLen(ctx context.Context, key string) *redisv9.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) HGet(ctx context.Context, key, field string) *redisv9.StringCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redisv9.StringCmd)
}

func (m *MockRedisClient) HDel(ctx context.Context, key string, fields ...string) *redisv9.IntCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) HGetAll(ctx context.Context, key string) *redisv9.MapStringStringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.MapStringStringCmd)
}

func (m *MockRedisClient) HExists(ctx context.Context, key, field string) *redisv9.BoolCmd {
	args := m.Called(ctx, key, field)
	return args.Get(0).(*redisv9.BoolCmd)
}

func (m *MockRedisClient) HLen(ctx context.Context, key string) *redisv9.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) HMGet(ctx context.Context, key string, fields ...string) *redisv9.SliceCmd {
	args := m.Called(ctx, key, fields)
	return args.Get(0).(*redisv9.SliceCmd)
}

func (m *MockRedisClient) HMSet(ctx context.Context, key string, values ...interface{}) *redisv9.BoolCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redisv9.BoolCmd)
}

func (m *MockRedisClient) HIncrBy(ctx context.Context, key, field string, incr int64) *redisv9.IntCmd {
	args := m.Called(ctx, key, field, incr)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) LPush(ctx context.Context, key string, values ...interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) RPush(ctx context.Context, key string, values ...interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, values)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) LPop(ctx context.Context, key string) *redisv9.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.StringCmd)
}

func (m *MockRedisClient) RPop(ctx context.Context, key string) *redisv9.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.StringCmd)
}

func (m *MockRedisClient) LRange(ctx context.Context, key string, start, stop int64) *redisv9.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redisv9.StringSliceCmd)
}

func (m *MockRedisClient) LRem(ctx context.Context, key string, count int64, value interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, count, value)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) LTrim(ctx context.Context, key string, start, stop int64) *redisv9.StatusCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redisv9.StatusCmd)
}

func (m *MockRedisClient) LLen(ctx context.Context, key string) *redisv9.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) SRem(ctx context.Context, key string, members ...interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) *redisv9.StringSliceCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redisv9.StringSliceCmd)
}

func (m *MockRedisClient) SIsMember(ctx context.Context, key string, member interface{}) *redisv9.BoolCmd {
	args := m.Called(ctx, key, member)
	return args.Get(0).(*redisv9.BoolCmd)
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...redisv9.Z) *redisv9.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) ZRem(ctx context.Context, key string, members ...interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) ZRange(ctx context.Context, key string, start, stop int64) *redisv9.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redisv9.StringSliceCmd)
}

func (m *MockRedisClient) ZRemRangeByScore(ctx context.Context, key, min, max string) *redisv9.IntCmd {
	args := m.Called(ctx, key, min, max)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redisv9.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redisv9.IntCmd)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redisv9.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redisv9.PubSub)
}

func (m *MockRedisClient) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redisv9.Cmd {
	callArgs := m.Called(ctx, script, keys, args)

	val := callArgs.Get(0)
	if val == nil {
		return nil
	}
	return val.(*redisv9.Cmd)
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redisv9.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redisv9.BoolCmd)
}

func (m *MockRedisClient) Ping(ctx context.Context) *redisv9.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redisv9.StatusCmd)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
