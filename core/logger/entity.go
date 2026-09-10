package logger

import "io"

type LoggerProvider string // LoggerProvider represents the logging framework/library to be used.
type LogLevel string       // LogLevel represents the severity level of log messages.
type LoggerType uint8      // LoggerType represents the type of logger (e.g., file, stdout).

// LoggerConfig contains configuration details used to initialize a logger.
// It specifies the logger provider, name, type, log level, base directory for log files,
// and the time format to be used in log entries.
//
// Fields:
//   - Provider: LoggerProvider indicating the logging framework to be used (e.g., Zerolog, Logrus).
//   - Name: A unique name for the logger. This name is used to retrieve the logger instance.
//   - Type: A uint8 value representing the type of logger (e.g., file, console).
//   - Level: A string that sets the minimum level of messages that should be logged (e.g., INFO, DEBUG).
//   - BaseDir: The directory where log files will be stored if the logger type supports file output.
//   - TimeFormat: The format string for timestamping log entries. This should be compatible with Go's time formatting rules.
//   - Writer: An io.Writer interface for custom output destinations. This is optional and can be nil.
type LoggerConfig struct {
	Provider   LoggerProvider
	Name       string
	Type       LoggerType
	Level      LogLevel
	BaseDir    string
	TimeFormat string
	Writer     io.Writer
}
