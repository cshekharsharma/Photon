package middleware

import (
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OTelTracerMiddleware is a middleware that wraps an HTTP handler with OpenTelemetry tracing.
// It creates a span for the incoming request and sets various attributes on the span.
// The span is ended after the request is processed, and the latency is recorded.
func OTelTracerMiddleware(operationName string, tracer trace.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			respWriter := NewWrapResponseWriter(w)

			ctx, span := tracer.Start(r.Context(), operationName)
			defer func() {
				latency := time.Since(start)

				span.SetAttributes(
					attribute.String(telemetry.AttributeHttpRequestID, GetRequestID(ctx)),
					attribute.String(telemetry.AttributeHttpHost, r.Host),
					attribute.String(telemetry.AttributeHttpMethod, r.Method),
					attribute.String(telemetry.AttributeHttpPath, r.URL.Path),
					attribute.String(telemetry.AttributeHttpQuery, r.URL.RawQuery),
					attribute.String(telemetry.AttributeHttpUserAgent, r.UserAgent()),
					attribute.String(telemetry.AttributeHttpClientIP, r.RemoteAddr),
					attribute.Int(telemetry.AttributeHttpStatusCode, respWriter.Status()),
					attribute.Int(telemetry.AttributeHttpBytesWritten, respWriter.BytesWritten()),
					attribute.Float64(telemetry.AttributeHttpLatencyMS, float64(latency.Milliseconds())),
				)

				span.End()
			}()

			next.ServeHTTP(respWriter, r.WithContext(ctx))
		})
	}
}
