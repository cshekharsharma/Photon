package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetBuildEnvironment(t *testing.T) {
	originalEnv := os.Getenv(buildEnvKey)
	defer func() {
		assert.NoError(t, os.Setenv(buildEnvKey, originalEnv))
	}()

	tests := []struct {
		name        string
		mode        BuildEnv
		expectedEnv string
	}{
		{name: "SetToProduction", mode: EnvProduction, expectedEnv: "production"},
		{name: "SetToDevelopment", mode: EnvDevelopment, expectedEnv: "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, SetBuildEnvironment(tt.mode))
			assert.Equal(t, tt.mode, buildEnv, "Expected build environment to be set correctly")
			assert.Equal(t, tt.expectedEnv, os.Getenv(buildEnvKey), "Expected environment variable to be set correctly")
		})
	}
}

func TestGetBuildEnvironment(t *testing.T) {
	originalBuildEnv := buildEnv
	defer func() { buildEnv = originalBuildEnv }()
	originalEnv := os.Getenv(buildEnvKey)
	defer func() {
		assert.NoError(t, os.Setenv(buildEnvKey, originalEnv))
	}()

	tests := []struct {
		name        string
		initialEnv  BuildEnv
		expectedEnv BuildEnv
	}{
		{name: "Get Production Environment", initialEnv: EnvProduction, expectedEnv: EnvProduction},
		{name: "Get Development Environment", initialEnv: EnvDevelopment, expectedEnv: EnvDevelopment},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildEnv = tt.initialEnv
			assert.Equal(t, tt.expectedEnv, GetBuildEnvironment(), "Expected to get correct build environment")
		})
	}
}

func TestGetBuildEnvironment_FallsBackToEnv(t *testing.T) {
	originalBuildEnv := buildEnv
	originalEnv := os.Getenv(buildEnvKey)
	defer func() {
		buildEnv = originalBuildEnv
		assert.NoError(t, os.Setenv(buildEnvKey, originalEnv))
	}()

	buildEnv = ""
	assert.NoError(t, os.Setenv(buildEnvKey, string(EnvProduction)))

	assert.Equal(t, EnvProduction, GetBuildEnvironment(), "Expected build environment to fall back to env var")
}

func TestSetEnvAndGetEnv(t *testing.T) {
	key := "TEST_ENV_KEY"
	value := "test_value"

	assert.NoError(t, os.Unsetenv(key))

	err := SetEnv(key, value)
	if err != nil {
		t.Fatalf("expected no error from SetEnv, got: %v", err)
	}

	got := GetEnv(key)
	if got != value {
		t.Errorf("expected value %q from GetEnv, got %q", value, got)
	}
}

func TestGetEnv_NotSet(t *testing.T) {
	key := "UNSET_ENV_KEY"
	assert.NoError(t, os.Unsetenv(key))

	got := GetEnv(key)
	if got != "" {
		t.Errorf("expected empty string for unset key, got: %q", got)
	}
}
