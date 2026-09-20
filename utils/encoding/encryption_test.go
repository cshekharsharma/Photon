package encoding

import (
	"bytes"
	"errors"
	"testing"
)

func TestEncryptDecryptAES_GCMRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plain := []byte("not block aligned plaintext")

	ciphertext, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}

	if !bytes.HasPrefix(ciphertext, aesGCMCiphertextPrefix) {
		t.Fatalf("ciphertext prefix = %q, want %q", ciphertext[:len(aesGCMCiphertextPrefix)], aesGCMCiphertextPrefix)
	}
	if len(ciphertext) <= len(plain) {
		t.Fatalf("ciphertext len = %d, want > plaintext len %d", len(ciphertext), len(plain))
	}
	if bytes.Contains(ciphertext, plain) {
		t.Fatalf("ciphertext contains plaintext %q", plain)
	}

	decrypted, err := DecryptAES(ciphertext, key)
	if err != nil {
		t.Fatalf("DecryptAES failed: %v", err)
	}
	if !bytes.Equal(decrypted, plain) {
		t.Fatalf("decrypted = %q, want %q", decrypted, plain)
	}
}

func TestEncryptAES_UsesFreshNonce(t *testing.T) {
	key := []byte("0123456789abcdef")
	plain := []byte("same plaintext")

	first, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("first EncryptAES failed: %v", err)
	}
	second, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("second EncryptAES failed: %v", err)
	}

	if bytes.Equal(first, second) {
		t.Fatalf("EncryptAES produced identical ciphertext for same plaintext/key")
	}
}

func TestDecryptAES_TamperingFails(t *testing.T) {
	key := []byte("0123456789abcdef01234567")
	plain := []byte("authenticated message")

	ciphertext, err := EncryptAES(plain, key)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}

	tests := map[string]func([]byte){
		"Prefix": func(data []byte) {
			data[0] = 'X'
		},
		"Nonce": func(data []byte) {
			data[len(aesGCMCiphertextPrefix)] ^= 0xff
		},
		"Ciphertext": func(data []byte) {
			data[len(data)-2] ^= 0xff
		},
		"Tag": func(data []byte) {
			data[len(data)-1] ^= 0xff
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			tampered := append([]byte(nil), ciphertext...)
			mutate(tampered)

			decrypted, err := DecryptAES(tampered, key)
			if err == nil {
				t.Fatalf("DecryptAES unexpectedly succeeded with plaintext %q", decrypted)
			}
		})
	}
}

func TestDecryptAES_WrongKeyFails(t *testing.T) {
	ciphertext, err := EncryptAES([]byte("secret"), []byte("0123456789abcdef"))
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}

	decrypted, err := DecryptAES(ciphertext, []byte("fedcba9876543210"))
	if err == nil {
		t.Fatalf("DecryptAES unexpectedly succeeded with plaintext %q", decrypted)
	}
}

func TestEncryptDecryptAES_InvalidKeySizes(t *testing.T) {
	if _, err := EncryptAES([]byte("data"), []byte("short")); err == nil {
		t.Fatal("EncryptAES accepted invalid key size")
	}
	if _, err := DecryptAES([]byte("PAG1too-short"), []byte("short")); err == nil {
		t.Fatal("DecryptAES accepted invalid key size")
	}
}

func TestEncryptAES_RandFailure(t *testing.T) {
	original := aesRandRead
	aesRandRead = func([]byte) (int, error) {
		return 0, errors.New("entropy unavailable")
	}
	t.Cleanup(func() {
		aesRandRead = original
	})

	ciphertext, err := EncryptAES([]byte("data"), []byte("0123456789abcdef"))
	if err == nil {
		t.Fatalf("EncryptAES unexpectedly succeeded with ciphertext %x", ciphertext)
	}
}

func TestDecryptAES_RejectsMalformedAndLegacyCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef")

	tests := map[string][]byte{
		"Empty":      nil,
		"Short":      []byte("PAG1short"),
		"LegacyCBC":  bytes.Repeat([]byte{0x7a}, 32),
		"BadPrefix":  append([]byte("BAD1"), bytes.Repeat([]byte{0x01}, 28)...),
		"PrefixOnly": []byte("PAG1"),
	}

	for name, ciphertext := range tests {
		t.Run(name, func(t *testing.T) {
			plain, err := DecryptAES(ciphertext, key)
			if err == nil {
				t.Fatalf("DecryptAES unexpectedly succeeded with plaintext %q", plain)
			}
		})
	}
}
