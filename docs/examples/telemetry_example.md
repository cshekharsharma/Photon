# Telemetry

Initialize telemetry once during service startup. Photon wires OpenTelemetry tracer, meter, and logger providers from one options object.

```go
package examples

import (
	"context"

	"github.com/cshekharsharma/photon/telemetry"
)

func InitTelemetry(ctx context.Context) error {
	tel, err := telemetry.InitTelemetry(&telemetry.Options{
		ServiceName:    "checkout-api",
		Environment:    telemetry.EnvProduction,
		SampleRate:     1.0,
		TracerExporter: telemetry.TracerExporterOTLP,
		TracerEndpoint: &telemetry.RemoteEndpoint{
			Host:     "otel-collector",
			Port:     "4317",
			Insecure: true,
		},
		MeterExporter: telemetry.MeterExporterOTLP,
		MeterEndpoint: &telemetry.RemoteEndpoint{
			Host:     "otel-collector",
			Port:     "4317",
			Insecure: true,
		},
		LoggerExporter: telemetry.LoggerExporterOTLP,
		LoggerEndpoint: &telemetry.RemoteEndpoint{
			Host:     "otel-collector",
			Port:     "4317",
			Insecure: true,
		},
	})
	if err != nil {
		return err
	}

	ctx, span := tel.Tracer.Start(ctx, "checkout.startup")
	defer span.End()

	counter, err := tel.Meter.Int64Counter("checkout_startups_total")
	if err != nil {
		return err
	}
	counter.Add(ctx, 1)

	return nil
}
```

Use `NoOpTracer`, `NoOpMeter`, and `NoOpLogger` for local tests that should not emit telemetry.

