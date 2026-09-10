package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	})

	fastHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name        string
		handler     http.Handler
		timeout     time.Duration
		wantStatus  int
		expectDelay bool
	}{
		{"TimeoutOccurs", slowHandler, 200 * time.Millisecond, http.StatusRequestTimeout, true},
		{"NoTimeout", fastHandler, 200 * time.Millisecond, http.StatusOK, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			rec := httptest.NewRecorder()

			timeoutMiddleware := Timeout(tt.timeout)
			wrappedHandler := timeoutMiddleware(tt.handler)
			wrappedHandler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.expectDelay && rec.Code != http.StatusRequestTimeout {
				t.Errorf("expected a timeout and service unavailable status")
			}
		})
	}
}

func TestTimeoutMiddlewareDiscardLateWrites(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Late-Write", "should-not-leak")
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("late response"))
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rec := httptest.NewRecorder()

	Timeout(5*time.Millisecond)(handler).ServeHTTP(rec, req)
	time.Sleep(75 * time.Millisecond)

	if rec.Code != http.StatusRequestTimeout {
		t.Fatalf("expected status %d, got %d", http.StatusRequestTimeout, rec.Code)
	}
	if rec.Header().Get("X-Late-Write") != "" {
		t.Fatalf("late handler header leaked into timeout response")
	}
	if strings.Contains(rec.Body.String(), "late response") {
		t.Fatalf("late handler body leaked into timeout response")
	}
}

func TestTimeoutMiddlewareNonPositiveDurationRunsDirectly(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rec := httptest.NewRecorder()

	Timeout(0)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if rec.Body.String() != "accepted" {
		t.Fatalf("expected body %q, got %q", "accepted", rec.Body.String())
	}
}

func TestTimeoutMiddlewarePropagatesHandlerPanic(t *testing.T) {
	wantPanic := errors.New("handler panic")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(wantPanic)
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if recovered := recover(); recovered != wantPanic {
			t.Fatalf("expected panic %v, got %v", wantPanic, recovered)
		}
	}()

	Timeout(time.Second)(handler).ServeHTTP(rec, req)
}

func TestTimeoutResponseWriterWriteDefaultsStatusOK(t *testing.T) {
	tw := newTimeoutResponseWriter()

	n, err := tw.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != len("ok") {
		t.Fatalf("expected %d bytes written, got %d", len("ok"), n)
	}
	if !tw.wroteHeader {
		t.Fatal("expected implicit header write")
	}
	if tw.statusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, tw.statusCode)
	}
}

func TestTimeoutResponseWriterWriteAfterTimeout(t *testing.T) {
	tw := newTimeoutResponseWriter()
	tw.markTimedOut()

	n, err := tw.Write([]byte("too late"))
	if !errors.Is(err, errHandlerTimedOut) {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if n != 0 {
		t.Fatalf("expected no bytes written, got %d", n)
	}
}

func TestTimeoutResponseWriterCopyTo(t *testing.T) {
	tw := newTimeoutResponseWriter()
	tw.Header().Add("X-Test", "one")
	tw.Header().Add("X-Test", "two")
	tw.WriteHeader(http.StatusCreated)
	_, err := tw.Write([]byte("created"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	tw.WriteHeader(http.StatusAccepted)

	rec := httptest.NewRecorder()
	tw.copyTo(rec)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if rec.Body.String() != "created" {
		t.Fatalf("expected body %q, got %q", "created", rec.Body.String())
	}
	if got := rec.Header().Values("X-Test"); len(got) != 2 || got[0] != "one" || got[1] != "two" {
		t.Fatalf("unexpected copied headers: %#v", got)
	}
}

func TestTimeoutResponseWriterCopyToAfterTimeoutIsNoop(t *testing.T) {
	tw := newTimeoutResponseWriter()
	tw.Header().Set("X-Test", "should-not-copy")
	tw.WriteHeader(http.StatusCreated)
	tw.markTimedOut()

	rec := httptest.NewRecorder()
	tw.copyTo(rec)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected untouched recorder status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("X-Test") != "" {
		t.Fatalf("expected no copied headers, got %q", rec.Header().Get("X-Test"))
	}
}
