package elasticsearch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/stretchr/testify/assert"
)

// MockTransport is a mock implementation of http.RoundTripper
type mockTransport struct{}

func (m mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path
	method := req.Method

	headers := http.Header{
		"X-Elastic-Product": []string{"Elasticsearch"},
		"Content-Type":      []string{"application/json"},
	}

	// Handle Scroll API request (POST /_search/scroll)
	if method == http.MethodPost && strings.Contains(path, "_search/scroll") {
		if strings.Contains(req.URL.RawQuery, "scroll_id=force-400") {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Header:     headers,
				Body: io.NopCloser(strings.NewReader(`{
                "error": "scroll request failed"
            }`)),
			}, nil
		}

		if strings.Contains(req.URL.RawQuery, "scroll-id-page-1") {
			return &http.Response{
				StatusCode: 200,
				Header:     headers,
				Body: io.NopCloser(strings.NewReader(`{
					"_scroll_id": "scroll-id-page-2",
					"hits": {
						"total": {"value": 3},
						"hits": [
							{"_id": "1", "_source": {"field": "value1"}},
							{"_id": "2", "_source": {"field": "value2"}}
						]
					}
				}`)),
			}, nil
		}

		if strings.Contains(req.URL.RawQuery, "scroll-id-page-2") {
			return &http.Response{
				StatusCode: 200,
				Header:     headers,
				Body: io.NopCloser(strings.NewReader(`{
					"_scroll_id": "scroll-id-page-3",
					"hits": {
						"total": {"value": 3},
						"hits": [
							{"_id": "3", "_source": {"field": "value3"}}
						]
					}
				}`)),
			}, nil
		}

		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
            "hits": {
                "total": {"value": 1},
                "hits": [{"_id": "1", "_source": {"field": "value"}}]
            },
            "_scroll_id": "next-scroll"
        }`)),
		}, nil
	}

	// Handle Search (POST /index/_search)
	if method == http.MethodPost && strings.Contains(path, "_search") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"hits": {
					"total": {"value": 1},
					"hits": [{"_id": "1", "_source": {"field": "value"}}]
				}
			}`)),
		}, nil
	}

	// Handle Get by Doc ID (GET /index/_doc/{id})
	if method == http.MethodGet && strings.Contains(path, "_doc") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"_index": "my-index",
				"_id": "1",
				"_source": {"field": "value"}
			}`)),
		}, nil
	}

	// Handle Index (PUT /index/_doc/{id})
	if method == http.MethodPut && strings.Contains(path, "_doc") {
		return &http.Response{
			StatusCode: 201,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"result": "created"
			}`)),
		}, nil
	}

	// Handle Update (POST /index/_update/{id})
	if method == http.MethodPost && strings.Contains(path, "_update") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"result": "updated"
			}`)),
		}, nil
	}

	// Handle Delete (DELETE /index/_doc/{id})
	if method == http.MethodDelete && strings.Contains(path, "_doc") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"result": "deleted"
			}`)),
		}, nil
	}

	// Handle Count (POST /index/_count)
	if method == http.MethodPost && strings.Contains(path, "_count") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"count": 1,
				"_shards": {
					"total": 1,
					"successful": 1,
					"skipped": 0,
					"failed": 0
				}
			}`)),
		}, nil
	}

	// Handle UpdateByQuery (POST /index/_update_by_query)
	if method == http.MethodPost && strings.Contains(path, "_update_by_query") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"updated": 1
			}`)),
		}, nil
	}

	// Handle DeleteByQuery (POST /index/_delete_by_query)
	if method == http.MethodPost && strings.Contains(path, "_delete_by_query") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"deleted": 1
			}`)),
		}, nil
	}

	// Handle Bulk (POST /_bulk)
	if method == http.MethodPost && strings.HasSuffix(path, "/_bulk") {
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
				"errors": false,
				"items": []
			}`)),
		}, nil
	}

	// Handle Scroll API request (POST /_scroll)
	if method == http.MethodPost && strings.Contains(path, "_scroll") {
		bodyBytes, _ := io.ReadAll(req.Body)
		bodyStr := string(bodyBytes)

		// Simulate 400 error for specific scroll ID
		if strings.Contains(bodyStr, "force-400") {
			return &http.Response{
				StatusCode: 400,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		}

		// Simulate valid scroll response
		return &http.Response{
			StatusCode: 200,
			Header:     headers,
			Body: io.NopCloser(strings.NewReader(`{
			"_scroll_id": "next-scroll",
			"hits": {
				"hits": [{"_id": "1", "_source": {"field": "value"}}]
			}
		}`)),
		}, nil
	}

	// Default fallback
	return &http.Response{
		StatusCode: 404,
		Header:     headers,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil
}

func init() {
	SetConnectionConfig("test-cluster", &ConnectionConfig{
		Addresses:  []string{"http://localhost:9200"},
		Username:   "user",
		Password:   "pass",
		MaxRetries: 1,
		Transport:  mockTransport{},
	})
}

func setupTestClusterConfig() {
	SetConnectionConfig("test-cluster", &ConnectionConfig{
		Addresses:  []string{"http://localhost:9200"},
		Username:   "user",
		Password:   "pass",
		MaxRetries: 1,
		Transport:  mockTransport{},
	})
}

func TestSearch(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	resp, err := es.Search(context.Background(), "test-cluster", "test-index", Query{
		Query: BoolQuery{
			Bool: Bool{
				Must: []interface{}{
					map[string]interface{}{"match_all": map[string]interface{}{}},
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestSearchByDocId(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	resp, err := es.SearchByDocId(context.Background(), "test-cluster", "test-index", "123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIndex(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	doc := map[string]string{"name": "test"}
	resp, err := es.Index(context.Background(), "test-cluster", "test-index", "123", doc)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestUpdate(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	doc := map[string]string{"name": "updated"}
	resp, err := es.Update(context.Background(), "test-cluster", "test-index", "123", doc)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDelete(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	resp, err := es.Delete(context.Background(), "test-cluster", "test-index", "123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestCount(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	resp, err := es.Count(context.Background(), "test-cluster", "test-index")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Count)
}

func TestCountByQuery(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}

	resp, err := es.CountByQuery(context.Background(), "test-cluster", "test-index", Query{
		Query: BoolQuery{
			Bool: Bool{
				Must: []interface{}{
					map[string]interface{}{"match_all": map[string]interface{}{}},
				},
			},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Count)
}

func TestUpdateByQuery(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	query := map[string]interface{}{"script": "ctx._source.field = 'new_value'"}
	resp, err := es.UpdateByQuery(context.Background(), "test-cluster", "test-index", query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestDeleteByQuery(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	query := map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}}
	resp, err := es.DeleteByQuery(context.Background(), "test-cluster", "test-index", query)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestRawSearchQuery(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	resp, err := es.RawSearchQuery(context.Background(), "test-cluster", "test-index", `{"query":{"match_all":{}}}`)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestRawCountQuery(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	resp, err := es.RawCountQuery(context.Background(), "test-cluster", "test-index", `{"query":{"match_all":{}}}`)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp.Count)
}

func TestBulk(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	ops := []BulkOperation{
		{Action: "index", Index: "test-index", ID: strPtr("1"), Document: map[string]interface{}{"name": "test"}},
	}
	resp, err := es.Bulk(context.Background(), "test-cluster", ops)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func strPtr(s string) *string {
	return &s
}

func TestIncrementalSearch(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}

	t.Run("invalid from/size with FetchAllDocs=false", func(t *testing.T) {
		_, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			FetchAllDocs: false,
			From:         -1,
			Size:         10,
			Query:        map[string]interface{}{},
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid parameters")
	})

	t.Run("nil query should return error", func(t *testing.T) {
		_, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			FetchAllDocs: false,
			From:         0,
			Size:         10,
			Query:        nil,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query cannot be nil")
	})

	t.Run("FetchAllDocs forces From and Size to 0 and uses bulk logic", func(t *testing.T) {
		resp, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			FetchAllDocs: true,
			From:         10,
			Size:         20,
			Query:        map[string]interface{}{},
			Sort:         []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Hits.Total.Value)
	})

	t.Run("From+Size exceeds MaxFetchSize -> bulk logic", func(t *testing.T) {
		resp, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			FetchAllDocs: false,
			From:         30000,
			Size:         100,
			Query:        map[string]interface{}{},
			Sort:         []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.Hits.Total.Value, int64(0))
	})

	t.Run("Within max fetch size, sorted query -> regular search path", func(t *testing.T) {
		resp, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			FetchAllDocs: false,
			From:         0,
			Size:         10,
			Query:        map[string]interface{}{},
			Sort:         []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Hits.Total.Value)
	})

	t.Run("Unsorted query with small size uses scroll logic", func(t *testing.T) {
		resp, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			FetchAllDocs: false,
			From:         0,
			Size:         10,
			Query:        map[string]interface{}{},
			Sort:         []map[string]interface{}{},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Hits.Total.Value)
	})
}

func Test_getDocumentsWithScrollApi(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}

	t.Run("Valid Scroll API Flow - FetchAllDocs false", func(t *testing.T) {
		resp, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			From:         0,
			Size:         1,
			FetchAllDocs: false,
			Query: map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			Sort: nil, // This ensures getDocumentsWithScrollApi is used
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Hits.Total.Value)
	})

	t.Run("FetchAllDocs true", func(t *testing.T) {
		resp, err := es.IncrementalSearch(context.Background(), "test-cluster", "test-index", SearchQuery{
			From:         0,
			Size:         0,
			FetchAllDocs: true,
			Query: map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Hits.Total.Value)
	})

	t.Run("Client failure case", func(t *testing.T) {
		SetConnectionConfig("broken-cluster", &ConnectionConfig{
			Addresses: []string{"http://invalidhost:9200000"},
			Transport: mockTransport{},
		})

		delete(connectionConfigMap, "broken-cluster")
		_, err := es.getDocumentsWithScrollApi(context.Background(), "broken-cluster", "test-index", SearchQuery{
			From:         0,
			Size:         1,
			FetchAllDocs: false,
			Query: map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		})
		assert.Error(t, err)
	})
}

func Test_initiateScrollApiSearchRequest(t *testing.T) {
	es := ElasticSearchHelper{}
	client, _ := getClient("test-cluster")

	t.Run("Valid request returns 200", func(t *testing.T) {
		query := map[string]interface{}{
			"query": map[string]interface{}{"match_all": map[string]interface{}{}},
		}
		resp, err := es.initiateScrollApiSearchRequest(context.Background(), client, "test-index", query, 1*time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Marshal error", func(t *testing.T) {
		query := map[string]interface{}{
			"bad": func() {}, // functions can't be marshaled
		}
		resp, err := es.initiateScrollApiSearchRequest(context.Background(), client, "test-index", query, 1*time.Second)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func Test_getDocumentsWithScrollSession(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}

	client, err := getClient("test-cluster")
	assert.NoError(t, err)

	t.Run("ValidScrollWithFetchAllDocs", func(t *testing.T) {
		searchQuery := SearchQuery{
			FetchAllDocs: true,
			Query: map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		scrollID := "initial-scroll"
		scrollTimeout := 1 * time.Minute
		fetchSize := int64(10)

		docs, sid, err := es.getDocumentsWithScrollSession(context.Background(), client, searchQuery, scrollID, scrollTimeout, fetchSize)
		assert.NoError(t, err)
		assert.NotEmpty(t, docs)
		assert.Equal(t, "next-scroll", sid)
	})

	t.Run("BadRequestFromES", func(t *testing.T) {
		searchQuery := SearchQuery{
			FetchAllDocs: true,
			Query: map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		scrollID := "force-400"
		scrollTimeout := 1 * time.Minute
		fetchSize := int64(10)

		docs, sid, err := es.getDocumentsWithScrollSession(context.Background(), client, searchQuery, scrollID, scrollTimeout, fetchSize)
		assert.NoError(t, err) // scroll API 400 is treated as non-fatal in your logic
		assert.Empty(t, docs)
		assert.Equal(t, "force-400", sid)
	})

	t.Run("EmptyScrollID", func(t *testing.T) {
		searchQuery := SearchQuery{
			FetchAllDocs: true,
			Query: map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}

		scrollID := ""
		scrollTimeout := 1 * time.Minute
		fetchSize := int64(10)

		docs, sid, err := es.getDocumentsWithScrollSession(context.Background(), client, searchQuery, scrollID, scrollTimeout, fetchSize)
		assert.NoError(t, err)
		assert.Empty(t, docs)
		assert.Equal(t, "", sid)
	})

	t.Run("PaginationLogic", func(t *testing.T) {
		client, err := es8.NewTypedClient(es8.Config{
			Transport: mockTransport{},
		})
		assert.NoError(t, err)

		helper := &ElasticSearchHelper{}
		query := SearchQuery{
			From:         1,
			Size:         2,
			FetchAllDocs: false,
		}

		scrollID := "scroll-id-page-1"
		docs, sid, err := helper.getDocumentsWithScrollSession(context.Background(), client, query, scrollID, 5*time.Minute, 2)

		assert.NoError(t, err)
		assert.Len(t, docs, 2) // Expected: doc[1] from page-1 + doc[0] from page-2
		assert.Equal(t, "scroll-id-page-2", sid)
	})

}

// ---------------- Exec error test cases ---------------------------

type errorTransport struct {
	status int
	body   string
	path   string
}

type execErrorTransport struct {
	targetPath string
}

func (e errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if e.path != "" && !strings.Contains(req.URL.Path, e.path) {
		return &http.Response{
			StatusCode: 404,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	}
	return &http.Response{
		StatusCode: e.status,
		Header:     http.Header{"Content-Type": []string{"application/json"}, "X-Elastic-Product": []string{"Elasticsearch"}},
		Body:       io.NopCloser(strings.NewReader(e.body)),
	}, nil
}

func TestSearch_MarshalError(t *testing.T) {
	SetConnectionConfig("bad-search", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: mockTransport{},
	})
	es := ElasticSearchHelper{}
	_, err := es.Search(context.Background(), "bad-search", "idx", Query{
		Query: BoolQuery{
			Bool: Bool{
				Must: []interface{}{make(chan int)},
			},
		},
	})
	assert.Error(t, err)
}

func TestSearch_StatusBadRequest(t *testing.T) {
	SetConnectionConfig("bad-search-status", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusBadRequest, body: `{}`, path: "_search"},
	})
	es := ElasticSearchHelper{}
	resp, err := es.Search(context.Background(), "bad-search-status", "idx", Query{
		Query: BoolQuery{Bool: Bool{Must: []interface{}{map[string]interface{}{"match_all": map[string]interface{}{}}}}},
	})
	assert.NoError(t, err)
	assert.Equal(t, SearchResponse{}, resp)
}

func TestSearchByDocId_NotFound(t *testing.T) {
	SetConnectionConfig("not-found", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusNotFound, body: `{}`, path: "_doc"},
	})
	es := ElasticSearchHelper{}
	resp, err := es.SearchByDocId(context.Background(), "not-found", "idx", "1")
	assert.NoError(t, err)
	assert.Equal(t, SearchByDocIdResponse{}, resp)
}

func TestSearchByDocId_CastError(t *testing.T) {
	SetConnectionConfig("bad-doc-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{"_index":`, path: "_doc"},
	})
	es := ElasticSearchHelper{}
	_, err := es.SearchByDocId(context.Background(), "bad-doc-cast", "idx", "1")
	assert.Error(t, err)
}

func TestIndex_MarshalError(t *testing.T) {
	SetConnectionConfig("bad-index", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: mockTransport{},
	})
	es := ElasticSearchHelper{}
	_, err := es.Index(context.Background(), "bad-index", "idx", "1", map[string]interface{}{"x": make(chan int)})
	assert.Error(t, err)
}

func TestIndex_CastError(t *testing.T) {
	SetConnectionConfig("bad-index-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_doc"},
	})
	es := ElasticSearchHelper{}
	_, err := es.Index(context.Background(), "bad-index-cast", "idx", "1", map[string]interface{}{"x": "y"})
	assert.Error(t, err)
}

func TestUpdate_CastError(t *testing.T) {
	SetConnectionConfig("bad-update-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_update"},
	})
	es := ElasticSearchHelper{}
	_, err := es.Update(context.Background(), "bad-update-cast", "idx", "1", map[string]interface{}{"x": "y"})
	assert.Error(t, err)
}

func TestDelete_CastError(t *testing.T) {
	SetConnectionConfig("bad-delete-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_doc"},
	})
	es := ElasticSearchHelper{}
	_, err := es.Delete(context.Background(), "bad-delete-cast", "idx", "1")
	assert.Error(t, err)
}

func TestCount_CastError(t *testing.T) {
	SetConnectionConfig("bad-count-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_count"},
	})
	es := ElasticSearchHelper{}
	_, err := es.Count(context.Background(), "bad-count-cast", "idx")
	assert.Error(t, err)
}

