package logger

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

var (
	instances sync.Map
	mutex     sync.Mutex
)

// Get retrieves a Logger instance by name. It checks if the Logger has been initialized and stored in the instances map.
// If the Logger has not been initialized, it outputs a message indicating that the Logger must be initialized first.
//
// Parameters:
//   - name: The name of the Logger to retrieve.
//
// Returns:
//   - Logger: The Logger instance associated with the given name.
func Get(name string) Logger {
	logger, ok := instances.Load(name)
	if !ok {
		log.Printf("Logger `%s` must be initialised using Init() before use\n", name)
	}
	return castToLogger(logger)
}

// InitE initializes a Logger instance using the specified configuration.
// It stores the logger by name and returns the existing instance on duplicate init.
func InitE(config *LoggerConfig) (Logger, error) {
	if config == nil {
		return nil, errors.New("logger config is required")
	}
	if config.Name == "" {
		return nil, errors.New("logger name is required")
	}

	provider := config.Provider
	if provider == "" {
		provider = LoggerProviderZerolog
	}
	if provider != LoggerProviderZerolog {
		return nil, fmt.Errorf("unsupported logger provider: %s", provider)
	}

	instance, ok := instances.Load(config.Name)

	if !ok {
		mutex.Lock()
		defer mutex.Unlock()

		instance, ok = instances.Load(config.Name)
		if !ok {
			var err error
			instance, err = newZerolog(config)
			if err != nil {
				return nil, err
			}

			instances.Store(config.Name, instance)
		}
	}

	return castToLogger(instance), nil
}

// Init initializes a Logger instance and panics on configuration errors.
// Use InitE when the caller can return initialization errors.
func Init(config *LoggerConfig) Logger {
	instance, err := InitE(config)
	if err != nil {
		panic(fmt.Sprintf("Error while loading the logger: %v", err))
	}
	return instance
}

// castToLogger attempts to cast a generic interface value to a Logger type.
// This function is used internally to convert the values retrieved from the instances map into Logger type.
//
// Parameters:
//   - value: The interface{} value to cast to a Logger.
//
// Returns:
//   - Logger: The Logger instance cast from the interface value.
func castToLogger(value interface{}) Logger {
	logger, _ := value.(Logger)
	return logger
}
