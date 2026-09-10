package middleware

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cshekharsharma/photon/utils/rest"
)

func TestNewCompressor(t *testing.T) {
	compressor := NewCompressor(5)
	if len(compressor.allowedTypes) != len(defaultCompressibleContentTypes) {
		t.Errorf("Expected %d default content types, got %d", len(defaultCompressibleContentTypes), len(compressor.allowedTypes))
	}

	customTypes := []string{rest.ContentTypeXML, rest.ContentTypeJPEG}
	compressor = NewCompressor(5, customTypes...)
	if len(compressor.allowedTypes) != len(customTypes) {
		t.Errorf("Expected %d custom content types, got %d", len(customTypes), len(compressor.allowedTypes))
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid content type pattern, did not panic")
		}
	}()
	NewCompressor(5, "application/abcd*")
}

func TestNewCompressor_Wildcard(t *testing.T) {
	compressor := NewCompressor(5, "text/*")
	if _, ok := compressor.allowedWildcards["text"]; !ok {
		t.Errorf("expected wildcard for text to be set")
	}
}

func TestSetEncoder(t *testing.T) {
	compressor := NewCompressor(5)
	compressor.SetEncoder("br", encoderGzip) // Example of adding a Brotli encoder
	if _, exists := compressor.pooledEncoders["br"]; !exists {
		t.Errorf("Expected 'br' encoder to be set")
	}
}

func TestSetEncoder_Panics(t *testing.T) {
	compressor := NewCompressor(5)
	if !assertPanics(t, func() { compressor.SetEncoder("", encoderGzip) }) {
		t.Errorf("expected panic for empty encoding")
	}
	if !assertPanics(t, func() { compressor.SetEncoder("gzip", nil) }) {
		t.Errorf("expected panic for nil encoder")
	}
}

func TestSelectEncoder_NonPooled(t *testing.T) {
	compressor := NewCompressor(5)
	custom := func(w io.Writer, level int) io.Writer {
		if w == io.Discard {
			return nil
		}
		return w
	}
	compressor.SetEncoder("custom", custom)

	header := http.Header{}
	header.Set("Accept-Encoding", "custom")
	encoder, encoding, cleanup := compressor.selectEncoder(header, bytes.NewBuffer(nil))
	defer cleanup()
	if encoder == nil || encoding != "custom" {
		t.Errorf("expected non-pooled encoder to be selected")
	}
}

func TestSelectEncoder_NoMatch(t *testing.T) {
	compressor := NewCompressor(5)
	header := http.Header{}
	header.Set("Accept-Encoding", "br")
	encoder, encoding, cleanup := compressor.selectEncoder(header, io.Discard)
	defer cleanup()
	if encoder != nil || encoding != "" {
		t.Errorf("expected no encoder to be selected")
	}
}

func TestMiddlewareHandler(t *testing.T) {
	compressor := NewCompressor(5, "text/plain")
	compressor.SetEncoder("gzip", encoderGzip)

	handler := compressor.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(rest.HeaderContentType, rest.ContentTypePlainText)
		if _, err := w.Write([]byte("Hello, world")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected 'Content-Encoding: gzip', got '%s'", w.Header().Get(rest.HeaderContentEncoding))
	}

	gz, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal("Failed to create gzip reader:", err)
	}
	defer func() {
		if err := gz.Close(); err != nil {
			t.Fatalf("failed to close gzip reader: %v", err)
		}
	}()

	decompressed, err := io.ReadAll(gz)
	if err != nil {
		t.Fatal("Failed to decompress response:", err)
	}

	if string(decompressed) != "Hello, world" {
		t.Errorf("Expected decompressed response to be 'Hello, world', got '%s'", string(decompressed))
	}
}

func TestEncoders(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	gzipWriter := encoderGzip(buffer, gzip.DefaultCompression)
	if gzipWriter == nil {
		t.Error("Failed to create gzip writer")
	}

	deflateWriter := encoderDeflate(buffer, flate.DefaultCompression)
	if deflateWriter == nil {
		t.Error("Failed to create deflate writer")
	}
}

