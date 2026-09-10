package features

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitOptions_IsValid(t *testing.T) {
	t.Run("ValidRawBytesSource", func(t *testing.T) {
		opts := &InitOptions{
			SourceType: FeatureSourceRawBytes,
			Input:      `{"dummy": "json"}`,
		}
		err := opts.IsValid()
		assert.NoError(t, err)
	})

	t.Run("InvalidSourceType", func(t *testing.T) {
		opts := &InitOptions{
			SourceType: 99, // Invalid source type
			Input:      "",
		}
		err := opts.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid config source")
	})

	t.Run("FileSourceUnreadable", func(t *testing.T) {
		opts := &InitOptions{
			SourceType: FeatureSourceFile,
			Input:      "/path/to/nonexistent/file.json",
		}
		err := opts.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be a readable file")
	})

	t.Run("EnableWatchWithoutWatcherOptions", func(t *testing.T) {
		opts := &InitOptions{
			SourceType:     FeatureSourceRawBytes,
			Input:          `{}`,
			EnableWatch:    true,
			WatcherOptions: nil,
		}
		err := opts.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "watcher options must be set")
	})

	t.Run("ValidFileSource", func(t *testing.T) {
		// Create a temporary file
		tmpFile, err := os.CreateTemp("", "test-config-*.json")
		assert.NoError(t, err)
		defer func() {
			assert.NoError(t, os.Remove(tmpFile.Name()))
		}()

		opts := &InitOptions{
			SourceType: FeatureSourceFile,
			Input:      tmpFile.Name(),
		}
		err = opts.IsValid()
		assert.NoError(t, err)
	})
}
