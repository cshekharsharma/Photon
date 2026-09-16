package types

import (
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var cryptoRandRead = cryptorand.Read

// ToString changes any type of input argument to a string
func ToString(arg interface{}, timeFormat ...string) string {
	if arg == nil {
		return ""
	}

	// Handle custom types and nil pointers/interfaces
	val := reflect.ValueOf(arg)
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}
	if val.Type() == reflect.TypeOf(reflect.Value{}) {
		rv := val.Interface().(reflect.Value)
		if !rv.IsValid() {
			return ""
		}
		arg = rv.Interface()
		val = reflect.ValueOf(arg)
	}
	arg = val.Interface()

	switch v := arg.(type) {
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	case []byte:
		return string(v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		if len(timeFormat) > 0 {
			return v.Format(timeFormat[0])
		}
		return v.Format("2006-01-02 15:04:05")
	case fmt.Stringer:
		return v.String()
	default:
		// Handle unsupported or custom types gracefully
		return fmt.Sprintf("%v", arg)
	}
}

// takes string slice as an input and convert that into more
// generic slice of interface{}
func ToInterfaceSlice(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, v := range strings {
		result[i] = v
	}
	return result
}

// Encodes input value into base64 encoded string
func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// Decodes base64 string and return the decoded value
func Base64Decode(str string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(str)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Capitalize the first character of provided string
func UCFirst(value string) string {
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

// NullableString converts an interface{} to a string.
// returning an empty string if the value is nil or not a string.
func NullableString(value interface{}) string {
	if value == nil {
		return ""
	}

	if vstr, ok := value.(string); ok {
		return vstr
	}

	return ""
}

// IsASCII checks if a string contains only ASCII characters.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// GetRandomString generates a random string of the specified length.
func GetRandomString(length int64) string {
	if length <= 0 {
		return ""
	}

	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	randomBytes := make([]byte, length)
	if _, err := cryptoRandRead(randomBytes); err != nil {
		return ""
	}

	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[int(randomBytes[i])%len(letterBytes)]
	}
	return string(b)
}

// GetCryptoSafeRandomString generates a random string of the specified length using the crypto/rand package.
// The string is URL-safe and does not contain any padding characters.
func GetCryptoSafeRandomString(length uint8) (string, error) {
	bytes := make([]byte, length)
	_, err := cryptoRandRead(bytes)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
