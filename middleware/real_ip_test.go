package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealIP(t *testing.T) {
	tests := []struct {
		name              string
		headers           map[string]string
		wantRemote        string
		initialRemoteAddr string // Add an explicit initial RemoteAddr field for clarity
	}{
		{"True-Client-IP", map[string]string{trueClientIP: "192.0.2.1"}, "192.0.2.1", ""},
		{"X-Real-IP", map[string]string{xRealIP: "198.51.100.1"}, "198.51.100.1", ""},
		{"X-Forwarded-For single", map[string]string{xForwardedFor: "203.0.113.1"}, "203.0.113.1", ""},
		{"X-Forwarded-For multiple", map[string]string{xForwardedFor: "203.0.113.2, 192.0.2.2"}, "203.0.113.2", ""},
		{"Invalid IP in True-Client-IP", map[string]string{trueClientIP: "not_an_ip"}, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.RemoteAddr = tt.initialRemoteAddr // Explicitly set RemoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.RemoteAddr != tt.wantRemote {
					t.Errorf("RemoteAddr = %v, want = %v", r.RemoteAddr, tt.wantRemote)
				}
				w.WriteHeader(http.StatusOK)
			})

			handler := RealIP(nextHandler)
			handler.ServeHTTP(rec, req)
		})
	}
}
