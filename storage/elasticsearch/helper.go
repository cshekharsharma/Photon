package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ElasticSearchHelper provides methods to interact with an Elasticsearch cluster.1
type ElasticSearchHelper struct {
}

var (
	getClientHook = func(clusterName string) (*es8.TypedClient, error) {
		return getClient(clusterName)
	}
	castSearchResponseHook = func(resp *esapi.Response) (SearchResponse, error) {
		return CastEsApiResponseToSearchResponse(resp)
	}
	searchRequestWithQueryHook = func(ctx context.Context, es *ElasticSearchHelper, cluster, index string, query map[string]interface{}) (SearchResponse, error) {
		return es.searchRequestWithQuery(ctx, cluster, index, query)
	}
	jumpToActiveBatchHook = func(ctx context.Context, es *ElasticSearchHelper, cluster, index string, searchQuery SearchQuery) (JumpedDocumentResponse, error) {
		return es.jumpToActiveBatchOfDocuments(ctx, cluster, index, searchQuery)
	}
	initiateScrollApiSearchRequestHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, index string, query map[string]interface{}, scrollTimeout time.Duration) (*esapi.Response, error) {
		return es.initiateScrollApiSearchRequest(ctx, client, index, query, scrollTimeout)
	}
	getDocumentsWithScrollSessionHook = func(ctx context.Context, es *ElasticSearchHelper, client *es8.TypedClient, searchQuery SearchQuery, scrollID string, scrollTimeout time.Duration, fetchSize int64) ([]Hit, string, error) {
		return es.getDocumentsWithScrollSession(ctx, client, searchQuery, scrollID, scrollTimeout, fetchSize)
	}
	marshalBulkMetaHook = func(v interface{}) ([]byte, error) {
		return json.Marshal(v)
	}
	clearScrollRequestHook = func(ctx context.Context, req esapi.ClearScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return req.Do(ctx, client)
	}
	scrollRequestDoHook = func(ctx context.Context, req esapi.ScrollRequest, client *es8.TypedClient) (*esapi.Response, error) {
		return req.Do(ctx, client)
	}
)

// getClient retrieves an Elasticsearch client for the specified cluster name.
// This function assumes that the connection configuration has been set using SetConnectionConfig.
func getClient(clusterName string) (*es8.TypedClient, error) {
	client, err := Connect(clusterName) // Assume Connect is implemented to return the client
	if err != nil {
		return nil, fmt.Errorf("failed to connect to cluster %s: %w", clusterName, err)
	}
	return client, nil
}

// doSearchRequest provides method to do elastic search request by raw query
func doSearchRequest(ctx context.Context, cluster, index string, query []byte) (SearchResponse, error) {
	var searchResponse SearchResponse
	client, err := getClientHook(cluster)
	if err != nil {
		return searchResponse, fmt.Errorf("error in getting client: %w", err)
	}
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(query)),
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return searchResponse, fmt.Errorf("error executing search: %w", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return searchResponse, nil
	}

	searchResponse, err = castSearchResponseHook(resp)
	if err != nil {
		return searchResponse, fmt.Errorf("error in casting esapi response to search response: %w", err)
	}
	return searchResponse, nil
}

// Search performs a search query against the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index to search.
//   - query: The search query as a Query struct.
//
// Returns:
//   - SearchResponse: The response from the Elasticsearch search request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) Search(ctx context.Context, cluster string, index string, query Query) (SearchResponse, error) {
	var searchResponse SearchResponse
	body, err := json.Marshal(query)
	if err != nil {
		return searchResponse, fmt.Errorf("error marshaling query: %w", err)
	}

	return doSearchRequest(ctx, cluster, index, body)
}

// SearchByDocId performs a search by docId against the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index to search.
//   - docId: documentId of the document in ES.
//
// Returns:
//   - SearchResponse: The response from the Elasticsearch search request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) SearchByDocId(ctx context.Context, cluster string, index string, docId string) (SearchByDocIdResponse, error) {
	var searchResponse SearchByDocIdResponse
	client, err := getClient(cluster)
	if err != nil {
		return searchResponse, fmt.Errorf("error in getting client: %w", err)
	}

	req := esapi.GetRequest{
		Index:      index,
		DocumentID: docId,
	}

	resp, err := req.Do(ctx, client)

	if err != nil {
		return searchResponse, fmt.Errorf("error executing search: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return searchResponse, nil
	}

	searchResponse, err = CastEsApiResponseToSearchByDocIdResponse(resp)
	if err != nil {
		return searchResponse, fmt.Errorf("error in casting esapi response to search response: %w", err)
	}

	return searchResponse, nil
}

