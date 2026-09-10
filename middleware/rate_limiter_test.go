package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockLimiter struct {
	mock.Mock
}

func (m *MockLimiter) Allow(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	limiter := new(MockLimiter)
	log := logger.Init(&logger.LoggerConfig{
		Provider: logger.LoggerProviderZerolog,
		Name:     "test",
		Type:     logger.LoggerTypeStdout,
	})

	// Use a fake IP in the header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "123.123.123.123")
	w := httptest.NewRecorder()

	expectedKey := "ratelimit:http:123.123.123.123"
	limiter.On("Allow", mock.Anything, expectedKey).Return(true, nil)

	handler := RateLimitMiddleware(limiter, log, 500*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("success")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	limiter.AssertExpectations(t)
}

func TestRateLimitMiddleware_TooManyRequests(t *testing.T) {
	limiter := new(MockLimiter)
	log := logger.Init(&logger.LoggerConfig{
		Provider: logger.LoggerProviderZerolog,
		Name:     "test",
		Type:     logger.LoggerTypeStdout,
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:5678"
	w := httptest.NewRecorder()

	expectedKey := "ratelimit:http:1.2.3.4"
	limiter.On("Allow", mock.Anything, expectedKey).Return(false, nil)

	handler := RateLimitMiddleware(limiter, log, 500*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), "false"))
	limiter.AssertExpectations(t)
}

func TestRateLimitMiddleware_InternalError(t *testing.T) {
	limiter := new(MockLimiter)
	log := logger.Init(&logger.LoggerConfig{
		Provider: logger.LoggerProviderZerolog,
		Name:     "test",
		Type:     logger.LoggerTypeStdout,
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "5.6.7.8:9999"
	w := httptest.NewRecorder()

	expectedKey := "ratelimit:http:5.6.7.8"
	limiter.On("Allow", mock.Anything, expectedKey).Return(false, errors.New("redis failure"))

	handler := RateLimitMiddleware(limiter, log, 500*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.True(t, strings.Contains(w.Body.String(), "false"))
	limiter.AssertExpectations(t)
}
