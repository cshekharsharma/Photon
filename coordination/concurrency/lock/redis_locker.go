// Package concurrency provides primitives to manage distributed concurrency patterns
// such as distributed locks.
package concurrency

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	lockStore               sync.Map // Map of [Key -> redisLockState] for local tracking
}

type redisLockState struct {
	value string
	token uint64
}

// Lock attempts to acquire a distributed lock on the specified key with the given expiry.
//
// Parameters:
//   - ctx: Context with cancellation or deadline for retries.
//   - key: Unique key representing the lock.
//   - expiry: How long the lock should live.
//   - retryInterval: Time to wait before retrying if the lock is already held.
//
// Returns the fencing token and an error if the lock could not be acquired before
// context timeout/cancellation.
//
// Internally uses Redis Lua scripting to atomically acquire ownership and issue
// a fencing token.
func (d *RedisLocker) Lock(
	ctx context.Context,
	key string,
	expiry time.Duration,
	retryInterval time.Duration,
) (uint64, error) {
	if expiry <= 0 {
		return 0, errors.New("lock expiry must be positive")
	}
	if retryInterval <= 0 {
		retryInterval = defaultLockRetryInterval
	}
	owner := uuid.NewString()
	expiryMillis := durationMilliseconds(expiry)

	// Ensure context has a deadline to avoid infinite looping.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		timeout := d.defaultLockRetryTimeout
		if timeout <= 0 {
			timeout = defaultLockRetryTimeout
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		token, err := d.tryAcquire(ctx, key, owner, expiryMillis)
		if err != nil {
			return 0, fmt.Errorf("distributed lock acquisition failed for key '%s': %w", key, err)
		}
		if token > 0 {
			d.lockStore.Store(key, redisLockState{
				value: lockValue(token, owner),
				token: token,
			})
			return token, nil
		}

		timer := time.NewTimer(retryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return 0, ErrLockNotAcquired
		case <-timer.C:
		}
	}
}

// FencingToken returns the locally held fencing token for key.
func (d *RedisLocker) FencingToken(key string) (uint64, bool) {
	state, ok := d.lockState(key)
	if !ok || state.token == 0 {
		return 0, false
	}
	return state.token, true
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
	state, ok := redisLockStateFromValue(valueRaw)
	if !ok {
		return ErrLockNotHeld
	}

	const luaScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	res, err := d.storageclient.GetClient().Eval(ctx, luaScript, []string{key}, state.value).Result()
	if err != nil {
		return err
	}

	intRes, ok := res.(int64)
	if !ok {
		return errors.New("unexpected result type from Eval")
	}
	if intRes == 0 {
		d.lockStore.Delete(key)
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
	if extension <= 0 {
		return errors.New("lock extension must be positive")
	}

	valueRaw, ok := d.lockStore.Load(key)
	if !ok {
		return ErrLockNotHeld
	}
	state, ok := redisLockStateFromValue(valueRaw)
	if !ok {
		return ErrLockNotHeld
	}

	const luaScript = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	millis := durationMilliseconds(extension)
	cmd := d.storageclient.GetClient().Eval(ctx, luaScript, []string{key}, state.value, millis)

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
		d.lockStore.Delete(key)
		return ErrLockNotHeld
	}

	return nil
}

func (d *RedisLocker) tryAcquire(ctx context.Context, key string, owner string, expiryMillis int64) (uint64, error) {
	const luaScript = `
		if redis.call("EXISTS", KEYS[1]) == 1 then
			return 0
		end
		local token = redis.call("INCR", KEYS[2])
		redis.call("PSETEX", KEYS[1], ARGV[1], token .. ":" .. ARGV[2])
		return token
	`

	res, err := d.storageclient.GetClient().
		Eval(ctx, luaScript, []string{key, fencingCounterKey(key)}, expiryMillis, owner).
		Result()
	if err != nil {
		return 0, err
	}
	return positiveUint64(res)
}

func (d *RedisLocker) lockState(key string) (redisLockState, bool) {
	valueRaw, ok := d.lockStore.Load(key)
	if !ok {
		return redisLockState{}, false
	}
	return redisLockStateFromValue(valueRaw)
}

func redisLockStateFromValue(value any) (redisLockState, bool) {
	switch typed := value.(type) {
	case redisLockState:
		return typed, typed.value != ""
	case string:
		return redisLockState{value: typed}, typed != ""
	default:
		return redisLockState{}, false
	}
}

func positiveUint64(value any) (uint64, error) {
	switch typed := value.(type) {
	case int64:
		if typed < 0 {
			return 0, errors.New("unexpected negative fencing token")
		}
		return strconv.ParseUint(strconv.FormatInt(typed, 10), 10, 64)
	case uint64:
		return typed, nil
	case int:
		if typed < 0 {
			return 0, errors.New("unexpected negative fencing token")
		}
		return strconv.ParseUint(strconv.Itoa(typed), 10, 64)
	default:
		return 0, errors.New("unexpected result type from Eval")
	}
}

func durationMilliseconds(duration time.Duration) int64 {
	millis := duration / time.Millisecond
	if duration%time.Millisecond != 0 {
		millis++
	}
	return int64(millis)
}

func lockValue(token uint64, owner string) string {
	return strconv.FormatUint(token, 10) + ":" + owner
}

func fencingCounterKey(key string) string {
	return key + ":fence"
}
