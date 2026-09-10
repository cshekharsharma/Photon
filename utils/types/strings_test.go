package types

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	now := time.Now()
	ptrData := "somedata"
	type customStringType string
	customString := customStringType("custom string")
	invalidRV := reflect.Value{}
	validRV := reflect.ValueOf("ref")

	cases := []struct {
		name       string
		input      interface{}
		timeFormat []string
		expected   string
	}{
		{"int to string", 123, nil, "123"},
		{"int8 to string", int8(127), nil, "127"},
		{"int16 to string", int16(32767), nil, "32767"},
		{"int32 to string", int32(2147483647), nil, "2147483647"},
		{"int64 to string", int64(9223372036854775807), nil, "9223372036854775807"},
		{"uint to string", uint(123), nil, "123"},
		{"uint8 to string", uint8(255), nil, "255"},
		{"uint16 to string", uint16(65535), nil, "65535"},
		{"uint32 to string", uint32(4294967295), nil, "4294967295"},
		{"uint64 to string", uint64(8446744073709551615), nil, "8446744073709551615"},
		{"string to string", "hello", nil, "hello"},
		{"byte slice to string", []byte("byte slice"), nil, "byte slice"},
		{"bool to string", true, nil, "true"},
		{"float32 to string", float32(3.14159), nil, "3.14159"},
		{"float64 to string", float64(3.141592653589793), nil, "3.141592653589793"},
		{"time to default string", now, nil, now.Format("2006-01-02 15:04:05")},
		{"time with format string", now, []string{"Jan 2, 2006"}, now.Format("Jan 2, 2006")},
		{"custom Stringer", customStringer{}, nil, "I am stringer"},
		{"non handled type", struct{}{}, nil, "{}"},
		{"nil pointer", (*time.Time)(nil), nil, ""},
		{"nil interface", (interface{})(nil), nil, ""},
		{"valid interface", (interface{})("dummy"), nil, "dummy"},
		{"pointer to string", &ptrData, nil, "somedata"},
		{"custom string type", customString, nil, "custom string"},
		{"invalid reflect value pointer", &invalidRV, nil, ""},
		{"valid reflect value pointer", &validRV, nil, "ref"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToString(tc.input, tc.timeFormat...)
			assert.Equal(t, tc.expected, result, fmt.Sprintf("Failed on %s", tc.name))
		})
	}

	// Test with nil input
	t.Run("nil input", func(t *testing.T) {
		assert.Equal(t, "", ToString(nil), "nil should return empty string")
	})
}

type customStringer struct{}

func (cs customStringer) String() string {
	return "I am stringer"
}

func TestBase64Encode(t *testing.T) {
	decoded, err := Base64Decode("YWJjZGVm")

	assert.Equal(t, "abcdef", decoded)
	assert.Nil(t, err)
}

func TestBase64Decode(t *testing.T) {
	assert.Equal(t, "YWJjZGVm", Base64Encode("abcdef"))
}

func TestBase64Decode_Error(t *testing.T) {
	_, err := Base64Decode("%%%")
	assert.Error(t, err)
}

func TestUCFirst(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"EmptyString", "", ""},
		{"FirstLetterLowercase", "example", "Example"},
		{"FirstLetterUppercase", "Example", "Example"},
		{"NonLetterChar", "1234", "1234"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := UCFirst(tc.input)
			if result != tc.expected {
				t.Errorf("UCFirst(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestNullableString(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil value", nil, ""},
		{"non-string type", 12345, ""},
		{"empty string", "", ""},
		{"normal string", "hello", "hello"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := NullableString(tc.input)
			if result != tc.expected {
				t.Errorf("NullableString(%v) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestToInterfaceSlice(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []interface{}
	}{
		{name: "Empty slice", input: []string{}, expected: []interface{}{}},
		{name: "Single element", input: []string{"hello"}, expected: []interface{}{"hello"}},
		{name: "Multiple elements", input: []string{"hello", "world"}, expected: []interface{}{"hello", "world"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToInterfaceSlice(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Test %s failed: expected %v, got %v", tc.name, tc.expected, result)
			}
		})
	}
}

func TestGetRandomString(t *testing.T) {
	lengths := []int64{0, 1, 10, 100}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("Length_%d", length), func(t *testing.T) {
			result := GetRandomString(length)

			if int64(len(result)) != length {
				t.Errorf("Expected length %d, got %d", length, len(result))
			}

			const validChars = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			for _, char := range result {
				if !strings.ContainsRune(validChars, char) {
					t.Errorf("Unexpected character '%c' in result", char)
				}
			}
		})
	}
}

func TestGetCryptoSafeRandomString(t *testing.T) {
	lengths := []uint8{0, 1, 10, 32}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("Length_%d", length), func(t *testing.T) {
			result, err := GetCryptoSafeRandomString(length)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if uint8(len(result)) != length {
				t.Errorf("Expected length %d, got %d", length, len(result))
			}

			const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
			for _, char := range result {
				if !strings.ContainsRune(validChars, char) {
					t.Errorf("Unexpected character '%c' in result", char)
				}
			}
		})
	}
}

func TestGetCryptoSafeRandomString_Error(t *testing.T) {
	orig := cryptoRandRead
	defer func() { cryptoRandRead = orig }()

	cryptoRandRead = func([]byte) (int, error) {
		return 0, fmt.Errorf("boom")
	}

	_, err := GetCryptoSafeRandomString(8)
	assert.Error(t, err)
}

func BenchmarkGetRandomString(b *testing.B) {
	length := int64(32)
	for i := 0; i < b.N; i++ {
		_ = GetRandomString(length)
	}
}

func BenchmarkGetCryptoSafeRandomString(b *testing.B) {
	length := uint8(32)
	for i := 0; i < b.N; i++ {
		_, _ = GetCryptoSafeRandomString(length)
	}
}