func TestEncoders_InvalidLevel(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	if encoderGzip(buffer, 100) != nil {
		t.Errorf("expected nil gzip writer for invalid level")
	}
	if encoderDeflate(buffer, 100) != nil {
		t.Errorf("expected nil deflate writer for invalid level")
	}
}

func TestCompressResponseWriter_Push(t *testing.T) {
	t.Run("Successful Push", func(t *testing.T) {
		mockPusher := &mockPusher{}
		cw := &compressResponseWriter{ResponseWriter: mockPusher}

		err := cw.Push("/example", nil)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Push Not Supported", func(t *testing.T) {
		cw := &compressResponseWriter{ResponseWriter: httptest.NewRecorder()}
		err := cw.Push("/example", nil)
		if err == nil {
			t.Errorf("Expected error for missing pusher")
		}
	})
}

type mockPusher struct {
	httptest.ResponseRecorder
}

func (m *mockPusher) Push(target string, opts *http.PushOptions) error {
	return nil
}

type mockFlushWriter struct {
	httptest.ResponseRecorder
	flushed int
}

func (m *mockFlushWriter) Flush() {
	m.flushed++
}

func (m *mockFlushWriter) Write(p []byte) (int, error) {
	return m.ResponseRecorder.Write(p)
}

type mockCompressFlusher struct {
	httptest.ResponseRecorder
	flushed int
}

func (m *mockCompressFlusher) Flush() error {
	m.flushed++
	return nil
}

func (m *mockCompressFlusher) Write(p []byte) (int, error) {
	return m.ResponseRecorder.Write(p)
}

type errorCompressFlusher struct{}

func (e *errorCompressFlusher) Header() http.Header {
	return make(http.Header)
}

func (e *errorCompressFlusher) Write([]byte) (int, error) {
	return 0, nil
}

func (e *errorCompressFlusher) WriteHeader(int) {}

func (e *errorCompressFlusher) Flush() error {
	return io.ErrClosedPipe
}

type mockHijacker struct {
	httptest.ResponseRecorder
}

func (m *mockHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, bufio.NewReadWriter(bufio.NewReader(bytes.NewBuffer(nil)), bufio.NewWriter(bytes.NewBuffer(nil))), nil
}

type mockWriteCloser struct {
	httptest.ResponseRecorder
	closed bool
}

func (m *mockWriteCloser) Close() error {
	m.closed = true
	return nil
}

func assertPanics(t *testing.T, fn func()) bool {
	t.Helper()
	defer func() {
		_ = recover()
	}()
	defer func() {
		_ = recover()
	}()
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		fn()
	}()
	return panicked
}

func TestCompressResponseWriter_WriteHeaderAndFlags(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{
		ResponseWriter: rec,
		w:              rec,
		contentTypes:   map[string]struct{}{rest.ContentTypePlainText: {}},
		encoding:       "gzip",
	}

	cw.Header().Set(rest.HeaderContentType, rest.ContentTypePlainText+"; charset=utf-8")
	cw.WriteHeader(http.StatusOK)
	if !cw.compressable {
		t.Errorf("expected compressable true")
	}
	if cw.Header().Get(rest.HeaderContentEncoding) != "gzip" {
		t.Errorf("expected content-encoding gzip")
	}

	cw.WriteHeader(http.StatusCreated) // second call should pass through
	if rec.Code == 0 {
		t.Errorf("expected status to be set")
	}
}

func TestCompressResponseWriter_SkipCompression(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{
		ResponseWriter: rec,
		w:              rec,
		contentTypes:   map[string]struct{}{rest.ContentTypePlainText: {}},
		encoding:       "gzip",
	}

	cw.Header().Set(rest.HeaderContentType, "application/octet-stream")
	cw.WriteHeader(http.StatusOK)
	if cw.compressable {
		t.Errorf("expected compressable false")
	}
}

func TestCompressResponseWriter_IsCompressableWildcard(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{
		ResponseWriter:   rec,
		w:                rec,
		contentWildcards: map[string]struct{}{"text": {}},
	}

	cw.Header().Set(rest.HeaderContentType, "text/plain")
	if !cw.isCompressable() {
		t.Errorf("expected wildcard content type to be compressable")
	}
}

