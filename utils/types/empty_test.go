package types

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	type CustomType struct {
		ID   int
		Name string
	}

	var testCases = []struct {
		value    interface{}
		expected bool
	}{
		// Test Case 1: Empty string
		{value: "", expected: true},

		// Test Case 2: Non-empty string
		{value: "hello", expected: false},

		// Test Case 3: Nil pointer
		{value: (*int)(nil), expected: true},

		// Test Case 4: Non-nil pointer
		{value: new(int), expected: false},

		// Test Case 5: Nil interface
		{value: interface{}(nil), expected: true},

		// Test Case 6: Non-nil interface
		{value: interface{}(42), expected: false},

		// Test Case 7: Empty custom type
		{value: CustomType{}, expected: true},

		// Test Case 8: Non-empty custom type
		{value: CustomType{ID: 1, Name: "test"}, expected: false},

		// Test Case 9: Empty Map
		{value: make(map[string]string), expected: true},

		// Test Case 10: Non-Empty Map
		{value: map[string]string{"A": "B"}, expected: false},

		// Test Case 11: Empty Slice
		{value: []string{}, expected: true},

		// Test Case 12: Non-Empty Slice
		{value: []string{"A", "B"}, expected: false},
	}
	for idx, tc := range testCases {
		actual := IsEmpty(tc.value)
		if actual != tc.expected {
			t.Errorf("Test case %d failed: expected %v, got %v", idx+1, tc.expected, actual)
		}
	}
}
