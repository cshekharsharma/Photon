// Package apiresponse contains all the data models that are expected to be
// used by the application for core workflows.
package apiresponse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/cshekharsharma/photon/utils/rest/minifier"
)

// Api response entity to standardize all HTTP api response
// and handle the pre/post hooks for response dispatch.
type ApiResponse struct {
	Success bool                 `json:"success"`
	Code    InternalResponseCode `json:"code"`
	Data    interface{}          `json:"data"`
	Message string               `json:"message"`
}

// Create new ApiResponse and return
func New(success bool, code InternalResponseCode, data interface{}, message string) *ApiResponse {
	if strings.TrimSpace(message) == "" {
		message = codesMap[code].message
	}

	r := &ApiResponse{
		Success: success,
		Code:    code,
		Data:    data,
		Message: message,
	}

	return r
}

// Get instance of ApiResponse as an struct
func (r *ApiResponse) GetStruct(success bool, code InternalResponseCode, data interface{}, message string) *ApiResponse {
	if strings.TrimSpace(message) == "" {
		message = codesMap[code].message
	}

	r.Success = success
	r.Code = code
	r.Data = data
	r.Message = message

	return r
}

// Convert ApiResponse struct to byte array
func (r *ApiResponse) ToByteArray() []byte {
	jsonResponse, _ := json.Marshal(r)
	return jsonResponse
}

// Send HTTP response over the network.
func (r *ApiResponse) Send(w http.ResponseWriter, httpCode int) {
	contentType := fmt.Sprintf("%s; charset=UTF-8", rest.ContentTypeJSON)
	w.Header().Set(rest.HeaderContentType, contentType)

	if httpCode == 0 {
		httpCode = ToHttpCode(r.Code)
	}
	w.WriteHeader(httpCode)

	minifyHeaderVal := w.Header().Get(rest.HeaderXApiMinifier)

	if minifier.IsApiKeyMinifierEnabled() && minifyHeaderVal == rest.XAPIMinifierValue {
		minified, err := minifier.MinifyApiKeys(r.Data)
		if err == nil {
			r.Data = minified
		}
	}

	if _, err := fmt.Fprintf(w, "%s", r.ToByteArray()); err != nil {
		return
	}
}
