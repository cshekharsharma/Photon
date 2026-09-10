package concurrency

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"
	"github.com/cshekharsharma/photon/utils/testutil/mocks"
	redisv9 "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRedis struct {
	mock.Mock
	client redis.RedisClientInterface
}

func (m *mockRedis) GetClient() redis.RedisClientInterface {
	args := m.Called()
	return args.Get(0).(redis.RedisClientInterface)
}

func (m *mockRedis) GetRawClient() *redisv9.Client {
	args := m.Called()
	return args.Get(0).(*redisv9.Client)
}

func (m *mockRedis) SetClient(client redis.RedisClientInterface) {
	m.client = client
}

func (m *mockRedis) Close() error {
	args := m.Called()
	return args.Error(0)
}

// ----------------------------- Tests for RedisLocker -----------------------------
func TestRedisLock_Success(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	mockCmd := redisv9.NewBoolResult(true, nil)
	mockClient.On("SetNX", mock.Anything, "mylock", mock.Anything, time.Second).Return(mockCmd)

	ctx := context.Background()
	err := locker.Lock(ctx, "mylock", time.Second, time.Millisecond*10)
	assert.NoError(t, err)
}

func TestRedisLock_WithStorageFailure(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	cmd := redisv9.NewBoolResult(false, errors.New("some error"))
	mockClient.On("SetNX", mock.Anything, "mylock", mock.Anything, time.Second).Return(cmd)

	ctx := context.Background()
	err := locker.Lock(ctx, "mylock", time.Second, time.Millisecond*10)
	assert.Error(t, err)
}

func TestRedisLock_ContextTimeout(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient) // Fix: Add this
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	mockCmd := redisv9.NewBoolResult(false, nil)
	mockClient.On("SetNX", mock.Anything, "lockfail", mock.Anything, time.Second).Return(mockCmd)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := locker.Lock(ctx, "lockfail", time.Second, 10*time.Millisecond)
	assert.Equal(t, ErrLockNotAcquired, err)
}

func TestRedisUnlock_Success(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)

	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})
	locker.((*RedisLocker)).lockStore.Store("unlockkey", "lockuuid")

	cmd := redisv9.NewCmd(context.TODO())
	cmd.SetVal(int64(1))
	mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd, nil)

	ctx := context.Background()
	err := locker.Unlock(ctx, "unlockkey")
	assert.NoError(t, err)
}

func TestRedisUnlock_LockNotHeld(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	ctx := context.Background()
	err := locker.Unlock(ctx, "nonexistent")
	assert.Equal(t, ErrLockNotHeld, err)
}

func TestRedisUnlock_EvalFails(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	locker.(*RedisLocker).lockStore.Store("unlockkey", "lockuuid")

	mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(redisv9.NewCmdResult(nil, errors.New("eval error")))

	ctx := context.Background()
	err := locker.Unlock(ctx, "unlockkey")
	assert.EqualError(t, err, "eval error")
}

func TestRedisUnlock_EvalReturnsZero(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	locker.(*RedisLocker).lockStore.Store("unlockkey", "lockuuid")

	cmd := redisv9.NewCmd(context.Background())
	cmd.SetVal(int64(0))
	mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd)

	err := locker.Unlock(context.Background(), "unlockkey")
	assert.Equal(t, ErrLockNotHeld, err)
}

func TestRedisExtend_Success(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})
	locker.(*RedisLocker).lockStore.Store("extendkey", "lockuuid")

	mockCmd := redisv9.NewCmdResult(int64(1), nil)
	mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockCmd)

	ctx := context.Background()
	err := locker.Extend(ctx, "extendkey", 10*time.Second)
	assert.NoError(t, err)
}

func TestRedisExtend_LockNotHeld(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})

	ctx := context.Background()
	err := locker.Extend(ctx, "unknown", 10*time.Second)
	assert.Equal(t, ErrLockNotHeld, err)
}

func TestRedisExtend_EvalFails(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)

	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})
	locker.((*RedisLocker)).lockStore.Store("extendkey", "lockuuid")

	mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("eval error"))

	ctx := context.Background()
	err := locker.Extend(ctx, "extendkey", 10*time.Second)

	assert.EqualError(t, err, "internal error in redis command")
}

func TestRedisExtend_Errors(t *testing.T) {
	mockClient := &mocks.MockRedisClient{}
	mockRedis := &mockRedis{client: mockClient}
	mockRedis.On("GetClient").Return(mockClient)

	locker, _ := GetDistributedLocker(&LockOptions{LockerProvider: RedisLockProvider, StorageClient: mockRedis})
	locker.(*RedisLocker).lockStore.Store("extendkey", "lockuuid")

	t.Run("ResultError", func(t *testing.T) {
		cmd := redisv9.NewCmd(context.Background())
		cmd.SetErr(errors.New("result failed"))

		mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd).Once()

		err := locker.Extend(context.Background(), "extendkey", 10*time.Second)
		assert.EqualError(t, err, "result failed")
	})

	t.Run("WrongType", func(t *testing.T) {
		cmd := redisv9.NewCmd(context.Background())
		cmd.SetVal("not-int64")

		mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd).Once()

		err := locker.Extend(context.Background(), "extendkey", 10*time.Second)
		assert.EqualError(t, err, "unexpected result type from Eval")
	})

	t.Run("ZeroReturnValue", func(t *testing.T) {
		cmd := redisv9.NewCmd(context.Background())
		cmd.SetVal(int64(0))

		mockClient.On("Eval", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd).Once()

		err := locker.Extend(context.Background(), "extendkey", 10*time.Second)
		assert.Equal(t, ErrLockNotHeld, err)
	})
}
