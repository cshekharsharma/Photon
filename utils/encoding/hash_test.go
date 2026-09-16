package encoding

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash/adler32"
	"hash/crc32"
	"hash/fnv"
	"testing"
)

func TestHashSHA256(t *testing.T) {
	input := []byte("hello")
	expected := sha256.Sum256(input)
	if out := HashSHA256(input); out != hex.EncodeToString(expected[:]) {
		t.Errorf("expected %q, got %q", hex.EncodeToString(expected[:]), out)
	}
}

func TestHashSHA512(t *testing.T) {
	input := []byte("hello")
	expected := sha512.Sum512(input)
	if out := HashSHA512(input); out != hex.EncodeToString(expected[:]) {
		t.Errorf("expected %q, got %q", hex.EncodeToString(expected[:]), out)
	}
}

func TestHashFNV32(t *testing.T) {
	input := []byte("hello")
	expected := fnv.New32()
	expected.Write(input)
	if out := HashFNV32(input); out != hex.EncodeToString(expected.Sum(nil)) {
		t.Errorf("expected %q, got %q", hex.EncodeToString(expected.Sum(nil)), out)
	}
}

func TestHashFNV64(t *testing.T) {
	input := []byte("hello")
	expected := fnv.New64()
	expected.Write(input)
	if out := HashFNV64(input); out != hex.EncodeToString(expected.Sum(nil)) {
		t.Errorf("expected %q, got %q", hex.EncodeToString(expected.Sum(nil)), out)
	}
}

func TestChecksumCRC32(t *testing.T) {
	input := []byte("hello")
	expected := crc32.ChecksumIEEE(input)
	if out := ChecksumCRC32(input); out != expected {
		t.Errorf("expected %v, got %v", expected, out)
	}
}

func TestChecksumAdler32(t *testing.T) {
	input := []byte("hello")
	expected := adler32.Checksum(input)
	if out := ChecksumAdler32(input); out != expected {
		t.Errorf("expected %v, got %v", expected, out)
	}
}
