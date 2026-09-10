package config

import (
	"bytes"
	"context"
	"testing"

	"github.com/cshekharsharma/photon/coordination/network/watcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rawManagerOptions(content string) *Options {
	return &Options{
		Source:    SourceRawBytes,
		Format:    FormatJson,
		Content:   []byte(content),
		Delimiter: ".",
	}
}

func TestManagerLifecycle(t *testing.T) {
	var nilManager *Manager
	cfg, err := nilManager.Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	nilManager.Close()

	opts := rawManagerOptions(`{"key":"value"}`)
	mgr, err := NewManager("", opts)
	require.NoError(t, err)
	assert.Equal(t, ConfigProviderKoanf, mgr.provider)

	opts.Content[0] = '['
	loaded, err := mgr.Load()
	require.NoError(t, err)
	assert.Equal(t, "value", loaded.GetString("key"))

	again, err := mgr.Load()
	require.NoError(t, err)
	assert.Same(t, loaded, again)

	replacement := &Koanf{}
	mgr.setConfig(replacement)
	assert.Same(t, replacement, mgr.config)

	mgr.Close()
	assert.Nil(t, mgr.config)
}

func TestManagerErrorsAndBranches(t *testing.T) {
	_, err := NewManager("unknown", rawManagerOptions(`{"key":"value"}`))
	require.Error(t, err)

	assert.Nil(t, cloneOptions(nil))

	mgr := &Manager{provider: "unknown", options: rawManagerOptions(`{"key":"value"}`)}
	cfg, err := mgr.newConfigLocked()
	require.Error(t, err)
	assert.Nil(t, cfg)

	(&Manager{}).startWatcherLocked(&Koanf{})
	(&Manager{options: &Options{EnableWatch: true}}).startWatcherLocked(&Koanf{})
}

func TestManagerStartWatcherUsesOwnedLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	opts := rawManagerOptions(`{"key":"value"}`)
	opts.EnableWatch = true
	opts.WatcherOptions = &watcher.WatcherOptions{
		WatchContext:  ctx,
		WatcherType:   "unsupported",
		Logger:        getConsoleLogger("manager-watch", &bytes.Buffer{}),
		UpdateChannel: make(chan *watcher.UpdaterSchema, 1),
	}

	mgr, err := NewManager(ConfigProviderKoanf, opts)
	require.NoError(t, err)
	defer mgr.Close()

	loaded, err := mgr.Load()
	require.NoError(t, err)
	assert.Equal(t, "value", loaded.GetString("key"))
	assert.NotNil(t, mgr.cancel)
	assert.NotSame(t, opts.WatcherOptions, mgr.options.WatcherOptions)
	assert.NotEqual(t, opts.WatcherOptions.UpdateChannel, mgr.options.WatcherOptions.UpdateChannel)
}

func TestManagerStartWatcherDefaultsContext(t *testing.T) {
	opts := rawManagerOptions(`{"key":"value"}`)
	opts.EnableWatch = true
	opts.WatcherOptions = &watcher.WatcherOptions{
		Logger: getConsoleLogger("manager-watch-default-context", &bytes.Buffer{}),
	}

	mgr, err := NewManager(ConfigProviderKoanf, opts)
	require.NoError(t, err)
	defer mgr.Close()

	loaded, err := mgr.Load()
	require.NoError(t, err)
	assert.Equal(t, "value", loaded.GetString("key"))
	assert.NotNil(t, mgr.cancel)
}

func TestInitReplacesExistingManager(t *testing.T) {
	resetConfigTestState()

	require.NoError(t, Init(ConfigProviderKoanf, rawManagerOptions(`{"key":"first"}`)))
	firstManager := cManager
	require.NotNil(t, firstManager)

	require.NoError(t, Init(ConfigProviderKoanf, rawManagerOptions(`{"key":"second"}`)))
	assert.NotSame(t, firstManager, cManager)
}

func TestKoanfApplyUpdateOwnedReloadAndChannels(t *testing.T) {
	callbackCalled := false
	onReloadCalled := false
	privateUpdates := make(chan *watcher.UpdaterSchema, 1)

	k := &Koanf{
		opts: &Options{
			Source:    SourceRawBytes,
			Format:    FormatJson,
			Content:   []byte(`{"key":"value"}`),
			Delimiter: ".",
			WatcherOptions: &watcher.WatcherOptions{
				UpdateChannel: privateUpdates,
				Logger:        getConsoleLogger("koanf-owned", &bytes.Buffer{}),
				OnUpdateCallback: func() {
					callbackCalled = true
				},
			},
		},
		onReload: func(cfg Config) {
			onReloadCalled = true
			assert.Equal(t, "next", cfg.GetString("key"))
		},
	}

	var expectedUpdates <-chan *watcher.UpdaterSchema = privateUpdates
	assert.Equal(t, expectedUpdates, k.updateChannel())
	k.applyUpdate(nil)
	k.applyUpdate(&watcher.UpdaterSchema{
		ContentSource: uint8(SourceRawBytes),
		Content:       `{"key":"next"}`,
	})

	assert.True(t, onReloadCalled)
	assert.True(t, callbackCalled)
	assert.Equal(t, []byte(`{"key":"next"}`), k.opts.Content)
}
