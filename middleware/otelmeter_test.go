package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/telemetry/meter"
)

func TestNormalizePath(tt *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/api/user/123", "/api/user/:id"},
		{"/api/v1/order/98765/item", "/api/v1/order/:id/item"},
		{"/", "/"},
		{"", ""},
	}

	for _, test := range tests {
		if got := normalizePath(test.input); got != test.expected {
			tt.Errorf("normalizePath(%q) = %q; want %q", test.input, got, test.expected)
		}
	}
}

func TestFormatHelper(tt *testing.T) {
	result := t("%s_test_metric", "myservice")
	expected := "myservice_test_metric"

	if result != expected {
		tt.Errorf("t() returned %q, expected %q", result, expected)
	}
}

func TestOTelMeterMiddleware_Success(t *testing.T) {
	middleware := OTelMeterMiddleware("testservice")

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/user/123", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !handlerCalled {
		t.Errorf("handler was not called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestOTelMeterMiddleware_ErrorStatus(t *testing.T) {
	middleware := OTelMeterMiddleware("testservice")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orders/5678", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestOTelMeterMiddleware_PanicRecovery(t *testing.T) {
	middleware := OTelMeterMiddleware("testservice")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic to propagate, but it did not")
		}
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("panic occurred"))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/fail/999", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)
}

func TestOTelMeterMiddleware_LatencyTracking(t *testing.T) {
	orig := meter.RecordLatency
	called := false

	meter.RecordLatencyFunc = func(name string, start time.Time, attr map[string]interface{}) {
		called = true
		orig(name, start, attr)
	}
	defer func() { meter.RecordLatencyFunc = orig }()

	middleware := OTelMeterMiddleware("testservice")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/delay/321", nil)
	rr := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rr, req)

	if !called {
		t.Errorf("expected RecordLatency to be called")
	}
}