func TestUpdateByQuery_CastError(t *testing.T) {
	SetConnectionConfig("bad-update-by-query-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_update_by_query"},
	})
	es := ElasticSearchHelper{}
	_, err := es.UpdateByQuery(context.Background(), "bad-update-by-query-cast", "idx", map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}})
	assert.Error(t, err)
}

func TestDeleteByQuery_CastError(t *testing.T) {
	SetConnectionConfig("bad-delete-by-query-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_delete_by_query"},
	})
	es := ElasticSearchHelper{}
	_, err := es.DeleteByQuery(context.Background(), "bad-delete-by-query-cast", "idx", map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}})
	assert.Error(t, err)
}

func TestBulk_CastError(t *testing.T) {
	SetConnectionConfig("bad-bulk-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_bulk"},
	})
	es := ElasticSearchHelper{}
	ops := []BulkOperation{
		{Action: "index", Index: "idx", ID: strPtr("1"), Document: map[string]interface{}{"x": "y"}},
	}
	_, err := es.Bulk(context.Background(), "bad-bulk-cast", ops)
	assert.Error(t, err)
}

func TestUpdateByQuery_MarshalError(t *testing.T) {
	SetConnectionConfig("bad-update-by-query", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: mockTransport{},
	})
	es := ElasticSearchHelper{}
	_, err := es.UpdateByQuery(context.Background(), "bad-update-by-query", "idx", map[string]interface{}{"q": make(chan int)})
	assert.Error(t, err)
}

