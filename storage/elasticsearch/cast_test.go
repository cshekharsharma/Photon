package elasticsearch

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/stretchr/testify/assert"
)

type elasticsearchErrCloser struct{}

func (elasticsearchErrCloser) Close() error {
	return errors.New("close failed")
}

func TestCloseBody_IgnoresCloseError(t *testing.T) {
	closeBody(elasticsearchErrCloser{})
}

// Test Cases for Search Response
func TestCastEsApiResponseToSearchResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		expectedHits  int64
	}{
		{
			name:          "Success case",
			mockResponse:  `{"hits":{"total":{"value":2,"relation":"eq"},"hits":[{"_index":"test","_id":"1","_score":1.0,"_source":{}},{"_index":"test","_id":"2","_score":1.0,"_source":{}}]}}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			expectedHits:  2,
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  `{"error"`,
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToSearchResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedHits, response.Hits.Total.Value)
			}
		})
	}
}

// Test Cases for Count Response
func TestCastEsApiResponseToCountResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		expectedCount int64
	}{
		{
			name:          "Success case",
			mockResponse:  `{"count":10,"_shards":{"total":2,"successful":2,"failed":0}}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			expectedCount: 10,
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToCountResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, response.Count)
			}
		})
	}
}

// Test Cases for Index Response
func TestCastEsApiResponseToIndexResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		expectedID    string
	}{
		{
			name:          "Success case",
			mockResponse:  `{"_index":"test","_id":"1","_version":1,"result":"created","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":1,"_primary_term":1}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			expectedID:    "1",
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToIndexResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, response.ID)
			}
		})
	}
}

// Test Cases for Update Response
func TestCastEsApiResponseToUpdateResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		expectedID    string
	}{
		{
			name:          "Success case",
			mockResponse:  `{"_index":"test","_id":"1","_version":2,"result":"updated","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":2,"_primary_term":1}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			expectedID:    "1",
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToUpdateResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, response.ID)
			}
		})
	}
}

// Test Cases for Delete Response
func TestCastEsApiResponseToDeleteResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		expectedID    string
	}{
		{
			name:          "Success case",
			mockResponse:  `{"_index":"test","_id":"1","_version":1,"result":"deleted","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":1,"_primary_term":1}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			expectedID:    "1",
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToDeleteResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, response.ID)
			}
		})
	}
}

// Test Cases for Bulk Response
func TestCastEsApiResponseToBulkResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		expectedItems int
	}{
		{
			name:          "Success case",
			mockResponse:  `{"took":3,"errors":false,"items":[{"index":{"_index":"test","_id":"1","_version":1,"status":201}},{"update":{"_index":"test","_id":"2","_version":2,"status":200}}]}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			expectedItems: 2,
		},
		{
			name:          "Error response",
			mockResponse:  `{"errors":true,"items":[{"index":{"error":{"type":"index_not_found_exception","reason":"no such index"}}}]}`,
			statusCode:    http.StatusBadRequest,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToBulkResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedItems, len(response.Items))
			}
		})
	}
}

// Test Cases for Update by Query Response
func TestCastEsApiResponseToUpdateByQueryResponse(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		statusCode      int
		expectedError   bool
		expectedUpdated int64
	}{
		{
			name:            "Success case",
			mockResponse:    `{"took":5,"timed_out":false,"total":10,"updated":2,"deleted":0,"batches":1,"version_conflicts":0}`,
			statusCode:      http.StatusOK,
			expectedError:   false,
			expectedUpdated: 2,
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToUpdateByQueryResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUpdated, response.Updated)
			}
		})
	}
}

// Test Cases for Delete by Query Response
func TestCastEsApiResponseToDeleteByQueryResponse(t *testing.T) {
	tests := []struct {
		name            string
		mockResponse    string
		statusCode      int
		expectedError   bool
		expectedDeleted int64
	}{
		{
			name:            "Success case",
			mockResponse:    `{"took":5,"timed_out":false,"total":10,"updated":0,"deleted":3,"batches":1,"version_conflicts":0}`,
			statusCode:      http.StatusOK,
			expectedError:   false,
			expectedDeleted: 3,
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "Parsing error",
			mockResponse:  "invalid json",
			statusCode:    http.StatusOK,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToDeleteByQueryResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDeleted, response.Deleted)
			}
		})
	}
}

// Test Cases for Search Response

func TestCastEsApiResponseToSearchByDocIdResponse(t *testing.T) {
	tests := []struct {
		name          string
		mockResponse  string
		statusCode    int
		expectedError bool
		docFound      bool
	}{
		{
			name:          "Success case",
			mockResponse:  `{"_index":"test","_id":"1","found":true,"_source":{}}`,
			statusCode:    http.StatusOK,
			expectedError: false,
			docFound:      true,
		},
		{
			name:          "Error response",
			mockResponse:  `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			statusCode:    http.StatusNotFound,
			expectedError: true,
			docFound:      false,
		},
		{
			name:          "Parsing error",
			mockResponse:  `{"error"`,
			statusCode:    http.StatusOK,
			expectedError: true,
			docFound:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &esapi.Response{
				StatusCode: tt.statusCode,
				Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
			}

			response, err := CastEsApiResponseToSearchByDocIdResponse(res)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.docFound, response.Found)
			}
		})
	}
}
