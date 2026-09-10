package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanPath(t *testing.T) {
	assert := assert.New(t)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Test Handler")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	req, err := http.NewRequest("GET", "/example//path//with////double///slashes/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	middleware := CleanPath(testHandler)

	middleware.ServeHTTP(recorder, req)

	expectedPath := "/example/path/with/double/slashes"
	assert.Equal(expectedPath, req.URL.Path, "URL path is not cleaned properly")

	expectedBody := "Test Handler"
	assert.Equal(expectedBody, recorder.Body.String(), "Response body is not as expected")
}
