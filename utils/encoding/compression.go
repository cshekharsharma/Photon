package encoding

import (
	"bytes"
	"compress/gzip"
	"io"
)

var (
	gzipWriteFn = func(w *gzip.Writer, data []byte) (int, error) {
		return w.Write(data)
	}
	gzipCloseFn       = func(w *gzip.Writer) error { return w.Close() }
	gzipReaderCloseFn = func(r *gzip.Reader) error { return r.Close() }
)

// CompressGzip compresses data using gzip compression.
// It returns the compressed data as a byte slice.
func CompressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	_, err := gzipWriteFn(writer, data)
	if err != nil {
		return nil, err
	}
	if err := gzipCloseFn(writer); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecompressGzip decompresses gzip-compressed data.
// It returns the decompressed data as a byte slice.
func DecompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	decompressed, err := io.ReadAll(reader)
	if closeErr := gzipReaderCloseFn(reader); closeErr != nil {
		return decompressed, closeErr
	}
	return decompressed, err
}
