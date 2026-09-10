package apiresponse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/cshekharsharma/photon/utils/rest/minifier"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var mapping = &sync.Map{}
	minifier.SetupApiKeyMinifierConfig(false, mapping)
	data := map[string]string{"key": "value"}
	apiResp := New(true, AllOk, data, "")

	assert.Equal(t, true, apiResp.Success)
	assert.Equal(t, AllOk, apiResp.Code)
	assert.Equal(t, data, apiResp.Data)
	assert.Equal(t, http.StatusText(http.StatusOK), apiResp.Message)
}

func TestGetStruct(t *testing.T) {
	r := &ApiResponse{}
	updated := r.GetStruct(true, AllOk, "data", "All good")

	if updated.Message != "All good" {
		t.Errorf("Expected message 'All good', got '%s'", updated.Message)
	}
}

func TestGetStruct_DefaultMessageWhenEmpty(t *testing.T) {
	r := &ApiResponse{}
	updated := r.GetStruct(true, AllOk, "data", "   ")

	assert.Equal(t, http.StatusText(http.StatusOK), updated.Message)
}

func TestToByteArray(t *testing.T) {
	apiResp := &ApiResponse{
		Success: true,
		Code:    AllOk,
		Data:    "data",
		Message: "Success",
	}
	byteArray := apiResp.ToByteArray()
	expectedJSON, _ := json.Marshal(apiResp)

	if !bytes.Equal(byteArray, expectedJSON) {
		t.Errorf("Expected JSON %s, got %s", expectedJSON, byteArray)
	}
}

func TestSend(t *testing.T) {
	apiResp := &ApiResponse{
		Success: true,
		Code:    AllOk,
		Data:    map[string]string{"data": "mydata"},
		Message: "Success",
	}

	minifiedMap := &sync.Map{}
	minifiedMap.Store("data", "d")

	minifier.SetupApiKeyMinifierConfig(true, minifiedMap)

	_ = httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	w.Header().Set(rest.HeaderXApiMinifier, rest.XAPIMinifierValue)

	apiResp.Send(w, 0)

	resp := w.Result()
	body, _ := json.Marshal(apiResp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}

	responseBody, _ := json.Marshal(apiResp)
	if string(body) != string(responseBody) {
		t.Errorf("Expected body %s, got %s", responseBody, body)
	}

	expectedCT := fmt.Sprintf("%s; charset=UTF-8", rest.ContentTypeJSON)

	if contentType := resp.Header.Get(rest.HeaderContentType); contentType != expectedCT {
		t.Errorf("Expected Content-Type '%s', got '%s'", expectedCT, contentType)
	}
}

type apiResponseErrorWriter struct {
	header http.Header
}

func (w *apiResponseErrorWriter) Header() http.Header {
	return w.header
}

func (w *apiResponseErrorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *apiResponseErrorWriter) WriteHeader(int) {}

func TestSend_WriteError(t *testing.T) {
	apiResp := New(true, AllOk, map[string]string{"data": "value"}, "")
	apiResp.Send(&apiResponseErrorWriter{header: make(http.Header)}, http.StatusOK)
}
