package apiresponse

import (
	"net/http"
	"testing"
)

func TestToHttpCode_KnownCodes(t *testing.T) {
	tests := []struct {
		name         string
		code         InternalResponseCode
		expectedCode int
	}{
		{"ALL_OK", AllOk, http.StatusOK},
		{"NO_CONTENT", NoContent, http.StatusNoContent},
		{"RESET_CONTENT", ResetContent, http.StatusResetContent},
		{"PARTIAL_CONTENT", PartialContent, http.StatusPartialContent},
		{"MOVED_PERMANENTLY", MovedPermanently, http.StatusMovedPermanently},
		{"FOUND", Found, http.StatusFound},
		{"REDIRECT_TEMPORARILY", RedirectTemporarily, http.StatusTemporaryRedirect},
		{"REDIRECT_PERMANENTLY", RedirectPermanently, http.StatusPermanentRedirect},
		{"BAD_REQUEST", BadRequest, http.StatusBadRequest},
		{"UNAUTHORIZED", Unauthorized, http.StatusUnauthorized},
		{"FORBIDDEN", Forbidden, http.StatusForbidden},
		{"NOT_FOUND", NotFound, http.StatusNotFound},
		{"REQUEST_TIMEOUT", RequestTimeout, http.StatusRequestTimeout},
		{"TOO_MANY_REQUEST", TooManyRequests, http.StatusTooManyRequests},
		{"SERVER_PANIC", InternalServerError, http.StatusInternalServerError},
		{"BAD_GATEWAY", BadGateway, http.StatusBadGateway},
		{"SERVICE_UNAVAILABLE", ServiceUnavailable, http.StatusServiceUnavailable},
		{"GATEWAY_TIMEOUT", GatewayTimeout, http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ToHttpCode(tt.code)
			if code != tt.expectedCode {
				t.Errorf("ToHttpCode(%s) = %d, expected %d", tt.code, code, tt.expectedCode)
			}
		})
	}
}

func TestToHttpCode_UnknownCode(t *testing.T) {
	unknownCode := InternalResponseCode("UNKNOWN")
	expected := http.StatusOK

	if code := ToHttpCode(unknownCode); code != expected {
		t.Errorf("ToHttpCode(UNKNOWN) = %d, expected %d", code, expected)
	}
}
