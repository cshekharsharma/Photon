package logger

import (
	"context"
	"time"

	"github.com/cshekharsharma/photon/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

const AttributeLogSource = "log.source" // The source of the log (e.g., access, error).
const AttributeSpanId = "span.id"       // The ID of the span associated with the log.
const AttributeTraceId = "trace.id"     // 	The ID of the trace associated with the log.

// LogEmitter wraps the OTel logger with a specific log source name (e.g., accces, error).
type LogEmitter struct {
	source string
	logger log.Logger
}

// NewLogEmitter returns a logger pre-tagged with the given source.
func NewLogEmitter(source string) *LogEmitter {
	return &LogEmitter{
		source: source,
		logger: telemetry.Get().Logger,
	}
}

// EmitRaw emits a raw log message, and assumes the message is already formatted.
// It does not add any attributes or severity level to the log record.
// This is useful for logging messages that are already structured or formatted.
func (s *LogEmitter) EmitRaw(ctx context.Context, bytes []byte) {
	record := log.Record{}
	record.SetBody(attribute.StringValue(string(bytes)))
	s.logger.Emit(ctx, record)
}

// EmitWithContext emits a log message with the given severity and attributes.
// It also extracts trace/span information from the context and adds it to the log record.
func (s *LogEmitter) EmitWithContext(ctx context.Context, severity log.Severity, message string, attrs ...attribute.KeyValue) {
	record := s.newRecord(ctx, severity, message, attrs...)
	s.logger.Emit(ctx, record)
}

// Debug logs a debug message.
func (s *LogEmitter) Trace(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	s.EmitWithContext(ctx, log.SeverityTrace, msg, attrs...)
}

// Debug logs a debug message.
func (s *LogEmitter) Debug(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	s.EmitWithContext(ctx, log.SeverityDebug, msg, attrs...)
}

// Info logs an informational message.
func (s *LogEmitter) Info(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	s.EmitWithContext(ctx, log.SeverityInfo, msg, attrs...)
}

// Warn logs a warning message.
func (s *LogEmitter) Warn(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	s.EmitWithContext(ctx, log.SeverityWarn, msg, attrs...)
}

// Error logs an error message.
func (s *LogEmitter) Error(ctx context.Context, msg string, err error, attrs ...attribute.KeyValue) {
	record := s.newRecord(ctx, log.SeverityError, msg, attrs...)
	record.SetErr(err)
	s.logger.Emit(ctx, record)
}

// Fatal logs a fatal error message
func (s *LogEmitter) Fatal(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	s.EmitWithContext(ctx, log.SeverityFatal, msg, attrs...)
}

// Fatal logs a fatal error message
func (s *LogEmitter) Panic(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	// since otel doesn't have a panic level, we use fatal4 and panic
	s.EmitWithContext(ctx, log.SeverityFatal4, msg, attrs...)
	panic(msg)
}

func (s *LogEmitter) newRecord(ctx context.Context, severity log.Severity, message string, attrs ...attribute.KeyValue) log.Record {
	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetBody(attribute.StringValue(message))

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		attrs = append(attrs,
			attribute.String(AttributeTraceId, spanCtx.TraceID().String()),
			attribute.String(AttributeSpanId, spanCtx.SpanID().String()),
		)
	}

	attrs = append(attrs, attribute.String(AttributeLogSource, s.source))
	record.AddAttributes(attrs...)

	return record
}
