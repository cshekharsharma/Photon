package watcher

import (
	"context"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
)

// WatcherType represents the supported external systems that can be watched
// for configuration or content changes.
type WatcherType string

const (
	// WatcherTypeAwsAppConfig represents AWS AppConfig as the source of configuration.
	WatcherTypeAwsAppConfig WatcherType = "AwsAppConfig"
)

// WatcherOptions defines the behavior for a config/content watcher.
// It supports contextual polling, automatic update callbacks, and source metadata.
type WatcherOptions struct {
	// WatcherType specifies which backend or system to watch.
	// Example: AwsAppConfig.
	WatcherType WatcherType

	// OnUpdateCallback is invoked when a change is detected in the content source.
	// This is a user-defined function typically used to reload content on disk/memory
	OnUpdateCallback func()

	// WatchContext controls the lifecycle of the watcher. When the context is cancelled,
	// polling is stopped gracefully.
	WatchContext context.Context

	// Application represents the name of the application whose content is being watched.
	// Applicable to systems like AWS AppConfig where configs are tied to applications.
	Application string

	// ClientID is a unique identifier for the consuming service instance.
	// Required by systems like AWS AppConfig to track and differentiate clients.
	ClientID string

	// ContentScope is a flexible field used to specify profile, namespace, or config path
	// depending on the backend system (e.g., config profile in AWS AppConfig,
	// KV path or prefix in Consul, etc.).
	ContentScope string

	// Environment defines the environment for which config is being fetched,
	// such as "dev", "staging", or "prod".
	Environment string

	// PollInterval specifies how often the content source should be checked
	// for changes. A shorter duration provides faster reactivity but consumes
	// more resources.
	PollInterval time.Duration

	// Logger is the logger instance used to log watcher activity, errors,
	// and updates.
	Logger logger.Logger

	// ContentSource is an optional backend-specific identifier for the content source.
	// It can be used to differentiate between content repositories or buckets if needed.
	ContentSource uint8

	// ContentFormat is used to describe the format of the fetched content
	// (e.g., JSON, YAML, etc.), allowing appropriate parsing logic to be applied.
	ContentFormat uint8

	// UpdateChannel receives content updates for this watcher. If nil, updates
	// are sent to the package-level ContentUpdateChannel for legacy callers.
	UpdateChannel chan *UpdaterSchema
}
