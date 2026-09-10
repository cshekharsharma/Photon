package features

import (
	"fmt"

	"github.com/cshekharsharma/photon/coordination/network/watcher"
	"github.com/cshekharsharma/photon/utils/filesys"
	"github.com/cshekharsharma/photon/utils/types"
)

// InitOptions defines the input parameters for initializing the feature flag system.
// It specifies the source of the configuration file (file path or raw string),
// along with the content or path based on the selected source type.
type InitOptions struct {

	// SourceType determines the origin of the feature flag configuration.
	// Supported values are:
	//   - FeatureSourceFile: The configuration is read from a file on disk.
	//   - FeatureSourceRawBytes: The configuration is provided directly as a raw JSON string.
	SourceType FeatureSourceType

	// Input contains either the full JSON content (when SourceType is SourceRawBytes)
	// or the absolute/relative file path to the configuration file (when SourceType is SourceFile).
	Input string

	// EnableWatch indicates whether the watch functionality is enabled.
	// When set to true, the system will monitor and respond to changes in real-time.
	// When EnableWatch is true, the WatcherOptions field must be set to a valid watcher configuration.
	// This allows the system to automatically reload or update the feature flags
	EnableWatch bool

	// WatcherOptions represents the configuration options for the watcher component.
	// It allows customization of the behavior and settings of the watcher, which is
	// responsible for monitoring and responding to specific events or changes.
	//
	// Fields in WatcherOptions may include parameters such as polling intervals,
	// event filters, logging preferences, and other settings that influence how
	// the watcher operates.
	WatcherOptions *watcher.WatcherOptions
}

// EvaluationOptions holds the runtime context used to evaluate feature flags.
// This struct is passed to the evaluation engine when checking if a feature
// is enabled or determining which variant to return.
type EvaluationOptions struct {

	// Platform represents the platform context (e.g., "web", "mobile") from which
	// the evaluation is being made. This helps apply platform-specific rollouts.
	Platform string

	// Environment indicates the current deployment environment (e.g., "production", "staging").
	// It's used to evaluate rollout targeting based on environment.
	Environment string

	// Region indicates the geographical or logical region of the request (e.g., "us", "eu").
	// It helps evaluate region-specific rollout rules.
	Region string

	// EvaluationBucketKey is an optional identifier used to ensure deterministic bucketing during feature evaluation.
	//
	// If provided, the feature flag engine uses this ID (e.g., user ID, session ID, device ID) to calculate
	// a consistent hash-based bucket value. This ensures that the same entity (user/device/session) will
	// consistently receive the same variant across multiple evaluations, enabling reliable A/B testing,
	// canary rollouts, or gradual feature exposure.
	//
	// If EvaluationBucketKey is not provided, the engine falls back to random bucketing using a pseudo-random number generator.
	// This is useful for anonymous users or stateless scenarios but will not guarantee consistency across requests.
	EvaluationBucketKey string

	// RuleContext contains dynamic key-value pairs that represent runtime facts
	// about the user or request, such as "userTier", "userId", or "subscriptionLevel".
	// These are used to evaluate rule-based conditions within the feature definition.
	RuleContext map[string]interface{}
}

// IsValid validates the InitOptions struct to ensure that all required fields
// are properly set and contain valid values. It performs the following checks:
//
//  1. Verifies that the SourceType field is one of the allowed feature source types.
//     If the SourceType is not valid, an error is returned.
//
//  2. If the SourceType is FeatureSourceFile, it checks whether the Input field
//     points to a readable file. If the file is not readable, an error is returned.
//
//  3. If the EnableWatch field is set to true, it ensures that the WatcherOptions
//     field is not nil. If WatcherOptions is nil, an error is returned.
//
// Returns an error if any of the validation checks fail, otherwise returns nil.
func (o *InitOptions) IsValid() error {
	if o == nil {
		return fmt.Errorf("feature init options are required")
	}

	allowedFeatureSources := []FeatureSourceType{
		FeatureSourceFile,
		FeatureSourceRawBytes,
	}

	if exists, _ := types.ExistsInList(o.SourceType, allowedFeatureSources); !exists {
		return fmt.Errorf("invalid config source %v provided", o.SourceType)
	}

	if o.SourceType == FeatureSourceFile {
		if isReadable, _ := filesys.IsReadableFile(o.Input); !isReadable {
			return fmt.Errorf("filePath must be a readable file when feature source is of type file")
		}
	}

	if o.EnableWatch {
		if o.WatcherOptions == nil {
			return fmt.Errorf("watcher options must be set when EnableWatch is true")
		}
	}

	return nil
}
