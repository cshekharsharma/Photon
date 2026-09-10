package elasticsearch

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func closeBody(body io.Closer) {
	if err := body.Close(); err != nil {
		return
	}
}

// CastEsapiResponseToSearchResponse converts an Elasticsearch esapi.Response into a SearchResponse struct.
func CastEsApiResponseToSearchResponse(res *esapi.Response) (SearchResponse, error) {
	var searchResponse SearchResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return searchResponse, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return searchResponse, fmt.Errorf("error parsing response body: %s", err)
	}

	return searchResponse, nil
}

// CastEsapiResponseToCountResponse converts an Elasticsearch esapi.Response into a CountResponse struct.
func CastEsApiResponseToCountResponse(res *esapi.Response) (CountResponse, error) {
	var countResponse CountResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return countResponse, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&countResponse); err != nil {
		return countResponse, fmt.Errorf("error parsing response body: %s", err)
	}

	return countResponse, nil
}

// CastEsapiResponseToBulkResponse converts an Elasticsearch esapi.Response into a BulkResponse struct.
func CastEsApiResponseToBulkResponse(res *esapi.Response) (BulkResponse, error) {
	var bulkResponse BulkResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return bulkResponse, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&bulkResponse); err != nil {
		return bulkResponse, fmt.Errorf("error parsing response body: %s", err)
	}

	return bulkResponse, nil
}

// CastEsapiResponseToIndexResponse converts an Elasticsearch esapi.Response into an IndexResponse struct.
func CastEsApiResponseToIndexResponse(res *esapi.Response) (IndexResponse, error) {
	var response IndexResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return response, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("error parsing response body: %s", err)
	}

	return response, nil
}

// CastEsapiResponseToUpdateResponse converts an Elasticsearch esapi.Response into an UpdateResponse struct.
func CastEsApiResponseToUpdateResponse(res *esapi.Response) (UpdateResponse, error) {
	var response UpdateResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return response, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("error parsing response body: %s", err)
	}

	return response, nil
}

// CastEsapiResponseToDeleteResponse converts an Elasticsearch esapi.Response into a DeleteResponse struct.
func CastEsApiResponseToDeleteResponse(res *esapi.Response) (DeleteResponse, error) {
	var response DeleteResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return response, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("error parsing response body: %s", err)
	}

	return response, nil
}

// CastEsapiResponseToUpdateByQueryResponse converts an Elasticsearch esapi.Response into an UpdateByQueryResponse struct.
func CastEsApiResponseToUpdateByQueryResponse(res *esapi.Response) (UpdateByQueryResponse, error) {
	var response UpdateByQueryResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return response, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("error parsing response body: %s", err)
	}

	return response, nil
}

// CastEsapiResponseToDeleteByQueryResponse converts an Elasticsearch esapi.Response into a DeleteByQueryResponse struct.
func CastEsApiResponseToDeleteByQueryResponse(res *esapi.Response) (DeleteByQueryResponse, error) {
	var response DeleteByQueryResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return response, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return response, fmt.Errorf("error parsing response body: %s", err)
	}

	return response, nil
}

// CastEsApiResponseToSearchByDocIdResponse converts an Elasticsearch esapi.Response into a SearchByDocIdResponse struct.
func CastEsApiResponseToSearchByDocIdResponse(res *esapi.Response) (SearchByDocIdResponse, error) {
	var searchResponse SearchByDocIdResponse
	defer closeBody(res.Body)

	if res.IsError() {
		return searchResponse, fmt.Errorf("error response: %s", res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResponse); err != nil {
		return searchResponse, fmt.Errorf("error parsing response body: %s", err)
	}

	return searchResponse, nil
}