func TestDeleteByQuery_MarshalError(t *testing.T) {
	SetConnectionConfig("bad-delete-by-query", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: mockTransport{},
	})
	es := ElasticSearchHelper{}
	_, err := es.DeleteByQuery(context.Background(), "bad-delete-by-query", "idx", map[string]interface{}{"q": make(chan int)})
	assert.Error(t, err)
}

func TestCountByQuery_MarshalError(t *testing.T) {
	SetConnectionConfig("bad-count-by-query", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: mockTransport{},
	})
	es := ElasticSearchHelper{}
	_, err := es.CountByQuery(context.Background(), "bad-count-by-query", "idx", Query{
		Query: BoolQuery{Bool: Bool{Must: []interface{}{make(chan int)}}},
	})
	assert.Error(t, err)
}

func TestBulk_MarshalError(t *testing.T) {
	SetConnectionConfig("bad-bulk", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: mockTransport{},
	})
	es := ElasticSearchHelper{}
	ops := []BulkOperation{
		{Action: "index", Index: "idx", ID: strPtr("1"), Document: map[string]interface{}{"bad": make(chan int)}},
	}
	_, err := es.Bulk(context.Background(), "bad-bulk", ops)
	assert.Error(t, err)
}

func TestConnect_MissingConfig(t *testing.T) {
	instances = nil
	_, err := Connect("missing-cluster")
	assert.Error(t, err)
}

