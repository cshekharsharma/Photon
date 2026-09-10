package types

import "reflect"

func IsEmpty(value interface{}) bool {
	switch value := value.(type) {
	case string:
		return value == ""
	case nil:
		return true
	default:
		// Check for other types
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			return v.IsNil()
		case reflect.Map:
			return v.IsNil() || v.Len() == 0 // Handle maps
		case reflect.Array, reflect.Slice:
			return v.Len() == 0
		default:
			return reflect.DeepEqual(value, reflect.Zero(v.Type()).Interface())
		}
	}
}
