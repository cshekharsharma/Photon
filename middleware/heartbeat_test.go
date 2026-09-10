package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeartbeat(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestPath    string
		expectedStatus int
		expectedBody   string
		handlerStatus  int // Status the handler should write if reached
	}{
		{"Heartbeat on correct path GET", "GET", "/ping", http.StatusOK, ".", 0},
		{"Heartbeat on correct path HEAD", "HEAD", "/ping", http.StatusOK, "", 0},
		{"Heartbeat on incorrect path", "GET", "/wrong", http.StatusTeapot, "", http.StatusTeapot},
		{"Heartbeat with incorrect method", "POST", "/ping", http.StatusTeapot, "", http.StatusTeapot},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := httptest.NewRequest(tt.method, tt.requestPath, nil)
			rr := httptest.NewRecorder()
			baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// This is only reached if the middleware passes through
				w.WriteHeader(tt.handlerStatus)
			})

			// Act
			heartbeatMiddleware := Heartbeat("/ping")
			heartbeatHandler := heartbeatMiddleware(baseHandler)
			heartbeatHandler.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rr.Body.String(), "Unexpected body content")
			}
		})
	}
}

type heartbeatErrorWriter struct {
	header http.Header
}

func (w *heartbeatErrorWriter) Header() http.Header {
	return w.header
}

func (w *heartbeatErrorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *heartbeatErrorWriter) WriteHeader(int) {}

func TestHeartbeat_WriteError(t *testing.T) {
	calledNext := false
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	writer := &heartbeatErrorWriter{header: make(http.Header)}

	Heartbeat("/ping")(baseHandler).ServeHTTP(writer, req)

	assert.False(t, calledNext)
}
