package types

import (
	"reflect"
	"strings"
	"testing"
)

type TestStruct struct {
	Name                string             `json:"name"`
	Age                 int                `json:"age"`
	Score               *float64           `json:"score,omitempty"`
	Meta                map[string]string  `json:"meta,omitempty"`
	ValidIdentification []ValidIdentifiers `json:"validIdentifiers,omitempty"`
}

type ValidIdentifiers string

type AllTypesStruct struct {
	Int64Field   int64            `json:"int64Field"`
	Float32Field float32          `json:"float32Field"`
	BoolField    bool             `json:"boolField"`
	MapField     map[string]int64 `json:"mapField"`
}

type OmitStruct struct {
	Required string `json:"req"`
	Optional string `json:"opt,omitempty"`
	Skip     string `json:"-"`
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestConvertToStruct_Success(t *testing.T) {
	data := map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"score": 95.5,
		"meta":  map[string]interface{}{"role": "admin", "active": "true"},
	}

	var target TestStruct
	err := convertToStruct(data, &target)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := TestStruct{
		Name:  "Alice",
		Age:   30,
		Score: floatPtr(95.5),
		Meta:  map[string]string{"role": "admin", "active": "true"},
	}

	if !reflect.DeepEqual(target, expected) {
		t.Errorf("Expected %v, got %v", expected, target)
	}
}

func TestConvertToStruct_WithAllTypesStruct(t *testing.T) {
	data := map[string]interface{}{
		"int64Field":   int64(987654321),
		"float32Field": float32(23.45),
		"boolField":    true,
		"mapField":     map[string]interface{}{"key1": int64(100), "key2": int64(200)},
	}

	var target AllTypesStruct
	err := convertToStruct(data, &target)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if target.Int64Field != 987654321 {
		t.Errorf("Expected Int64Field 987654321, got %v", target.Int64Field)
	}

	if target.Float32Field != 23.45 {
		t.Errorf("Expected Float32Field 23.45, got %v", target.Float32Field)
	}

	if target.BoolField != true {
		t.Errorf("Expected BoolField true, got %v", target.BoolField)
	}

	expectedMap := map[string]int64{"key1": 100, "key2": 200}
	if !reflect.DeepEqual(target.MapField, expectedMap) {
		t.Errorf("Expected MapField %v, got %v", expectedMap, target.MapField)
	}
}

func TestConvertField_StringConversion(t *testing.T) {
	var target string
	value := "example"

	field := reflect.ValueOf(&target).Elem()
	fieldType := reflect.TypeOf(target)

	err := convertField(field, reflect.StructField{Type: fieldType}, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if target != "example" {
		t.Errorf("Expected 'example', got %v", target)
	}
}

func TestConvertField_Int64Conversion(t *testing.T) {
	var target int64
	value := int64(123456789)

	field := reflect.ValueOf(&target).Elem()
	fieldType := reflect.TypeOf(target)

	err := convertField(field, reflect.StructField{Type: fieldType}, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if target != 123456789 {
		t.Errorf("Expected 123456789, got %v", target)
	}
}

func TestConvertField_Float32Conversion(t *testing.T) {
	var target float32
	value := 42.5 // value in float64, which should be converted to float32

	field := reflect.ValueOf(&target).Elem()
	fieldType := reflect.TypeOf(target)

	err := convertField(field, reflect.StructField{Type: fieldType}, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if target != float32(42.5) {
		t.Errorf("Expected 42.5, got %v", target)
	}
}

func TestConvertField_BoolConversion(t *testing.T) {
	var target bool
	value := true

	field := reflect.ValueOf(&target).Elem()
	fieldType := reflect.TypeOf(target)

	err := convertField(field, reflect.StructField{Type: fieldType}, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if target != true {
		t.Errorf("Expected true, got %v", target)
	}
}

func TestConvertField_MapConversion(t *testing.T) {
	var target map[string]interface{}
	value := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	field := reflect.ValueOf(&target).Elem()
	fieldType := reflect.TypeOf(target)

	err := convertField(field, reflect.StructField{Type: fieldType}, value)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	if !reflect.DeepEqual(target, expected) {
		t.Errorf("Expected %v, got %v", expected, target)
	}
}

func TestConvertToStruct_MissingField(t *testing.T) {
	data := map[string]interface{}{
		"name": "Alice",
	}

	var target TestStruct
	err := convertToStruct(data, &target)
	if err == nil {
		t.Errorf("Expected error for missing field, got nil")
	}
}

func TestConvertToStruct_ErrorOnMarshal(t *testing.T) {
	var target TestStruct
	err := convertToStruct(make(chan int), &target)
	if err == nil {
		t.Errorf("expected error for unsupported marshal type")
	}
}

func TestUnmarshalCustom_InvalidJSON(t *testing.T) {
	var target TestStruct
	err := UnmarshalCustom([]byte("{invalid"), &target)
	if err == nil {
		t.Errorf("expected error for invalid json")
	}
}

func TestConvertField_MapErrors(t *testing.T) {
	type mapHolder struct {
		Values map[string]int64 `json:"values"`
	}
	var target mapHolder
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, "not-a-map")
	if err == nil {
		t.Errorf("expected error for non-map value")
	}

	err = convertField(field, fieldType, map[string]interface{}{"k": "bad"})
	if err == nil {
		t.Errorf("expected error for type mismatch in map value")
	}
}

func TestConvertField_MapStructValues(t *testing.T) {
	type inner struct {
		ID int `json:"id"`
	}
	type outer struct {
		Items map[string]inner `json:"items"`
	}

	value := map[string]interface{}{
		"items": map[string]interface{}{
			"a": map[string]interface{}{"id": 5},
		},
	}

	var target outer
	err := convertToStruct(value, &target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Items["a"].ID != 5 {
		t.Fatalf("expected id=5, got %v", target.Items["a"].ID)
	}
}

func TestConvertField_SliceErrors(t *testing.T) {
	type sliceHolder struct {
		Nums []int `json:"nums"`
	}
	var target sliceHolder
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, "not-a-slice")
	if err == nil {
		t.Errorf("expected error for non-slice")
	}
}

func TestConvertField_SliceElementError(t *testing.T) {
	type sliceHolder struct {
		Nums []int `json:"nums"`
	}
	var target sliceHolder
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, []interface{}{"bad"})
	if err == nil {
		t.Errorf("expected error for bad slice element")
	}
}

func TestConvertField_UnsupportedType(t *testing.T) {
	type badStruct struct {
		C complex64 `json:"c"`
	}
	var target badStruct
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)
	err := convertField(field, fieldType, complex64(1))
	if err == nil {
		t.Errorf("expected error for unsupported type")
	}
}

func TestConvertField_MapFloat64ToInt64(t *testing.T) {
	type holder struct {
		Values map[string]int64 `json:"values"`
	}
	var target holder
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, map[string]interface{}{"a": float64(12)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Values["a"] != 12 {
		t.Fatalf("expected 12, got %v", target.Values["a"])
	}
}

func TestConvertField_MapUnsupportedElem(t *testing.T) {
	type holder struct {
		Values map[string][]int `json:"values"`
	}
	var target holder
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, map[string]interface{}{"a": []int{1}})
	if err == nil {
		t.Fatalf("expected error for unsupported map element type")
	}
}

func TestConvertField_StructError(t *testing.T) {
	type holder struct {
		Inner struct {
			ID int `json:"id"`
		} `json:"inner"`
	}
	var target holder
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, make(chan int))
	if err == nil {
		t.Fatalf("expected error for struct conversion")
	}
}

