package concurrency

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"
)

var (
	defaultLockRetryTimeout  = 10 * time.Second       // default retry timeout for lock acquisition
	defaultLockRetryInterval = 100 * time.Millisecond // default wait between acquisition attempts
)

// DistributedLocker is an interface for acquiring and releasing distributed locks.
// It provides methods to lock, unlock, and extend the expiry of a lock.
type DistributedLocker interface {
	// Lock attempts to acquire a distributed lock and returns a monotonically
	// increasing fencing token for the key. Callers must pass this token to the
	// guarded resource, which should reject writes with tokens older than the
	// greatest token it has already accepted.
	Lock(ctx context.Context, key string, expiry time.Duration, retryInterval time.Duration) (uint64, error)

	// FencingToken returns the locally held fencing token for key.
	FencingToken(key string) (uint64, bool)

	// Unlock releases the lock held by the caller on the specified key.
	// It checks if the caller is the owner of the lock before releasing it.
	Unlock(ctx context.Context, key string) error

	// Extend extends the expiry time of the lock held by the caller.
	Extend(ctx context.Context, key string, extension time.Duration) error
}

// GetDistributedLocker creates a new instance of DistributedLocker based on the provided options.
// It supports different locker providers, such as Redis.
func GetDistributedLocker(opts *LockOptions) (DistributedLocker, error) {
	if opts == nil {
		return nil, errors.New("options cannot be nil")
	}

	if opts.DefaultLockRetryTimeout == 0 {
		opts.DefaultLockRetryTimeout = defaultLockRetryTimeout
	}

	switch opts.LockerProvider {
	case RedisLockProvider:
		storageclient, ok := opts.StorageClient.(redis.RedisInterface)
		if !ok {
			return nil, errors.New("invalid storage client type for Redis locker")
		}
		return &RedisLocker{
			storageclient:           storageclient,
			defaultLockRetryTimeout: opts.DefaultLockRetryTimeout,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported locker provider: %s", opts.LockerProvider)
	}
}
