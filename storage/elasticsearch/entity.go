package elasticsearch

import "net/http"

type ConnectionConfig struct {
	Addresses  []string
	Username   string
	Password   string
	MaxRetries int64
	Transport  http.RoundTripper
}

// BulkOperation represents a single operation in a bulk request.
type BulkOperation struct {
	Action   string                 `json:"action"`             // Action type: "index", "update", "delete"
	Index    string                 `json:"index"`              // Index name
	ID       *string                `json:"id,omitempty"`       // Document ID (optional)
	Document map[string]interface{} `json:"document,omitempty"` // Document data (for index/update)
}

const (
	ActionIndex  = "index"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

type Query struct {
	Query        BoolQuery                `json:"query"`
	Sort         []map[string]interface{} `json:"sort,omitempty"`
	From         int64                    `json:"from,omitempty"`
	Size         int64                    `json:"size,omitempty"`
	Aggregations map[string]interface{}   `json:"aggregations,omitempty"`
	SearchAfter  []interface{}            `json:"search_after,omitempty"`
	Source       []string                 `json:"_source,omitempty"`
}

// BoolQuery represents a boolean query structure.
type BoolQuery struct {
	Bool Bool `json:"bool"`
}

// Bool represents the components of a boolean query.
type Bool struct {
	Must    []interface{} `json:"must,omitempty"`
	Filter  []interface{} `json:"filter,omitempty"`
	Should  []interface{} `json:"should,omitempty"`
	MustNot []interface{} `json:"must_not,omitempty"`
}

// MatchQuery represents a match query structure.
type MatchQuery struct {
	Match map[string]interface{} `json:"match"`
}

// MatchAllQuery represents a match all query structure.
type MatchAllQuery struct {
	MatchAll struct{} `json:"match_all"` // Empty struct as no parameters are needed
}

// TermQuery represents a term query structure.
type TermQuery struct {
	Term map[string]interface{} `json:"term"`
}

// TermsQuery represents a terms query structure.
type TermsQuery struct {
	Terms map[string][]interface{} `json:"terms"`
}

// NestedQuery represents a nested query structure.
type NestedQuery struct {
	Nested map[string]interface{} `json:"nested"`
}

// SearchResponse represents the structure of the response returned by a search query.
type SearchResponse struct {
	Hits         Hits                   `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Took         int64                  `json:"took"`
	TimedOut     bool                   `json:"timed_out"`
	ScrollID     string                 `json:"_scroll_id,omitempty"`
}

// Hit represents an individual hit returned in the search response.
type Hit struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Score  float64                `json:"_score,omitempty"`
	Source map[string]interface{} `json:"_source"`
	Sort   []interface{}          `json:"sort,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

type Hits struct {
	Total struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	Hits []Hit `json:"hits"`
}

// CountResponse represents the response from both Count and CountByQuery methods.
type CountResponse struct {
	Count  int64     `json:"count"`
	Shards ShardInfo `json:"_shards"`
}

// IndexResponse represents the response from Insert, Update, and Delete methods.
type IndexResponse struct {
	Index       string    `json:"_index"`
	ID          string    `json:"_id"`
	Version     int64     `json:"_version"`
	Result      string    `json:"result"`
	Shards      ShardInfo `json:"_shards"`
	SeqNo       int64     `json:"_seq_no"`
	PrimaryTerm int64     `json:"_primary_term"`
}

type UpdateResponse struct {
	IndexResponse
}

type DeleteResponse struct {
	IndexResponse
}

// DeleteByQueryResponse represent the response from the DeleteByQueryResponse method.
type DeleteByQueryResponse struct {
	Took                 int64          `json:"took"`
	TimedOut             bool           `json:"timed_out"`
	Total                int64          `json:"total"`
	Updated              int64          `json:"updated,omitempty"`
	Deleted              int64          `json:"deleted"`
	Batches              int64          `json:"batches"`
	VersionConflicts     int64          `json:"version_conflicts"`
	Noops                int64          `json:"noops"`
	Retries              Retries        `json:"retries"`
	ThrottledMillis      int64          `json:"throttled_millis"`
	RequestsPerSecond    float64        `json:"requests_per_second"`
	ThrottledUntilMillis int64          `json:"throttled_until_millis"`
	Failures             []ErrorDetails `json:"failures"`
}

type UpdateByQueryResponse struct {
	DeleteByQueryResponse
}

// Retries represents the number of retries attempted during the update by query operation.
type Retries struct {
	Bulk   int64 `json:"bulk"`
	Search int64 `json:"search"`
}

// BulkResponse represents the response from the Bulk method.
type BulkResponse struct {
	Took   int64      `json:"took"`
	Errors bool       `json:"errors"`
	Items  []BulkItem `json:"items"`
}

// BulkItem represents an individual item in a BulkResponse.
type BulkItem struct {
	Index  *BulkItemDetails `json:"index,omitempty"`
	Update *BulkItemDetails `json:"update,omitempty"`
}

// BulkItemDetails contains details about an individual bulk operation result.
type BulkItemDetails struct {
	Index   string        `json:"_index"`
	ID      string        `json:"_id"`
	Version int64         `json:"_version"`
	Status  int64         `json:"status"`
	Error   *ErrorDetails `json:"error,omitempty"`
}

// ErrorDetails provides information about errors that occur during bulk operations.
type ErrorDetails struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// ShardInfo contains shard-related information.
type ShardInfo struct {
	Total      int64 `json:"total"`
	Successful int64 `json:"successful"`
	Failed     int64 `json:"failed"`
}

type SearchByDocIdResponse struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Source map[string]interface{} `json:"_source"`
	Found  bool                   `json:"found"`
}

type JumpedDocumentResponse struct {
	query                 map[string]interface{}
	insufficientDocuments bool
	from                  int64
}

type SearchQuery struct {
	Query         map[string]interface{}   `json:"query"`
	Sort          []map[string]interface{} `json:"sort,omitempty"`
	From          int64                    `json:"from,omitempty"`
	Size          int64                    `json:"size,omitempty"`
	FetchAllDocs  bool                     `json:"fetchAllDocs,omitempty"`
	ScrollTimeout int64                    `json:"scrollTimeout,omitempty"`
}
