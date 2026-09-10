package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

const (
	AttributeHttpHost         = "http.host"
	AttributeHttpMethod       = "http.method"
	AttributeHttpPath         = "http.path"
	AttributeHttpQuery        = "http.query"
	AttributeHttpUserAgent    = "http.userAgent"
	AttributeHttpClientIP     = "http.clientIp"
	AttributeHttpStatusCode   = "http.statusCode"
	AttributeHttpStatusClass  = "http.statusClass"
	AttributeHttpBytesWritten = "http.bytesWritten"
	AttributeHttpRequestID    = "http.requestId"
	AttributeHttpLatencyMS    = "http.latencyMS"
	AttributeExceptionMessage = "exception.message"
	AttributeExceptionType    = "exception.type"
	AttributeExceptionStack   = "exception.stacktrace"
	AttributeExceptionCode    = "exception.code"
)

// ConvertToAttributes converts a plain map[string]interface{} into OTEL []attribute.KeyValue,
// supporting all common primitive types.
func ConvertToAttributes(attrMap map[string]interface{}) []attribute.KeyValue {
	if len(attrMap) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(attrMap))
	for k, v := range attrMap {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
	return attrs
}
