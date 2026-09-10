package mocks

import (
	"context"

	"github.com/cshekharsharma/photon/storage/elasticsearch"
	"github.com/stretchr/testify/mock"
)

type MockElasticSearchHelper struct {
	mock.Mock
}

func (m *MockElasticSearchHelper) RawSearchQuery(ctx context.Context, cluster, index, query string) (elasticsearch.SearchResponse, error) {
	args := m.Called(ctx, cluster, index, query)
	return args.Get(0).(elasticsearch.SearchResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) RawCountQuery(ctx context.Context, cluster, index, query string) (elasticsearch.CountResponse, error) {
	args := m.Called(ctx, cluster, index, query)
	return args.Get(0).(elasticsearch.CountResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) Search(ctx context.Context, cluster string, index string, query elasticsearch.Query) (elasticsearch.SearchResponse, error) {
	args := m.Called(ctx, cluster, index, query)
	return args.Get(0).(elasticsearch.SearchResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) SearchByDocId(ctx context.Context, cluster string, index string, docId string) (elasticsearch.SearchByDocIdResponse, error) {
	args := m.Called(ctx, cluster, index, docId)
	return args.Get(0).(elasticsearch.SearchByDocIdResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) Index(ctx context.Context, cluster string, index string, docID string, document interface{}) (elasticsearch.IndexResponse, error) {
	args := m.Called(ctx, cluster, index, docID, document)
	return args.Get(0).(elasticsearch.IndexResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) Update(ctx context.Context, cluster string, index string, docID string, document interface{}) (elasticsearch.UpdateResponse, error) {
	args := m.Called(ctx, cluster, index, docID, document)
	return args.Get(0).(elasticsearch.UpdateResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) Delete(ctx context.Context, cluster string, index string, docID string) (elasticsearch.DeleteResponse, error) {
	args := m.Called(ctx, cluster, index, docID)
	return args.Get(0).(elasticsearch.DeleteResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) Count(ctx context.Context, cluster string, index string) (elasticsearch.CountResponse, error) {
	args := m.Called(ctx, cluster, index)
	return args.Get(0).(elasticsearch.CountResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) CountByQuery(ctx context.Context, cluster string, index string, query elasticsearch.Query) (elasticsearch.CountResponse, error) {
	args := m.Called(ctx, cluster, index, query)
	return args.Get(0).(elasticsearch.CountResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) UpdateByQuery(ctx context.Context, cluster string, index string, query interface{}) (elasticsearch.UpdateByQueryResponse, error) {
	args := m.Called(ctx, cluster, index, query)
	return args.Get(0).(elasticsearch.UpdateByQueryResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) DeleteByQuery(ctx context.Context, cluster string, index string, query interface{}) (elasticsearch.DeleteByQueryResponse, error) {
	args := m.Called(ctx, cluster, index, query)
	return args.Get(0).(elasticsearch.DeleteByQueryResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) Bulk(ctx context.Context, cluster string, operations []elasticsearch.BulkOperation) (elasticsearch.BulkResponse, error) {
	args := m.Called(ctx, cluster, operations)
	return args.Get(0).(elasticsearch.BulkResponse), args.Error(1)
}

func (m *MockElasticSearchHelper) IncrementalSearch(ctx context.Context, cluster string, index string, searchQuery elasticsearch.SearchQuery) (elasticsearch.SearchResponse, error) {
	args := m.Called(ctx, cluster, index, searchQuery)
	return args.Get(0).(elasticsearch.SearchResponse), args.Error(1)
}