func TestConvertField_ParseErrors(t *testing.T) {
	var i64 int64
	err := convertField(reflect.ValueOf(&i64).Elem(), reflect.StructField{Type: reflect.TypeOf(i64)}, "bad")
	if err == nil {
		t.Fatalf("expected int64 parse error")
	}

	var f64 float64
	err = convertField(reflect.ValueOf(&f64).Elem(), reflect.StructField{Type: reflect.TypeOf(f64)}, "bad")
	if err == nil {
		t.Fatalf("expected float64 parse error")
	}

	var f32 float32
	err = convertField(reflect.ValueOf(&f32).Elem(), reflect.StructField{Type: reflect.TypeOf(f32)}, "bad")
	if err == nil {
		t.Fatalf("expected float32 parse error")
	}

	var b bool
	err = convertField(reflect.ValueOf(&b).Elem(), reflect.StructField{Type: reflect.TypeOf(b)}, "bad")
	if err == nil {
		t.Fatalf("expected bool parse error")
	}
}

func TestConvertField_IntParsedSet(t *testing.T) {
	var i int
	err := convertField(reflect.ValueOf(&i).Elem(), reflect.StructField{Type: reflect.TypeOf(i)}, "7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if i != 7 {
		t.Fatalf("expected 7, got %d", i)
	}
}

func TestConvertField_MapStructConvertError(t *testing.T) {
	type inner struct {
		ID int `json:"id"`
	}
	type outer struct {
		Items map[string]inner `json:"items"`
	}

	var target outer
	field := reflect.ValueOf(&target).Elem().Field(0)
	fieldType := reflect.TypeOf(target).Field(0)

	err := convertField(field, fieldType, map[string]interface{}{"a": make(chan int)})
	if err == nil {
		t.Fatalf("expected error for map struct conversion")
	}
}

func TestUnmarshalCustom_Success(t *testing.T) {
	data := `{"name": "Bob", "age": 25, "score": 88.7, "meta": {"role": "user"}, "validIdentifiers": ["aadhar", "pan card", "passport"]}`
	var target TestStruct
	err := UnmarshalCustom([]byte(data), &target)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	expected := TestStruct{
		Name:                "Bob",
		Age:                 25,
		Score:               floatPtr(88.7),
		Meta:                map[string]string{"role": "user"},
		ValidIdentification: []ValidIdentifiers{"aadhar", "pan card", "passport"},
	}

	if !reflect.DeepEqual(target, expected) {
		t.Errorf("Expected %v, got %v", expected, target)
	}
}

func TestUnmarshalCustom_MissingRequiredField(t *testing.T) {
	data := `{"name": "Charlie"}` // Age is missing
	var target TestStruct
	err := UnmarshalCustom([]byte(data), &target)

	if err == nil || err.Error() != "missing required field 'age' in JSON" {
		t.Errorf("Expected missing field error, got %v", err)
	}
}

func TestUnmarshalCustom_ConvertFieldError(t *testing.T) {
	data := `{"name": "Bob", "age": "bad"}`
	var target TestStruct
	err := UnmarshalCustom([]byte(data), &target)
	if err == nil || !strings.Contains(err.Error(), "error converting field Age") {
		t.Fatalf("expected conversion error, got %v", err)
	}
}

func TestUnmarshalCustom_OmitEmptyAndSkip(t *testing.T) {
	data := `{"req":"ok"}`
	var target OmitStruct
	err := UnmarshalCustom([]byte(data), &target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Required != "ok" {
		t.Fatalf("expected required=ok, got %q", target.Required)
	}
	if target.Optional != "" || target.Skip != "" {
		t.Fatalf("expected optional/skip to be empty")
	}
}
