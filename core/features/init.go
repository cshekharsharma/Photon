package features

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/cshekharsharma/photon/core/logger"
)

//go:embed schema.json
var schemaJSON []byte // schemaJSON is the embedded JSON schema for feature validation

var (
	mutex                  sync.Mutex     // mutex to ensure thread-safe access to the feature flag engine
	areFeaturesInitialized bool           // indicates whether the feature flag engine has been initialized
	configCache            *FeatureConfig // holds the currently loaded and validated feature configuration
	featureManager         *Manager
	readFileFn             = os.ReadFile
	validateFeatureConfig  = ValidateFeatureConfig
	unmarshalFeatureConfig = json.Unmarshal
)

// Init initializes the feature flag engine by loading and validating the configuration.
//
// It accepts an InitOptions struct that specifies the source of the configuration (either a file path or raw JSON).
// The initialization process performs the following steps:
//  1. Validates the InitOptions.
//  2. Reads the configuration from the specified source.
//  3. Validates the config against the embedded JSON schema.
//  4. Unmarshals the config into a strongly typed FeatureConfig object.
//  5. Optionally initializes a file or remote watcher to monitor for updates.
//
// Init is guarded by a mutex to ensure thread safety, and it can only be executed once;
// subsequent calls will return immediately without reinitialization.
//
// Returns an error if:
//   - The options are invalid.
//   - Reading from the file source fails.
//   - Schema validation fails.
//   - JSON unmarshalling fails.
func Init(opts *InitOptions) error {
	mutex.Lock()
	defer mutex.Unlock()

	if areFeaturesInitialized {
		return nil
	}

	manager, err := NewManager(opts)
	if err != nil {
		return err
	}

	if featureManager != nil {
		featureManager.Close()
	}

	featureManager = manager
	configCache = manager.config
	areFeaturesInitialized = true
	return nil
}

func loadFeatureConfig(opts *InitOptions) (*FeatureConfig, *InitOptions, error) {
	var configData []byte
	var err error

	opts = populateRequiredOptionsProperties(opts)

	if err := opts.IsValid(); err != nil {
		return nil, nil, fmt.Errorf("invalid options provided: %w", err)
	}

	switch opts.SourceType {
	case FeatureSourceFile:
		configData, err = readFileFn(opts.Input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read config file: %w", err)
		}

	case FeatureSourceRawBytes:
		configData = []byte(opts.Input)
	}

	if err := validateFeatureConfig(configData, schemaJSON); err != nil {
		return nil, nil, fmt.Errorf("invalid feature config: %w", err)
	}

	cfg := &FeatureConfig{}
	if err := unmarshalFeatureConfig(configData, cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal validated feature config: %w", err)
	}

	return cfg, opts, nil
}

// GetFeatureConfigStore returns the currently loaded and validated feature configuration.
//
// The returned configuration reflects the most recent state, either from initial load or a dynamic update.
// If Init has not been called successfully, this function returns an error.
//
// Returns:
//   - *FeatureConfig: the current feature configuration in memory.
//   - error: if the configuration has not been initialized yet.
func GetFeatureConfigStore() (*FeatureConfig, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if !areFeaturesInitialized {
		return &FeatureConfig{}, fmt.Errorf("feature config is not initialized")
	}

	return configCache, nil
}

// populateRequiredOptionsProperties populates required properties in the options.
// It sets default values for the logger if not provided and ensures the watcher options are initialized.
// This function is called during the initialization of the feature config to ensure
func populateRequiredOptionsProperties(options *InitOptions) *InitOptions {
	if options == nil {
		return nil
	}
	if options.EnableWatch {
		if options.WatcherOptions.Logger == nil {
			options.WatcherOptions.Logger = logger.Init(&logger.LoggerConfig{
				Name:     "FeatureFlagWatcher",
				Level:    logger.LogLevelInfo,
				Provider: logger.LoggerProviderZerolog,
				Type:     logger.LoggerTypeStdout,
			})
		}
	}

	return options
}
