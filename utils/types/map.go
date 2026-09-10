package types

import (
	"fmt"
	"reflect"
)

// ConvertToMapStringInterface converts a map with interface{} keys to a map with string keys.
func ConvertToMapStringInterface(m map[interface{}]interface{}) map[string]interface{} {
	converted := make(map[string]interface{})

	for k, v := range m {
		strKey := fmt.Sprintf("%v", k) // Convert the key to a string
		switch v := v.(type) {
		case map[interface{}]interface{}:
			converted[strKey] = ConvertToMapStringInterface(v)
		default:
			converted[strKey] = v
		}
	}

	return converted
}

// ConvertToMapInterface converts a map with string keys to a map with interface{} keys.
func ConvertToMapInterface(m map[string]interface{}) map[interface{}]interface{} {
	converted := make(map[interface{}]interface{})

	for k, v := range m {
		switch v := v.(type) {
		case map[string]interface{}:
			converted[k] = ConvertToMapInterface(v)
		default:
			converted[k] = v
		}
	}

	return converted
}

// MergeMaps merges two maps of the same type into one.
func MergeMaps[T any](map1, map2 map[string]T) map[string]T {
	merged := make(map[string]T)

	for k, v := range map1 {
		merged[k] = v
	}

	for k, v := range map2 {
		merged[k] = v
	}

	return merged
}

// CopyStructFieldsRecursive copies fields from one struct to another recursively.
func CopyStructFieldsRecursive(dst interface{}, src interface{}) {
	dstVal := reflect.ValueOf(dst)
	srcVal := reflect.ValueOf(src)

	if dstVal.Kind() != reflect.Pointer || srcVal.Kind() != reflect.Pointer {
		panic("both dst and src must be pointers to structs")
	}

	dstVal = dstVal.Elem()
	srcVal = srcVal.Elem()

	if dstVal.Kind() != reflect.Struct || srcVal.Kind() != reflect.Struct {
		panic("both dst and src must be pointers to structs")
	}

	dstType := dstVal.Type()

	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Field(i)
		dstFieldType := dstType.Field(i)

		if !dstField.CanSet() {
			continue
		}

		srcField := srcVal.FieldByName(dstFieldType.Name)
		if !srcField.IsValid() {
			continue
		}

		// If it's a struct, recurse
		if dstField.Kind() == reflect.Struct && srcField.Kind() == reflect.Struct {
			CopyStructFieldsRecursive(dstField.Addr().Interface(), srcField.Addr().Interface())
		} else if srcField.Type().AssignableTo(dstField.Type()) {
			dstField.Set(srcField)
		}
	}
}
