package session

import (
	"fmt"
	"net/http"
	"time"
)

const DefaultSessionName string = "GFRRSESSID"

// StoreType defines the backing store type.
type StoreType string

const (
	StoreMemory StoreType = "memory"
	StoreRedis  StoreType = "redis"
)

// Encoding defines the session encoding format.
type Encoding string

const (
	EncodingGob  Encoding = "gob"
	EncodingJSON Encoding = "json"
)

// ErrorHandler is invoked when the session middleware encounters an error.
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// Config controls session initialization.
type Config struct {
	Store            StoreConfig
	Cookie           CookieConfig
	Lifetime         time.Duration
	IdleTimeout      time.Duration
	HashTokenInStore *bool
	Encoding         Encoding
	ErrorHandler     ErrorHandler
}

// StoreConfig controls the session store configuration.
type StoreConfig struct {
	Type    StoreType
	Options StoreOptions
}

// StoreOptions holds store-specific configuration values.
type StoreOptions struct {
	Name       string
	Address    string
	Username   string
	Password   string
	DB         int
	KeyPrefix  string
	PoolSize   int
	PingOnInit *bool
}

// CookieConfig controls the session cookie settings.
type CookieConfig struct {
	Name        string
	Domain      string
	Path        string
	SameSite    http.SameSite
	Secure      *bool
	HttpOnly    *bool
	Persist     *bool
	Partitioned *bool
}

func applyDefaults(cfg *Config) (*Config, error) {
	if cfg == nil {
		return nil, ErrNilConfig
	}

	copyCfg := *cfg

	if copyCfg.Store.Type == "" {
		copyCfg.Store.Type = StoreMemory
	}

	if copyCfg.Encoding == "" {
		copyCfg.Encoding = EncodingGob
	}

	if !isValidEncoding(copyCfg.Encoding) {
		return nil, ErrInvalidEncoding
	}

	if copyCfg.Lifetime == 0 {
		copyCfg.Lifetime = 24 * time.Hour
	}

	if copyCfg.HashTokenInStore == nil {
		copyCfg.HashTokenInStore = boolPtr(true)
	}

	if err := applyCookieDefaults(&copyCfg.Cookie); err != nil {
		return nil, err
	}

	if copyCfg.Store.Type == StoreRedis {
		applyRedisDefaults(&copyCfg.Store.Options)
	}

	if !isValidStore(copyCfg.Store.Type) {
		return nil, ErrInvalidStore
	}

	return &copyCfg, nil
}

func isValidStore(store StoreType) bool {
	switch store {
	case StoreMemory, StoreRedis:
		return true
	default:
		return false
	}
}

func isValidEncoding(enc Encoding) bool {
	switch enc {
	case EncodingGob, EncodingJSON:
		return true
	default:
		return false
	}
}

func applyCookieDefaults(cookie *CookieConfig) error {
	if cookie.Name == "" {
		cookie.Name = DefaultSessionName
	}
	if cookie.Path == "" {
		cookie.Path = "/"
	}
	if cookie.SameSite == 0 {
		cookie.SameSite = http.SameSiteLaxMode
	}
	if cookie.Secure == nil {
		cookie.Secure = boolPtr(true)
	}
	if cookie.HttpOnly == nil {
		cookie.HttpOnly = boolPtr(true)
	}
	if cookie.Persist == nil {
		cookie.Persist = boolPtr(true)
	}
	if cookie.Partitioned == nil {
		cookie.Partitioned = boolPtr(false)
	}

	if cookie.SameSite == http.SameSiteNoneMode && !*cookie.Secure {
		return ErrInvalidCookie
	}

	return nil
}

func applyRedisDefaults(redisCfg *StoreOptions) {
	if redisCfg.Name == "" {
		redisCfg.Name = "default"
	}
	if redisCfg.Address == "" {
		redisCfg.Address = "127.0.0.1:6379"
	}
	if redisCfg.PingOnInit == nil {
		redisCfg.PingOnInit = boolPtr(true)
	}
	if redisCfg.KeyPrefix == "" {
		redisCfg.KeyPrefix = fmt.Sprintf("%s:", DefaultSessionName)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

// Bool returns a pointer to the provided bool value.
func Bool(v bool) *bool {
	return &v
}
