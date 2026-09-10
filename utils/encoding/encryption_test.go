package encoding

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"testing"
)

func pad(data []byte) []byte {
	padLen := aes.BlockSize - (len(data) % aes.BlockSize)
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}

func unpad(data []byte) []byte {
	length := len(data)
	padLen := int(data[length-1])
	return data[:(length - padLen)]
}

func TestEncryptDecryptAES(t *testing.T) {
	key := make([]byte, 16) // AES-128
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	plain := pad([]byte("Hello AES CBC Encryption!"))

	ciphertext, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}
	if len(ciphertext) == 0 {
		t.Fatal("EncryptAES returned empty ciphertext")
	}

	decrypted, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAES failed: %v", err)
	}
	decrypted = unpad(decrypted)
	if !bytes.Equal(decrypted, []byte("Hello AES CBC Encryption!")) {
		t.Errorf("Decrypted data mismatch. Got %q", decrypted)
	}
}

func TestEncryptAES_InvalidBlockSize(t *testing.T) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	plain := []byte("not-multiple-of-block")
	_, err = EncryptAES(plain, key)
	if err == nil {
		t.Error("expected error for non-block-multiple plaintext, got nil")
	}
}

func TestEncryptAES_EmptyKey(t *testing.T) {
	key := []byte("") // Empty key
	plain := pad([]byte("someplaintext"))
	_, err := EncryptAES(plain, key)
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
}

func TestDecryptAES_InvalidBlockSize(t *testing.T) {
	key := make([]byte, 16)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	ciphertext := []byte("short") // Not multiple of block size
	_, err = DecryptAES(ciphertext, key)
	if err == nil {
		t.Error("expected error for non-block-multiple ciphertext, got nil")
	}
}

func TestDecryptAES_EmptyKey(t *testing.T) {
	key := []byte("") // Empty key
	ciphertext := []byte("someciphertext")
	_, err := DecryptAES(ciphertext, key)
	if err == nil {
		t.Error("expected error for empty key, got nil")
	}
}
