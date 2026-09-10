package encoding

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	Name string
	Age  int
}

func TestEncodeDecodeGOB(t *testing.T) {
	original := TestStruct{Name: "Alice", Age: 30}

	// Encode
	data, err := EncodeGOB(original)
	if err != nil {
		t.Fatalf("EncodeGOB failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("EncodeGOB returned empty byte slice")
	}

	// Decode
	var decoded TestStruct
	err = DecodeGOB(data, &decoded)
	if err != nil {
		t.Fatalf("DecodeGOB failed: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("Decoded value mismatch. Expected %+v, got %+v", original, decoded)
	}
}

func TestDecodeGOB_EmptyInput(t *testing.T) {
	var out TestStruct
	err := DecodeGOB([]byte{}, &out)
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
}

func TestEncodeGOB_UnsupportedType(t *testing.T) {
	_, err := EncodeGOB(func() {})
	if err == nil {
		t.Error("Expected error for unsupported gob type, got nil")
	}
}
