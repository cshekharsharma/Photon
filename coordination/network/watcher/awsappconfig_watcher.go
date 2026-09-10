package watcher

import (
	"context"

	"github.com/cshekharsharma/photon/cloud"
	"github.com/cshekharsharma/photon/cloud/entity/appconfig"
)

// AwsAppConfigWatcher is responsible for watching configuration changes.
type AwsAppConfigWatcher struct {
	watcherOptions *WatcherOptions
}

// NewAwsAppConfigWatcher creates a new watcher with the config and update handler.
func NewAwsAppConfigWatcher(cfg *WatcherOptions, configSource string) *AwsAppConfigWatcher {

	return &AwsAppConfigWatcher{
		watcherOptions: cfg,
	}
}

// Watch begins watching AppConfig. It triggers the updater on change.
func (a *AwsAppConfigWatcher) Watch(ctx context.Context) {
	watchConfigInput := &appconfig.WatchConfigInput{
		UniqueConfigID: a.generateUniqueID(),
		Application:    a.watcherOptions.Application,
		Environment:    a.watcherOptions.Environment,
		ConfigProfile:  a.watcherOptions.ContentScope,
		PollInterval:   a.watcherOptions.PollInterval,
		ClientID:       a.watcherOptions.ClientID,
		Logger:         a.watcherOptions.Logger,
	}

	appConfigSvc, err := cloud.GetAppConfigService()
	if err != nil {
		a.watcherOptions.Logger.Error("[AppConfigWatcher] failed to initialize service: %v\n", err)
		return
	}

	err = appConfigSvc.WatchConfig(ctx, watchConfigInput, a.OnContentUpdate)

	if err != nil {
		a.watcherOptions.Logger.Error("[AppConfigWatcher] WatchConfig failed: %v\n", err)
		return
	}
}

func (a *AwsAppConfigWatcher) OnContentUpdate(result *appconfig.FetchConfigResult) {
	if result == nil {
		a.watcherOptions.Logger.Error("[AppConfigWatcher] received nil config result to update")
		return
	}

	if result.Content == "" {
		a.watcherOptions.Logger.Error("[AppConfigWatcher] received nil config file content from AWS appconfig")
		return
	}

	a.watcherOptions.Logger.Debug("[AppConfigWatcher] Pushing configuration update to updater channel")

	update := &UpdaterSchema{
		ContentSource: a.watcherOptions.ContentSource,
		ContentFormat: a.watcherOptions.ContentFormat,
		Content:       result.Content,
	}
	if a.watcherOptions.UpdateChannel != nil {
		a.watcherOptions.UpdateChannel <- update
		return
	}

	PushToContentUpdateChannelFn(update, a.watcherOptions.ContentSource)
}