func TestGetClientErrors(t *testing.T) {
	es := ElasticSearchHelper{}
	cluster := "missing-cluster-2"

	_, err := es.Search(context.Background(), cluster, "idx", Query{Query: BoolQuery{Bool: Bool{Must: []interface{}{map[string]interface{}{"match_all": map[string]interface{}{}}}}}})
	assert.Error(t, err)

	_, err = es.SearchByDocId(context.Background(), cluster, "idx", "1")
	assert.Error(t, err)

	_, err = es.Index(context.Background(), cluster, "idx", "1", map[string]interface{}{"x": "y"})
	assert.Error(t, err)

	_, err = es.Update(context.Background(), cluster, "idx", "1", map[string]interface{}{"x": "y"})
	assert.Error(t, err)

	_, err = es.Delete(context.Background(), cluster, "idx", "1")
	assert.Error(t, err)

	_, err = es.Count(context.Background(), cluster, "idx")
	assert.Error(t, err)

	_, err = es.CountByQuery(context.Background(), cluster, "idx", Query{Query: BoolQuery{Bool: Bool{Must: []interface{}{map[string]interface{}{"match_all": map[string]interface{}{}}}}}})
	assert.Error(t, err)

	_, err = es.UpdateByQuery(context.Background(), cluster, "idx", map[string]interface{}{"script": "x"})
	assert.Error(t, err)

	_, err = es.DeleteByQuery(context.Background(), cluster, "idx", map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}})
	assert.Error(t, err)

	_, err = es.Bulk(context.Background(), cluster, []BulkOperation{{Action: "index", Index: "idx", ID: strPtr("1"), Document: map[string]interface{}{"x": "y"}}})
	assert.Error(t, err)
}

