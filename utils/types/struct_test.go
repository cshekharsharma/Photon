package types

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type SampleStruct struct {
	Name    string     `json:"name"`
	Age     int        `json:"age"`
	Active  bool       `json:"active"`
	Birth   *time.Time `json:"birth"`
	Rate    float64    `json:"rate"`
	Profile SubStruct  `json:"profile"`
}

type SubStruct struct {
	City string `json:"city"`
	Zip  int    `json:"zip"`
}

type NestedStruct struct {
	BasicInfo SampleStruct `json:"basicInfo"`
	Level     string       `json:"level"`
}

func TestPopulateStructFromMap(t *testing.T) {
	nowStr := "2023-05-20 14:45:00"
	nowTime, _ := time.Parse(time.DateTime, nowStr)

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected SampleStruct
		wantErr  bool
	}{
		{
			name: "basic fields",
			input: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"active": true,
			},
			expected: SampleStruct{
				Name:   "John",
				Age:    30,
				Active: true,
			},
		},
		{
			name: "string to int and bool",
			input: map[string]interface{}{
				"name":   "Jane",
				"age":    "42",
				"active": "true",
			},
			expected: SampleStruct{
				Name:   "Jane",
				Age:    42,
				Active: true,
			},
		},
		{
			name: "string to time.Time pointer",
			input: map[string]interface{}{
				"birth": nowStr,
			},
			expected: SampleStruct{
				Birth: &nowTime,
			},
		},
		{
			name: "nested struct mapping",
			input: map[string]interface{}{
				"profile": map[string]interface{}{
					"city": "Mumbai",
					"zip":  "400001",
				},
			},
			expected: SampleStruct{
				Profile: SubStruct{
					City: "Mumbai",
					Zip:  400001,
				},
			},
		},
		{
			name: "float to int fallback",
			input: map[string]interface{}{
				"age": 30.0,
			},
			expected: SampleStruct{
				Age: 30,
			},
		},
		{
			name: "type mismatch error",
			input: map[string]interface{}{
				"age": []string{"invalid"},
			},
			wantErr: true,
		},
		{
			name: "all fields set",
			input: map[string]interface{}{
				"name":   "Alice",
				"age":    "28",
				"active": "false",
				"birth":  nowStr,
				"rate":   "4.5",
				"profile": map[string]interface{}{
					"city": "Delhi",
					"zip":  110001,
				},
			},
			expected: SampleStruct{
				Name:   "Alice",
				Age:    28,
				Active: false,
				Birth:  &nowTime,
				Rate:   4.5,
				Profile: SubStruct{
					City: "Delhi",
					Zip:  110001,
				},
			},
		},
		{
			name: "missing fields",
			input: map[string]interface{}{
				"name": "Bob",
			},
			expected: SampleStruct{
				Name: "Bob",
			},
		},
		{
			name: "zero values",
			input: map[string]interface{}{
				"name":   "",
				"age":    0,
				"active": false,
				"rate":   0.0,
				"profile": map[string]interface{}{
					"city": "",
					"zip":  0,
				},
			},
			expected: SampleStruct{
				Name:   "",
				Age:    0,
				Active: false,
				Rate:   0.0,
				Profile: SubStruct{
					City: "",
					Zip:  0,
				},
			},
		},
		{
			name: "invalid time format",
			input: map[string]interface{}{
				"birth": "not-a-date",
			},
			wantErr: true,
		},
		{
			name: "profile as nil",
			input: map[string]interface{}{
				"profile": nil,
			},
			expected: SampleStruct{},
		},
		{
			name: "profile as wrong type",
			input: map[string]interface{}{
				"profile": "not-a-map",
			},
			wantErr: true,
		},
		{
			name: "rate as float",
			input: map[string]interface{}{
				"rate": 7.77,
			},
			expected: SampleStruct{
				Rate: 7.77,
			},
		},
		{
			name: "rate as string",
			input: map[string]interface{}{
				"rate": "8.88",
			},
			expected: SampleStruct{
				Rate: 8.88,
			},
		},
		{
			name: "extra fields ignored",
			input: map[string]interface{}{
				"name":   "Extra",
				"foobar": "should be ignored",
			},
			expected: SampleStruct{
				Name: "Extra",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SampleStruct
			err := PopulateStructFromMap(&result, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Compare result fields individually
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %v; want %v", result.Name, tt.expected.Name)
			}
			if result.Age != tt.expected.Age {
				t.Errorf("Age = %v; want %v", result.Age, tt.expected.Age)
			}
			if result.Active != tt.expected.Active {
				t.Errorf("Active = %v; want %v", result.Active, tt.expected.Active)
			}
			if tt.expected.Birth != nil {
				if result.Birth == nil || !result.Birth.Equal(*tt.expected.Birth) {
					t.Errorf("Birth = %v; want %v", result.Birth, tt.expected.Birth)
				}
			}
			if result.Rate != tt.expected.Rate {
				t.Errorf("Rate = %v; want %v", result.Rate, tt.expected.Rate)
			}
			if !reflect.DeepEqual(result.Profile, tt.expected.Profile) {
				t.Errorf("Profile = %+v; want %+v", result.Profile, tt.expected.Profile)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	now := time.Now()
	input := SampleStruct{
		Name:   "Jane",
		Age:    25,
		Active: false,
		Rate:   3.75,
		Birth:  &now,
	}

	result, err := StructToMap(input)
	assert.NoError(t, err)

	assert.Equal(t, input.Name, result["Name"])
	assert.Equal(t, input.Age, result["Age"])
	assert.Equal(t, input.Active, result["Active"])
	assert.Equal(t, input.Rate, result["Rate"])
	assert.Equal(t, input.Birth, result["Birth"])
}

func TestStructToMapErrors(t *testing.T) {
	_, err := StructToMap(123) // Not a struct
	assert.Error(t, err)
}

func TestStructToMap_Pointer(t *testing.T) {
	input := &SampleStruct{Name: "Ptr"}
	result, err := StructToMap(input)
	assert.NoError(t, err)
	assert.Equal(t, "Ptr", result["Name"])
}

func TestPopulateStructFromMap_Errors(t *testing.T) {
	err := PopulateStructFromMap(nil, map[string]interface{}{})
	assert.Error(t, err)

	var notStruct int
	err = PopulateStructFromMap(&notStruct, map[string]interface{}{})
	assert.Error(t, err)

	err = PopulateStructFromMap(notStruct, map[string]interface{}{})
	assert.Error(t, err)
}

func TestPopulateStructFromMap_MapInterfaceKey(t *testing.T) {
	type inner struct {
		ID int `json:"id"`
	}
	type outer struct {
		Items map[string]inner `json:"items"`
	}

	input := map[string]interface{}{
		"items": map[interface{}]interface{}{
			"a": map[interface{}]interface{}{"id": 7},
		},
	}

	var out outer
	err := PopulateStructFromMap(&out, input)
	assert.NoError(t, err)
	assert.Equal(t, 7, out.Items["a"].ID)
}

func TestPopulateStructFromMap_SkipsMissingAndUnexported(t *testing.T) {
	type target struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		secret string
	}

	var out target
	err := PopulateStructFromMap(&out, map[string]interface{}{"name": "ok", "secret": "nope"})
	assert.NoError(t, err)
	assert.Equal(t, "ok", out.Name)
	assert.Equal(t, 0, out.Age)
	assert.Equal(t, "", out.secret)
}

func TestPopulateStructFromMap_TagWithComma(t *testing.T) {
	type target struct {
		Name string `json:"name,omitempty"`
	}
	var out target
	err := PopulateStructFromMap(&out, map[string]interface{}{"name": "ok"})
	assert.NoError(t, err)
	assert.Equal(t, "ok", out.Name)
}

func TestSetValueForStructToMap_MapConversion(t *testing.T) {
	type inner struct {
		ID int `json:"id"`
	}
	type holder struct {
		Items map[string]inner
	}

	h := &holder{}
	field := reflect.ValueOf(h).Elem().Field(0)

	raw := map[string]interface{}{
		"a": map[string]interface{}{"id": 9},
	}
	err := setValueForStructToMap(field, reflect.ValueOf(raw))
	assert.NoError(t, err)
	assert.Equal(t, 9, h.Items["a"].ID)
}

func TestSetValueForStructToMap_MapConvertibleElem(t *testing.T) {
	type holder struct {
		Items map[string]int
	}
	h := &holder{}
	field := reflect.ValueOf(h).Elem().Field(0)

	raw := map[string]interface{}{
		"a": int64(5),
	}
	err := setValueForStructToMap(field, reflect.ValueOf(raw))
	assert.NoError(t, err)
	assert.Equal(t, 5, h.Items["a"])
}

func TestSetValueForStructToMap_MapStructError(t *testing.T) {
	type inner struct {
		ID int `json:"id"`
	}
	type holder struct {
		Items map[string]inner
	}
	h := &holder{}
	field := reflect.ValueOf(h).Elem().Field(0)

	raw := map[string]interface{}{
		"a": map[string]interface{}{"id": "not-int"},
	}
	err := setValueForStructToMap(field, reflect.ValueOf(raw))
	assert.Error(t, err)
}

func TestSetValueForStructToMap_InvalidValue(t *testing.T) {
	type holder struct {
		Count int
	}
	h := &holder{Count: 7}
	field := reflect.ValueOf(h).Elem().Field(0)

	err := setValueForStructToMap(field, reflect.Value{})
	assert.NoError(t, err)
	assert.Equal(t, 0, h.Count)
}

func TestSetValueForStructToMap_PointerField(t *testing.T) {
	type holder struct {
		Count *int
	}
	h := &holder{}
	field := reflect.ValueOf(h).Elem().Field(0)

	err := setValueForStructToMap(field, reflect.ValueOf("12"))
	assert.NoError(t, err)
	if h.Count == nil || *h.Count != 12 {
		t.Fatalf("expected Count=12, got %#v", h.Count)
	}
}

func TestSetValueForStructToMap_MapTypeMismatch(t *testing.T) {
	type holder struct {
		Items map[string]int
	}
	h := &holder{}
	field := reflect.ValueOf(h).Elem().Field(0)

	raw := map[int]interface{}{1: "bad"}
	err := setValueForStructToMap(field, reflect.ValueOf(raw))
	assert.Error(t, err)
}

func TestSetValueForStructToMap_StringParseErrors(t *testing.T) {
	t.Run("bool error", func(t *testing.T) {
		var v bool
		err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf("nope"))
		assert.Error(t, err)
	})

	t.Run("int error", func(t *testing.T) {
		var v int
		err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf("bad"))
		assert.Error(t, err)
	})

	t.Run("uint error", func(t *testing.T) {
		var v uint
		err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf("-1"))
		assert.Error(t, err)
	})

	t.Run("float error", func(t *testing.T) {
		var v float64
		err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf("bad"))
		assert.Error(t, err)
	})

	t.Run("time error", func(t *testing.T) {
		var v time.Time
		err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf("bad-time"))
		assert.Error(t, err)
	})
}

