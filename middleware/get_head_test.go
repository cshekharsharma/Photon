package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHead(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Test Handler")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	assert := assert.New(t)

	testCases := []struct {
		name             string
		method           string
		path             string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "GET Request",
			method:           http.MethodGet,
			path:             "/",
			expectedStatus:   http.StatusOK,
			expectedResponse: "Test Handler",
		},
		{
			name:             "HEAD Request with Allowed Path",
			method:           http.MethodHead,
			path:             "/",
			expectedStatus:   http.StatusOK,
			expectedResponse: "Test Handler",
		},
		{
			name:             "HEAD Request with Not Allowed Path",
			method:           http.MethodHead,
			path:             "/example",
			expectedStatus:   http.StatusOK,
			expectedResponse: "Test Handler",
		},
		{
			name:             "Unknown Path",
			method:           http.MethodHead,
			path:             "/unknown",
			expectedStatus:   http.StatusOK,
			expectedResponse: "Test Handler",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			assert.NoError(err)

			recorder := httptest.NewRecorder()

			middleware := GetHead(testHandler)
			middleware.ServeHTTP(recorder, req)

			assert.Equal(tc.expectedStatus, recorder.Code)
			assert.Equal(tc.expectedResponse, recorder.Body.String())
		})
	}
}
