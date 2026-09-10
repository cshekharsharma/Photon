package logger

import (
	"context"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelLog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// InMemoryExporter stores logs in memory for inspection.
type InMemoryExporter struct {
	mu      sync.Mutex
	records []sdklog.Record
}

func (e *InMemoryExporter) Export(_ context.Context, rec []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.records = append(e.records, rec...)
	return nil
}

func (e *InMemoryExporter) ForceFlush(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.records = nil
	return nil
}

func (e *InMemoryExporter) Shutdown(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.records = nil
	return nil
}

func (e *InMemoryExporter) Records() []sdklog.Record {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]sdklog.Record(nil), e.records...)
}

func newTestLogEmitter(source string, exporter *InMemoryExporter) *LogEmitter {
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)),
	)
	return &LogEmitter{
		source: source,
		logger: provider.Logger(source),
	}
}

func TestNewLogEmitter(t *testing.T) {
	emitter := NewLogEmitter("test-source")

	if emitter == nil {
		t.Fatal("expected non-nil emitter")
	}
	if emitter.source != "test-source" {
		t.Errorf("expected source 'test-source', got '%s'", emitter.source)
	}
}

func TestEmitRaw(t *testing.T) {
	exporter := &InMemoryExporter{}
	emitter := newTestLogEmitter("raw-test", exporter)

	ctx := context.Background()
	emitter.EmitRaw(ctx, []byte("raw message"))

	records := exporter.Records()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Body().AsString() != "raw message" {
		t.Errorf("expected 'raw message', got '%s'", records[0].Body().AsString())
	}
}

func TestEmitWithContextAddsSource(t *testing.T) {
	ctx := context.Background()
	exporter := &InMemoryExporter{}
	emitter := newTestLogEmitter("emit-source-test", exporter)

	emitter.EmitWithContext(ctx, otelLog.SeverityInfo, "message", attribute.String("x", "y"))

	records := exporter.Records()
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.Severity() != otelLog.SeverityInfo {
		t.Errorf("expected severity Info, got %v", rec.Severity())
	}
	if rec.Body().AsString() != "message" {
		t.Errorf("expected body 'message', got '%s'", rec.Body().AsString())
	}

	var hasSource, hasKey bool
	rec.WalkAttributes(func(kv attribute.KeyValue) bool {
		if kv.Key == AttributeLogSource && kv.Value.AsString() == "emit-source-test" {
			hasSource = true
		}
		if kv.Key == "x" && kv.Value.AsString() == "y" {
			hasKey = true
		}
		return true // continue walking
	})

	if !hasSource {
		t.Error("log.source attribute missing")
	}
	if !hasKey {
		t.Error("attribute x=y missing")
	}
}

func TestEmitWithValidSpanContext(t *testing.T) {
	ctx := context.Background()
	exporter := &InMemoryExporter{}

	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("test")
	ctxWithSpan, span := tracer.Start(ctx, "log-span-test")
	defer span.End()

	emitter := newTestLogEmitter("span-log", exporter)
	emitter.Info(ctxWithSpan, "message with span")

	records := exporter.Records()
	if len(records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(records))
	}

	var foundTraceID, foundSpanID bool
	records[0].WalkAttributes(func(kv attribute.KeyValue) bool {
		switch kv.Key {
		case AttributeTraceId:
			foundTraceID = kv.Value.AsString() == span.SpanContext().TraceID().String()
		case AttributeSpanId:
			foundSpanID = kv.Value.AsString() == span.SpanContext().SpanID().String()
		}
		return true
	})

	if !foundTraceID {
		t.Error("trace_id attribute not found or invalid")
	}
	if !foundSpanID {
		t.Error("span_id attribute not found or invalid")
	}
}

func TestAllLogLevels(t *testing.T) {
	ctx := context.Background()
	exporter := &InMemoryExporter{}
	emitter := newTestLogEmitter("multi", exporter)

	emitter.Trace(ctx, "trace")
	emitter.Debug(ctx, "debug")
	emitter.Info(ctx, "info")
	emitter.Warn(ctx, "warn")
	emitter.Error(ctx, "error", nil)
	emitter.Fatal(ctx, "fatal")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic, but none occurred")
		}

		expected := []otelLog.Severity{
			otelLog.SeverityTrace,
			otelLog.SeverityDebug,
			otelLog.SeverityInfo,
			otelLog.SeverityWarn,
			otelLog.SeverityError,
			otelLog.SeverityFatal,
			otelLog.SeverityFatal4, // panic
		}

		records := exporter.Records()
		if len(records) != len(expected) {
			t.Fatalf("expected %d records, got %d", len(expected), len(records))
		}

		for i, want := range expected {
			if records[i].Severity() != want {
				t.Errorf("record %d: expected severity %v, got %v", i, want, records[i].Severity())
			}
		}

		last := records[len(records)-1]
		if last.Body().AsString() != "panic test" {
			t.Errorf("expected panic message 'panic test', got '%s'", last.Body().AsString())
		}
	}()

	emitter.Panic(ctx, "panic test")
}