func TestSetValueForStructToMap_StringToUintSuccess(t *testing.T) {
	var v uint
	err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf("5"))
	assert.NoError(t, err)
	assert.Equal(t, uint(5), v)
}

func TestSetValueForStructToMap_FloatToIntAndIntToBool(t *testing.T) {
	var iv int
	err := setValueForStructToMap(reflect.ValueOf(&iv).Elem(), reflect.ValueOf(float64(3)))
	assert.NoError(t, err)
	assert.Equal(t, 3, iv)

	var bv bool
	err = setValueForStructToMap(reflect.ValueOf(&bv).Elem(), reflect.ValueOf(1))
	assert.NoError(t, err)
	assert.True(t, bv)
}

func TestSetValueForStructToMap_TypeMismatch(t *testing.T) {
	var v int
	err := setValueForStructToMap(reflect.ValueOf(&v).Elem(), reflect.ValueOf(struct{}{}))
	assert.Error(t, err)
}

func TestIsASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", true},
		{"All ASCII", "Hello, world!", true},
		{"Contains non-ASCII", "Café", false},
		{"Non-ASCII character", "你好", false},
		{"Mixed characters", "abc你好xyz", false},
		{"Long ASCII", "This is a long string with only ASCII characters...", true},
		{"Numerical", "1234567890", true},
		{"Special characters", "!@#$%^&*()", true},
		{"Whitespaces", " \t\n\r\v\f", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsASCII(tc.input), "Test Failed: %s", tc.name)
		})
	}
}