// Index adds a new document to the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index where the document will be inserted.
//   - docID: The document ID; can be used to overwrite an existing document.
//   - document: The document to insert as a JSON-serializable object.
//
// Returns:
//   - IndexResponse: The response from the Elasticsearch index request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) Index(ctx context.Context, cluster string, index string, docID string, document interface{}) (IndexResponse, error) {
	client, err := getClient(cluster)
	var indexResponse IndexResponse
	if err != nil {
		return indexResponse, err
	}

	body, err := json.Marshal(document)
	if err != nil {
		return indexResponse, fmt.Errorf("error marshaling document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return indexResponse, fmt.Errorf("error executing insert: %w", err)
	}

	indexResponse, err = CastEsApiResponseToIndexResponse(resp)
	if err != nil {
		return indexResponse, fmt.Errorf("error in casting esapi response to index response: %w", err)
	}

	return indexResponse, nil
}

// Update modifies an existing document in the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index containing the document to update.
//   - docID: The document ID of the document to update.
//   - document: The updated document as a JSON-serializable object.
//
// Returns:
//   - UpdateResponse: The response from the Elasticsearch update request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) Update(ctx context.Context, cluster string, index string, docID string, document interface{}) (UpdateResponse, error) {
	client, err := getClient(cluster)
	var updateResponse UpdateResponse
	if err != nil {
		return updateResponse, err
	}

	updatedDocument := map[string]interface{}{
		"doc": document,
	}

	body, err := json.Marshal(updatedDocument)
	if err != nil {
		return updateResponse, fmt.Errorf("error marshaling document: %w", err)
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: docID,
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return updateResponse, fmt.Errorf("error executing update: %w", err)
	}
	updateResponse, err = CastEsApiResponseToUpdateResponse(resp)
	if err != nil {
		return updateResponse, fmt.Errorf("error in casting esapi response to update response: %w", err)
	}

	return updateResponse, nil
}

// Delete removes a document from the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index containing the document to delete.
//   - docID: The document ID of the document to delete.
//
// Returns:
//   - DeleteResponse: The response from the Elasticsearch delete request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) Delete(ctx context.Context, cluster string, index string, docID string) (DeleteResponse, error) {
	client, err := getClient(cluster)
	var deleteResponse DeleteResponse
	if err != nil {
		return deleteResponse, err
	}

	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: docID,
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return deleteResponse, fmt.Errorf("error executing delete: %w", err)
	}

	deleteResponse, err = CastEsApiResponseToDeleteResponse(resp)
	if err != nil {
		return deleteResponse, fmt.Errorf("error in casting esapi response to delete response: %w", err)
	}

	return deleteResponse, nil
}

// Count returns the number of documents in the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index to count documents in.
//
// Returns:
//   - CountResponse: The response from the Elasticsearch count request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) Count(ctx context.Context, cluster string, index string) (CountResponse, error) {
	client, err := getClient(cluster)
	var countResponse CountResponse
	if err != nil {
		return countResponse, err
	}

	req := esapi.CountRequest{
		Index: []string{index},
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return countResponse, fmt.Errorf("error executing count: %w", err)
	}

	countResponse, err = CastEsApiResponseToCountResponse(resp)
	if err != nil {
		return countResponse, fmt.Errorf("error in casting esapi response to count response: %w", err)
	}

	return countResponse, nil
}

// doCountByQueryRequest fetches count based on raw query
func doCountByQueryRequest(ctx context.Context, cluster, index string, query []byte) (CountResponse, error) {
	client, err := getClient(cluster)
	var countResponse CountResponse
	if err != nil {
		return countResponse, err
	}

	req := esapi.CountRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(query)),
	}

	// Execute the count request
	resp, err := req.Do(ctx, client)
	if err != nil {
		return countResponse, fmt.Errorf("error executing count by query: %w", err)
	}

	countResponse, err = CastEsApiResponseToCountResponse(resp)
	if err != nil {
		return countResponse, fmt.Errorf("error in casting esapi response to count response: %w", err)
	}

	return countResponse, nil
}

