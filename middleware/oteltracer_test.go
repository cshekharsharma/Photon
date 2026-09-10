package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cshekharsharma/photon/middleware"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey string

const ctxKeyRequestID ctxKey = "requestID"

func TestOTelTracerMiddleware(t *testing.T) {
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider()
	tp.RegisterSpanProcessor(spanRecorder)
	tracer := tp.Tracer("test-service")

	mw := middleware.OTelTracerMiddleware("test-operation", tracer)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		if _, err := w.Write([]byte("OK")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	wrapped := mw(handler)

	req := httptest.NewRequest("GET", "http://example.com/test?query=value", nil)
	req.RemoteAddr = "127.0.0.1:2345"
	req.Header.Set("User-Agent", "UnitTest/1.0")
	ctx := context.WithValue(req.Context(), ctxKeyRequestID, "req-id-123")
	req = req.WithContext(ctx)

	resp := httptest.NewRecorder()

	wrapped.ServeHTTP(resp, req)

	spans := spanRecorder.Ended()
	assert.Len(t, spans, 1)

	span := spans[0]
	attrMap := mapAttributes(span.Attributes())

	assert.Equal(t, "test-operation", span.Name())
	assert.Equal(t, trace.SpanKindInternal, span.SpanKind())

	expected := map[attribute.Key]any{
		"http.host":         "example.com",
		"http.method":       "GET",
		"http.path":         "/test",
		"http.query":        "query=value",
		"http.userAgent":    "UnitTest/1.0",
		"http.clientIp":     "127.0.0.1:2345",
		"http.statusCode":   int64(http.StatusAccepted),
		"http.bytesWritten": int64(2),
	}

	for k, v := range expected {
		assert.Equal(t, v, attrMap[k], "attribute %s", k)
	}
}

func mapAttributes(attrs []attribute.KeyValue) map[attribute.Key]any {
	m := make(map[attribute.Key]any)
	for _, kv := range attrs {
		switch kv.Value.Type() {
		case attribute.STRING:
			m[kv.Key] = kv.Value.AsString()
		case attribute.INT64:
			m[kv.Key] = kv.Value.AsInt64()
		case attribute.FLOAT64:
			m[kv.Key] = kv.Value.AsFloat64()
		case attribute.BOOL:
			m[kv.Key] = kv.Value.AsBool()
		default:
			m[kv.Key] = kv.Value.String() // fallback string
		}
	}
	return m
}