func TestSetValueForStructToMap(t *testing.T) {
	type MyStruct struct {
		BoolVal    bool
		IntVal     int
		Int64Val   int64
		FloatVal   float64
		TimeVal    time.Time
		IntFromF64 int
		DirectConv string
		PtrInt     *int
	}

	s := &MyStruct{}
	v := reflect.ValueOf(s).Elem()

	validTime := "2020-01-01 15:00:00"
	expectedTime, _ := time.Parse(time.DateTime, validTime)

	tests := []struct {
		name        string
		fieldName   string
		input       any
		expectErr   bool
		expectedVal any
	}{
		// ✅ Valid conversions
		{"Valid Bool", "BoolVal", "true", false, true},
		{"Valid Int", "IntVal", "42", false, 42},
		{"Valid Int64", "Int64Val", "98765", false, int64(98765)},
		{"Valid Float64", "FloatVal", "3.14", false, 3.14},
		{"Valid Time", "TimeVal", validTime, false, expectedTime},
		{"Float64 to Int", "IntFromF64", float64(7.0), false, 7},
		{"Direct convertible", "DirectConv", "hello", false, "hello"},
		{"Pointer Int", "PtrInt", "123", false, 123},

		// ❌ Error cases
		{"Invalid Bool", "BoolVal", "notabool", true, nil},
		{"Invalid Int", "IntVal", "notanint", true, nil},
		{"Invalid Int64", "Int64Val", "64bitfail", true, nil},
		{"Invalid Float64", "FloatVal", "notafloat", true, nil},
		{"Invalid Time", "TimeVal", "13-02-2024", true, nil},
		{"Unconvertible Type", "IntVal", struct{}{}, true, nil},

		// ✅ Additional valid conversions
		{"Bool from int (1)", "BoolVal", 1, false, true},
		{"Bool from int (0)", "BoolVal", 0, false, false},
		{"Int from int", "IntVal", 55, false, 55},
		{"Int64 from int64", "Int64Val", int64(123456), false, int64(123456)},
		{"Float64 from float64", "FloatVal", 2.718, false, 2.718},
		{"String to string", "DirectConv", "world", false, "world"},
		{"PtrInt from int", "PtrInt", 456, false, 456},

		// ✅ Nil and zero value handling
		{"Nil value sets zero int", "IntVal", nil, false, 0},
		{"Nil value sets zero bool", "BoolVal", nil, false, false},
		{"Nil value sets zero float", "FloatVal", nil, false, 0.0},
		{"Nil value sets zero string", "DirectConv", nil, false, ""},
		{"Nil value sets zero time", "TimeVal", nil, false, time.Time{}},

		// ❌ More error cases
		{"Unconvertible float to string", "DirectConv", 1.23, true, nil},
		{"Unconvertible bool to int", "IntVal", true, true, nil},
		{"Unconvertible string to int64", "Int64Val", "notanumber", true, nil},
		{"Unconvertible float to bool", "BoolVal", 3.14, true, nil},
		{"Unconvertible struct to float", "FloatVal", struct{}{}, true, nil},

		// ✅ Additional missing test cases
		{"Bool from string 'false'", "BoolVal", "false", false, false},
		{"Bool from string '1'", "BoolVal", "1", false, true},
		{"Bool from string '0'", "BoolVal", "0", false, false},
		{"Int from float64 with decimal", "IntVal", 123.99, false, 123},
		{"Float64 from int", "FloatVal", 42, false, 42.0},
		{"Float64 from string int", "FloatVal", "42", false, 42.0},
		{"Time from time.Time", "TimeVal", expectedTime, false, expectedTime},
		{"PtrInt from nil", "PtrInt", nil, false, 0},
		{"PtrInt from float64", "PtrInt", 321.0, false, 321},
		{"DirectConv from nil", "DirectConv", nil, false, ""},
		{"DirectConv from bool", "DirectConv", true, true, nil},
		{"IntVal from bool", "IntVal", false, true, nil},
		{"FloatVal from bool", "FloatVal", true, true, nil},
		{"TimeVal from int", "TimeVal", 1234567890, true, nil},
	}

	for _, tt := range tests {
		field := v.FieldByName(tt.fieldName)
		inputVal := reflect.ValueOf(tt.input)

		err := setValueForStructToMap(field, inputVal)

		if tt.expectErr {
			if err == nil {
				t.Errorf("[%s] Expected error, got none", tt.name)
			}
			continue
		}

		if err != nil {
			t.Errorf("[%s] Unexpected error: %v", tt.name, err)
			continue
		}

		// Validate actual value
		switch tt.fieldName {
		case "BoolVal":
			got := s.BoolVal
			want := tt.expectedVal.(bool)
			if got != want {
				t.Errorf("[%s] Expected %v, got %v", tt.name, want, got)
			}
		case "IntVal":
			got := s.IntVal
			want := tt.expectedVal.(int)
			if got != want {
				t.Errorf("[%s] Expected %v, got %v", tt.name, want, got)
			}
		case "Int64Val":
			got := s.Int64Val
			want := tt.expectedVal.(int64)
			if got != want {
				t.Errorf("[%s] Expected %v, got %v", tt.name, want, got)
			}
		case "FloatVal":
			got := s.FloatVal
			want := tt.expectedVal.(float64)
			if got != want {
				t.Errorf("[%s] Expected %v, got %v", tt.name, want, got)
			}
		case "TimeVal":
			got := s.TimeVal
			want := tt.expectedVal.(time.Time)
			if !got.Equal(want) {
				t.Errorf("[%s] Expected %v, got %v", tt.name, want, got)
			}
		case "IntFromF64":
			got := s.IntFromF64
			want := tt.expectedVal.(int)
			if got != want {
				t.Errorf("[%s] Expected %v, got %v", tt.name, want, got)
			}
		case "DirectConv":
			got := s.DirectConv
			want := tt.expectedVal.(string)
			if got != want {
				t.Errorf("[%s] Expected %q, got %q", tt.name, want, got)
			}
		case "PtrInt":
			if s.PtrInt == nil {
				if tt.input == nil {
					if tt.expectedVal.(int) != 0 {
						t.Errorf("[%s] Expected zero pointer value", tt.name)
					}
				} else {
					t.Errorf("[%s] Expected non-nil pointer", tt.name)
				}
			} else {
				want := tt.expectedVal.(int)
				if *s.PtrInt != want {
					t.Errorf("[%s] Expected %v, got %v", tt.name, want, *s.PtrInt)
				}
			}
		}
	}
}

