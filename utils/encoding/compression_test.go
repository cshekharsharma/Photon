package encoding

import (
	"bytes"
	"compress/gzip"
	"errors"
	"testing"
)

func TestCompressDecompressGzip(t *testing.T) {
	original := []byte("this is some test data to compress and decompress")

	compressed, err := CompressGzip(original)
	if err != nil {
		t.Fatalf("CompressGzip failed: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("CompressGzip returned empty output")
	}

	decompressed, err := DecompressGzip(compressed)
	if err != nil {
		t.Fatalf("DecompressGzip failed: %v", err)
	}
	if !bytes.Equal(original, decompressed) {
		t.Errorf("Decompressed data does not match original. Expected %q, got %q", original, decompressed)
	}
}

func TestDecompressGzip_InvalidInput(t *testing.T) {
	invalid := []byte("this is not gzipped")
	_, err := DecompressGzip(invalid)
	if err == nil {
		t.Error("Expected error for invalid gzip data, got nil")
	}
}

func TestCompressGzip_WriteError(t *testing.T) {
	orig := gzipWriteFn
	defer func() { gzipWriteFn = orig }()

	gzipWriteFn = func(_ *gzip.Writer, _ []byte) (int, error) {
		return 0, errors.New("write failed")
	}

	_, err := CompressGzip([]byte("data"))
	if err == nil {
		t.Error("Expected compression write error, got nil")
	}
}

func TestCompressGzip_CloseError(t *testing.T) {
	orig := gzipCloseFn
	defer func() { gzipCloseFn = orig }()

	gzipCloseFn = func(*gzip.Writer) error {
		return errors.New("close failed")
	}

	_, err := CompressGzip([]byte("data"))
	if err == nil {
		t.Error("Expected compression close error, got nil")
	}
}

func TestDecompressGzip_CloseError(t *testing.T) {
	orig := gzipReaderCloseFn
	defer func() { gzipReaderCloseFn = orig }()

	compressed, err := CompressGzip([]byte("data"))
	if err != nil {
		t.Fatalf("CompressGzip failed: %v", err)
	}

	gzipReaderCloseFn = func(*gzip.Reader) error {
		return errors.New("close failed")
	}

	got, err := DecompressGzip(compressed)
	if err == nil {
		t.Error("Expected decompression close error, got nil")
	}
	if !bytes.Equal(got, []byte("data")) {
		t.Errorf("Expected decompressed data before close error, got %q", got)
	}
}
