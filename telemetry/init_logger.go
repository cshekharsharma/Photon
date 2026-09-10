package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

var newOTLPLogExporterHook = otlploggrpc.New

// initLoggerProvider initializes the LoggerProvider using OTLP exporter and resource config.
func initLoggerProvider(ctx context.Context, options *Options, res *resource.Resource) (log.LoggerProvider, error) {
	switch options.LoggerExporter {
	case LoggerExporterOTLP:
		exporterOptions := []otlploggrpc.Option{
			otlploggrpc.WithEndpoint(options.LoggerEndpoint.AsString()),
		}

		if options.LoggerEndpoint.Insecure {
			exporterOptions = append(exporterOptions, otlploggrpc.WithInsecure())
		}

		exporter, err := newOTLPLogExporterHook(ctx, exporterOptions...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
		}

		processor := sdklog.NewBatchProcessor(exporter)

		lp := sdklog.NewLoggerProvider(
			sdklog.WithProcessor(processor),
			sdklog.WithResource(res),
		)
		return lp, nil

	default:
		return nil, fmt.Errorf("unsupported logger exporter: %v", options.LoggerExporter)
	}
}