func TestConvertMapToTypedMap(t *testing.T) {
	type args struct {
		raw any
	}
	type testCase[K ~string, V any] struct {
		name     string
		args     args
		expected map[K]V
	}

	intMap := map[string]int{"a": 1, "b": 2}
	intfMap := map[string]any{"a": 1, "b": 2}
	strMap := map[string]string{"x": "foo", "y": "bar"}
	strIntfMap := map[string]any{"x": "foo", "y": "bar"}
	mixedIntfMap := map[string]any{"x": "foo", "y": 123, "z": "baz"}
	emptyMap := map[string]any{}

	testsInt := []testCase[string, int]{
		{
			name:     "map[string]int input",
			args:     args{raw: intMap},
			expected: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "map[string]interface{} input with ints",
			args:     args{raw: intfMap},
			expected: map[string]int{"a": 1, "b": 2},
		},
		{
			name:     "map[string]interface{} with non-int values",
			args:     args{raw: mixedIntfMap},
			expected: map[string]int{"y": 123},
		},
		{
			name:     "empty map[string]interface{}",
			args:     args{raw: emptyMap},
			expected: map[string]int{},
		},
	}

	testsString := []testCase[string, string]{
		{
			name:     "map[string]string input",
			args:     args{raw: strMap},
			expected: map[string]string{"x": "foo", "y": "bar"},
		},
		{
			name:     "map[string]interface{} input with strings",
			args:     args{raw: strIntfMap},
			expected: map[string]string{"x": "foo", "y": "bar"},
		},
		{
			name:     "map[string]interface{} with mixed values",
			args:     args{raw: mixedIntfMap},
			expected: map[string]string{"x": "foo", "z": "baz"},
		},
		{
			name:     "empty map[string]interface{}",
			args:     args{raw: emptyMap},
			expected: map[string]string{},
		},
	}

	for _, tt := range testsInt {
		t.Run("int-"+tt.name, func(t *testing.T) {
			got := ConvertMapToTypedMap[string, int](tt.args.raw)
			assert.Equal(t, tt.expected, got)
		})
	}

	for _, tt := range testsString {
		t.Run("string-"+tt.name, func(t *testing.T) {
			got := ConvertMapToTypedMap[string, string](tt.args.raw)
			assert.Equal(t, tt.expected, got)
		})
	}

	// Test with custom key type
	type MyKey string
	customMap := map[string]int{"foo": 10}
	expectedCustom := map[MyKey]int{"foo": 10}
	t.Run("custom key type", func(t *testing.T) {
		got := ConvertMapToTypedMap[MyKey, int](customMap)
		assert.Equal(t, expectedCustom, got)
	})

	// Test with nil input
	t.Run("nil input", func(t *testing.T) {
		got := ConvertMapToTypedMap[string, int](nil)
		assert.Equal(t, map[string]int{}, got)
	})
}

