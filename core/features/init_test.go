package features

import (
	"errors"
	"os"
	"testing"

	"github.com/cshekharsharma/photon/coordination/network/watcher"
	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validConfig = `{
  "version": "1.0",
  "attributes": {
    "canary": true,
    "platforms": ["web"],
    "environments": ["production"],
    "regions": ["us"]
  },
  "features": {
    "featureA": {
      "name": "Feature A",
      "enabled": true,
      "datatype": "STRING",
      "defaultValue": "variantA",
      "variants": [
        {"name": "variantA", "weight": 100, "label": "v1"}
      ],
      "rollouts": [
        {"platform": "web", "environment": "production", "region": "us", "percentage": 100}
      ]
    }
  }
}`

const invalidConfig = `{}` // will fail schema validation

func resetInitState() {
	mutex.Lock()
	defer mutex.Unlock()
	areFeaturesInitialized = false
	configCache = nil
}

func TestInit_WithValidRawBytes(t *testing.T) {
	resetInitState()
	err := Init(&InitOptions{
		SourceType: FeatureSourceRawBytes,
		Input:      validConfig,
	})
	assert.NoError(t, err)

	cfg, err := GetFeatureConfigStore()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Contains(t, cfg.Features, "featureA")
}

func TestInit_WithInvalidRawBytes(t *testing.T) {
	resetInitState()
	err := Init(&InitOptions{
		SourceType: FeatureSourceRawBytes,
		Input:      invalidConfig,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid feature config")
}

func TestInit_WithValidSourceFile(t *testing.T) {
	resetInitState()

	tmpFile, err := os.CreateTemp("", "feature_config_*.json")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(tmpFile.Name()))
	}()

	_, err = tmpFile.WriteString(validConfig)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	opts := &InitOptions{
		SourceType: FeatureSourceFile,
		Input:      tmpFile.Name(),
	}
	err = Init(opts)
	require.NoError(t, err)

	cfg, err := GetFeatureConfigStore()
	require.NoError(t, err)
	require.Contains(t, cfg.Features, "featureA")
	require.Equal(t, "Feature A", cfg.Features["featureA"].Name)
}

func TestInit_SourceFileReadError(t *testing.T) {
	resetInitState()

	opts := &InitOptions{
		SourceType: FeatureSourceFile,
		Input:      "/non/existent/path/config.json",
	}

	err := Init(opts)

	require.Error(t, err)
	require.Contains(t, err.Error(), "filePath must be a readable file when feature source is of type file")
}

func TestInit_SourceFileReadError_AfterValidation(t *testing.T) {
	resetInitState()
	origReadFile := readFileFn
	defer func() { readFileFn = origReadFile }()

	tmpFile, err := os.CreateTemp("", "feature_config_*.json")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(tmpFile.Name()))
	}()
	_, err = tmpFile.WriteString(validConfig)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	readFileFn = func(name string) ([]byte, error) {
		return nil, errors.New("read failed")
	}

	err = Init(&InitOptions{
		SourceType: FeatureSourceFile,
		Input:      tmpFile.Name(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read config file")
}

func TestInit_WithInvalidSourceType(t *testing.T) {
	resetInitState()
	err := Init(&InitOptions{
		SourceType: 99, // invalid type
		Input:      validConfig,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config source 99 provided")
}

func TestGetFeatureConfigStore_WithoutInit(t *testing.T) {
	resetInitState()
	_, err := GetFeatureConfigStore()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestInit_UnmarshalFailureAfterValidation(t *testing.T) {
	resetInitState()
	origValidate := validateFeatureConfig
	origUnmarshal := unmarshalFeatureConfig
	defer func() {
		validateFeatureConfig = origValidate
		unmarshalFeatureConfig = origUnmarshal
	}()

	validateFeatureConfig = func(configJSON []byte, schemaJSON []byte) error {
		return nil
	}
	unmarshalFeatureConfig = func(data []byte, v interface{}) error {
		return errors.New("unmarshal failed")
	}

	err := Init(&InitOptions{
		SourceType: FeatureSourceRawBytes,
		Input:      validConfig,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal validated feature config")
}

func TestInit_Idempotent(t *testing.T) {
	resetInitState()
	err1 := Init(&InitOptions{SourceType: FeatureSourceRawBytes, Input: validConfig})
	assert.NoError(t, err1)
	err2 := Init(&InitOptions{SourceType: FeatureSourceRawBytes, Input: invalidConfig})
	assert.NoError(t, err2) // should not re-init or fail
}

func Test_populateRequiredOptionsProperties(t *testing.T) {
	t.Run("EnableWatchTrue_LoggerNotSet_ShouldInitializeLogger", func(t *testing.T) {
		opts := &InitOptions{
			EnableWatch:    true,
			WatcherOptions: &watcher.WatcherOptions{},
		}

		updated := populateRequiredOptionsProperties(opts)

		assert.NotNil(t, updated.WatcherOptions.Logger, "Logger should be initialized when EnableWatch is true and Logger is nil")
	})

	t.Run("EnableWatchTrue_LoggerAlreadySet_ShouldNotOverwriteLogger", func(t *testing.T) {
		existingLogger := logger.Init(&logger.LoggerConfig{
			Name:     "CustomLogger",
			Level:    logger.LogLevelDebug,
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		})

		opts := &InitOptions{
			EnableWatch: true,
			WatcherOptions: &watcher.WatcherOptions{
				Logger: existingLogger,
			},
		}

		updated := populateRequiredOptionsProperties(opts)

		assert.Equal(t, existingLogger, updated.WatcherOptions.Logger, "Logger should not be overwritten if already set")
	})

	t.Run("EnableWatchFalse_ShouldNotTouchLogger", func(t *testing.T) {
		opts := &InitOptions{
			EnableWatch:    false,
			WatcherOptions: &watcher.WatcherOptions{},
		}

		updated := populateRequiredOptionsProperties(opts)

		assert.Nil(t, updated.WatcherOptions.Logger, "Logger should remain nil when EnableWatch is false")
	})
}
