package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func TestConvertToAttributes_AllTypes(t *testing.T) {
	attrMap := map[string]interface{}{
		"str":   "hello",
		"int":   42,
		"int64": int64(64),
		"float": 3.14,
		"bool":  true,
	}

	attrs := ConvertToAttributes(attrMap)

	assert.Len(t, attrs, 5)

	attrMapFromResult := make(map[string]attribute.Value)
	for _, attr := range attrs {
		attrMapFromResult[string(attr.Key)] = attr.Value
	}

	assert.Equal(t, attribute.StringValue("hello"), attrMapFromResult["str"])
	assert.Equal(t, attribute.IntValue(42), attrMapFromResult["int"])
	assert.Equal(t, attribute.Int64Value(64), attrMapFromResult["int64"])
	assert.Equal(t, attribute.Float64Value(3.14), attrMapFromResult["float"])
	assert.Equal(t, attribute.BoolValue(true), attrMapFromResult["bool"])
}

func TestConvertToAttributes_EmptyInput(t *testing.T) {
	attrs := ConvertToAttributes(map[string]interface{}{})
	assert.Nil(t, attrs)
}

func TestConvertToAttributes_UnsupportedType(t *testing.T) {
	type custom struct{ val string }

	attrMap := map[string]interface{}{
		"custom": custom{"abc"},
	}

	attrs := ConvertToAttributes(attrMap)
	assert.Len(t, attrs, 1)
	assert.Equal(t, "custom", string(attrs[0].Key))
	assert.Equal(t, attribute.StringValue("{abc}"), attrs[0].Value)
}
