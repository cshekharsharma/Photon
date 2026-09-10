package tracer

import (
	"context"
	"runtime/debug"

	"github.com/cshekharsharma/photon/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AttachHttpAttributesToSpanWithError attaches HTTP attributes to the span and records an error if present.
func AddSpanAttribute(span trace.Span, key string, value interface{}) trace.Span {
	if !span.IsRecording() {
		return span
	}

	attrs := telemetry.ConvertToAttributes(map[string]interface{}{key: value})
	if len(attrs) > 0 {
		span.SetAttributes(attrs[0]) // only one attribute expected
	}

	return span
}

// AddSpanAttributeByContext retrieves the span from the context and adds an attribute to it.
// If the span is not recording, it returns nil.
// This is useful for adding attributes to spans in a middleware or handler context.
// It is important to note that this function does not create a new span.
func AddSpanAttributeByContext(ctx context.Context, key string, value any) trace.Span {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		return AddSpanAttribute(span, key, value)
	}

	return nil // span is not recording
}

// EventOptions contains options for adding events to spans.
func AddSpanEvent(span trace.Span, opts *EventOptions) trace.Span {
	if opts == nil {
		return nil
	}

	if !span.IsRecording() {
		return nil
	}

	eventName := opts.Name
	if eventName == "" {
		eventName = "event" // fallback default name
	}

	attrs := telemetry.ConvertToAttributes(opts.Attributes)

	if opts.WithStackTrace {
		stack := string(debug.Stack())
		attrs = append(attrs, attribute.String("exception.stacktrace", stack))
	}

	optsList := []trace.EventOption{
		trace.WithAttributes(attrs...),
	}

	if !opts.WithTimestamp.IsZero() {
		optsList = append(optsList, trace.WithTimestamp(opts.WithTimestamp))
	}

	span.AddEvent(eventName, optsList...)
	return span
}

// AddSpanEventByContext retrieves the span from the context and adds an event to it.
func AddSpanEventByContext(ctx context.Context, opts *EventOptions) trace.Span {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		return AddSpanEvent(span, opts)
	}

	return nil // span is not recording
}
