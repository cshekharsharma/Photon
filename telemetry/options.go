package telemetry

import "fmt"

type TracerExporterType string // e.g., "OTLP", "Jaeger", etc.
type MeterExporterType string  // e.g., "OTLP", "Prometheus", etc.
type LoggerExporterType string // e.g., "OTLP", "Prometheus", etc.
type EnvironmentType string    // e.g., "production", "staging", "development"

type Options struct {
	ServiceName string          // Name of the service
	SampleRate  float64         // Between 0.0 and 1.0 for trace sampling rate
	Environment EnvironmentType // e.g., "production", "staging", "development"

	NoOpTracer     bool               // If true, no tracing will be performed
	TracerExporter TracerExporterType // exporter type, e.g., grafana tempo
	TracerEndpoint *RemoteEndpoint    // remote endpoint for tracing

	NoOpMeter     bool              // If true, no metrics will be collected
	MeterExporter MeterExporterType // exporter type, e.g., otel-collector+prometheus
	MeterEndpoint *RemoteEndpoint   // remote endpoint for metrics collection

	NoOpLogger     bool               // If true, no logging will be performed
	LoggerExporter LoggerExporterType // exporter type, e.g., otel-collector+prometheus
	LoggerEndpoint *RemoteEndpoint    // remote endpoint for logging
}

type RemoteEndpoint struct {
	Host     string
	Port     string
	Insecure bool
}

func (oe *RemoteEndpoint) AsString() string {
	return fmt.Sprintf("%s:%s", oe.Host, oe.Port)
}

const (
	DefaultSampleRate     float64            = 0.5
	DefaultTracerExporter TracerExporterType = TracerExporterOTLP
	DefaultMeterExporter  MeterExporterType  = MeterExporterOTLP
	DefaultLoggerExporter LoggerExporterType = LoggerExporterOTLP

	// application environments
	EnvProduction  EnvironmentType = "production"
	EnvStaging     EnvironmentType = "staging"
	EnvDevelopment EnvironmentType = "development"

	// tracer exporters
	TracerExporterOTLP TracerExporterType = "OTLP"

	// meter exporters
	MeterExporterOTLP MeterExporterType = "OTLP"

	// logger exporters
	LoggerExporterOTLP LoggerExporterType = "otlp"
)
