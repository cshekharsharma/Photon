package discovery

import (
	"context"
	"testing"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewDiscoveryFactory(t *testing.T) {
	t.Run("ValidationFailed", func(t *testing.T) {
		discovery, err := New(context.Background(), &Options{
			Provider: "",
			Address:  "localhost:8500",
			Logger: logger.Init(&logger.LoggerConfig{
				Name:     "discovery-validation-test",
				Provider: logger.LoggerProviderZerolog,
				Type:     logger.LoggerTypeStdout}),
		})

		assert.Error(t, err)
		assert.Nil(t, discovery)
	})

	t.Run("SupportedProviders", func(t *testing.T) {
		discovery, err := New(context.Background(), &Options{
			Provider: ProviderConsul,
			Address:  "localhost:8500",
			Logger: logger.Init(&logger.LoggerConfig{
				Name:     "discovery-supported-test",
				Provider: logger.LoggerProviderZerolog,
				Type:     logger.LoggerTypeStdout}),
		})

		assert.NoError(t, err)
		assert.NotNil(t, discovery)
	})

	t.Run("UnsupportedProvider", func(t *testing.T) {
		discovery, err := New(context.Background(), &Options{
			Provider: "unknown",
			Address:  "localhost:8500",
			Logger: logger.Init(&logger.LoggerConfig{
				Name:     "discovery-unsupported-test",
				Provider: logger.LoggerProviderZerolog,
				Type:     logger.LoggerTypeStdout}),
		})

		assert.Error(t, err)
		assert.Nil(t, discovery)
		assert.Contains(t, err.Error(), "unsupported discovery provider")
	})
}
