package config

import (
	"fmt"
	"strings"

	"github.com/cshekharsharma/photon/coordination/network/watcher"
	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/utils/filesys"
	"github.com/cshekharsharma/photon/utils/types"
)

type ConfigSource uint8 // Source of the config file (file/raw bytes)
type ConfigFormat uint8 // Data format of the config (json/etc)

const (
	SourceFile     ConfigSource = 1 // Config source file
	SourceRawBytes ConfigSource = 2 // Config source raw bytes

	FormatJson ConfigFormat = 1 // Format JSON
)

// Options struct contains the necessary details to initialize and manage configurations.
// It defines parameters like source, format, file path, content, and delimiter for parsing configuration data.
//
// Fields:
//   - Source: ConfigSource enum that specifies the source of configuration data (e.g., file, raw bytes).
//   - Format: ConfigFormat enum that defines the format of the configuration data (e.g., JSON, XML).
//   - FilePath: The file path to the configuration file if the source is a file. This must be readable.
//   - Content: A byte slice that holds raw configuration data if the source is set to raw bytes.
//   - Delimiter: A string that specifies the delimiter used in the configuration data for separating keys.
//   - EnableWatch: A boolean flag indicating whether to enable file watching for changes in the configuration.
//   - WatcherOptions: A pointer to watcher.WatcherOptions that contains options for the config watcher.
type Options struct {
	Source    ConfigSource
	Format    ConfigFormat
	FilePath  string
	Content   []byte
	Delimiter string

	EnableWatch    bool
	WatcherOptions *watcher.WatcherOptions
}

// IsValid method validates the configuration options to ensure that all necessary parameters are correctly set and valid.
// It checks if the configuration source and format are supported and verifies file accessibility if the source is a file.
//
// Returns:
//   - nil if all config options are valid.
//   - An error detailing what is incorrect or missing in the options.
//
// Usage:
//   - Call IsValid before using the options to load the configuration to ensure all parameters meet the expected criteria.
func (o *Options) IsValid() error {
	return o.Validate()
}

// Validate validates the configuration options. It is equivalent to IsValid and
// exists as the clearer production-facing API for new callers.
func (o *Options) Validate() error {
	if o == nil {
		return fmt.Errorf("config options are required")
	}

	if exists, _ := types.ExistsInList(o.Source, []interface{}{SourceFile, SourceRawBytes}); !exists {
		return fmt.Errorf("invalid config source %v provided", o.Source)
	}

	if exists, _ := types.ExistsInList(o.Format, []interface{}{FormatJson}); !exists {
		return fmt.Errorf("invalid config format %v provided", o.Format)
	}

	if o.Source == SourceFile {
		if isReadable, _ := filesys.IsReadableFile(o.FilePath); !isReadable {
			return fmt.Errorf("filePath must be a readable file when Source is of type file")
		}
	}

	if o.Source == SourceRawBytes && strings.TrimSpace(string(o.Content)) == "" {
		return fmt.Errorf("content must be set when Source is of type raw bytes")
	}

	if o.EnableWatch {
		if o.WatcherOptions == nil {
			return fmt.Errorf("watcher options must be set when EnableWatch is true")
		}
	}

	return nil
}

func populateRequiredOptionsProperties(options *Options) *Options {
	if options == nil {
		return nil
	}
	if options.Delimiter == "" {
		options.Delimiter = DefaultConfigPathDelimiter
	}
	if options.EnableWatch {
		if options.WatcherOptions.Logger == nil {
			options.WatcherOptions.Logger = logger.Init(&logger.LoggerConfig{
				Name:     "ConfigWatcher",
				Level:    logger.LogLevelInfo,
				Provider: logger.LoggerProviderZerolog,
				Type:     logger.LoggerTypeStdout,
			})
		}
	}
	return options
}