func (e execErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, e.targetPath) {
		return nil, errors.New("transport error")
	}

	headers := http.Header{
		"X-Elastic-Product": []string{"Elasticsearch"},
		"Content-Type":      []string{"application/json"},
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     headers,
		Body: io.NopCloser(strings.NewReader(`{
			"hits": {"total": {"value": 0}, "hits": []},
			"_scroll_id": "sid",
			"count": 0,
			"updated": 0,
			"deleted": 0,
			"errors": false,
			"items": []
		}`)),
	}, nil
}

func setExecErrCluster(cluster, path string) {
	SetConnectionConfig(cluster, &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: execErrorTransport{targetPath: path},
	})
}

func TestDoErrorBranches(t *testing.T) {
	es := ElasticSearchHelper{}

	setExecErrCluster("es-search-do-error", "_search")
	_, err := es.Search(context.Background(), "es-search-do-error", "idx", Query{
		Query: BoolQuery{Bool: Bool{Must: []interface{}{map[string]interface{}{"match_all": map[string]interface{}{}}}}},
	})
	assert.Error(t, err)

	setExecErrCluster("es-doc-do-error", "_doc")
	_, err = es.SearchByDocId(context.Background(), "es-doc-do-error", "idx", "1")
	assert.Error(t, err)

	setExecErrCluster("es-index-do-error", "_doc")
	_, err = es.Index(context.Background(), "es-index-do-error", "idx", "1", map[string]interface{}{"a": "b"})
	assert.Error(t, err)

	setExecErrCluster("es-update-do-error", "_update")
	_, err = es.Update(context.Background(), "es-update-do-error", "idx", "1", map[string]interface{}{"a": "b"})
	assert.Error(t, err)

	setExecErrCluster("es-delete-do-error", "_doc")
	_, err = es.Delete(context.Background(), "es-delete-do-error", "idx", "1")
	assert.Error(t, err)

	setExecErrCluster("es-count-do-error", "_count")
	_, err = es.Count(context.Background(), "es-count-do-error", "idx")
	assert.Error(t, err)

	setExecErrCluster("es-count-query-do-error", "_count")
	_, err = es.RawCountQuery(context.Background(), "es-count-query-do-error", "idx", `{"query":{"match_all":{}}}`)
	assert.Error(t, err)

	setExecErrCluster("es-update-query-do-error", "_update_by_query")
	_, err = es.UpdateByQuery(context.Background(), "es-update-query-do-error", "idx", map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}})
	assert.Error(t, err)

	setExecErrCluster("es-delete-query-do-error", "_delete_by_query")
	_, err = es.DeleteByQuery(context.Background(), "es-delete-query-do-error", "idx", map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}})
	assert.Error(t, err)

	setExecErrCluster("es-bulk-do-error", "_bulk")
	_, err = es.Bulk(context.Background(), "es-bulk-do-error", []BulkOperation{{Action: "index", Index: "idx", ID: strPtr("1"), Document: map[string]interface{}{"a": "b"}}})
	assert.Error(t, err)
}

func TestSearchRequestWithQueryAndInitiate_DoErrors(t *testing.T) {
	es := ElasticSearchHelper{}

	_, err := es.searchRequestWithQuery(context.Background(), "test-cluster", "idx", map[string]interface{}{
		"bad": func() {},
	})
	assert.Error(t, err)

	setExecErrCluster("es-init-scroll-do-error", "_search")
	client, err := getClient("es-init-scroll-do-error")
	assert.NoError(t, err)

	_, err = es.initiateScrollApiSearchRequest(context.Background(), client, "idx", map[string]interface{}{"query": map[string]interface{}{"match_all": map[string]interface{}{}}}, time.Second)
	assert.Error(t, err)

	_, err = es.initiateScrollApiSearchRequest(context.Background(), client, "idx", map[string]interface{}{"bad": func() {}}, time.Second)
	assert.Error(t, err)
}

func TestGetDocumentsWithScrollSession_DoError(t *testing.T) {
	setExecErrCluster("es-scroll-session-do-error", "_search/scroll")
	es := ElasticSearchHelper{}
	client, err := getClient("es-scroll-session-do-error")
	assert.NoError(t, err)

	_, _, err = es.getDocumentsWithScrollSession(context.Background(), client, SearchQuery{
		From:         0,
		Size:         10,
		FetchAllDocs: true,
		Query:        map[string]interface{}{"match_all": map[string]interface{}{}},
	}, "scroll-id", time.Second, 10)
	assert.Error(t, err)
}

