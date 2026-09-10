package types

import (
	"fmt"
)

// ToBool converts an interface{} value to a bool.
func ToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		if v != 0 {
			return true, nil
		}
		return false, nil
	case int64:
		if v != 0 {
			return true, nil
		}
		return false, nil
	case string:
		// Convert common string representations of truth values.
		switch v {
		case "1", "true", "TRUE", "True":
			return true, nil
		case "0", "false", "FALSE", "False":
			return false, nil
		default:
			return false, fmt.Errorf("invalid string format for bool: %v", v)
		}
	default:
		return false, fmt.Errorf("unsupported type: %T", value)
	}
}
