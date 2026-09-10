package concurrency

import (
	"testing"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"
	"github.com/cshekharsharma/photon/utils/testutil/mocks"

	redisv9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestGetDistributedLocker_Success(t *testing.T) {
	mockRedis := new(mockRedisInterface)
	mockRedisClient := new(mocks.MockRedisClient)
	mockRedis.On("GetClient").Return(mockRedisClient)
	mockRedis.On("GetRawClient").Return(nil)

	opts := &LockOptions{
		LockerProvider:          RedisLockProvider,
		StorageClient:           mockRedis,
		DefaultLockRetryTimeout: 5 * time.Second,
	}

	locker, err := GetDistributedLocker(opts)
	assert.NoError(t, err)
	assert.NotNil(t, locker)
}

func TestGetDistributedLocker_NilOptions(t *testing.T) {
	locker, err := GetDistributedLocker(nil)
	assert.Error(t, err)
	assert.Nil(t, locker)
	assert.Equal(t, "options cannot be nil", err.Error())
}

func TestGetDistributedLocker_InvalidRedisClient(t *testing.T) {
	opts := &LockOptions{
		LockerProvider: RedisLockProvider,
		StorageClient:  "invalid-type",
	}

	locker, err := GetDistributedLocker(opts)
	assert.Error(t, err)
	assert.Nil(t, locker)
	assert.Equal(t, "invalid storage client type for Redis locker", err.Error())
}

func TestGetDistributedLocker_DefaultTimeoutApplied(t *testing.T) {
	mockRedis := new(mockRedisInterface)
	mockRedisClient := new(mocks.MockRedisClient)
	mockRedis.On("GetClient").Return(mockRedisClient)
	mockRedis.On("GetRawClient").Return(nil)

	opts := &LockOptions{
		LockerProvider: RedisLockProvider,
		StorageClient:  mockRedis,
	}

	locker, err := GetDistributedLocker(opts)
	assert.NoError(t, err)
	assert.NotNil(t, locker)
	assert.Equal(t, defaultLockRetryTimeout, opts.DefaultLockRetryTimeout)
}

func TestGetDistributedLocker_InvalidLockerType(t *testing.T) {
	opts := &LockOptions{
		LockerProvider: "invalid-locker-type",
	}

	locker, err := GetDistributedLocker(opts)
	assert.Error(t, err)
	assert.Nil(t, locker)
	assert.Equal(t, "unsupported locker provider: invalid-locker-type", err.Error())
}
