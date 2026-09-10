package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToMiddlewareFn(t *testing.T) {
	mockMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Mock-Middleware", "active")
			next.ServeHTTP(w, r)
		})
	}

	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Mock-Handler", "reached")
		w.WriteHeader(http.StatusOK)
	}

	adaptedMiddleware := ToMiddlewareFn(mockMiddleware)
	finalHandler := adaptedMiddleware(mockHandler)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("GET", "/", nil)

	finalHandler.ServeHTTP(recorder, request)

	if recorder.Header().Get("X-Mock-Middleware") != "active" {
		t.Errorf("Expected middleware header 'X-Mock-Middleware' to be 'active'")
	}

	if recorder.Header().Get("X-Mock-Handler") != "reached" {
		t.Errorf("Expected handler header 'X-Mock-Handler' to be 'reached'")
	}

	if status := recorder.Result().StatusCode; status != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", status)
	}
}
