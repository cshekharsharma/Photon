package writers

import (
	"testing"

	"github.com/cshekharsharma/photon/telemetry/logger"
)

func TestNewOtelWriter(t *testing.T) {
	writer := NewOtelWriter("test-source")
	if writer == nil {
		t.Fatal("expected non-nil OtelWriter")
	}
	if writer.emitter == nil {
		t.Fatal("expected non-nil emitter")
	}
}

func TestOtelWriter_Write(t *testing.T) {

	ow := &OtelWriter{
		emitter: logger.NewLogEmitter("test-source"),
	}

	data := []byte("test log message")
	n, err := ow.Write(data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}
}
