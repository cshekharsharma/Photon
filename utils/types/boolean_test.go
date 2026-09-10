package types

import (
	"testing"
)

func TestToBool(t *testing.T) {
	cases := []struct {
		name     string
		input    interface{}
		expected bool
		hasError bool
	}{
		{"valid bool true", true, true, false},                    // bool -> bool
		{"valid bool false", false, false, false},                 // bool -> bool
		{"int 1 to true", 1, true, false},                         // int -> true
		{"int 0 to false", 0, false, false},                       // int -> false
		{"int64 1 to true", int64(1), true, false},                // int64 -> true
		{"int64 0 to false", int64(0), false, false},              // int64 -> false
		{"string 'true' to true", "true", true, false},            // string -> true
		{"string 'false' to false", "false", false, false},        // string -> false
		{"string '1' to true", "1", true, false},                  // string -> true
		{"string '0' to false", "0", false, false},                // string -> false
		{"invalid string 'yes'", "yes", false, true},              // invalid string -> error
		{"empty string", "", false, true},                         // empty string -> error
		{"unsupported type: float", 3.14, false, true},            // unsupported type -> error
		{"unsupported type: complex", complex(1, 1), false, true}, // unsupported type -> error
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToBool(tc.input)
			if tc.hasError {
				if err == nil {
					t.Errorf("ToBool(%v) expected error but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ToBool(%v) unexpected error: %v", tc.input, err)
				}
				if result != tc.expected {
					t.Errorf("ToBool(%v) = %v, want %v", tc.input, result, tc.expected)
				}
			}
		})
	}
}
