package aws

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/cloud/entity/appconfig"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
)

const restartDelay = 5 * time.Second

// awsAppConfigClientInterface defines the methods for AWS AppConfigData client.
type awsAppConfigClientInterface interface {
	StartConfigurationSession(context.Context, *appconfigdata.StartConfigurationSessionInput, ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error)
	GetLatestConfiguration(context.Context, *appconfigdata.GetLatestConfigurationInput, ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error)
}

// AppConfig provides AWS AppConfig service operations.
type AppConfig struct {
	AcClient       awsAppConfigClientInterface
	watchers       map[string]context.CancelFunc
	mutex          sync.RWMutex
	watchedConfigs map[string]*appconfig.FetchConfigResult
	lastHashes     map[string]string
	restartDelay   time.Duration

	watchConfigOnceFn         func(ctx context.Context, input *appconfig.WatchConfigInput, configToken *string, callback func(*appconfig.FetchConfigResult)) (*string, error)
	beforeWatchConfigOnceLock func(a *AppConfig)
}

// FetchConfig fetches a configuration from AWS AppConfig.
func (a *AppConfig) FetchConfig(ctx context.Context, input *appconfig.FetchConfigInput) (*appconfig.FetchConfigResult, error) {
	sessionInput := &appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:          aws.String(input.ApplicationName),
		ConfigurationProfileIdentifier: aws.String(input.ConfigProfile),
		EnvironmentIdentifier:          aws.String(input.EnvironmentName),
	}

	sessionResp, err := a.AcClient.StartConfigurationSession(ctx, sessionInput)
	if err != nil {
		return nil, fmt.Errorf("failed to start configuration session: %w", err)
	}

	if sessionResp.InitialConfigurationToken == nil {
		return nil, errors.New("received nil InitialConfigurationToken")
	}

	latestConfigInput := &appconfigdata.GetLatestConfigurationInput{
		ConfigurationToken: sessionResp.InitialConfigurationToken,
	}

	configResp, err := a.AcClient.GetLatestConfiguration(ctx, latestConfigInput)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch configuration: %w", err)
	}

	return &appconfig.FetchConfigResult{
		ConfigProfile: input.ConfigProfile,
		Content:       string(configResp.Configuration),
		ContentType:   aws.ToString(configResp.ContentType),
	}, nil
}

// WatchConfig starts watching a configuration for changes.
func (a *AppConfig) WatchConfig(ctx context.Context, input *appconfig.WatchConfigInput, callback func(*appconfig.FetchConfigResult)) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.watchers == nil {
		a.watchers = make(map[string]context.CancelFunc)
	}

	if a.watchedConfigs == nil {
		a.watchedConfigs = make(map[string]*appconfig.FetchConfigResult)
	}

	// If a watcher for this config ID is already running, refuse to start another.
	if _, exists := a.watchers[input.UniqueConfigID]; exists {
		return fmt.Errorf("watcher for config %s[%s] already exists", input.ConfigProfile, input.UniqueConfigID)
	}

	// Create a derived context for this watcher
	watchCtx, cancel := context.WithCancel(ctx)
	a.watchers[input.UniqueConfigID] = cancel

	sessionInput := &appconfigdata.StartConfigurationSessionInput{
		ApplicationIdentifier:          aws.String(input.Application),
		ConfigurationProfileIdentifier: aws.String(input.ConfigProfile),
		EnvironmentIdentifier:          aws.String(input.Environment),
	}

	sessionResp, err := a.AcClient.StartConfigurationSession(watchCtx, sessionInput)
	if err != nil {
		delete(a.watchers, input.UniqueConfigID)
		cancel()
		return fmt.Errorf("failed to start configuration session for %s: %v", input.UniqueConfigID, err)
	}

	if sessionResp.InitialConfigurationToken == nil {
		delete(a.watchers, input.UniqueConfigID)
		cancel()
		return fmt.Errorf("received nil InitialConfigurationToken for %s", input.UniqueConfigID)
	}

	go a.startConfigWatcherLoop(watchCtx, input, sessionResp.InitialConfigurationToken, callback)

	return nil
}

