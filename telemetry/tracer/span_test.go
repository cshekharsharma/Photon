package tracer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func setupTestTracer() (*tracetest.InMemoryExporter, *trace.TracerProvider, oteltrace.Tracer) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(exporter)),
	)
	tracer := tp.Tracer("test")
	return exporter, tp, tracer
}

func TestAddSpanAttribute(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(exporter)),
	)

	tracer := tp.Tracer("test-tracer")

	tests := []struct {
		name     string
		key      string
		value    interface{}
		expected attribute.KeyValue
	}{
		{"String attribute", "attr.string", "value", attribute.String("attr.string", "value")},
		{"Int attribute", "attr.int", 42, attribute.Int("attr.int", 42)},
		{"Float attribute", "attr.float", 3.14, attribute.Float64("attr.float", 3.14)},
		{"Bool attribute", "attr.bool", true, attribute.Bool("attr.bool", true)},
		{"Default fallback (slice)", "attr.slice", []string{"x", "y"}, attribute.String("attr.slice", fmt.Sprint([]string{"x", "y"}))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, span := tracer.Start(context.Background(), "test-span")
			AddSpanAttribute(span, tt.key, tt.value)
			span.End()

			spans := exporter.GetSpans()
			assert.Len(t, spans, 1)

			got := spans[0].Attributes
			attrMap := make(map[string]interface{})
			for _, kv := range got {
				attrMap[string(kv.Key)] = kv.Value.AsInterface()
			}

			assert.Equal(t, tt.expected.Value.AsInterface(), attrMap[tt.key])
			exporter.Reset()
		})
	}
}

func TestAddSpanAttribute_NotRecording(t *testing.T) {
	span := oteltrace.SpanFromContext(context.Background())
	result := AddSpanAttribute(span, "key", "value")
	assert.Equal(t, span, result)
}

func TestAddSpanAttributeByContext_WhenSpanIsRecording(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSpanProcessor(trace.NewSimpleSpanProcessor(exporter)),
	)
	otel.SetTracerProvider(tp)
	tr := otel.Tracer("test")

	ctx, span := tr.Start(context.Background(), "test-span")
	defer span.End()

	resultSpan := AddSpanAttributeByContext(ctx, "foo", "bar")

	assert.NotNil(t, resultSpan)
	span.End()

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	gotSpan := spans[0]

	found := false
	for _, attr := range gotSpan.Attributes {
		if string(attr.Key) == "foo" && attr.Value.AsString() == "bar" {
			found = true
			break
		}
	}
	assert.True(t, found, "attribute foo=bar not found in span")
}

func TestAddSpanAttributeByContext_WhenNoSpan(t *testing.T) {
	ctx := context.Background()

	span := AddSpanAttributeByContext(ctx, "key", "value")

	assert.Nil(t, span, "expected nil when no span in context")
}

func TestAddSpanAttributeByContext_WhenSpanNotRecording(t *testing.T) {
	// Create non-recording span
	ctx := context.Background()
	ctx = oteltrace.ContextWithSpan(ctx, oteltrace.SpanFromContext(context.Background()))

	span := AddSpanAttributeByContext(ctx, "key", "value")

	assert.Nil(t, span, "expected nil when span is not recording")
}

func TestAddSpanEvent_NilOptions(t *testing.T) {
	_, tp, tracer := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	result := AddSpanEvent(span, nil)
	assert.Nil(t, result)
}

func TestAddSpanEvent_NotRecording(t *testing.T) {
	span := oteltrace.SpanFromContext(context.Background()) // non-recording span
	result := AddSpanEvent(span, &EventOptions{Name: "noop"})
	assert.Nil(t, result)
}

func TestAddSpanEvent_AddsAttributesAndEvent(t *testing.T) {
	exporter, tp, tracer := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	eventTime := time.Now()
	opts := &EventOptions{
		Name: "",
		Attributes: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": true,
			"key4": []string{"x", "y"},
			"key5": 3.14,
			"key6": int64(456),
		},
		WithStackTrace: true,
		WithTimestamp:  eventTime,
	}

	returned := AddSpanEvent(span, opts)
	assert.Equal(t, span, returned)

	span.End()
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)

	events := spans[0].Events
	assert.Len(t, events, 1)
	event := events[0]

	assert.Equal(t, "event", event.Name) // default name
	assert.WithinDuration(t, eventTime, event.Time, time.Second)

	// Assert attributes exist
	attrMap := map[string]attribute.Value{}
	for _, attr := range event.Attributes {
		attrMap[string(attr.Key)] = attr.Value
	}

	assert.Equal(t, "value1", attrMap["key1"].AsString())
	assert.Equal(t, int64(123), attrMap["key2"].AsInt64())
	assert.Equal(t, true, attrMap["key3"].AsBool())
	assert.Contains(t, attrMap, "exception.stacktrace") // from WithStackTrace
}

func TestAddSpanEvent_NoStackNoTimestamp(t *testing.T) {
	exporter, tp, tracer := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	opts := &EventOptions{
		Name: "custom",
		Attributes: map[string]interface{}{
			"k": "v",
		},
		WithStackTrace: false,
	}

	AddSpanEvent(span, opts)
	span.End()

	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	events := spans[0].Events
	assert.Len(t, events, 1)
	assert.Equal(t, "custom", events[0].Name)
}

func TestAddSpanEventByContext(t *testing.T) {
	exporter, tp, tracer := setupTestTracer()
	defer func() { _ = tp.Shutdown(context.Background()) }()

	ctx, span := tracer.Start(context.Background(), "span-from-ctx")
	defer span.End()

	opts := &EventOptions{
		Name: "ctx-event",
		Attributes: map[string]interface{}{
			"ctxkey": "ctxval",
		},
	}

	ctxSpan := AddSpanEventByContext(ctx, opts)
	assert.Equal(t, span, ctxSpan)

	span.End()
	spans := exporter.GetSpans()
	assert.Len(t, spans, 1)
	assert.Len(t, spans[0].Events, 1)

	assert.Equal(t, "ctx-event", spans[0].Events[0].Name)
}

func TestAddSpanEventByContext_NoSpan(t *testing.T) {
	ctx := context.Background()
	span := AddSpanEventByContext(ctx, &EventOptions{Name: "noop"})
	assert.Nil(t, span)
}
