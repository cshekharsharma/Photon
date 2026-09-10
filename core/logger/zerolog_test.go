package logger

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/core/logger/writers"
	zerologLib "github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewZerolog(t *testing.T) {
	tests := []struct {
		name          string
		config        *LoggerConfig
		configFactory func(t *testing.T) *LoggerConfig
		expectError   bool
	}{
		{
			name: "ConfigureConsoleLogger",
			config: &LoggerConfig{
				Name:       "consoleLogger",
				Provider:   LoggerProviderZerolog,
				Type:       LoggerTypeStdout,
				TimeFormat: time.RFC3339,
				Level:      LogLevelDebug,
			},
			expectError: false,
		},
		{
			name: "ConfigureOtelLogger",
			config: &LoggerConfig{
				Name:       "otelLogger",
				Provider:   LoggerProviderZerolog,
				Type:       LoggerTypeOtel,
				TimeFormat: time.RFC3339,
				Level:      LogLevelDebug,
			},
			expectError: false,
		},
		{
			name: "ConfigureFileLogger",
			config: &LoggerConfig{
				Name:       "fileLogger",
				Provider:   LoggerProviderZerolog,
				Type:       LoggerTypeFile,
				TimeFormat: time.RFC3339,
				Level:      LogLevelDebug,
				BaseDir:    t.TempDir(),
			},
			expectError: false,
		},
		{
			name: "ConfigureFileLoggerWithoutBaseDir",
			config: &LoggerConfig{
				Name:       "fileLogger",
				Provider:   LoggerProviderZerolog,
				Type:       LoggerTypeFile,
				TimeFormat: time.RFC3339,
				Level:      LogLevelDebug,
			},
			expectError: false,
		},
		{
			name: "ConfigureFileLoggerWithInvalidBaseDir",
			config: &LoggerConfig{
				Name:       "fileLogger",
				Provider:   LoggerProviderZerolog,
				Type:       LoggerTypeFile,
				TimeFormat: time.RFC3339,
				Level:      LogLevelDebug,
				BaseDir:    os.TempDir() + "/invalid_dir",
			},
			expectError: false,
		},
		{
			name: "ConfigureFileLoggerWithBaseDirFileError",
			configFactory: func(t *testing.T) *LoggerConfig {
				tmpFile, err := os.CreateTemp("", "logger_file")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				t.Cleanup(func() { _ = os.Remove(tmpFile.Name()) })
				_ = tmpFile.Close()
				return &LoggerConfig{
					Name:       "fileLogger",
					Provider:   LoggerProviderZerolog,
					Type:       LoggerTypeFile,
					TimeFormat: time.RFC3339,
					Level:      LogLevelDebug,
					BaseDir:    tmpFile.Name(),
				}
			},
			expectError: true,
		},
		{
			name: "NoLoggerConfigured",
			config: &LoggerConfig{
				Name:       "noLogger",
				Provider:   LoggerProviderZerolog,
				Type:       0,
				TimeFormat: time.RFC3339,
				Level:      LogLevelDebug,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			if tt.configFactory != nil {
				cfg = tt.configFactory(t)
			}
			logger, err := newZerolog(cfg)
			if tt.expectError {
				assert.Error(t, err)
				assert.NotNil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestConfigureConsoleLogger(t *testing.T) {
	writer := configureConsoleLogger(os.Stdout)
	assert.NotNil(t, writer)
	assert.IsType(t, zerologLib.ConsoleWriter{}, writer)
}

func TestConfigureFileLogger(t *testing.T) {
	t.Run("CreateFileWithoutBaseDir", func(t *testing.T) {
		config := &LoggerConfig{
			Name: "TempLogTest",
		}

		writer, err := configureFileLogger(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if writer == nil {
			t.Fatal("expected non-nil writer")
		}

		expectedFile := filepath.Join(config.BaseDir, "templogtest.log")
		if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
			t.Errorf("expected log file %s to be created", expectedFile)
		}

		_ = os.Remove(expectedFile)
	})

	t.Run("CreateMissingDir", func(t *testing.T) {
		tempRoot := t.TempDir()
		missingDir := filepath.Join(tempRoot, "missing-dir")

		config := &LoggerConfig{
			Name:    "MissingDirTest",
			BaseDir: missingDir,
		}

		if _, err := os.Stat(missingDir); err == nil {
			t.Fatalf("directory %s unexpectedly exists before test", missingDir)
		}

		writer, err := configureFileLogger(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if writer == nil {
			t.Fatal("expected non-nil writer")
		}

		if _, err := os.Stat(missingDir); os.IsNotExist(err) {
			t.Errorf("expected directory %s to be created", missingDir)
		}

		logFile := filepath.Join(missingDir, "missingdirtest.log")
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Errorf("expected log file %s to be created", logFile)
		}

		_ = os.Remove(logFile)
	})

	t.Run("WriteContent", func(t *testing.T) {
		dir := t.TempDir()
		config := &LoggerConfig{
			Name:    "WriteLogTest",
			BaseDir: dir,
		}

		writer, err := configureFileLogger(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		msg := "log entry test\n"
		n, err := writer.Write([]byte(msg))
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
		if n != len(msg) {
			t.Errorf("expected %d bytes written, got %d", len(msg), n)
		}

		logFile := filepath.Join(dir, "writelogtest.log")
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("failed to read written file: %v", err)
		}

		if !strings.Contains(string(content), msg) {
			t.Errorf("log file content mismatch, expected to contain %q", msg)
		}
	})

	t.Run("MkdirError", func(t *testing.T) {
		config := &LoggerConfig{
			Name:    "NoPerm",
			BaseDir: "/root/forbidden_dir",
		}

		_, err := configureFileLogger(config)
		if err == nil {
			t.Fatalf("expected error for mkdir failure")
		}
	})

	t.Run("OpenFileError", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "logger_file")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer func() {
			assert.NoError(t, os.Remove(tmpFile.Name()))
		}()

		config := &LoggerConfig{
			Name:    "BadDir",
			BaseDir: tmpFile.Name(),
		}

		_, err = configureFileLogger(config)
		if err == nil {
			t.Fatalf("expected error for open file in non-directory")
		}
	})
}

type mockWriter struct{}

func (m *mockWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestConfigureOtelLogger(t *testing.T) {
	t.Run("returns existing OtelWriter if provided", func(t *testing.T) {
		existing := writers.NewOtelWriter("existing")
		config := &LoggerConfig{
			Name:   "existing",
			Writer: existing,
		}

		got := configureOtelLogger(config)
		if got != existing {
			t.Errorf("expected to return the existing OtelWriter")
		}
	})

	t.Run("creates new OtelWriter if Writer is nil", func(t *testing.T) {
		config := &LoggerConfig{Name: "new"}
		got := configureOtelLogger(config)

		if got == nil {
			t.Fatal("expected non-nil writer")
		}

		if _, ok := got.(*writers.OtelWriter); !ok {
			t.Errorf("expected *OtelWriter, got %T", got)
		}
	})

	t.Run("creates new OtelWriter if Writer is non-Otel io.Writer", func(t *testing.T) {
		config := &LoggerConfig{
			Name:   "fallback",
			Writer: &mockWriter{},
		}
		got := configureOtelLogger(config)

		if _, ok := got.(*writers.OtelWriter); !ok {
			t.Errorf("expected *OtelWriter to be returned, got %T", got)
		}
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		level         string
		expectedLevel zerologLib.Level
	}{
		{"trace", zerologLib.TraceLevel},
		{"debug", zerologLib.DebugLevel},
		{"info", zerologLib.InfoLevel},
		{"warn", zerologLib.WarnLevel},
		{"error", zerologLib.ErrorLevel},
		{"fatal", zerologLib.FatalLevel},
		{"panic", zerologLib.PanicLevel},
		{"invalid", zerologLib.WarnLevel}, // Default log level
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			assert.Equal(t, tt.expectedLevel, parseLogLevel(tt.level))
		})
	}
}

func TestIsValidLogLevel(t *testing.T) {
	tests := []struct {
		level   string
		isValid bool
	}{
		{"trace", true},
		{"debug", true},
		{"info", true},
		{"warn", true},
		{"error", true},
		{"fatal", true},
		{"panic", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			assert.Equal(t, tt.isValid, isValidLogLevel(tt.level))
		})
	}
}

func TestZerologMethods(t *testing.T) {
	config := &LoggerConfig{
		Name:     "testLogger",
		Provider: LoggerProviderZerolog,
		Type:     LoggerTypeStdout,
		Level:    LogLevelDebug,
	}

	logger, err := newZerolog(config)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	fields := map[string]interface{}{"key": "value"}
	const message = "test message"

	t.Run("TraceWithFields", func(t *testing.T) {
		logger.TraceWithFields(fields, message)
	})

	t.Run("Trace", func(t *testing.T) {
		logger.Trace(message)
	})

	t.Run("DebugWithFields", func(t *testing.T) {
		logger.DebugWithFields(fields, message)
	})

	t.Run("Debug", func(t *testing.T) {
		logger.Debug(message)
	})

	t.Run("InfoWithFields", func(t *testing.T) {
		logger.InfoWithFields(fields, message)
	})

	t.Run("Info", func(t *testing.T) {
		logger.Info(message)
	})

	t.Run("With", func(t *testing.T) {
		logger.With(fields)
	})

	t.Run("WarnWithFields", func(t *testing.T) {
		logger.WarnWithFields(fields, message)
	})

	t.Run("Warn", func(t *testing.T) {
		logger.Warn(message)
	})

	t.Run("ErrorWithFields", func(t *testing.T) {
		logger.ErrorWithFields(fields, message)
	})

	t.Run("Error", func(t *testing.T) {
		logger.Error(message)
	})

	t.Run("PanicWithFields", func(t *testing.T) {
		assert.Panics(t, func() {
			logger.PanicWithFields(fields, message)
		})
	})

	t.Run("Panic", func(t *testing.T) {
		assert.Panics(t, func() {
			logger.Panic(message)
		})
	})

	t.Run("LogWithFields", func(t *testing.T) {
		logger.LogWithFields(LogLevelInfo, fields, message)
	})

	t.Run("Log", func(t *testing.T) {
		logger.Log(LogLevelInfo, message)
	})
}

func TestZerologFatal(t *testing.T) {
	if os.Getenv("TEST_FATAL") == "1" {
		logger, err := newZerolog(&LoggerConfig{
			Name:     "fatalLogger",
			Provider: LoggerProviderZerolog,
			Type:     LoggerTypeStdout,
		})
		if err != nil {
			os.Exit(2)
		}
		logger.Fatal("fatal message")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestZerologFatal")
	cmd.Env = append(os.Environ(), "TEST_FATAL=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit code")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatalf("expected ExitError, got %T", err)
	}
}

func TestZerologFatalWithFields(t *testing.T) {
	if os.Getenv("TEST_FATAL_FIELDS") == "1" {
		logger, err := newZerolog(&LoggerConfig{
			Name:     "fatalFieldsLogger",
			Provider: LoggerProviderZerolog,
			Type:     LoggerTypeStdout,
		})
		if err != nil {
			os.Exit(2)
		}
		logger.FatalWithFields(map[string]interface{}{"k": "v"}, "fatal with fields")
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestZerologFatalWithFields")
	cmd.Env = append(os.Environ(), "TEST_FATAL_FIELDS=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit code")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() != 1 {
			t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Fatalf("expected ExitError, got %T", err)
	}
}
