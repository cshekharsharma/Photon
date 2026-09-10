package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestId(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ridFromCtx := GetRequestID(r.Context())
		ridFromHeader := w.Header().Get("X-Request-ID")

		if ridFromCtx != ridFromHeader {
			t.Errorf("Request ID from context and header do not match: ctx %v, header %v", ridFromCtx, ridFromHeader)
		}
	})

	testCases := []struct {
		name          string
		provideHeader bool
		headerValue   string
		expectNewId   bool
	}{
		{"Without existing Request ID", false, "", true},
		{"With existing Request ID", true, "existing-id", false},
		{"With multiple Request IDs", true, "first-id,second-id", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.provideHeader {
				req.Header.Set("X-Request-ID", tc.headerValue)
			}

			rec := httptest.NewRecorder()

			middleware := RequestId(handler)
			middleware.ServeHTTP(rec, req)

			responseRid := rec.Header().Get("X-Request-ID")
			if tc.expectNewId && responseRid == tc.headerValue {
				t.Errorf("Expected a new request ID to be generated, but got the same: %v", responseRid)
			}
			if !tc.expectNewId && responseRid != tc.headerValue {
				t.Errorf("Expected the existing request ID to be preserved, but got a different one: %v", responseRid)
			}
		})
	}
}
