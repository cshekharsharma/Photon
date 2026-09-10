package config

import "os"

// BuildEnv defines a type for representing build environment modes.
type BuildEnv string

// buildEnv holds the current build environment mode.
var buildEnv BuildEnv

const (
	// buildEnvKey is the environment variable key used to store the build environment mode.
	buildEnvKey string = "BuildEnvMode"

	// EnvProduction represents the production build environment mode.
	EnvProduction BuildEnv = "production"

	// EnvDevelopment represents the development build environment mode.
	EnvDevelopment BuildEnv = "development"
)

// GetEnv retrieves the value of the environment variable identified by the given key.
// If the specified environment variable is not set, an empty string is returned.
//
// Parameters:
//   - key: A string representing the environment variable name.
//
// Returns:
//   - string: The value of the environment variable, or an empty string if not set.
func GetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv sets the value of the environment variable identified by the given key.
// If the environment variable does not exist, it is created.
//
// Parameters:
//   - key: A string representing the environment variable name.
//   - value: A string representing the value to be set.
//
// Returns:
//   - error: An error if the environment variable could not be set.
func SetEnv(key string, value string) error {
	return os.Setenv(key, value)
}

// SetBuildEnvironment sets the current build environment mode and updates the
// corresponding environment variable. It takes a BuildEnv value and sets it
// as the current build environment.
//
// Parameters:
//   - mode: The BuildEnv value representing the desired build environment mode.
//
// Usage:
//   - SetBuildEnvironment(EnvProduction) sets the environment to production mode.
//   - SetBuildEnvironment(EnvDevelopment) sets the environment to development mode.
func SetBuildEnvironment(mode BuildEnv) error {
	buildEnv = mode
	return os.Setenv(buildEnvKey, string(mode))
}

// GetBuildEnvironment returns the current build environment mode.
//
// Returns:
//   - BuildEnv: The current build environment mode.
//
// Usage:
//   - env := GetBuildEnvironment() retrieves the current build environment mode.
func GetBuildEnvironment() BuildEnv {
	if buildEnv == "" {
		return BuildEnv(os.Getenv(buildEnvKey))
	}
	return buildEnv
}
