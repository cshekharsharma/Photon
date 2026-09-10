package elasticsearch

import "context"

// ElasticSearchInterface defines the methods available for interacting with Elasticsearch.
type ElasticSearchInterface interface {
	// Search performs a search query against the specified index.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index to search.
	//   - query: The search query as a Query struct.
	//
	// Returns:
	//   - SearchResponse: The response from the Elasticsearch search request.
	//   - error: An error object describing any issues encountered during the operation.
	Search(context.Context, string, string, Query) (SearchResponse, error)

	// SearchByDocId performs a search by docId against the specified index.
	// Parameters:
	//   - cluster: Elastic cluster to be used
	//   - index: The name of the index to search.
	//   - docId: documentId of the document in ES.
	//
	// Returns:
	//   - SearchResponse: The response from the Elasticsearch search request.
	//   - error: An error object describing any issues encountered during the operation.
	SearchByDocId(context.Context, string, string, string) (SearchByDocIdResponse, error)
	// Index adds a new document to the specified index.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index where the document will be inserted.
	//   - docID: The document ID; can be used to overwrite an existing document.
	//   - document: The document to insert as a JSON-serializable object.
	//
	// Returns:
	//   - IndexResponse: The response from the Elasticsearch index request.
	//   - error: An error object describing any issues encountered during the operation.
	Index(context.Context, string, string, string, interface{}) (IndexResponse, error)

	// Update modifies an existing document in the specified index.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index containing the document to update.
	//   - docID: The document ID of the document to update.
	//   - document: The updated document as a JSON-serializable object.
	//
	// Returns:
	//   - UpdateResponse: The response from the Elasticsearch update request.
	//   - error: An error object describing any issues encountered during the operation.
	Update(context.Context, string, string, string, interface{}) (UpdateResponse, error)

	// Delete removes a document from the specified index.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index containing the document to delete.
	//   - docID: The document ID of the document to delete.
	//
	// Returns:
	//   - DeleteResponse: The response from the Elasticsearch delete request.
	//   - error: An error object describing any issues encountered during the operation.
	Delete(context.Context, string, string, string) (DeleteResponse, error)

	// Count returns the number of documents in the specified index.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index to count documents in.
	//
	// Returns:
	//   - CountResponse: The response from the Elasticsearch count request.
	//   - error: An error object describing any issues encountered during the operation.
	Count(context.Context, string, string) (CountResponse, error)

	// CountByQuery counts the documents in the specified index based on the provided query.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index to count documents in.
	//   - query: The query to match documents for counting, as a Query struct.
	//
	// Returns:
	//   - CountResponse: The response from the Elasticsearch count by query request.
	//   - error: An error object describing any issues encountered during the operation.
	CountByQuery(context.Context, string, string, Query) (CountResponse, error)

	// UpdateByQuery updates documents in the specified index based on the provided query.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index to update documents in.
	//   - query: The query to match documents for updating, as a JSON-serializable object.
	//
	// Returns:
	//   - UpdateByQueryResponse: The response from the Elasticsearch update by query request.
	//   - error: An error object describing any issues encountered during the operation.
	UpdateByQuery(context.Context, string, string, interface{}) (UpdateByQueryResponse, error)

	// DeleteByQuery deletes documents in the specified index based on the provided query.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - index: The name of the index to delete documents from.
	//   - query: The query to match documents for deletion, as a JSON-serializable object.
	//
	// Returns:
	//   - DeleteByQueryResponse: The response from the Elasticsearch delete by query request.
	//   - error: An error object describing any issues encountered during the operation.
	DeleteByQuery(context.Context, string, string, interface{}) (DeleteByQueryResponse, error)

	// Bulk performs bulk operations on the specified index.
	// Parameters:
	//	 - cluster: Elastic cluster to be used
	//   - operations: A slice of BulkOperation representing the operations to perform.
	//
	// Returns:
	//   - BulkResponse: The response from the Elasticsearch bulk request.
	//   - error: An error object describing any issues encountered during the operation.
	Bulk(context.Context, string, []BulkOperation) (BulkResponse, error)

	RawSearchQuery(context.Context, string, string, string) (SearchResponse, error)

	RawCountQuery(context.Context, string, string, string) (CountResponse, error)

	IncrementalSearch(context.Context, string, string, SearchQuery) (SearchResponse, error)
}
