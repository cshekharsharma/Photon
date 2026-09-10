package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteger(t *testing.T) {

	assert.True(t, IsNumber(1))
	assert.True(t, IsNumber(1.3))
	assert.True(t, IsNumber(-5))
	assert.True(t, IsNumber(-5.5))
	assert.False(t, IsNumber("abcd"))
	assert.False(t, IsNumber(true))
}

func TestToInt(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected int
		hasError bool
	}{
		{"valid integer", 123, 123, false},
		{"negative integer", -123, -123, false},
		{"zero", 0, 0, false},
		{"string integer", "123", 123, false},
		{"negative string", "-123", -123, false},
		{"leading zeros string", "007", 7, false},
		{"float32", float32(42.5), 43, false},   // Rounded to nearest integer
		{"float64", 42.7, 43, false},            // Rounded to nearest integer
		{"float64 negative", -42.5, -43, false}, // Rounded to nearest integer
		{"int64", int64(100), 100, false},
		{"invalid non-numeric string", "abc", 0, true},
		{"empty string", "", 0, true},
		{"spaces", " 123 ", 0, true}, // strconv.Atoi would fail with leading/trailing spaces
		{"unsupported type: nil", nil, 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToInt(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("ToInt(%v) expected error but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ToInt(%v) unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("ToInt(%v) = %d, want %d", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestToInt_Int64Overflow(t *testing.T) {
	origMax := maxIntValue
	origMin := minIntValue
	defer func() {
		maxIntValue = origMax
		minIntValue = origMin
	}()

	maxIntValue = 10
	minIntValue = -11

	_, err := ToInt(int64(100))
	if err == nil {
		t.Fatalf("expected overflow error")
	}
}

func TestToInt64(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected int64
		hasError bool
	}{
		{"valid int", 123, 123, false},
		{"valid int64", int64(456), 456, false},
		{"valid string", "789", 789, false},
		{"negative int", -123, -123, false},
		{"negative string", "-456", -456, false},
		{"zero as int", 0, 0, false},
		{"zero as string", "0", 0, false},
		{"leading zeros", "007", 7, false},
		{"float32", float32(3.14), 3, false},   // Convert float32 to int64 (rounded)
		{"float64", 3.99, 4, false},            // Convert float64 to int64 (rounded)
		{"float64 negative", -3.99, -4, false}, // Convert float64 to int64 (rounded)
		{"int64 to int64", int64(1000), 1000, false},
		{"invalid non-numeric string", "abc", 0, true},
		{"empty string", "", 0, true},
		{"spaces around string", " 123 ", 0, true},
		{"unsupported type: nil", nil, 0, true},
		{"unsupported type: bool", true, 0, true},             // Unsupported bool type
		{"unsupported type: complex", complex(1, 1), 0, true}, // Unsupported complex type
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToInt64(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("ToInt64(%v) expected error but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ToInt64(%v) unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("ToInt64(%v) = %d, want %d", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestToFloat32(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected float32
		hasError bool
	}{
		{"valid int", 123, 123.0, false},
		{"valid int64", int64(456), 456.0, false},
		{"valid float32", float32(3.14), 3.14, false},
		{"valid float64", 3.99, 3.99, false},
		{"negative int", -123, -123.0, false},
		{"negative float64", -3.99, -3.99, false},
		{"string number", "789", 789.0, false},
		{"string with decimal", "3.14", 3.14, false},
		{"leading zeros in string", "007", 7.0, false},
		{"invalid non-numeric string", "abc", 0, true},
		{"empty string", "", 0, true},
		{"spaces around string", " 123 ", 0, true},
		{"unsupported type: nil", nil, 0, true},
		{"unsupported type: bool", true, 0, true},
		{"unsupported type: complex", complex(1, 1), 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToFloat32(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("ToFloat32(%v) expected error but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ToFloat32(%v) unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("ToFloat32(%v) = %f, want %f", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected float64
		hasError bool
	}{
		{"valid int", 123, 123.0, false},
		{"valid int64", int64(456), 456.0, false},
		{"valid float64", 3.99, 3.99, false},
		{"valid float32", float32(1.25), 1.25, false},
		{"negative int", -123, -123.0, false},
		{"negative float64", -3.99, -3.99, false},
		{"string number", "789", 789.0, false},
		{"string with decimal", "3.14", 3.14, false},
		{"leading zeros in string", "007", 7.0, false},
		{"invalid non-numeric string", "abc", 0, true},
		{"empty string", "", 0, true},
		{"spaces around string", " 123 ", 0, true},
		{"unsupported type: nil", nil, 0, true},
		{"unsupported type: bool", true, 0, true},
		{"unsupported type: complex", complex(1, 1), 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToFloat64(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("ToFloat64(%v) expected error but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ToFloat64(%v) unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("ToFloat64(%v) = %f, want %f", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestMaxUint64(t *testing.T) {
	tests := []struct {
		a, b uint
		want uint
	}{
		{0, 0, 0},
		{1, 0, 1},
		{0, 1, 1},
		{1, 1, 1},
		{1, 2, 2},
		{2, 1, 2},
		{2, 2, 2},
		{2, 3, 3},
		{3, 2, 3},
		{3, 3, 3},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := MaxUint64(uint64(tt.a), uint64(tt.b)); got != uint64(tt.want) {
				t.Errorf("MaxUint64() = %v, want %v", got, tt.want)
			}
		})
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := MaxUint8(uint8(tt.a), uint8(tt.b)); got != uint8(tt.want) {
				t.Errorf("MaxUint8() = %v, want %v", got, tt.want)
			}
		})
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := MaxInt64(int64(tt.a), int64(tt.b)); got != int64(tt.want) {
				t.Errorf("MaxInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPackAndUnpackNumbers(t *testing.T) {
	tests := []struct {
		num1 int32
		num2 int32
	}{
		{123456789, 987654321},
		{-123456789, 987654321},
		{123456789, -987654321},
		{-123456789, -987654321},
		{0, 0},
		{1, -1},
	}

	for _, test := range tests {
		packed := PackNumbers(test.num1, test.num2)

		unpackedNum1, unpackedNum2 := UnpackNumbers(packed)
		if unpackedNum1 != test.num1 || unpackedNum2 != test.num2 {
			t.Errorf("UnpackNumbers(%d, %d) = %d, %d; want %d, %d", test.num1, test.num2, unpackedNum1, unpackedNum2, test.num1, test.num2)
		}

		firstNum := UnpackFirstNumber(packed)
		if firstNum != test.num1 {
			t.Errorf("UnpackFirstNumber(%d, %d) = %d; want %d", test.num1, test.num2, firstNum, test.num1)
		}

		secondNum := UnpackSecondNumber(packed)
		if secondNum != test.num2 {
			t.Errorf("UnpackSecondNumber(%d, %d) = %d; want %d", test.num1, test.num2, secondNum, test.num2)
		}
	}
}