// CountByQuery counts the number of documents in the specified index based on the provided query.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index to count documents in.
//   - query: The query to match documents for counting, as a Query struct.
//
// Returns:
//   - CountResponse: The response from the Elasticsearch count request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) CountByQuery(ctx context.Context, cluster string, index string, query Query) (CountResponse, error) {

	var countResponse CountResponse
	// Marshal the query to JSON
	body, err := json.Marshal(query)
	if err != nil {
		return countResponse, fmt.Errorf("error marshaling query: %w", err)
	}

	return doCountByQueryRequest(ctx, cluster, index, body)
}

// UpdateByQuery updates documents in the specified index based on the provided query.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index to update documents in.
//   - query: The query to match documents for updating, as a JSON-serializable object.
//
// Returns:
//   - UpdateByQueryResponse: The response from the Elasticsearch update by query request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) UpdateByQuery(ctx context.Context, cluster string, index string, query interface{}) (UpdateByQueryResponse, error) {
	client, err := getClient(cluster)
	var updateResponse UpdateByQueryResponse
	if err != nil {
		return updateResponse, err
	}

	body, err := json.Marshal(query)
	if err != nil {
		return updateResponse, fmt.Errorf("error marshaling query: %w", err)
	}

	req := esapi.UpdateByQueryRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(body)),
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return updateResponse, fmt.Errorf("error executing update by query: %w", err)
	}

	updateResponse, err = CastEsApiResponseToUpdateByQueryResponse(resp)
	if err != nil {
		return updateResponse, fmt.Errorf("error in casting esapi response to update by query response: %w", err)
	}

	return updateResponse, nil
}

// DeleteByQuery deletes documents in the specified index based on the provided query.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - index: The name of the index to delete documents from.
//   - query: The query to match documents for deletion, as a JSON-serializable object.
//
// Returns:
//   - DeleteByQuery: The response from the Elasticsearch delete by query request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) DeleteByQuery(ctx context.Context, cluster string, index string, query interface{}) (DeleteByQueryResponse, error) {
	client, err := getClient(cluster)
	var deleteResponse DeleteByQueryResponse
	if err != nil {
		return deleteResponse, err
	}

	body, err := json.Marshal(query)
	if err != nil {
		return deleteResponse, fmt.Errorf("error marshaling query: %w", err)
	}

	req := esapi.DeleteByQueryRequest{
		Index: []string{index},
		Body:  strings.NewReader(string(body)),
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return deleteResponse, fmt.Errorf("error executing delete by query: %w", err)
	}

	deleteResponse, err = CastEsApiResponseToDeleteByQueryResponse(resp)
	if err != nil {
		return deleteResponse, fmt.Errorf("error in casting esapi response to update by delete response: %w", err)
	}

	return deleteResponse, nil
}

// Bulk performs bulk operations on the specified index.
// Parameters:
//   - cluster: Elastic cluster to be used
//   - operations: A slice of BulkOperation representing the operations to perform.
//
// Returns:
//   - BulkResponse: The response from the Elasticsearch bulk request.
//   - error: An error object describing any issues encountered during the operation.
func (es *ElasticSearchHelper) Bulk(ctx context.Context, cluster string, operations []BulkOperation) (BulkResponse, error) {
	client, err := getClient(cluster)
	var bulkResponse BulkResponse
	if err != nil {
		return bulkResponse, err
	}

	var buf bytes.Buffer

	for _, op := range operations {
		// Initialize metadata for the action
		meta := map[string]interface{}{
			op.Action: map[string]interface{}{
				"_index": op.Index,
			},
		}

		// Add ID if present
		if op.ID != nil {
			meta[op.Action].(map[string]interface{})["_id"] = *op.ID
		}

		// Marshal metadata
		metaLine, err := marshalBulkMetaHook(meta)
		if err != nil {
			return bulkResponse, fmt.Errorf("error marshaling metadata: %w", err)
		}
		buf.Write(metaLine)
		buf.WriteByte('\n')

		// Marshal document if action is index or update
		if op.Action == ActionIndex || op.Action == ActionUpdate {
			docLine, err := json.Marshal(op.Document)
			if err != nil {
				return bulkResponse, fmt.Errorf("error marshaling document: %w", err)
			}
			buf.Write(docLine)
			buf.WriteByte('\n')
		}
	}

	// Perform the bulk request
	req := esapi.BulkRequest{
		Body: strings.NewReader(buf.String()),
	}

	resp, err := req.Do(ctx, client)
	if err != nil {
		return bulkResponse, fmt.Errorf("error executing bulk operation: %w", err)
	}

	bulkResponse, err = CastEsApiResponseToBulkResponse(resp)
	if err != nil {
		return bulkResponse, fmt.Errorf("error in casting esapi response to update by delete response: %w", err)
	}

	return bulkResponse, nil
}

