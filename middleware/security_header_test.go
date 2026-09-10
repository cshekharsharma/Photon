package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	rec := httptest.NewRecorder()

	handlerToTest := SecurityHeaders(nextHandler)

	handlerToTest.ServeHTTP(rec, req)

	tests := []struct {
		header string
		want   string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-XSS-Protection", "1; mode=block"},
		{"X-Content-Type-Options", "nosniff"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			if got := rec.Header().Get(tt.header); got != tt.want {
				t.Errorf("Header %s = %v, want %v", tt.header, got, tt.want)
			}
		})
	}

	if status := rec.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
