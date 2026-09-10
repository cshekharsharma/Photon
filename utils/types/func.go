package types

import (
	"reflect"
	"runtime"
)

// GetFunctionName takes a function pointer and returns its fully qualified name.
//
// It uses reflection to get the function pointer, then uses the runtime package
// to retrieve the function information, and finally returns the fully qualified
// function name.
//
// Parameters:
// - f: The function pointer.
//
// Returns:
// - string: The fully qualified name of the function.
func GetFunctionName(f interface{}) string {
	fnPtr := reflect.ValueOf(f).Pointer()
	fn := runtime.FuncForPC(fnPtr)

	if fn == nil {
		return "unknown"
	}

	return fn.Name()
}
