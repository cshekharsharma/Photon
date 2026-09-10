package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var newOTLPMeterExporterHook = otlpmetricgrpc.New

// initMeterProvider initializes the MeterProvider based on the options.
func initMeterProvider(ctx context.Context, options *Options, res *resource.Resource) (metric.MeterProvider, error) {
	switch options.MeterExporter {
	case MeterExporterOTLP:

		meterOptions := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithEndpoint(options.MeterEndpoint.AsString()),
		}

		if options.MeterEndpoint.Insecure {
			meterOptions = append(meterOptions, otlpmetricgrpc.WithInsecure())
		}

		metricExporter, err := newOTLPMeterExporterHook(ctx, meterOptions...)

		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
		}

		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
			sdkmetric.WithResource(res),
		)
		return mp, nil

	default:
		return nil, fmt.Errorf("unsupported meter exporter: %v", options.MeterExporter)
	}
}
