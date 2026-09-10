package middleware

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/utils/rest/apiresponse"
)

var errHandlerTimedOut = errors.New("http handler timed out")

// Timeout creates a middleware that sets a timeout for HTTP requests.
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if duration <= 0 {
				h.ServeHTTP(w, r)
				return
			}

			ctx, cancelFn := context.WithTimeout(r.Context(), duration)
			defer cancelFn()

			tw := newTimeoutResponseWriter()
			done := make(chan struct{}, 1)
			panicCh := make(chan any, 1)

			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicCh <- p
						return
					}
					done <- struct{}{}
				}()
				h.ServeHTTP(tw, r.WithContext(ctx))
			}()

			select {
			case p := <-panicCh:
				panic(p)
			case <-done:
				tw.copyTo(w)
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					tw.markTimedOut()
					apiresponse.New(false, apiresponse.RequestTimeout, nil, "").Send(w, http.StatusRequestTimeout)
					return
				}
			}
		})
	}
}

type timeoutResponseWriter struct {
	header      http.Header
	body        bytes.Buffer
	statusCode  int
	wroteHeader bool
	timedOut    bool
	mu          sync.Mutex
}

func newTimeoutResponseWriter() *timeoutResponseWriter {
	return &timeoutResponseWriter{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (t *timeoutResponseWriter) Header() http.Header {
	return t.header
}

func (t *timeoutResponseWriter) WriteHeader(statusCode int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timedOut || t.wroteHeader {
		return
	}

	t.statusCode = statusCode
	t.wroteHeader = true
}

func (t *timeoutResponseWriter) Write(body []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timedOut {
		return 0, errHandlerTimedOut
	}
	if !t.wroteHeader {
		t.statusCode = http.StatusOK
		t.wroteHeader = true
	}

	return t.body.Write(body)
}

func (t *timeoutResponseWriter) markTimedOut() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.timedOut = true
}

func (t *timeoutResponseWriter) copyTo(w http.ResponseWriter) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timedOut {
		return
	}

	for key, values := range t.header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(t.statusCode)
	_, _ = w.Write(t.body.Bytes())
}

// Flush is intentionally a no-op because timeout responses are buffered until
// the handler completes. Implementing it keeps ResponseController-based code
// from failing just because the timeout middleware is present.
func (t *timeoutResponseWriter) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()
}

func (t *timeoutResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("http.Hijacker is unavailable on timeout response writer")
}

func (t *timeoutResponseWriter) Push(string, *http.PushOptions) error {
	return errors.New("http.Pusher is unavailable on timeout response writer")
}

func (t *timeoutResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.timedOut {
		return 0, errHandlerTimedOut
	}
	if !t.wroteHeader {
		t.statusCode = http.StatusOK
		t.wroteHeader = true
	}

	return t.body.ReadFrom(r)
}
