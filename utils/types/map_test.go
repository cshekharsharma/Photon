package types

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToMapInterface(t *testing.T) {
	input := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"nested": map[string]interface{}{
			"nestedKey1": "nestedValue1",
			"nestedKey2": 100,
		},
	}

	expected := map[interface{}]interface{}{
		"key1": "value1",
		"key2": 42,
		"nested": map[interface{}]interface{}{
			"nestedKey1": "nestedValue1",
			"nestedKey2": 100,
		},
	}

	result := ConvertToMapInterface(input)
	assert.Equal(t, expected, result)
}

func TestConvertToMapStringInterface(t *testing.T) {
	input := map[interface{}]interface{}{
		"key1": "value1",
		"key2": 42,
		"nested": map[interface{}]interface{}{
			"nestedKey1": "nestedValue1",
			"nestedKey2": 100,
		},
	}

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"nested": map[string]interface{}{
			"nestedKey1": "nestedValue1",
			"nestedKey2": 100,
		},
	}

	result := ConvertToMapStringInterface(input)
	assert.Equal(t, expected, result)
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name     string
		map1     map[string]int
		map2     map[string]int
		expected map[string]int
	}{
		{
			name:     "both maps are empty",
			map1:     map[string]int{},
			map2:     map[string]int{},
			expected: map[string]int{},
		},
		{
			name:     "first map is empty",
			map1:     map[string]int{},
			map2:     map[string]int{"a": 1, "b": 2},
			expected: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "second map is empty",
			map1:     map[string]int{"c": 3, "d": 4},
			map2:     map[string]int{},
			expected: map[string]int{"c": 3, "d": 4},
		},
		{
			name:     "no overlapping keys",
			map1:     map[string]int{"a": 1, "b": 2},
			map2:     map[string]int{"c": 3, "d": 4},
			expected: map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		},
		{
			name:     "with overlapping keys",
			map1:     map[string]int{"a": 1, "b": 2},
			map2:     map[string]int{"b": 3, "c": 4},
			expected: map[string]int{"a": 1, "b": 3, "c": 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeMaps(tt.map1, tt.map2)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

type nestedSource struct {
	City string
	Code int
}

type nestedTarget struct {
	City string
	Code int
}

type sourceRec struct {
	Name   string
	Age    int
	Nested nestedSource
	Ignore string
}

type targetRec struct {
	Name   string
	Age    int
	Nested nestedTarget
}

func TestCopyStructFieldsRecursive(t *testing.T) {
	src := &sourceRec{
		Name: "John",
		Age:  42,
		Nested: nestedSource{
			City: "New York",
			Code: 10001,
		},
		Ignore: "not copied",
	}

	dst := &targetRec{}

	CopyStructFieldsRecursive(dst, src)

	if dst.Name != "John" {
		t.Errorf("expected Name = John, got %s", dst.Name)
	}
	if dst.Age != 42 {
		t.Errorf("expected Age = 42, got %d", dst.Age)
	}
	if dst.Nested.City != "New York" {
		t.Errorf("expected Nested.City = New York, got %s", dst.Nested.City)
	}
	if dst.Nested.Code != 10001 {
		t.Errorf("expected Nested.Code = 10001, got %d", dst.Nested.Code)
	}
}

func TestCopyStructFieldsRecursive_MissingAndUnsettable(t *testing.T) {
	type src struct {
		A int
	}
	type dst struct {
		A      int
		C      string
		hidden int
	}

	s := &src{A: 10}
	d := &dst{hidden: 5}

	CopyStructFieldsRecursive(d, s)

	if d.A != 10 {
		t.Fatalf("expected A=10, got %d", d.A)
	}
	if d.C != "" {
		t.Fatalf("expected C to remain empty, got %q", d.C)
	}
	if d.hidden != 5 {
		t.Fatalf("expected hidden to remain 5, got %d", d.hidden)
	}
}

func TestCopyStructFieldsRecursive_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for non-pointer")
		}
	}()
	var src sourceRec
	var dst targetRec
	CopyStructFieldsRecursive(dst, &src)
}

func TestCopyStructFieldsRecursive_PanicsOnNonStruct(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for non-struct")
		}
	}()
	var src int
	var dst int
	CopyStructFieldsRecursive(&dst, &src)
}