// startConfigWatcherLoop runs "watchConfigOnce" in a loop, handling errors and restarts.
// If ctx is canceled, or the loop function returns normally, we exit.
func (a *AppConfig) startConfigWatcherLoop(
	ctx context.Context, input *appconfig.WatchConfigInput,
	configToken *string, callback func(*appconfig.FetchConfigResult),
) {

	delay := a.restartDelay
	if delay == 0 {
		delay = restartDelay
	}

	defer func() {
		if r := recover(); r != nil {
			logWatchError(input, "Watcher loop for %s panicked: %v. Attempting restart...", input.UniqueConfigID, r)

			select {
			case <-ctx.Done():
				logWatchWarn(input, "Context canceled, not restarting watcher loop for %s.", input.UniqueConfigID)
				return
			case <-time.After(delay):
				go a.startConfigWatcherLoop(ctx, input, configToken, callback)
			}
		}
	}()

	var watchOnce func(context.Context, *appconfig.WatchConfigInput, *string, func(*appconfig.FetchConfigResult)) (*string, error)
	if a.watchConfigOnceFn != nil {
		watchOnce = a.watchConfigOnceFn
	} else {
		watchOnce = a.watchConfigOnce
	}

	for {
		select {
		case <-ctx.Done():
			logWatchWarn(input, "Context canceled for %s[%s], stopping watcher loop", input.ConfigProfile, input.UniqueConfigID)
			return
		default:
		}

		// Attempt a single watch cycle
		newToken, err := watchOnce(ctx, input, configToken, callback)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logWatchWarn(input, "watchConfigOnce canceled for %s, stopping", input.UniqueConfigID)
				return
			}

			logWatchError(input, "Watcher for %s failed: %v. Restarting in %s", input.UniqueConfigID, err, delay)

			select {
			case <-ctx.Done():
				logWatchWarn(input, "Context canceled during restart delay for %s, stopping", input.UniqueConfigID)
				return
			case <-time.After(delay):
				// proceed to restart
			}

			configToken = newToken

		} else {
			// No error means watchConfigOnce ended "normally"—which in practice, usually won't happen
			// unless you decide to break out once some condition is met.
			logWatchInfo(input, "Watcher for %s ended normally", input.UniqueConfigID)
			return
		}
	}
}

// watchConfigOnce performs a single continuous watch loop using the given token.
// It returns either when ctx is canceled, or upon error/panic.
//
// If it panics, we recover and return an error so the caller can decide to restart.
func (a *AppConfig) watchConfigOnce(ctx context.Context, input *appconfig.WatchConfigInput, configToken *string, callback func(*appconfig.FetchConfigResult)) (newToken *string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered in watcher for %s: %v", input.UniqueConfigID, r)
		}
	}()

	ticker := time.NewTicker(input.PollInterval)
	defer ticker.Stop()

	newToken = configToken

	for {
		select {
		case <-ctx.Done():
			return newToken, ctx.Err()

		case <-ticker.C:
			configResp, fetchErr := a.AcClient.GetLatestConfiguration(ctx, &appconfigdata.GetLatestConfigurationInput{
				ConfigurationToken: newToken,
			})

			if fetchErr != nil {
				return newToken, fmt.Errorf("failed to fetch config for %s: %w", input.UniqueConfigID, fetchErr)
			}

			newToken = configResp.NextPollConfigurationToken

			if len(configResp.Configuration) > 0 {
				hash := sha256.Sum256(configResp.Configuration)
				hashStr := fmt.Sprintf("%x", hash[:])
				result := &appconfig.FetchConfigResult{
					ConfigProfile: input.ConfigProfile,
					ConfigID:      input.UniqueConfigID,
					Content:       string(configResp.Configuration),
					ContentType:   aws.ToString(configResp.ContentType),
				}

				a.mutex.Lock()
				if a.beforeWatchConfigOnceLock != nil {
					a.beforeWatchConfigOnceLock(a)
				}

				if a.lastHashes == nil {
					a.lastHashes = make(map[string]string)
				}

				if lastHash, ok := a.lastHashes[input.UniqueConfigID]; ok && lastHash == hashStr {
					a.mutex.Unlock()
					continue // Same content as before, skip update
				}

				a.lastHashes[input.UniqueConfigID] = hashStr

				if a.watchedConfigs == nil {
					a.watchedConfigs = make(map[string]*appconfig.FetchConfigResult)
				}

				a.watchedConfigs[input.UniqueConfigID] = result
				a.mutex.Unlock()

				if callback != nil {
					callback(result)
				}
			}
		}
	}
}

// GetCurrentConfiguration returns the current configuration for a given uniqueConfigId.
func (a *AppConfig) GetCurrentConfiguration(uniqueConfigId string) (*appconfig.FetchConfigResult, error) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	configResult, exists := a.watchedConfigs[uniqueConfigId]
	if !exists {
		return nil, fmt.Errorf("no configuration found for %s", uniqueConfigId)
	}

	return configResult, nil
}

// StopWatching stops watching a configuration.
func (a *AppConfig) StopWatching(configID string) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if cancel, exists := a.watchers[configID]; exists {
		cancel()
		delete(a.watchers, configID)
		return nil
	}
	return fmt.Errorf("no watcher found for config %s", configID)
}

// StopAllWatching stops all active watchers.
func (a *AppConfig) StopAllWatching() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	for id, cancel := range a.watchers {
		cancel()
		delete(a.watchers, id)
	}

	a.watchedConfigs = make(map[string]*appconfig.FetchConfigResult) // free memory
}

func logWatchInfo(input *appconfig.WatchConfigInput, message string, args ...interface{}) {
	if input != nil && input.Logger != nil {
		input.Logger.Info(message, args...)
	}
}

func logWatchWarn(input *appconfig.WatchConfigInput, message string, args ...interface{}) {
	if input != nil && input.Logger != nil {
		input.Logger.Warn(message, args...)
	}
}

func logWatchError(input *appconfig.WatchConfigInput, message string, args ...interface{}) {
	if input != nil && input.Logger != nil {
		input.Logger.Error(message, args...)
	}
}
