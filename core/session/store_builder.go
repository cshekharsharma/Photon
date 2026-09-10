package session

import "github.com/alexedwards/scs/v2"

type storeWithCloser struct {
	store   scs.Store
	closeFn func() error
}

func buildStore(cfg *Config) (scs.Store, func() error, error) {
	switch cfg.Store.Type {
	case StoreMemory:
		mem := newMemoryStore()
		return mem.store, mem.closeFn, nil

	case StoreRedis:
		applyRedisDefaults(&cfg.Store.Options)
		redisStore, closeFn, err := newRedisStore(cfg.Store.Options)
		if err != nil {
			return nil, nil, err
		}
		return redisStore, closeFn, nil

	default:
		return nil, nil, ErrInvalidStore
	}
}