func TestCompressResponseWriter_IsCompressable_NoSlash(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{
		ResponseWriter: rec,
		w:              rec,
	}

	cw.Header().Set(rest.HeaderContentType, "plain")
	if cw.isCompressable() {
		t.Errorf("expected content type without slash to be non-compressable")
	}
}

func TestCompressResponseWriter_ContentEncodingAlreadySet(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{
		ResponseWriter: rec,
		w:              rec,
		contentTypes:   map[string]struct{}{rest.ContentTypePlainText: {}},
		encoding:       "gzip",
	}

	cw.Header().Set(rest.HeaderContentType, rest.ContentTypePlainText)
	cw.Header().Set(rest.HeaderContentEncoding, "br")
	cw.WriteHeader(http.StatusOK)
	if cw.compressable {
		t.Errorf("expected compressable false when already encoded")
	}
}

func TestCompressResponseWriter_FlushHijackCloseUnwrap(t *testing.T) {
	rec := httptest.NewRecorder()
	fw := &mockFlushWriter{}
	cw := &compressResponseWriter{
		ResponseWriter: rec,
		w:              fw,
		compressable:   true,
	}

	cw.Flush()
	if fw.flushed == 0 {
		t.Errorf("expected flusher to be called")
	}

	cw2 := &compressResponseWriter{
		ResponseWriter: &mockHijacker{},
		w:              &mockHijacker{},
		compressable:   true,
	}
	if _, _, err := cw2.Hijack(); err != nil {
		t.Errorf("expected hijack to succeed")
	}

	cw3 := &compressResponseWriter{
		ResponseWriter: &mockWriteCloser{},
		w:              &mockWriteCloser{},
		compressable:   true,
	}
	if err := cw3.Close(); err != nil {
		t.Errorf("expected close to succeed")
	}

	cw4 := &compressResponseWriter{ResponseWriter: rec}
	if cw4.Unwrap() != rec {
		t.Errorf("expected unwrap to return response writer")
	}
}

func TestCompressResponseWriter_FlushWithCompressFlusher(t *testing.T) {
	cfw := &mockCompressFlusher{}
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{
		ResponseWriter: rec,
		w:              cfw,
		compressable:   true,
	}
	cw.Flush()
	if cfw.flushed == 0 {
		t.Errorf("expected compress flusher to be called")
	}
}

func TestCompressResponseWriter_FlushWithCompressFlusherError(t *testing.T) {
	cw := &compressResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		w:              &errorCompressFlusher{},
		compressable:   true,
	}

	cw.Flush()
}

func TestCompressResponseWriter_HijackAndCloseErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	cw := &compressResponseWriter{ResponseWriter: rec}
	if _, _, err := cw.Hijack(); err == nil {
		t.Errorf("expected hijack error")
	}
	if err := cw.Close(); err == nil {
		t.Errorf("expected close error")
	}
}

func TestCompressMiddleware_Gzip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte("Hello, compressed world!")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	compressedHandler := Compress(gzip.BestSpeed)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()
	compressedHandler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("expected gzip encoding, got %s", resp.Header.Get("Content-Encoding"))
	}

	// Try decompressing to validate the content
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			t.Fatalf("failed to close gzip reader: %v", err)
		}
	}()

	body, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatalf("failed to decompress body: %v", err)
	}

	expected := "Hello, compressed world!"
	if string(body) != expected {
		t.Errorf("expected body %q, got %q", expected, string(body))
	}
}

func TestCompressMiddleware_NonCompressibleContentType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write([]byte("Binary content")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	compressedHandler := Compress(gzip.BestSpeed)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()
	compressedHandler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	if resp.Header.Get("Content-Encoding") != "" {
		t.Errorf("expected no content-encoding, got %s", resp.Header.Get("Content-Encoding"))
	}
}

func TestCompressMiddleware_NoAcceptEncoding(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte("No encoding expected")); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})

	compressedHandler := Compress(gzip.BestSpeed)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	compressedHandler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("failed to close response body: %v", err)
		}
	}()

	if resp.Header.Get("Content-Encoding") != "" {
		t.Errorf("expected no content-encoding, got %s", resp.Header.Get("Content-Encoding"))
	}
}