// RawSearchQuery provides support for search using raw query
func (es *ElasticSearchHelper) RawSearchQuery(ctx context.Context, cluster, index, query string) (SearchResponse, error) {
	var body = []byte(query)
	return doSearchRequest(ctx, cluster, index, body)
}

// RawCountQuery provides support for count using raw query
func (es *ElasticSearchHelper) RawCountQuery(ctx context.Context, cluster, index, query string) (CountResponse, error) {
	var body = []byte(query)
	return doCountByQueryRequest(ctx, cluster, index, body)
}

func (es *ElasticSearchHelper) IncrementalSearch(ctx context.Context, cluster string, index string, searchQuery SearchQuery) (SearchResponse, error) {
	var response SearchResponse

	if !searchQuery.FetchAllDocs && (searchQuery.From < 0 || searchQuery.Size <= 0) {
		return response, fmt.Errorf("invalid parameters: 'from' must be >= 0 and 'size' must be > 0, given from: %d, size: %d", searchQuery.From, searchQuery.Size)
	}

	if searchQuery.Query == nil {
		return response, fmt.Errorf("query cannot be nil")
	}

	if searchQuery.FetchAllDocs {
		searchQuery.From = 0
		searchQuery.Size = 0
	}

	if searchQuery.FetchAllDocs || searchQuery.From+searchQuery.Size > EsMaxFetchSize {
		return es.getBulkDocuments(ctx, cluster, index, searchQuery)
	}

	searchQuery.Query["size"] = searchQuery.Size
	searchQuery.Query["from"] = searchQuery.From

	if len(searchQuery.Sort) > 0 {
		searchQuery.Query["sort"] = searchQuery.Sort
	}

	return es.searchRequestWithQuery(ctx, cluster, index, searchQuery.Query)
}

func (es *ElasticSearchHelper) getBulkDocuments(ctx context.Context, cluster string, index string, searchQuery SearchQuery) (SearchResponse, error) {
	var response SearchResponse
	var docs []Hit
	var err error

	if len(searchQuery.Sort) > 0 {
		docs, err = es.getDocumentsWithSearchAfter(ctx, cluster, index, searchQuery)
	} else {
		docs, err = es.getDocumentsWithScrollApi(ctx, cluster, index, searchQuery)
	}

	if err != nil {
		return response, fmt.Errorf("error while getting documents : %w", err)
	}

	response.Hits.Hits = docs
	response.Hits.Total.Value = int64(len(docs))

	return response, nil
}

