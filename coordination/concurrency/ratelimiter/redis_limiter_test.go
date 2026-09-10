package ratelimiter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"

	redisv9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockScript struct {
	mock.Mock
}

func (m *mockScript) Run(ctx context.Context, client redisv9.Scripter, keys []string, args ...interface{}) *redisv9.Cmd {
	call := m.Called(ctx, client, keys, args)
	cmd := redisv9.NewCmd(ctx)
	if val, ok := call.Get(0).(error); ok {
		cmd.SetErr(val)
	} else {
		cmd.SetVal(call.Get(0))
	}
	return cmd
}

type mockRedisInterface struct {
	mock.Mock
}

func (m *mockRedisInterface) GetClient() redis.RedisClientInterface {
	args := m.Called()
	return args.Get(0).(redis.RedisClientInterface)
}

func (m *mockRedisInterface) GetRawClient() *redisv9.Client {
	args := m.Called()
	return args.Get(0).(*redisv9.Client)
}

func (m *mockRedisInterface) SetClient(client redis.RedisClientInterface) {
	m.Called(client)
}

func (m *mockRedisInterface) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRedisLimiter_Allow_WithMockedScript(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(mockRedisInterface)
	mockScript := new(mockScript)

	mockRedis.On("GetRawClient").Return(&redisv9.Client{})

	limiter := NewRedisLimiter(mockRedis, 3, time.Second)
	limiter.script = &redisv9.Script{}

	limiter.script = mockScript
	mockScript.On("Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1))

	allowed, err := limiter.Allow(ctx, "ratelimit:unit:test")
	assert.NoError(t, err)
	assert.True(t, allowed)
	mockRedis.AssertExpectations(t)
	mockScript.AssertExpectations(t)
}

func TestRedisLimiter_Allow_ScriptError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(mockRedisInterface)
	rds := redisv9.NewClient(&redisv9.Options{Addr: "localhost:16379"})
	mockRedis.On("GetRawClient").Return(rds)

	limiter := NewRedisLimiter(mockRedis, 3, time.Second)
	mockScript := &mockScript{}
	mockScript.On("Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	limiter.script = mockScript

	_, err := limiter.Allow(ctx, "ratelimit:unit:err")
	assert.Error(t, err)
	mockRedis.AssertExpectations(t)
}

func TestRedisLimiter_Allow_RunReturnsError(t *testing.T) {
	ctx := context.Background()
	mockRedis := new(mockRedisInterface)
	mockScript := new(mockScript)

	mockRedis.On("GetRawClient").Return(&redisv9.Client{})

	limiter := NewRedisLimiter(mockRedis, 3, time.Second)
	limiter.script = mockScript

	mockScript.
		On("Run", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("lua exploded"))

	allowed, err := limiter.Allow(ctx, "ratelimit:unit:script-err")
	assert.False(t, allowed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis script error")
	mockRedis.AssertExpectations(t)
	mockScript.AssertExpectations(t)
}
