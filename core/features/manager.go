package features

import (
	"context"
	"sync"

	"github.com/cshekharsharma/photon/coordination/network/watcher"
)

// Manager owns one feature-config store and its watcher lifecycle.
type Manager struct {
	mu      sync.RWMutex
	options *InitOptions
	config  *FeatureConfig

	updates chan *watcher.UpdaterSchema
	cancel  context.CancelFunc
}

func NewManager(opts *InitOptions) (*Manager, error) {
	config, populatedOpts, err := loadFeatureConfig(cloneInitOptions(opts))
	if err != nil {
		return nil, err
	}

	m := &Manager{
		options: populatedOpts,
		config:  config,
		updates: make(chan *watcher.UpdaterSchema, 1),
	}
	m.startWatcher()
	return m, nil
}

func (m *Manager) GetFeatureConfigStore() (*FeatureConfig, error) {
	if m == nil {
		return &FeatureConfig{}, errFeatureManagerNotInitialized()
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return &FeatureConfig{}, errFeatureManagerNotInitialized()
	}
	return m.config, nil
}

func (m *Manager) Close() {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.config = nil
}

func (m *Manager) reload(entry *watcher.UpdaterSchema) error {
	if err := applyFeatureUpdateToOptions(m.options, entry); err != nil {
		return err
	}

	config, opts, err := loadFeatureConfig(m.options)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.options = opts
	m.config = config
	m.mu.Unlock()

	return nil
}

func (m *Manager) startWatcher() {
	if m == nil || m.options == nil || !m.options.EnableWatch || m.options.WatcherOptions == nil {
		return
	}

	opts := m.options.WatcherOptions
	opts.ContentSource = uint8(m.options.SourceType)
	opts.ContentFormat = uint8(FeatureContentFormatJson)
	opts.UpdateChannel = m.updates

	watchCtx := opts.WatchContext
	if watchCtx == nil {
		watchCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(watchCtx)
	m.cancel = cancel
	opts.WatchContext = ctx

	go m.watchUpdaterChannel(ctx)

	w := watcher.NewAwsAppConfigWatcher(opts, string(m.options.SourceType))
	go w.Watch(ctx)
}

func (m *Manager) watchUpdaterChannel(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case entry, ok := <-m.updates:
			if !ok {
				return
			}
			if err := m.reload(entry); err != nil {
				if m.options != nil && m.options.WatcherOptions != nil && m.options.WatcherOptions.Logger != nil {
					m.options.WatcherOptions.Logger.Error("[FeatureWatch] Error reloading features: %v", err)
				}
				continue
			}
			updateDefaultFeatureStore(m.config)
			callFeatureUpdateCallback(m.options)
			if m.options != nil && m.options.WatcherOptions != nil && m.options.WatcherOptions.Logger != nil {
				m.options.WatcherOptions.Logger.Info("[FeatureWatch] Features updated successfully through updater channel: "+
					"BytesWritten=%d", len(entry.Content))
			}
		}
	}
}

func cloneInitOptions(opts *InitOptions) *InitOptions {
	if opts == nil {
		return nil
	}

	copied := *opts
	if opts.WatcherOptions != nil {
		watcherOptions := *opts.WatcherOptions
		copied.WatcherOptions = &watcherOptions
	}
	return &copied
}

func errFeatureManagerNotInitialized() error {
	return featureConfigNotInitializedError{}
}

type featureConfigNotInitializedError struct{}

func (featureConfigNotInitializedError) Error() string {
	return "feature config is not initialized"
}
