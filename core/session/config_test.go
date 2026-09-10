package session

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplyDefaults_NilConfig(t *testing.T) {
	cfg, err := applyDefaults(nil)
	assert.Nil(t, cfg)
	assert.ErrorIs(t, err, ErrNilConfig)
}

func TestApplyDefaults_Defaults(t *testing.T) {
	cfg, err := applyDefaults(&Config{})
	assert.NoError(t, err)
	assert.Equal(t, StoreMemory, cfg.Store.Type)
	assert.Equal(t, EncodingGob, cfg.Encoding)
	assert.Equal(t, 24*time.Hour, cfg.Lifetime)
	assert.True(t, *cfg.HashTokenInStore)
	assert.Equal(t, DefaultSessionName, cfg.Cookie.Name)
	assert.Equal(t, "/", cfg.Cookie.Path)
	assert.Equal(t, http.SameSiteLaxMode, cfg.Cookie.SameSite)
	assert.True(t, *cfg.Cookie.Secure)
	assert.True(t, *cfg.Cookie.HttpOnly)
	assert.True(t, *cfg.Cookie.Persist)
	assert.False(t, *cfg.Cookie.Partitioned)
}

func TestApplyDefaults_InvalidEncoding(t *testing.T) {
	_, err := applyDefaults(&Config{Encoding: Encoding("nope")})
	assert.ErrorIs(t, err, ErrInvalidEncoding)
}

func TestApplyDefaults_InvalidStore(t *testing.T) {
	_, err := applyDefaults(&Config{Store: StoreConfig{Type: StoreType("nope")}})
	assert.ErrorIs(t, err, ErrInvalidStore)
}

func TestApplyDefaults_InvalidCookie(t *testing.T) {
	secure := false
	cfg := &Config{Cookie: CookieConfig{SameSite: http.SameSiteNoneMode, Secure: &secure}}
	_, err := applyDefaults(cfg)
	assert.ErrorIs(t, err, ErrInvalidCookie)
}

func TestApplyDefaults_RedisDefaults(t *testing.T) {
	cfg, err := applyDefaults(&Config{Store: StoreConfig{Type: StoreRedis}})
	assert.NoError(t, err)
	assert.Equal(t, "default", cfg.Store.Options.Name)
	assert.Equal(t, "127.0.0.1:6379", cfg.Store.Options.Address)
	assert.Equal(t, fmt.Sprintf("%s:", DefaultSessionName), cfg.Store.Options.KeyPrefix)
	assert.True(t, *cfg.Store.Options.PingOnInit)
}

func TestApplyDefaults_RedisAddressValidation(t *testing.T) {
	cfg := &Config{Store: StoreConfig{Type: StoreRedis, Options: StoreOptions{Address: ""}}}
	normalized, err := applyDefaults(cfg)
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1:6379", normalized.Store.Options.Address)
}

func TestBoolHelper(t *testing.T) {
	v := Bool(true)
	assert.NotNil(t, v)
	assert.True(t, *v)
}
