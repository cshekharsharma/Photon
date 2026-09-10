package writers

import (
	"context"

	"github.com/cshekharsharma/photon/telemetry/logger"
)

// OtelWriter is a custom writer that emits logs to an OTLP exporter using a LogEmitter.
// It implements the io.Writer interface, allowing it to be used with any function that accepts an io.Writer.
// The OtelWriter is designed to be used with the OpenTelemetry Go SDK.
type OtelWriter struct {
	emitter *logger.LogEmitter
}

// NewOtelWriter creates an OtelWriter using an OTLP exporter and LogEmitter with log.source tag set.
func NewOtelWriter(source string) *OtelWriter {
	return &OtelWriter{
		emitter: logger.NewLogEmitter(source),
	}
}

// Write implements the io.Writer interface for OtelWriter.
// It emits the raw bytes to the OTLP exporter using the LogEmitter.
func (ow *OtelWriter) Write(p []byte) (n int, err error) {
	ow.emitter.EmitRaw(context.Background(), p)
	return len(p), nil
}
