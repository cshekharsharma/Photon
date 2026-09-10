package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/cshekharsharma/photon/utils/rest"
)

func TestRequestLogger(t *testing.T) {
	logger := logger.Init(&logger.LoggerConfig{
		Provider: logger.LoggerProviderZerolog,
		Name:     "ABCD",
		Type:     logger.LoggerTypeStdout,
	})

	middleware := RequestLogger(logger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	req.RemoteAddr = "127.0.0.1"
	req.Header.Set(rest.HeaderUserAgent, "TestAgent")

	rec := httptest.NewRecorder()

	finalHandler := middleware(handler)
	finalHandler.ServeHTTP(rec, req)

	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
}
