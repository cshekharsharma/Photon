package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cshekharsharma/photon/core/logger"
)

type recovererTestLogger struct {
	lastMessage string
	lastFields  map[string]interface{}
}

func (m *recovererTestLogger) With(fields map[string]interface{}) logger.Logger { return m }
func (m *recovererTestLogger) Trace(message string, args ...interface{})        {}
func (m *recovererTestLogger) Debug(message string, args ...interface{})        {}
func (m *recovererTestLogger) Info(message string, args ...interface{})         {}
func (m *recovererTestLogger) Warn(message string, args ...interface{})         {}
func (m *recovererTestLogger) Error(message string, args ...interface{})        {}
func (m *recovererTestLogger) Fatal(message string, args ...interface{})        {}
func (m *recovererTestLogger) Panic(message string, args ...interface{})        {}
func (m *recovererTestLogger) Log(level logger.LogLevel, message string, args ...interface{}) {
}
func (m *recovererTestLogger) TraceWithFields(fields map[string]interface{}, message string, args ...interface{}) {
}
func (m *recovererTestLogger) DebugWithFields(fields map[string]interface{}, message string, args ...interface{}) {
}
func (m *recovererTestLogger) InfoWithFields(fields map[string]interface{}, message string, args ...interface{}) {
}
func (m *recovererTestLogger) WarnWithFields(fields map[string]interface{}, message string, args ...interface{}) {
}
func (m *recovererTestLogger) ErrorWithFields(fields map[string]interface{}, message string, args ...interface{}) {
	m.lastFields = fields
	m.lastMessage = message
}
func (m *recovererTestLogger) FatalWithFields(fields map[string]interface{}, message string, args ...interface{}) {
}
func (m *recovererTestLogger) PanicWithFields(fields map[string]interface{}, message string, args ...interface{}) {
}
func (m *recovererTestLogger) LogWithFields(level logger.LogLevel, fields map[string]interface{}, message string, args ...interface{}) {
}

func TestRecoverer(t *testing.T) {
	mylogger := &recovererTestLogger{}

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodPost, "http://example.com/auth/login", nil)
	rec := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), ridKey, "my-sample-uuid")
	req = req.WithContext(ctx)

	handlerToTest := Recoverer(mylogger)(panicHandler)
	handlerToTest.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	if mylogger.lastMessage != "panic recovered" {
		t.Errorf("Expected log message 'panic recovered', got %q", mylogger.lastMessage)
	}

	if mylogger.lastFields["request_id"] != "my-sample-uuid" {
		t.Errorf("Expected request_id to be 'my-sample-uuid', got %v", mylogger.lastFields["request_id"])
	}

	if mylogger.lastFields["http_method"] != http.MethodPost {
		t.Errorf("Expected http_method to be %q, got %v", http.MethodPost, mylogger.lastFields["http_method"])
	}

	if mylogger.lastFields["http_path"] != "/auth/login" {
		t.Errorf("Expected http_path to be '/auth/login', got %v", mylogger.lastFields["http_path"])
	}

	if mylogger.lastFields["http_status"] != http.StatusInternalServerError {
		t.Errorf("Expected http_status to be %d, got %v", http.StatusInternalServerError, mylogger.lastFields["http_status"])
	}

	if mylogger.lastFields["module"] != "http" {
		t.Errorf("Expected module to be 'http', got %v", mylogger.lastFields["module"])
	}

	if mylogger.lastFields["operation"] != "middleware.recoverer" {
		t.Errorf("Expected operation to be 'middleware.recoverer', got %v", mylogger.lastFields["operation"])
	}

	if mylogger.lastFields["error"] != "panic: test panic" {
		t.Errorf("Expected error to be 'panic: test panic', got %v", mylogger.lastFields["error"])
	}

	stackTrace, ok := mylogger.lastFields["stack"].([]string)
	if !ok {
		t.Errorf("Expected stack to be []string, got %T", mylogger.lastFields["stack"])
	}
	if len(stackTrace) == 0 {
		t.Errorf("Expected stack to contain at least one frame")
	}
}
