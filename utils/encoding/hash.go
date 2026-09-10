package encoding

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash/adler32"
	"hash/crc32"
	"hash/fnv"
)

// HashMD5 computes the MD5 hash of the input data.
// It returns the hash as a hexadecimal string.
func HashMD5(input []byte) string {
	hash := md5.Sum(input)
	return hex.EncodeToString(hash[:])
}

// HashSHA1 computes the SHA-1 hash of the input data.
// It returns the hash as a hexadecimal string.
func HashSHA256(input []byte) string {
	hash := sha256.Sum256(input)
	return hex.EncodeToString(hash[:])
}

// HashSHA512 computes the SHA-512 hash of the input data.
// It returns the hash as a hexadecimal string.
func HashSHA512(input []byte) string {
	hash := sha512.Sum512(input)
	return hex.EncodeToString(hash[:])
}

// HashFNV32 computes the FNV-1a hash of the input data.
// It returns the hash as a hexadecimal string.
func HashFNV32(input []byte) string {
	h := fnv.New32()
	h.Write(input)
	return hex.EncodeToString(h.Sum(nil))
}

// HashFNV64 computes the FNV-1a hash of the input data.
// It returns the hash as a hexadecimal string.
func HashFNV64(input []byte) string {
	h := fnv.New64()
	h.Write(input)
	return hex.EncodeToString(h.Sum(nil))
}

// ChecksumCRC32 computes the CRC32 checksum of the input data.
// It returns the checksum as a uint32 value.
// This function is useful for data integrity checks and error detection.
func ChecksumCRC32(input []byte) uint32 {
	return crc32.ChecksumIEEE(input)
}

// ChecksumAdler32 computes the Adler-32 checksum of the input data.
// It returns the checksum as a uint32 value.
// Adler-32 is a checksum algorithm that is faster than CRC32 but less reliable.
// It is suitable for applications where speed is more important than reliability.
// This function is useful for data integrity checks and error detection.
// Note: The Adler-32 checksum is not suitable for cryptographic purposes.
func ChecksumAdler32(input []byte) uint32 {
	return adler32.Checksum(input)
}
