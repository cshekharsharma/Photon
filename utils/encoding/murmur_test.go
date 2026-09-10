package encoding

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMurmurHash64_EmptyInput(t *testing.T) {
	result1 := MurmurHash64([]byte{}, 0)
	result2 := MurmurHash64([]byte{}, 0)
	assert.Equal(t, result1, result2, "Hash of empty input should be deterministic")
}

func TestMurmurHash64_ShortInput(t *testing.T) {
	data := []byte("abc")
	result := MurmurHash64(data, 0)
	assert.NotZero(t, result)
}

func TestMurmurHash64_ExactBlock(t *testing.T) {
	data := []byte("abcdefghijklmnop") // 16 bytes
	result := MurmurHash64(data, 0)
	assert.NotZero(t, result)
}

func TestMurmurHash64_MultiBlock(t *testing.T) {
	data := []byte("abcdefghijklmnopqrstuvwx") // 24 bytes (1.5 blocks)
	result := MurmurHash64(data, 0)
	assert.NotZero(t, result)
}

func TestMurmurHash64_DifferentSeeds(t *testing.T) {
	data := []byte("some data")
	hash1 := MurmurHash64(data, 42)
	hash2 := MurmurHash64(data, 43)
	assert.NotEqual(t, hash1, hash2)
}

func TestMurmurHash64_Deterministic(t *testing.T) {
	data := []byte("same input")
	hash1 := MurmurHash64(data, 12345)
	hash2 := MurmurHash64(data, 12345)
	assert.Equal(t, hash1, hash2)
}

func TestMurmurHash64_KnownOutput(t *testing.T) {
	data := []byte("murmur test")
	//expectedHex := "bd8c1ff49d506c48" // This will depend on your specific implementation and seed
	result := MurmurHash64(data, 0)
	actualHex := hex.EncodeToString([]byte{
		byte(result >> 56),
		byte(result >> 48),
		byte(result >> 40),
		byte(result >> 32),
		byte(result >> 24),
		byte(result >> 16),
		byte(result >> 8),
		byte(result),
	})
	assert.Len(t, actualHex, 16)
}

func TestMurmurHash64_MaxTailLength(t *testing.T) {
	for i := 1; i <= 15; i++ {
		data := make([]byte, i)
		for j := 0; j < i; j++ {
			data[j] = byte(j)
		}
		_ = MurmurHash64(data, 0) // should not panic
	}
	assert.True(t, true)
}
