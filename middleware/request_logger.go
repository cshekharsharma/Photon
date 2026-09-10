package middleware

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
)

// RequestLogger middleware logs all the request specific parameters
// along with the profiling informatilon to the provided logger (file/stream).
func RequestLogger(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, r *http.Request) {
			ww := NewWrapResponseWriter(rw)
			start := time.Now()

			defer func() {
				fields := map[string]interface{}{
					"requestId": GetRequestID(r.Context()),
					"status":    ww.Status(),
					"bytes":     ww.BytesWritten(),
					"method":    r.Method,
					"path":      r.URL.Path,
					"ip":        r.RemoteAddr,
					"query":     r.URL.RawQuery,
					"userAgent": r.UserAgent(),
					"latency":   time.Since(start),
				}

				logger.DebugWithFields(fields, "Request Completed")
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

// WrapResponseWriter wraps http.ResponseWriter to capture status code and bytes written.
type WrapResponseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

// NewWrapResponseWriter creates a new WrapResponseWriter.
func NewWrapResponseWriter(w http.ResponseWriter) *WrapResponseWriter {
	return &WrapResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

// WriteHeader wraps the WriteHeader method to capture the status code.
func (ww *WrapResponseWriter) WriteHeader(code int) {
	if ww.wroteHeader {
		return
	}
	ww.status = code
	ww.wroteHeader = true
	ww.ResponseWriter.WriteHeader(code)
}

// Write wraps the Write method to capture the number of bytes written.
func (ww *WrapResponseWriter) Write(b []byte) (int, error) {
	if !ww.wroteHeader {
		ww.WriteHeader(http.StatusOK)
	}
	n, err := ww.ResponseWriter.Write(b)
	ww.bytes += n
	return n, err
}

// Status returns the captured status code.
func (ww *WrapResponseWriter) Status() int {
	return ww.status
}

// BytesWritten returns the number of bytes written.
func (ww *WrapResponseWriter) BytesWritten() int {
	return ww.bytes
}

// Flush preserves http.Flusher for handlers that stream responses.
func (ww *WrapResponseWriter) Flush() {
	if !ww.wroteHeader {
		ww.WriteHeader(http.StatusOK)
	}
	if flusher, ok := ww.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack preserves http.Hijacker for websocket and raw TCP upgrades.
func (ww *WrapResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := ww.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("http.Hijacker is unavailable on the writer")
	}
	return hijacker.Hijack()
}

// Push preserves http.Pusher when HTTP/2 server push is available.
func (ww *WrapResponseWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := ww.ResponseWriter.(http.Pusher)
	if !ok {
		return errors.New("http.Pusher is unavailable on the writer")
	}
	return pusher.Push(target, opts)
}

// ReadFrom preserves io.ReaderFrom fast paths while keeping byte counts correct.
func (ww *WrapResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if !ww.wroteHeader {
		ww.WriteHeader(http.StatusOK)
	}
	if readerFrom, ok := ww.ResponseWriter.(io.ReaderFrom); ok {
		n, err := readerFrom.ReadFrom(r)
		ww.bytes += int(n)
		return n, err
	}

	n, err := io.Copy(ww.ResponseWriter, r)
	ww.bytes += int(n)
	return n, err
}
