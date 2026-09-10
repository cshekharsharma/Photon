package types

import (
	"strings"
	"testing"
)

func exampleFunction() {
	// This is intentionally left empty.
}

func TestGetFunctionName(t *testing.T) {
	tests := []struct {
		funcPtr  interface{}
		expected string
	}{
		{exampleFunction, "github.com/cshekharsharma/photon/utils/types.exampleFunction"},
		{TestGetFunctionName, "github.com/cshekharsharma/photon/utils/types.TestGetFunctionName"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			functionName := GetFunctionName(tt.funcPtr)
			if !strings.Contains(functionName, tt.expected) {
				t.Errorf("GetFunctionName() = %v; want it to contain %v", functionName, tt.expected)
			}
		})
	}
}

func TestGetFunctionName_NilFunc(t *testing.T) {
	var fn func()
	name := GetFunctionName(fn)
	if name != "unknown" {
		t.Fatalf("expected unknown, got %s", name)
	}
}
