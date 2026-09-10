package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var newOTLPTraceExporterHook = otlptracegrpc.New

// initTracerProvider initializes the TracerProvider based on the options.
func initTracerProvider(ctx context.Context, options *Options, res *resource.Resource) (trace.TracerProvider, error) {
	switch options.TracerExporter {
	case TracerExporterOTLP:

		tracerOptions := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(options.TracerEndpoint.AsString()),
		}

		if options.TracerEndpoint.Insecure {
			tracerOptions = append(tracerOptions, otlptracegrpc.WithInsecure())
		}

		traceExporter, err := newOTLPTraceExporterHook(ctx, tracerOptions...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}

		// Create TracerProvider with the exporter.
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(options.SampleRate)),
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)
		return tp, nil

	default:
		return nil, fmt.Errorf("unsupported tracer exporter: %v", options.TracerExporter)
	}
}
