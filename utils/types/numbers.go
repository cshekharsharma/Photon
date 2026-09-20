package types

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// List of all valid numeric datatypes in GoLang
var NumberTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Int32,
	reflect.Int64,
	reflect.Float32,
	reflect.Float64,
}

var (
	maxIntValue = int(^uint(0) >> 1)
	minIntValue = -maxIntValue - 1
)

// Checks if the provided variable is a number.
// The number can be any of type, ie signed int, unsigned int or float.
func IsNumber(value interface{}) bool {
	datatype := reflect.TypeOf(value).Kind()

	var isNum = false
	for i := range NumberTypes {
		if NumberTypes[i] == datatype {
			isNum = true
			break
		}
	}

	return isNum
}

func ToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		// Convert int64 to int with potential overflow check
		if v > int64(maxIntValue) || v < int64(minIntValue) {
			return 0, fmt.Errorf("int64 value out of range for int type")
		}
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case float32:
		// Round float32 and convert to int
		return int(math.Round(float64(v))), nil
	case float64:
		// Round float64 and convert to int
		return int(math.Round(v)), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

// ToInt64 converts an interface{} value to int64.
func ToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		return int64(i), nil
	case float32:
		// Round float32 before converting to int64
		return int64(math.Round(float64(v))), nil
	case float64:
		// Round float64 before converting to int64
		return int64(math.Round(v)), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

// ToFloat32 converts an interface{} value to float32.
func ToFloat32(value interface{}) (float32, error) {
	switch v := value.(type) {
	case int:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	case string:
		// Try to parse the string into float32
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid string format: %v", err)
		}
		return float32(f), nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

// ToFloat64 converts an interface{} value to float64.
func ToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case string:
		// Try to parse the string into a float64
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid string format: %v", err)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

// MaxUint64 returns the maximum of two uint32 numbers.
func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// MaxUint8 returns the maximum of two uint8 numbers.
func MaxUint8(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

// MaxInt64 returns the maximum of two int64 numbers.
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func PackNumbers(num1 int32, num2 int32) int64 {
	return (int64(num1) << 32) | int64(uint32(num2)) // #nosec G115 -- intentional two's-complement bit packing.
}

// Unpack both numbers from the packed uint64
func UnpackNumbers(packed int64) (int32, int32) {
	num1 := int32(packed >> 32)        // #nosec G115 -- inverse of PackNumbers.
	num2 := int32(packed & 0xFFFFFFFF) // #nosec G115 -- inverse of PackNumbers.
	return num1, num2
}

// Get only the first number from the packed uint64
func UnpackFirstNumber(packed int64) int32 {
	return int32(packed >> 32) // #nosec G115 -- inverse of PackNumbers.
}

// Get only the second number from the packed uint64
func UnpackSecondNumber(packed int64) int32 {
	return int32(packed & 0xFFFFFFFF) // #nosec G115 -- inverse of PackNumbers.
}
