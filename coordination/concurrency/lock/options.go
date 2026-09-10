package concurrency

import (
	"time"
)

type DistributedLockProvider string

const (
	RedisLockProvider DistributedLockProvider = "redis"
)

// LockOptions is configuration struct for distributed lock implementation
type LockOptions struct {
	LockerProvider          DistributedLockProvider // type of distributed lock provider (e.g., Redis)
	StorageClient           interface{}             // storage client instance for in-memory lock storage
	DefaultLockRetryTimeout time.Duration           // default lock retry timeout, to be used if not provided with the Lock() op.
}