func (es *ElasticSearchHelper) getDocumentsWithSearchAfter(ctx context.Context, cluster, index string, searchQuery SearchQuery) ([]Hit, error) {
	if len(searchQuery.Sort) == 0 {
		return nil, fmt.Errorf("sort order should not be empty")
	}

	var docs []Hit
	searchQuery.Query["sort"] = searchQuery.Sort
	delete(searchQuery.Query, "from")

	jumpedDocs, err := jumpToActiveBatchHook(ctx, es, cluster, index, searchQuery)

	if err != nil {
		return docs, fmt.Errorf("error while jumping to active batch of documents: %w", err)
	}

	if jumpedDocs.insufficientDocuments {
		return docs, nil
	}

	searchQuery.Query = jumpedDocs.query
	fetchFrom := jumpedDocs.from
	var totalDocsToFetch = fetchFrom + searchQuery.Size

	for searchQuery.FetchAllDocs || totalDocsToFetch > 0 {
		batchSize := EsMaxFetchSize
		if !searchQuery.FetchAllDocs && totalDocsToFetch < EsMaxFetchSize {
			batchSize = totalDocsToFetch
		}
		searchQuery.Query["size"] = batchSize

		resp, err := searchRequestWithQueryHook(ctx, es, cluster, index, searchQuery.Query)

		if err != nil {
			return docs, fmt.Errorf("error executing search request: %w", err)
		}

		if int64(len(resp.Hits.Hits)) < batchSize {
			if int64(len(resp.Hits.Hits)) > fetchFrom {
				docs = append(docs, resp.Hits.Hits[fetchFrom:]...)
			}
			break
		}

		lastHit := resp.Hits.Hits[len(resp.Hits.Hits)-1]
		searchQuery.Query["search_after"] = lastHit.Sort

		if !searchQuery.FetchAllDocs {
			totalDocsToFetch -= batchSize
		}

		docs = append(docs, resp.Hits.Hits[fetchFrom:]...)
		fetchFrom = 0
	}

	return docs, nil
}

func (es *ElasticSearchHelper) jumpToActiveBatchOfDocuments(ctx context.Context, cluster, index string, searchQuery SearchQuery) (JumpedDocumentResponse, error) {
	var insufficientDocuments = false
	var response JumpedDocumentResponse
	var sourceToFetch []string
	var from = searchQuery.From
	var query = searchQuery.Query

	if source, exists := query["_source"]; exists {
		src, ok := source.([]string)
		if !ok {
			return response, fmt.Errorf("invalid type for _source, expected []string")
		}
		sourceToFetch = src
	}

	if from >= EsMaxFetchSize {
		query["_source"] = []string{"_id"}
		query["size"] = EsMaxFetchSize

		for from >= EsMaxFetchSize {
			resp, err := es.searchRequestWithQuery(ctx, cluster, index, query)
			if err != nil {
				return response, fmt.Errorf("error while executing search query: %w", err)
			}
			// use len
			if resp.Hits.Total.Value == 0 {
				insufficientDocuments = true
				break
			}

			lastHit := resp.Hits.Hits[len(resp.Hits.Hits)-1]
			query["search_after"] = lastHit.Sort

			from -= EsMaxFetchSize
		}

	}

	if sourceToFetch != nil {
		query["_source"] = sourceToFetch
	} else {
		delete(query, "_source")
	}

	response.from = from
	response.query = query
	response.insufficientDocuments = insufficientDocuments
	return response, nil
}

func (es *ElasticSearchHelper) getDocumentsWithScrollApi(ctx context.Context, cluster, index string, searchQuery SearchQuery) ([]Hit, error) {
	var docs []Hit
	var scrollTimeout = EsScrollApiDefaultTimeout * time.Second

	client, err := getClientHook(cluster)
	if err != nil {
		return docs, fmt.Errorf("error in getting client: %w", err)
	}

	if searchQuery.ScrollTimeout != 0 {
		scrollTimeout = time.Duration(searchQuery.ScrollTimeout) * time.Second
	}

	var numberOfIterations = math.Ceil(float64(searchQuery.From+searchQuery.Size) / float64(EsMaxFetchSize))

	var fetchSize = EsMaxFetchSize

	if !searchQuery.FetchAllDocs {
		fetchSize = int64(math.Ceil(float64(searchQuery.From+searchQuery.Size) / numberOfIterations))
	}

	searchQuery.Query["size"] = int(fetchSize)
	delete(searchQuery.Query, "from")

	res, err := initiateScrollApiSearchRequestHook(ctx, es, client, index, searchQuery.Query, scrollTimeout)

	if err != nil || res.StatusCode == http.StatusBadRequest {
		return docs, err
	}

	searchResponse, err := castSearchResponseHook(res)

	if err != nil {
		return docs, fmt.Errorf("error in casting esapi response: %v, to search response with error: %w", res, err)
	}

	var totalDocsFound = fetchSize

	if int64(len(searchResponse.Hits.Hits)) < fetchSize {
		totalDocsFound = int64(len(searchResponse.Hits.Hits))
	}

	if searchQuery.From < totalDocsFound {
		docs = append(docs, searchResponse.Hits.Hits[searchQuery.From:]...)
		searchQuery.Size -= (fetchSize - searchQuery.From)
		searchQuery.From = 0
	} else {
		if totalDocsFound < fetchSize {
			return docs, nil
		}
		searchQuery.From -= fetchSize
	}

	scrollID := searchResponse.ScrollID

	// Scroll through the remaining documents
	if totalDocsFound == fetchSize {
		resp, sId, err := getDocumentsWithScrollSessionHook(ctx, es, client, searchQuery, scrollID, scrollTimeout, fetchSize)

		if err != nil {
			return docs, err
		}

		docs = append(docs, resp...)
		scrollID = sId
	}
	// Clear the scroll context
	clearReq := esapi.ClearScrollRequest{
		ScrollID: []string{scrollID},
	}

	_, err = clearScrollRequestHook(ctx, clearReq, client)

	if err != nil {
		return docs, fmt.Errorf("error clearing scroll session: %w", err)
	}

	return docs, nil
}

