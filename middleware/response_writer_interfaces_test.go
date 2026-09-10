package middleware

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
)

type optionalResponseWriter struct {
	header      http.Header
	status      int
	body        strings.Builder
	flushed     bool
	pushed      string
	readFrom    bool
	hijackConn  net.Conn
	hijackOther net.Conn
}

func newOptionalResponseWriter() *optionalResponseWriter {
	return &optionalResponseWriter{header: make(http.Header)}
}

func (w *optionalResponseWriter) Header() http.Header {
	return w.header
}

func (w *optionalResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *optionalResponseWriter) Write(body []byte) (int, error) {
	return w.body.WriteString(string(body))
}

func (w *optionalResponseWriter) Flush() {
	w.flushed = true
}

func (w *optionalResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	left, right := net.Pipe()
	w.hijackConn = left
	w.hijackOther = right
	return left, bufio.NewReadWriter(bufio.NewReader(left), bufio.NewWriter(left)), nil
}

func (w *optionalResponseWriter) Push(target string, opts *http.PushOptions) error {
	w.pushed = target
	return nil
}

func (w *optionalResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	w.readFrom = true
	return io.Copy(&w.body, r)
}

func (w *optionalResponseWriter) closeHijackConns(t *testing.T) {
	t.Helper()
	if w.hijackConn != nil {
		_ = w.hijackConn.Close()
	}
	if w.hijackOther != nil {
		_ = w.hijackOther.Close()
	}
}

func TestWrapResponseWriterOptionalInterfaces(t *testing.T) {
	base := newOptionalResponseWriter()
	ww := NewWrapResponseWriter(base)

	ww.WriteHeader(http.StatusCreated)
	ww.WriteHeader(http.StatusAccepted)
	if ww.Status() != http.StatusCreated {
		t.Fatalf("expected first status to win, got %d", ww.Status())
	}

	ww.Flush()
	if !base.flushed {
		t.Fatal("expected Flush to delegate")
	}

	conn, rw, err := ww.Hijack()
	if err != nil {
		t.Fatalf("Hijack: %v", err)
	}
	if conn == nil || rw == nil {
		t.Fatal("expected hijacked connection and read-writer")
	}
	base.closeHijackConns(t)

	if err := ww.Push("/asset.js", nil); err != nil {
		t.Fatalf("Push: %v", err)
	}
	if base.pushed != "/asset.js" {
		t.Fatalf("unexpected pushed target %q", base.pushed)
	}

	n, err := ww.ReadFrom(strings.NewReader("abc"))
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	if n != 3 || ww.BytesWritten() != 3 || !base.readFrom {
		t.Fatalf("unexpected ReadFrom state: n=%d bytes=%d readFrom=%v", n, ww.BytesWritten(), base.readFrom)
	}
}

func TestWrapResponseWriterOptionalInterfaceFallbacks(t *testing.T) {
	ww := NewWrapResponseWriter(&basicResponseWriter{header: make(http.Header)})

	ww.Flush()
	if ww.Status() != http.StatusOK {
		t.Fatalf("expected Flush to write implicit OK, got %d", ww.Status())
	}

	if _, _, err := ww.Hijack(); err == nil {
		t.Fatal("expected unavailable hijacker error")
	}
	if err := ww.Push("/asset.js", nil); err == nil {
		t.Fatal("expected unavailable pusher error")
	}

	n, err := ww.ReadFrom(strings.NewReader("fallback"))
	if err != nil {
		t.Fatalf("ReadFrom fallback: %v", err)
	}
	if n != int64(len("fallback")) || ww.BytesWritten() != len("fallback") {
		t.Fatalf("unexpected fallback ReadFrom state: n=%d bytes=%d", n, ww.BytesWritten())
	}

	fresh := NewWrapResponseWriter(&basicResponseWriter{header: make(http.Header)})
	n, err = fresh.ReadFrom(strings.NewReader("implicit"))
	if err != nil {
		t.Fatalf("ReadFrom implicit fallback: %v", err)
	}
	if n != int64(len("implicit")) || fresh.Status() != http.StatusOK || fresh.BytesWritten() != len("implicit") {
		t.Fatalf("unexpected implicit ReadFrom state: n=%d status=%d bytes=%d", n, fresh.Status(), fresh.BytesWritten())
	}
}

func TestWrapResponseWriterWriteDefaultsStatus(t *testing.T) {
	base := &basicResponseWriter{header: make(http.Header)}
	ww := NewWrapResponseWriter(base)

	n, err := ww.Write([]byte("ok"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len("ok") || ww.Status() != http.StatusOK || ww.BytesWritten() != len("ok") {
		t.Fatalf("unexpected write state: n=%d status=%d bytes=%d", n, ww.Status(), ww.BytesWritten())
	}
}

type basicResponseWriter struct {
	header http.Header
	status int
	body   strings.Builder
}

func (w *basicResponseWriter) Header() http.Header {
	return w.header
}

func (w *basicResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *basicResponseWriter) Write(body []byte) (int, error) {
	if w.header == nil {
		return 0, errors.New("missing header")
	}
	return w.body.WriteString(string(body))
}

func TestTimeoutResponseWriterOptionalInterfaces(t *testing.T) {
	tw := newTimeoutResponseWriter()

	tw.Flush()

	if _, _, err := tw.Hijack(); err == nil {
		t.Fatal("expected hijacker unavailable error")
	}
	if err := tw.Push("/asset.js", nil); err == nil {
		t.Fatal("expected pusher unavailable error")
	}

	n, err := tw.ReadFrom(strings.NewReader("streamed"))
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}
	if n != int64(len("streamed")) || tw.body.String() != "streamed" || !tw.wroteHeader {
		t.Fatalf("unexpected ReadFrom state: n=%d body=%q wroteHeader=%v", n, tw.body.String(), tw.wroteHeader)
	}

	tw.markTimedOut()
	n, err = tw.ReadFrom(strings.NewReader("late"))
	if !errors.Is(err, errHandlerTimedOut) || n != 0 {
		t.Fatalf("expected timeout ReadFrom rejection, n=%d err=%v", n, err)
	}
}
