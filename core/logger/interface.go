package logger

// Logger interface defines a standard set of logging functions for structured logging across various levels (Trace, Debug, Info, Warn, Error, Fatal, Panic).
// It allows for logging with dynamic fields for enhanced log context, suitable for complex applications that require detailed and structured logging.
//
// Methods:
//
//   - With: Returns a new Logger instance that includes additional fields provided in the map. These fields are included in all subsequent log entries.
//
//   - Trace, Debug, Info, Warn, Error, Fatal, Panic: These methods log messages at their respective levels.
//     'args' allows for additional formatting or parameters similar to fmt.Printf.
//
//   - TraceWithFields, DebugWithFields, InfoWithFields, WarnWithFields, ErrorWithFields, FatalWithFields, PanicWithFields:
//     These methods function similarly to their simple counterparts but allow for additional context-specific fields
//     to be logged with each message, enhancing the granularity and usefulness of log entries.
//
// Usage:
//   - Use this interface to abstract logging mechanisms allowing for flexible logging backends and formats.
//   - Typical use involves logging messages throughout an application with optional context fields that provide more information about the log event.
//
// Example:
//
//	logger := GetLogger("application").With(map[string]interface{}{"userID": "12345"})
//	logger.Info("User logged in at time %v", time.Now())
//	logger.ErrorWithFields(map[string]interface{}{"reason": "timeout"}, "Failed to load user data")
//
// Note:
//   - The Fatal and Panic methods are expected to terminate the program or panic after logging, respectively.
//   - It's important to use the appropriate logging level to convey the correct importance and impact of the log message.
type Logger interface {
	With(fields map[string]interface{}) Logger // Enhances a logger with additional fields.

	// Base logging methods for different levels without additional fields.
	Trace(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	Fatal(message string, args ...interface{})
	Panic(message string, args ...interface{})
	Log(level LogLevel, message string, args ...interface{})

	// Logging methods for different levels with the ability to include contextual fields.
	TraceWithFields(fields map[string]interface{}, message string, args ...interface{})
	DebugWithFields(fields map[string]interface{}, message string, args ...interface{})
	InfoWithFields(fields map[string]interface{}, message string, args ...interface{})
	WarnWithFields(fields map[string]interface{}, message string, args ...interface{})
	ErrorWithFields(fields map[string]interface{}, message string, args ...interface{})
	FatalWithFields(fields map[string]interface{}, message string, args ...interface{})
	PanicWithFields(fields map[string]interface{}, message string, args ...interface{})
	LogWithFields(level LogLevel, fields map[string]interface{}, message string, args ...interface{})
}
