// Package concurrency provides primitives to manage distributed concurrency patterns
// such as distributed locks.
package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/storage/redis"
	"github.com/google/uuid"
)

// RedisLocker manages distributed locks using a Redis backend.
//
// It safely handles lock acquisition, renewal (extend), and release (unlock)
// operations by using Lua scripts to ensure atomicity and consistency across
// multiple nodes.
//
// It internally maintains an in-memory map to track currently held locks locally,
// making unlock and extend operations more efficient.
type RedisLocker struct {
	storageclient           redis.RedisInterface
	defaultLockRetryTimeout time.Duration
	lockStore               sync.Map // Map of [Key -> LockUUID] for local tracking
}

// Lock attempts to acquire a distributed lock on the specified key with the given expiry.
//
// Parameters:
//   - ctx: Context with cancellation or deadline for retries.
//   - key: Unique key representing the lock.
//   - expiry: How long the lock should live.
//   - retryInterval: Time to wait before retrying if the lock is already held.
//
// Returns an error if the lock could not be acquired before context timeout/cancellation.
//
// Internally uses Redis SETNX operation to ensure safe lock acquisition.
func (d *RedisLocker) Lock(ctx context.Context, key string, expiry time.Duration, retryInterval time.Duration) error {
	lockValue := uuid.NewString()

	// Ensuøre context has a deadline to avoid infinite looping
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second) // Default internal timeout if caller didn't specify
		defer cancel()
	}

	for {
		ok, err := d.storageclient.GetClient().SetNX(ctx, key, lockValue, expiry).Result()
		if err != nil {
			return fmt.Errorf("distributed lock acquisition failed for key '%s': %w", key, err)
		}
		if ok {
			d.lockStore.Store(key, lockValue)
			return nil
		}

		select {
		case <-ctx.Done():
			return ErrLockNotAcquired
		case <-time.After(retryInterval):
			// Retry after interval
		}
	}
}

// Unlock safely releases a lock on the specified key.
//
// It verifies that the caller holds the lock by checking the value before deletion,
// preventing accidental unlocks from other owners.
//
// Returns an error if the lock was not held locally or if the lock ownership check fails.
func (d *RedisLocker) Unlock(ctx context.Context, key string) error {
	valueRaw, ok := d.lockStore.Load(key)
	if !ok {
		return ErrLockNotHeld
	}
	lockValue, _ := valueRaw.(string)

	const luaScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	res, err := d.storageclient.GetClient().Eval(ctx, luaScript, []string{key}, lockValue).Result()
	if err != nil {
		return err
	}

	if res.(int64) == 0 {
		return ErrLockNotHeld
	}

	d.lockStore.Delete(key)
	return nil
}

// Extend renews (extends) the TTL (expiry) of a held lock on the specified key.
//
// It verifies that the caller still holds the lock before extending, avoiding
// extending locks held by other owners.
//
// Parameters:
//   - ctx: Context for the operation.
//   - key: Key representing the lock.
//   - extension: Additional time to add to the lock's expiry.
//
// Returns an error if the lock is not held locally or renewal fails.
func (d *RedisLocker) Extend(ctx context.Context, key string, extension time.Duration) error {
	valueRaw, ok := d.lockStore.Load(key)
	if !ok {
		return ErrLockNotHeld
	}
	lockValue, _ := valueRaw.(string)

	const luaScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	millis := int64(extension / time.Millisecond)
	cmd := d.storageclient.GetClient().Eval(ctx, luaScript, []string{key}, lockValue, millis)

	if cmd == nil {
		return errors.New("internal error in redis command")
	}

	res, err := cmd.Result()
	if err != nil {
		return err
	}

	intRes, ok := res.(int64)
	if !ok {
		return errors.New("unexpected result type from Eval")
	}

	if intRes == 0 {
		return ErrLockNotHeld
	}

	return nil
}