func TestGetBulkDocuments_ErrorBranch(t *testing.T) {
	es := ElasticSearchHelper{}
	_, err := es.getBulkDocuments(context.Background(), "test-cluster", "idx", SearchQuery{
		FetchAllDocs: false,
		From:         0,
		Size:         10,
		// Force a marshal error in searchRequestWithQuery path.
		Query: map[string]interface{}{"bad": func() {}},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.Error(t, err)
}

func TestGetDocumentsWithSearchAfter_Branches(t *testing.T) {
	es := ElasticSearchHelper{}

	_, err := es.getDocumentsWithSearchAfter(context.Background(), "test-cluster", "idx", SearchQuery{
		Query: map[string]interface{}{},
		Sort:  nil,
	})
	assert.Error(t, err)

	_, err = es.getDocumentsWithSearchAfter(context.Background(), "test-cluster", "idx", SearchQuery{
		From:  EsMaxFetchSize,
		Size:  10,
		Query: map[string]interface{}{"_source": "bad-type"},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.Error(t, err)
}

func TestJumpToActiveBatchAndScrollApi_Branches(t *testing.T) {
	es := ElasticSearchHelper{}
	setupTestClusterConfig()

	jumped, err := es.jumpToActiveBatchOfDocuments(context.Background(), "test-cluster", "idx", SearchQuery{
		From:  EsMaxFetchSize,
		Size:  1,
		Query: map[string]interface{}{"_source": []string{"field"}},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.NoError(t, err)
	assert.NotNil(t, jumped.query)

	setExecErrCluster("es-scroll-api-do-error", "_search")
	_, err = es.getDocumentsWithScrollApi(context.Background(), "es-scroll-api-do-error", "idx", SearchQuery{
		From:         0,
		Size:         10,
		FetchAllDocs: false,
		Query:        map[string]interface{}{"match_all": map[string]interface{}{}},
	})
	assert.Error(t, err)
}

func TestGetDocumentsWithScrollSession_NoFetchAllDocsBranch(t *testing.T) {
	client, err := es8.NewTypedClient(es8.Config{
		Transport: mockTransport{},
	})
	assert.NoError(t, err)

	es := ElasticSearchHelper{}
	docs, _, err := es.getDocumentsWithScrollSession(context.Background(), client, SearchQuery{
		From:         1,
		Size:         1,
		FetchAllDocs: false,
		Query:        map[string]interface{}{"match_all": map[string]interface{}{}},
	}, "scroll-id-page-1", time.Minute, 2)

	assert.NoError(t, err)
	assert.Len(t, docs, 1)
}

func TestSearch_CastError(t *testing.T) {
	SetConnectionConfig("bad-search-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_search"},
	})
	es := ElasticSearchHelper{}
	_, err := es.Search(context.Background(), "bad-search-cast", "idx", Query{
		Query: BoolQuery{Bool: Bool{Must: []interface{}{map[string]interface{}{"match_all": map[string]interface{}{}}}}},
	})
	assert.Error(t, err)
}

func TestUpdate_MarshalError(t *testing.T) {
	setupTestClusterConfig()
	es := ElasticSearchHelper{}
	_, err := es.Update(context.Background(), "test-cluster", "idx", "1", map[string]interface{}{"bad": make(chan int)})
	assert.Error(t, err)
}

func TestRawCountQuery_CastError(t *testing.T) {
	SetConnectionConfig("bad-raw-count-cast", &ConnectionConfig{
		Addresses: []string{"http://localhost:9200"},
		Transport: errorTransport{status: http.StatusOK, body: `{`, path: "_count"},
	})
	es := ElasticSearchHelper{}
	_, err := es.RawCountQuery(context.Background(), "bad-raw-count-cast", "idx", `{"query":{"match_all":{}}}`)
	assert.Error(t, err)
}

func TestBulk_MetadataMarshalError(t *testing.T) {
	setupTestClusterConfig()
	orig := marshalBulkMetaHook
	defer func() { marshalBulkMetaHook = orig }()
	marshalBulkMetaHook = func(v interface{}) ([]byte, error) {
		return nil, errors.New("meta marshal fail")
	}

	es := ElasticSearchHelper{}
	_, err := es.Bulk(context.Background(), "test-cluster", []BulkOperation{{Action: ActionDelete, Index: "idx"}})
	assert.Error(t, err)
}

func TestGetDocumentsWithScrollApi_FromBeyondAvailableDocs(t *testing.T) {
	es := &ElasticSearchHelper{}
	origGetClient := getClientHook
	origInitiate := initiateScrollApiSearchRequestHook
	origCast := castSearchResponseHook
	origClear := clearScrollRequestHook
	defer func() {
		getClientHook = origGetClient
		initiateScrollApiSearchRequestHook = origInitiate
		castSearchResponseHook = origCast
		clearScrollRequestHook = origClear
	}()

	getClientHook = func(clusterName string) (*es8.TypedClient, error) {
		return &es8.TypedClient{}, nil
	}
	initiateScrollApiSearchRequestHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, index string, query map[string]interface{}, scrollTimeout time.Duration) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return SearchResponse{
			ScrollID: "sid",
			Hits: Hits{
				Hits: []Hit{{ID: "1"}},
			},
		}, nil
	}
	clearScrollRequestHook = func(ctx context.Context, req esapi.ClearScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}

	docs, err := es.getDocumentsWithScrollApi(context.Background(), "c", "i", SearchQuery{
		From:  5,
		Size:  2,
		Query: map[string]interface{}{},
	})
	assert.NoError(t, err)
	assert.Empty(t, docs)
}

func TestGetDocumentsWithScrollApi_UsesDefaultScrollSessionHook(t *testing.T) {
	es := &ElasticSearchHelper{}
	origGetClient := getClientHook
	origInitiate := initiateScrollApiSearchRequestHook
	origCast := castSearchResponseHook
	origClear := clearScrollRequestHook
	origScrollDo := scrollRequestDoHook
	defer func() {
		getClientHook = origGetClient
		initiateScrollApiSearchRequestHook = origInitiate
		castSearchResponseHook = origCast
		clearScrollRequestHook = origClear
		scrollRequestDoHook = origScrollDo
	}()

	getClientHook = func(clusterName string) (*es8.TypedClient, error) {
		return &es8.TypedClient{}, nil
	}
	initiateScrollApiSearchRequestHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, index string, query map[string]interface{}, scrollTimeout time.Duration) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return SearchResponse{
			ScrollID: "sid",
			Hits: Hits{
				Hits: []Hit{{ID: "1"}, {ID: "2"}},
			},
		}, nil
	}
	usedScrollDo := false
	scrollRequestDoHook = func(ctx context.Context, req esapi.ScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		usedScrollDo = true
		return &esapi.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	clearScrollRequestHook = func(ctx context.Context, req esapi.ClearScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}

	_, err := es.getDocumentsWithScrollApi(context.Background(), "c", "i", SearchQuery{
		From:  0,
		Size:  2,
		Query: map[string]interface{}{},
	})
	assert.NoError(t, err)
	assert.True(t, usedScrollDo)
}

func TestGetDocumentsWithScrollSession_ShortPageBreak(t *testing.T) {
	es := &ElasticSearchHelper{}
	origDo := scrollRequestDoHook
	origCast := castSearchResponseHook
	defer func() {
		scrollRequestDoHook = origDo
		castSearchResponseHook = origCast
	}()

	scrollRequestDoHook = func(ctx context.Context, req esapi.ScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return SearchResponse{
			ScrollID: "sid2",
			Hits: Hits{
				Hits: []Hit{{ID: "1"}, {ID: "2"}},
			},
		}, nil
	}

	docs, sid, err := es.getDocumentsWithScrollSession(context.Background(), &es8.TypedClient{}, SearchQuery{
		From:  1,
		Size:  4,
		Query: map[string]interface{}{},
	}, "sid", time.Second, 5)
	assert.NoError(t, err)
	assert.Equal(t, "sid", sid)
	assert.Len(t, docs, 1)
}

// --- more test ---

func TestGetDocumentsWithSearchAfter_InternalBranches(t *testing.T) {
	es := &ElasticSearchHelper{}
	origJump := jumpToActiveBatchHook
	origSearch := searchRequestWithQueryHook
	defer func() {
		jumpToActiveBatchHook = origJump
		searchRequestWithQueryHook = origSearch
	}()

	jumpToActiveBatchHook = func(ctx context.Context, es *ElasticSearchHelper, cluster, index string, searchQuery SearchQuery) (JumpedDocumentResponse, error) {
		return JumpedDocumentResponse{insufficientDocuments: true, query: map[string]interface{}{}}, nil
	}
	docs, err := es.getDocumentsWithSearchAfter(context.Background(), "c", "i", SearchQuery{
		From:  0,
		Size:  1,
		Query: map[string]interface{}{},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.NoError(t, err)
	assert.Empty(t, docs)

	jumpToActiveBatchHook = func(ctx context.Context, es *ElasticSearchHelper, cluster, index string, searchQuery SearchQuery) (JumpedDocumentResponse, error) {
		return JumpedDocumentResponse{from: 0, query: map[string]interface{}{}}, nil
	}
	searchRequestWithQueryHook = func(ctx context.Context, es *ElasticSearchHelper, cluster, index string, query map[string]interface{}) (SearchResponse, error) {
		return SearchResponse{}, errors.New("search error")
	}
	_, err = es.getDocumentsWithSearchAfter(context.Background(), "c", "i", SearchQuery{
		From:  0,
		Size:  1,
		Query: map[string]interface{}{},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.Error(t, err)

	searchRequestWithQueryHook = func(ctx context.Context, es *ElasticSearchHelper, cluster, index string, query map[string]interface{}) (SearchResponse, error) {
		return SearchResponse{
			Hits: Hits{
				Hits: []Hit{
					{Sort: []interface{}{1}},
				},
			},
		}, nil
	}
	docs, err = es.getDocumentsWithSearchAfter(context.Background(), "c", "i", SearchQuery{
		From:  0,
		Size:  1,
		Query: map[string]interface{}{},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
}

func TestJumpToActiveBatchOfDocuments_InternalBranches(t *testing.T) {
	es := &ElasticSearchHelper{}
	setupTestClusterConfig()

	_, err := es.jumpToActiveBatchOfDocuments(context.Background(), "c", "i", SearchQuery{
		From:  EsMaxFetchSize,
		Size:  1,
		Query: map[string]interface{}{"bad": func() {}},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.Error(t, err)

	setExecErrCluster("es-jump-insufficient", "no-op-path")
	resp, err := es.jumpToActiveBatchOfDocuments(context.Background(), "es-jump-insufficient", "i", SearchQuery{
		From:  EsMaxFetchSize,
		Size:  1,
		Query: map[string]interface{}{},
		Sort:  []map[string]interface{}{{"_id": map[string]string{"order": "asc"}}},
	})
	assert.NoError(t, err)
	assert.True(t, resp.insufficientDocuments)
}

func TestGetDocumentsWithScrollApi_InternalBranches(t *testing.T) {
	es := &ElasticSearchHelper{}
	origGetClient := getClientHook
	origInitiate := initiateScrollApiSearchRequestHook
	origCast := castSearchResponseHook
	origGetSession := getDocumentsWithScrollSessionHook
	origClear := clearScrollRequestHook
	defer func() {
		getClientHook = origGetClient
		initiateScrollApiSearchRequestHook = origInitiate
		castSearchResponseHook = origCast
		getDocumentsWithScrollSessionHook = origGetSession
		clearScrollRequestHook = origClear
	}()

	getClientHook = func(clusterName string) (*es8.TypedClient, error) {
		return &es8.TypedClient{}, nil
	}

	initiateScrollApiSearchRequestHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, index string, query map[string]interface{}, scrollTimeout time.Duration) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	docs, err := es.getDocumentsWithScrollApi(context.Background(), "c", "i", SearchQuery{
		From:          0,
		Size:          10,
		FetchAllDocs:  false,
		Query:         map[string]interface{}{},
		ScrollTimeout: 1,
	})
	assert.NoError(t, err)
	assert.Empty(t, docs)

	initiateScrollApiSearchRequestHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, index string, query map[string]interface{}, scrollTimeout time.Duration) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return SearchResponse{}, errors.New("cast error")
	}
	_, err = es.getDocumentsWithScrollApi(context.Background(), "c", "i", SearchQuery{
		From:         0,
		Size:         10,
		FetchAllDocs: false,
		Query:        map[string]interface{}{},
	})
	assert.Error(t, err)

	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return SearchResponse{
			ScrollID: "sid",
			Hits: Hits{
				Hits: make([]Hit, EsMaxFetchSize),
			},
		}, nil
	}
	getDocumentsWithScrollSessionHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, searchQuery SearchQuery, scrollID string, scrollTimeout time.Duration, fetchSize int64) ([]Hit, string, error) {
		return nil, "", errors.New("scroll session error")
	}
	_, err = es.getDocumentsWithScrollApi(context.Background(), "c", "i", SearchQuery{
		From:         EsMaxFetchSize,
		Size:         1,
		FetchAllDocs: true,
		Query:        map[string]interface{}{},
	})
	assert.Error(t, err)

	getDocumentsWithScrollSessionHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, searchQuery SearchQuery, scrollID string, scrollTimeout time.Duration, fetchSize int64) ([]Hit, string, error) {
		return []Hit{{}}, "next", nil
	}
	clearScrollRequestHook = func(ctx context.Context, req esapi.ClearScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return nil, errors.New("clear error")
	}
	_, err = es.getDocumentsWithScrollApi(context.Background(), "c", "i", SearchQuery{
		From:         EsMaxFetchSize,
		Size:         1,
		FetchAllDocs: true,
		Query:        map[string]interface{}{},
	})
	assert.Error(t, err)
}

func TestGetDocumentsWithScrollSession_InternalBranches(t *testing.T) {
	es := &ElasticSearchHelper{}
	origDo := scrollRequestDoHook
	origCast := castSearchResponseHook
	defer func() {
		scrollRequestDoHook = origDo
		castSearchResponseHook = origCast
	}()

	scrollRequestDoHook = func(ctx context.Context, req esapi.ScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return nil, errors.New("scroll do error")
	}
	_, _, err := es.getDocumentsWithScrollSession(context.Background(), &es8.TypedClient{}, SearchQuery{FetchAllDocs: true}, "sid", time.Second, 2)
	assert.Error(t, err)

	scrollRequestDoHook = func(ctx context.Context, req esapi.ScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return &esapi.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
	}
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return SearchResponse{}, errors.New("cast error")
	}
	_, _, err = es.getDocumentsWithScrollSession(context.Background(), &es8.TypedClient{}, SearchQuery{FetchAllDocs: true}, "sid", time.Second, 2)
	assert.Error(t, err)

	sequence := 0
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		sequence++
		if sequence == 1 {
			res := SearchResponse{ScrollID: "sid2", Hits: Hits{Hits: []Hit{{}, {}}}}
			res.Hits.Total.Value = 2
			return res, nil
		}
		res := SearchResponse{ScrollID: "", Hits: Hits{Hits: []Hit{{}}}}
		res.Hits.Total.Value = 1
		return res, nil
	}
	_, _, err = es.getDocumentsWithScrollSession(context.Background(), &es8.TypedClient{}, SearchQuery{FetchAllDocs: true}, "sid", time.Second, 2)
	assert.NoError(t, err)

	sequence = 0
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		sequence++
		if sequence == 1 {
			res := SearchResponse{ScrollID: "sid2", Hits: Hits{Hits: []Hit{{}, {}}}}
			res.Hits.Total.Value = 2
			return res, nil
		}
		res := SearchResponse{ScrollID: "", Hits: Hits{Hits: []Hit{{}}}}
		res.Hits.Total.Value = 1
		return res, nil
	}
	_, _, err = es.getDocumentsWithScrollSession(context.Background(), &es8.TypedClient{}, SearchQuery{
		From:         3,
		Size:         1,
		FetchAllDocs: false,
		Query:        map[string]interface{}{},
	}, "sid", time.Second, 2)
	assert.NoError(t, err)
}
