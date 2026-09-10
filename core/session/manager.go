package session

import (
	"context"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

// Manager provides a vendor-neutral session API over SCS.
type Manager struct {
	sm      *scs.SessionManager
	closeFn func() error
}

// New creates a new session manager with the provided configuration.
func New(config *Config) (*Manager, error) {
	cfg, err := applyDefaults(config)
	if err != nil {
		return nil, err
	}

	store, closeFn, err := buildStore(cfg)
	if err != nil {
		return nil, err
	}

	sm := scs.New()
	sm.Store = store
	sm.Lifetime = cfg.Lifetime
	sm.IdleTimeout = cfg.IdleTimeout
	sm.HashTokenInStore = *cfg.HashTokenInStore
	sm.Cookie = scs.SessionCookie{
		Name:        cfg.Cookie.Name,
		Domain:      cfg.Cookie.Domain,
		Path:        cfg.Cookie.Path,
		SameSite:    cfg.Cookie.SameSite,
		Secure:      *cfg.Cookie.Secure,
		HttpOnly:    *cfg.Cookie.HttpOnly,
		Persist:     *cfg.Cookie.Persist,
		Partitioned: *cfg.Cookie.Partitioned,
	}

	if cfg.Encoding == EncodingJSON {
		sm.Codec = jsonCodec{}
	}

	if cfg.ErrorHandler != nil {
		sm.ErrorFunc = cfg.ErrorHandler
	}

	return &Manager{sm: sm, closeFn: closeFn}, nil
}

// Middleware returns a middleware using this manager.
func (m *Manager) Middleware() func(http.Handler) http.Handler {
	return m.sm.LoadAndSave
}

// Load manually loads session data into the context for non-middleware usage.
func (m *Manager) Load(ctx context.Context, token string) (context.Context, error) {
	if m == nil || m.sm == nil {
		return nil, ErrNotInitialized
	}
	return m.sm.Load(ctx, token)
}

// Close closes the underlying store if supported.
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

// Put stores a value in the session.
func (m *Manager) Put(ctx context.Context, key string, val interface{}) error {
	return m.guard(func() error {
		m.sm.Put(ctx, key, val)
		return nil
	})
}

// Get fetches a value from the session.
func (m *Manager) Get(ctx context.Context, key string) (interface{}, bool, error) {
	var (
		val interface{}
		ok  bool
	)
	err := m.guard(func() error {
		val = m.sm.Get(ctx, key)
		ok = m.sm.Exists(ctx, key)
		return nil
	})
	return val, ok, err
}

// GetString fetches a string value from the session.
func (m *Manager) GetString(ctx context.Context, key string) (string, error) {
	var val string
	return val, m.guard(func() error {
		val = m.sm.GetString(ctx, key)
		return nil
	})
}

// GetInt fetches an int value from the session.
func (m *Manager) GetInt(ctx context.Context, key string) (int, error) {
	var val int
	return val, m.guard(func() error {
		val = m.sm.GetInt(ctx, key)
		return nil
	})
}

// GetBool fetches a bool value from the session.
func (m *Manager) GetBool(ctx context.Context, key string) (bool, error) {
	var val bool
	return val, m.guard(func() error {
		val = m.sm.GetBool(ctx, key)
		return nil
	})
}

// Pop removes and returns a value from the session.
func (m *Manager) Pop(ctx context.Context, key string) (interface{}, bool, error) {
	var (
		val interface{}
		ok  bool
	)
	err := m.guard(func() error {
		ok = m.sm.Exists(ctx, key)
		val = m.sm.Pop(ctx, key)
		return nil
	})
	return val, ok, err
}

// Remove deletes a value from the session.
func (m *Manager) Remove(ctx context.Context, key string) error {
	return m.guard(func() error {
		m.sm.Remove(ctx, key)
		return nil
	})
}

// Clear removes all session values.
func (m *Manager) Clear(ctx context.Context) error {
	return m.guard(func() error {
		return m.sm.Clear(ctx)
	})
}

// Exists reports whether a key exists.
func (m *Manager) Exists(ctx context.Context, key string) (bool, error) {
	var ok bool
	return ok, m.guard(func() error {
		ok = m.sm.Exists(ctx, key)
		return nil
	})
}

// Renew rotates the session token.
func (m *Manager) Renew(ctx context.Context) error {
	return m.guard(func() error {
		return m.sm.RenewToken(ctx)
	})
}

// Destroy clears the session and deletes it from the store.
func (m *Manager) Destroy(ctx context.Context) error {
	return m.guard(func() error {
		return m.sm.Destroy(ctx)
	})
}

// Token returns the current session token.
func (m *Manager) Token(ctx context.Context) (string, error) {
	var token string
	return token, m.guard(func() error {
		token = m.sm.Token(ctx)
		return nil
	})
}

// Deadline returns the absolute expiry time.
func (m *Manager) Deadline(ctx context.Context) (time.Time, error) {
	var deadline time.Time
	return deadline, m.guard(func() error {
		deadline = m.sm.Deadline(ctx)
		return nil
	})
}

// guard wraps session operations to return a consistent error instead of panicking.
func (m *Manager) guard(fn func() error) (err error) {
	if m == nil || m.sm == nil {
		return ErrNotInitialized
	}
	defer func() {
		if r := recover(); r != nil {
			err = ErrNoSessionInContext
		}
	}()
	return fn()
}