func TestNormalizeInterfaceData(t *testing.T) {
	type testCase struct {
		name     string
		input    interface{}
		expected interface{}
	}

	tests := []testCase{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "simple map[string]interface{}",
			input:    map[string]interface{}{"a": 1, "b": "x"},
			expected: map[string]interface{}{"a": 1, "b": "x"},
		},
		{
			name:     "map[interface{}]interface{} with string keys",
			input:    map[interface{}]interface{}{"foo": 42, "bar": "baz"},
			expected: map[string]interface{}{"foo": 42, "bar": "baz"},
		},
		{
			name:     "map[interface{}]interface{} with non-string key",
			input:    map[interface{}]interface{}{"foo": 1, 123: "skip"},
			expected: map[string]any{"foo": 1, "123": "skip"},
		},
		{
			name: "nested map[interface{}]interface{}",
			input: map[interface{}]interface{}{
				"outer": map[interface{}]interface{}{
					"inner": 5,
				},
			},
			expected: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": 5,
				},
			},
		},
		{
			name:     "nested map[string]interface{}",
			input:    map[string]interface{}{"a": map[string]interface{}{"b": 2}},
			expected: map[string]interface{}{"a": map[string]interface{}{"b": 2}},
		},
		{
			name:     "slice of interface{}",
			input:    []interface{}{1, "x", map[interface{}]interface{}{"y": 2}},
			expected: []interface{}{1, "x", map[string]interface{}{"y": 2}},
		},
		{
			name: "deeply nested structure",
			input: map[interface{}]interface{}{
				"level1": []interface{}{
					map[interface{}]interface{}{"level2": map[string]interface{}{"val": 3}},
				},
			},
			expected: map[string]interface{}{
				"level1": []interface{}{
					map[string]interface{}{"level2": map[string]interface{}{"val": 3}},
				},
			},
		},
		{
			name:     "primitive value",
			input:    123,
			expected: 123,
		},
		{
			name:     "slice of primitives",
			input:    []any{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "empty map[interface{}]interface{}",
			input:    map[interface{}]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "empty map[string]interface{}",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "empty slice",
			input:    []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "slice of map[interface{}]interface{}",
			input:    []interface{}{map[interface{}]interface{}{"a": 1}, map[interface{}]interface{}{"b": 2}},
			expected: []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}},
		},
		{
			name:     "slice of map[string]interface{}",
			input:    []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}},
			expected: []interface{}{map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}},
		},
		{
			name:     "map[interface{}]interface{} with slice value",
			input:    map[interface{}]interface{}{"arr": []interface{}{1, 2, map[interface{}]interface{}{"x": "y"}}},
			expected: map[string]interface{}{"arr": []interface{}{1, 2, map[string]interface{}{"x": "y"}}},
		},
		{
			name:     "map[string]interface{} with slice value",
			input:    map[string]interface{}{"arr": []interface{}{1, 2, map[string]interface{}{"x": "y"}}},
			expected: map[string]interface{}{"arr": []interface{}{1, 2, map[string]interface{}{"x": "y"}}},
		},
		{
			name:     "slice of nils",
			input:    []interface{}{nil, nil},
			expected: []interface{}{nil, nil},
		},
		{
			name:     "map[interface{}]interface{} with nil value",
			input:    map[interface{}]interface{}{"a": nil},
			expected: map[string]interface{}{"a": nil},
		},
		{
			name:     "map[string]interface{} with nil value",
			input:    map[string]interface{}{"a": nil},
			expected: map[string]interface{}{"a": nil},
		},
		{
			name:     "slice of mixed types",
			input:    []interface{}{1, "a", map[interface{}]interface{}{"b": 2}, nil},
			expected: []interface{}{1, "a", map[string]interface{}{"b": 2}, nil},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeInterfaceData(tc.input)
			assert.Equal(t, tc.expected, got)
		})
	}
}