func (es *ElasticSearchHelper) initiateScrollApiSearchRequest(ctx context.Context, client *es8.TypedClient, index string, query map[string]interface{}, scrollTimeout time.Duration) (*esapi.Response, error) {
	queryBody, err := json.Marshal(query)

	if err != nil {
		return nil, fmt.Errorf("error while marshaling the query: %s, error: %w", string(queryBody), err)
	}

	searchReq := esapi.SearchRequest{
		Index:  []string{index},
		Body:   strings.NewReader(string(queryBody)),
		Scroll: scrollTimeout,
	}

	// Execute the initial search request
	res, err := searchReq.Do(ctx, client)

	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}

	return res, nil
}

func (es *ElasticSearchHelper) getDocumentsWithScrollSession(ctx context.Context, client *es8.TypedClient, searchQuery SearchQuery, scrollID string, scrollTimeout time.Duration, fetchSize int64) ([]Hit, string, error) {
	var docs []Hit

	for scrollID != "" {
		scrollReq := esapi.ScrollRequest{
			ScrollID: scrollID,
			Scroll:   scrollTimeout,
		}

		res, err := scrollRequestDoHook(ctx, scrollReq, client)
		if err != nil {
			return docs, scrollID, fmt.Errorf("error executing search: %w", err)
		}

		if res.StatusCode == http.StatusBadRequest {
			break
		}

		searchResponse, err := castSearchResponseHook(res)

		if err != nil {
			return docs, scrollID, fmt.Errorf("error in casting esapi response: %v, to search response with error: %w", res, err)
		}

		var totalDocsFound = fetchSize

		if int64(len(searchResponse.Hits.Hits)) < fetchSize {
			totalDocsFound = int64(len(searchResponse.Hits.Hits))
		}

		if searchQuery.FetchAllDocs {
			docs = append(docs, searchResponse.Hits.Hits...)
			scrollID = searchResponse.ScrollID
			if totalDocsFound < fetchSize {
				break
			}
			continue
		}

		if searchQuery.From < totalDocsFound {
			if searchQuery.From+searchQuery.Size <= totalDocsFound {
				docs = append(docs, searchResponse.Hits.Hits[searchQuery.From:searchQuery.From+searchQuery.Size]...)
				break
			} else {
				docs = append(docs, searchResponse.Hits.Hits[searchQuery.From:]...)
				if totalDocsFound < fetchSize {
					break
				}
				searchQuery.Size -= (fetchSize - searchQuery.From)
				searchQuery.From = 0
			}
		} else {
			if totalDocsFound < fetchSize {
				break
			}
			searchQuery.From -= fetchSize
		}

		// Update the scroll ID for the next request
		scrollID = searchResponse.ScrollID
	}

	return docs, scrollID, nil

}

func (es *ElasticSearchHelper) searchRequestWithQuery(ctx context.Context, cluster, index string, query map[string]interface{}) (SearchResponse, error) {
	var searchResponse SearchResponse

	body, err := json.Marshal(query)
	if err != nil {
		return searchResponse, fmt.Errorf("error while marshaling the query: %s, error: %w", string(body), err)
	}

	return doSearchRequest(ctx, cluster, index, body)
}
