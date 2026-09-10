package encoding

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// EncryptAES encrypts the given plaintext using AES encryption with the provided key.
// The key must be either 16, 24, or 32 bytes long.
func EncryptAES(plain []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(plain)%aes.BlockSize != 0 {
		return nil, errors.New("plaintext is not a multiple of block size")
	}
	ciphertext := make([]byte, len(plain))
	mode := cipher.NewCBCEncrypter(block, key[:aes.BlockSize])
	mode.CryptBlocks(ciphertext, plain)
	return ciphertext, nil
}

// DecryptAES decrypts the given ciphertext using AES decryption with the provided key.
// The key must be either 16, 24, or 32 bytes long.
func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, key[:aes.BlockSize])
	mode.CryptBlocks(plaintext, ciphertext)
	return plaintext, nil
}
