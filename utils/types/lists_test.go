package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChunks(t *testing.T) {

	// Test case 1: Valid input, size=2
	input1 := []interface{}{1, 2, 3, 4, 5}
	expectedOutput1 := [][]interface{}{{1, 2}, {3, 4}, {5}}
	chunks1, err1 := CreateChunks(input1, 2)

	assert.Nil(t, err1)
	assert.Equal(t, expectedOutput1, chunks1)

	// Test case 2: Valid input, size=1
	input2 := []interface{}{1, 2, 3, 4, 5}
	expectedOutput2 := [][]interface{}{{1}, {2}, {3}, {4}, {5}}
	chunks2, err2 := CreateChunks(input2, 1)

	assert.Nil(t, err2)
	assert.Equal(t, expectedOutput2, chunks2)

	// Test case 3: Invalid input, size=0
	input3 := []interface{}{1, 2, 3, 4, 5}
	_, err3 := CreateChunks(input3, 0)

	assert.NotNil(t, err3)
	assert.EqualError(t, err3, "size cannot be less than 1")
}

func TestExistsInList(t *testing.T) {
	testCases := []struct {
		name          string
		val           interface{}
		array         interface{}
		expected      bool
		expectedIndex int
	}{
		{
			name:          "ElementExistsInSlice",
			val:           3,
			array:         []int{1, 2, 3, 4, 5},
			expected:      true,
			expectedIndex: 2,
		},
		{
			name:          "ElementNotExistInSlice",
			val:           "test",
			array:         []string{"one", "two", "three"},
			expected:      false,
			expectedIndex: -1,
		},
		{
			name:          "EmptySlice",
			val:           42,
			array:         []int{},
			expected:      false,
			expectedIndex: -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exists, index := ExistsInList(tc.val, tc.array)
			assert.Equal(t, tc.expected, exists)
			assert.Equal(t, tc.expectedIndex, index)
		})
	}
}

func TestCastInterfaceSlice(t *testing.T) {
	tests := []struct {
		name      string
		input     []interface{}
		expect    []int
		expectErr bool
	}{
		{
			name:      "valid conversion",
			input:     []interface{}{1, 2, 3},
			expect:    []int{1, 2, 3},
			expectErr: false,
		},
		{
			name:      "mixed types with valid conversions",
			input:     []interface{}{1, 2, "3"},
			expect:    []int{1, 2},
			expectErr: true, // "3" cannot be converted to int
		},
		{
			name:      "empty slice",
			input:     []interface{}{},
			expect:    []int{},
			expectErr: false,
		},
		{
			name:      "nil slice",
			input:     nil,
			expect:    nil,
			expectErr: true, // nil input should cause an error
		},
		{
			name:      "invalid type in slice",
			input:     []interface{}{1, 2, 3.5},
			expect:    nil,
			expectErr: true, // 3.5 cannot be converted to int
		},
		{
			name:      "only invalid types",
			input:     []interface{}{"foo", 3.14},
			expect:    nil,
			expectErr: true, // No elements can be converted to int
		},
		{
			name:      "complex type",
			input:     []interface{}{1, 2, []int{3}},
			expect:    nil,
			expectErr: true, // []int cannot be converted to int
		},
		{
			name:      "all valid types with zero",
			input:     []interface{}{0, 1, 2},
			expect:    []int{0, 1, 2},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CastInterfaceSlice[int](tt.input)

			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if !tt.expectErr && (len(result) != len(tt.expect)) {
				t.Errorf("expected: %v, got: %v", tt.expect, result)
			}
		})
	}
}

// TestConvertToInterfaceSlice tests the ConvertToInterfaceSlice function.
func TestConvertToInterfaceSlice(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected []interface{}
	}{
		{
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			input:    []string{"a", "b", "c"},
			expected: []interface{}{"a", "b", "c"},
		},
		{
			input:    []float64{1.1, 2.2, 3.3},
			expected: []interface{}{1.1, 2.2, 3.3},
		},
		{
			input:    []bool{true, false, true},
			expected: []interface{}{true, false, true},
		},
		{
			input:    []struct{ Name string }{{"Alice"}, {"Bob"}},
			expected: []interface{}{struct{ Name string }{"Alice"}, struct{ Name string }{"Bob"}},
		},
		{
			input:    []int{},
			expected: []interface{}{},
		},
	}

	for _, test := range tests {
		// Use type assertion to get the slice type
		inputValue := reflect.ValueOf(test.input)

		// Use a type switch to call ConvertToInterfaceSlice with the correct type
		var outputValue []interface{}
		switch inputValue.Kind() {
		case reflect.Slice:
			switch inputValue.Type().Elem().Kind() {
			case reflect.Int:
				outputValue = ConvertToInterfaceSlice(inputValue.Interface().([]int))
			case reflect.String:
				outputValue = ConvertToInterfaceSlice(inputValue.Interface().([]string))
			case reflect.Float64:
				outputValue = ConvertToInterfaceSlice(inputValue.Interface().([]float64))
			case reflect.Bool:
				outputValue = ConvertToInterfaceSlice(inputValue.Interface().([]bool))
			case reflect.Struct:
				outputValue = ConvertToInterfaceSlice(inputValue.Interface().([]struct{ Name string }))
			}
		}

		// Compare output to expected
		if !reflect.DeepEqual(outputValue, test.expected) {
			t.Errorf("ConvertToInterfaceSlice(%v) = %v; want %v", test.input, outputValue, test.expected)
		}
	}
}
