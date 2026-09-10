package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetInstances() {
	instances.Range(func(key, _ interface{}) bool {
		instances.Delete(key)
		return true
	})
}

func TestInit(t *testing.T) {
	resetInstances()
	first := Init(&LoggerConfig{Name: "testLogger", Provider: LoggerProviderZerolog, Type: LoggerTypeStdout})
	second := Init(&LoggerConfig{Name: "testLogger", Provider: LoggerProviderZerolog, Type: LoggerTypeStdout})

	assert.NotNil(t, first)
	assert.Same(t, first, second)
}

func TestInitE(t *testing.T) {
	resetInstances()

	cfg := &LoggerConfig{Name: "init-e", Type: LoggerTypeStdout}
	instance, err := InitE(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Empty(t, cfg.Provider)

	again, err := InitE(&LoggerConfig{Name: "init-e", Type: LoggerTypeFile})
	assert.NoError(t, err)
	assert.Same(t, instance, again)
}

func TestGet(t *testing.T) {
	resetInstances()
	tests := []struct {
		name       string
		loggerName string
		config     *LoggerConfig
		wantInit   bool
	}{
		{
			name:       "Get Existing Logger",
			loggerName: "existingLogger",
			config:     &LoggerConfig{Name: "existingLogger", Provider: LoggerProviderZerolog, Type: LoggerTypeStdout},
			wantInit:   true,
		},
		{
			name:       "Get Non-Existing Logger",
			loggerName: "nonExistingLogger",
			wantInit:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantInit {
				Init(tt.config)
			}

			logger := Get(tt.loggerName)
			if tt.wantInit {
				assert.NotNil(t, logger)
			} else {
				assert.Nil(t, logger)
			}
		})
	}
}

func TestInit_PanicOnNoLoggerConfigured(t *testing.T) {
	resetInstances()
	cfg := &LoggerConfig{
		Name:     "bad",
		Provider: LoggerProviderZerolog,
		Type:     0,
	}

	assert.Panics(t, func() {
		Init(cfg)
	})
}

func TestInitE_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *LoggerConfig
	}{
		{name: "NilConfig"},
		{name: "EmptyName", config: &LoggerConfig{Provider: LoggerProviderZerolog, Type: LoggerTypeStdout}},
		{name: "UnsupportedProvider", config: &LoggerConfig{Name: "unknown", Provider: LoggerProvider("unknown"), Type: LoggerTypeStdout}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetInstances()
			instance, err := InitE(tt.config)
			assert.Error(t, err)
			assert.Nil(t, instance)
		})
	}
}

func TestInit_PanicsOnInvalidConfig(t *testing.T) {
	resetInstances()

	assert.Panics(t, func() {
		Init(&LoggerConfig{Name: "unknown", Provider: LoggerProvider("unknown"), Type: LoggerTypeStdout})
	})
}

func TestCastToLogger(t *testing.T) {
	mockLog := &zerolog{}

	tests := []struct {
		name  string
		input interface{}
		want  Logger
	}{
		{name: "Cast Valid Logger", input: mockLog, want: mockLog},
		{name: "Cast Invalid Logger", input: "notALogger", want: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := castToLogger(tt.input)
			assert.Equal(t, tt.want, logger)
		})
	}
}
