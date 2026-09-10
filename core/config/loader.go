package config

import (
	"fmt"
	"sync"
)

const (
	DefaultConfigFilePath      string = ""
	DefaultConfigPathDelimiter string = "."

	ConfigProviderKoanf string = "koanf"
)

var (
	cMutex    sync.Mutex
	instance  Config = nil
	cProvider string = ConfigProviderKoanf
	cManager  *Manager

	cOptions         *Options
	allowedProviders []string = []string{
		ConfigProviderKoanf,
	}
)

// Init initializes the configuration system with a specified provider and configuration options.
// This function sets up the configuration provider if it's listed as allowed and validates the provided options.
//
// Parameters:
//   - provider: A string representing the configuration provider to be used, which must be one of the allowed providers.
//   - options: A pointer to Options which contains necessary options that need to be validated and used during initialization.
//
// Usage:
//   - This function should be called at the start of an application to set up the configuration provider.
//   - If the options are not valid or the provider is not allowed, the application will terminate with a fatal log error.
func Init(provider string, options *Options) error {
	if provider == "" {
		provider = ConfigProviderKoanf
	}
	manager, err := NewManager(provider, options)
	if err != nil {
		return err
	}

	cMutex.Lock()
	defer cMutex.Unlock()

	if cManager != nil {
		cManager.Close()
	}
	cProvider = provider
	cOptions = manager.options
	instance = nil
	cManager = manager

	return nil
}

// Load returns a singleton instance of the Config interface, initializing it according to the configuration provider.
// This function ensures that the configuration instance is created only once (singleton pattern) using the specified provider options.
//
// Returns:
//   - An instance of the Config interface, ready to be used throughout the application.
//
// Usage:
//   - This function should be called to retrieve the configuration instance after it has been initialized with Init.
//   - It supports a thread-safe singleton pattern to ensure that only one configuration instance is active at any given time.
func Load() Config {
	cfg, err := LoadE()
	if err != nil {
		koanfFatalfHook("error loading config: %v", err)
		return nil
	}

	return cfg
}

// LoadE returns a singleton Config instance or an error. Prefer this function
// in libraries and services that should not terminate the process on config
// errors.
func LoadE() (Config, error) {
	cMutex.Lock()
	if instance != nil {
		cfg := instance
		cMutex.Unlock()
		return cfg, nil
	}
	if cOptions == nil {
		cMutex.Unlock()
		return nil, fmt.Errorf("config has not been initialized")
	}

	if cManager != nil {
		cManager.Close()
	}
	manager, err := NewManager(cProvider, cOptions)
	if err != nil {
		cMutex.Unlock()
		return nil, err
	}
	cManager = manager
	cMutex.Unlock()

	cfg, err := manager.Load()
	if err != nil {
		return nil, err
	}

	cMutex.Lock()
	instance = cfg
	cMutex.Unlock()

	return cfg, nil
}
