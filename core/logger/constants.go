package logger

const (
	DefaultLogLevel string = "warn"

	LoggerTypeFile   LoggerType = 1
	LoggerTypeStdout LoggerType = 2
	LoggerTypeOtel   LoggerType = 4

	LoggerProviderZerolog LoggerProvider = "zerolog"

	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)
