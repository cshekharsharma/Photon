package caching

import (
	"errors"
	"testing"

	aerov8 "github.com/aerospike/aerospike-client-go/v8"
	aero8type "github.com/aerospike/aerospike-client-go/v8/types"
	"github.com/cshekharsharma/photon/storage/aerospike"
	"github.com/cshekharsharma/photon/utils/testutil/mocks"
	"github.com/cshekharsharma/photon/utils/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockedAerospikeConnector struct {
	mock.Mock
}

func (mac *mockedAerospikeConnector) Connect(p *aerov8.ClientPolicy, hosts []*aerov8.Host, wp *aerov8.WritePolicy) (aerospike.AerospikeInterface, error) {
	args := mac.Called(p, hosts, wp)
	return args.Get(0).(aerospike.AerospikeInterface), args.Error(1)
}

// ---- Mocking of aerospike wrapper on top of aerospike client --------- //
type mockedAerospike struct {
	mock.Mock
	Client    aerospike.AerospikeClientInterface
	RawClient *aerov8.Client
}

func (ma *mockedAerospike) GetClient() aerospike.AerospikeClientInterface {
	return ma.Client
}

func (ma *mockedAerospike) GetRawClient() *aerov8.Client {
	return ma.RawClient
}

func (a *mockedAerospike) IsConnected() bool {
	args := a.Called()
	return args.Get(0).(bool)
}

func (a *mockedAerospike) Close() {
}

func getMockedAerospikeCache(t *testing.T) (*AerospikeCache, *mocks.MockAerospikeClient) {
	mockClient := new(mocks.MockAerospikeClient)
	mockAero := &mockedAerospike{Client: mockClient}
	mockAero.On("IsConnected").Return(true)

	mockedConnector := new(mockedAerospikeConnector)
	mockedConnector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(mockAero, nil)

	randomStr, _ := types.GetCryptoSafeRandomString(16)
	cache, err := NewAerospikeCache(&Options{
		Provider:    ProviderAerospike,
		Cluster:     "tc-" + randomStr,
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1:3000", "127.0.0.1:1099", "127.0.0.1:1099"},
		ConnTimeout: 10,
		DefaultTTL:  60,
	}, mockedConnector)
	if err != nil {
		t.Fatalf("failed to create AerospikeCache: %v", err)
	}

	return cache, mockClient
}

func TestAerospikeNewAerospikeCache(t *testing.T) {
	mockClient := new(mocks.MockAerospikeClient)
	mockedAero := &mockedAerospike{Client: mockClient}
	mockedAero.On("IsConnected").Return(true)

	mockedConnector := new(mockedAerospikeConnector)
	mockedConnector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(mockedAero, nil)

	cache, err := NewAerospikeCache(&Options{
		Provider:    ProviderAerospike,
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1:1099", "127.0.0.1:1099", "127.0.0.1:1099"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}, mockedConnector)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cache == nil {
		t.Fatalf("Expected non-nil cache, got nil")
	}

	if cache.namespace != "test-ns" {
		t.Fatalf("Expected namespace 'test-ns', got %v", cache.namespace)
	}
}

func TestAerospikeNewAerospikeCache_NegativeTTLAndNilConnector(t *testing.T) {
	origConnect := aerospikeConnect
	t.Cleanup(func() { aerospikeConnect = origConnect })

	aerospikeConnect = func(connector aerospike.AerospikeConnectorInterface, clusterName string) (aerospike.AerospikeInterface, error) {
		return nil, errors.New("connect failed")
	}

	_, err := NewAerospikeCache(&Options{
		Provider:    ProviderAerospike,
		Cluster:     "test-cluster-negttl",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1:3000", "127.0.0.2:3000"},
		ConnTimeout: 5,
		DefaultTTL:  -1,
	}, nil)

	if err == nil {
		t.Fatalf("expected error from connect")
	}
}

func TestAerospikeNewAerospikeCache_ConnectError(t *testing.T) {
	mockedConnector := new(mockedAerospikeConnector)
	mockedConnector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return((*mockedAerospike)(nil), errors.New("boom"))

	_, err := NewAerospikeCache(&Options{
		Provider:    ProviderAerospike,
		Cluster:     "test-cluster-error",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1:3000"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}, mockedConnector)

	if err == nil {
		t.Fatalf("expected error from connect")
	}
}

func TestAerospikeKeyCreationErrors(t *testing.T) {
	cache, _ := getMockedAerospikeCache(t)

	origNewKey := aerospikeNewKey
	t.Cleanup(func() { aerospikeNewKey = origNewKey })
	aerospikeNewKey = func(namespace, setName, key string) (*aerov8.Key, error) {
		return nil, errors.New("bad key")
	}

	_, err := cache.Exists(&ExistsRequest{Key: "k"})
	assert.Error(t, err)

	_, err = cache.Get(&GetRequest{Key: "k"})
	assert.Error(t, err)

	_, err = cache.Set(&SetRequest{Key: "k", Fields: map[string]any{"a": 1}})
	assert.Error(t, err)

	_, err = cache.Delete(&DeleteRequest{Key: "k"})
	assert.Error(t, err)

	err = cache.Increment(&IncrementRequest{Key: "k", Fields: map[string]int64{"a": 1}})
	assert.Error(t, err)

	err = cache.Decrement(&DecrementRequest{Key: "k", Fields: map[string]int64{"a": 1}})
	assert.Error(t, err)

	err = cache.Append(&AppendRequest{Key: "k", Fields: map[string]string{"a": "b"}})
	assert.Error(t, err)

	_, err = cache.GetTTL(&GetTTLRequest{Key: "k"})
	assert.Error(t, err)

	err = cache.SetTTL(&SetTTLRequest{Key: "k", TTL: 10})
	assert.Error(t, err)
}

func TestAerospikeNewMultipleKeyErrors(t *testing.T) {
	cache, _ := getMockedAerospikeCache(t)

	origNewKeys := aerospikeNewKeys
	t.Cleanup(func() { aerospikeNewKeys = origNewKeys })
	aerospikeNewKeys = func(namespace, setName string, keys []string) ([]*aerov8.Key, error) {
		return nil, errors.New("bad keys")
	}

	_, err := cache.MultiGet(&MultiGetRequest{Keys: []string{"k1"}})
	assert.Error(t, err)

	_, err = cache.MultiDelete(&MultiDeleteRequest{Keys: []string{"k1"}})
	assert.Error(t, err)
}

func TestAerospikeMultiSet_KeyCreationError(t *testing.T) {
	cache, _ := getMockedAerospikeCache(t)

	origNewKey := aerospikeNewKey
	t.Cleanup(func() { aerospikeNewKey = origNewKey })
	aerospikeNewKey = func(namespace, setName, key string) (*aerov8.Key, error) {
		if key == "bad" {
			return nil, errors.New("bad key")
		}
		return aerov8.NewKey(namespace, setName, key)
	}

	_, err := cache.MultiSet(&MultiSetRequest{
		FieldsMap: map[string]map[string]any{
			"bad": {"a": 1},
		},
	})
	assert.Error(t, err)
}

func TestAerospikeExists_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	mockClient.On("Exists", mock.Anything, expectedKey).Return(true, nil)

	exists, err := cache.Exists(&ExistsRequest{
		Key: "test-key",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !exists {
		t.Fatalf("Expected exists = true, got false")
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeExists_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)
	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Exists", mock.Anything, expectedKey).Return(false, &aerov8.AerospikeError{})

	exists, err := cache.Exists(&ExistsRequest{Key: "test-key"})
	assert.Error(t, err)
	assert.False(t, exists)

	cache.namespace = ""
	_, err = cache.Exists(&ExistsRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeGet_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	expectedBins := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}

	mockClient.On("Get", mock.Anything, expectedKey, mock.Anything).Return(
		&aerov8.Record{
			Bins: expectedBins,
		}, nil)

	data, err := cache.Get(&GetRequest{
		Key:    "test-key",
		Fields: nil, // all fields
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	binData, ok := data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{} type, got %T", data)
	}

	if binData["field1"] != "value1" || binData["field2"] != 123 {
		t.Fatalf("Unexpected data retrieved: %+v", binData)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeGet_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)
	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Get", mock.Anything, expectedKey, mock.Anything).Return(&aerov8.Record{}, &aerov8.AerospikeError{})

	data, err := cache.Get(&GetRequest{Key: "test-key"})
	assert.Error(t, err)
	assert.Nil(t, data)

	cache.namespace = ""
	_, err = cache.Get(&GetRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeSet_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	bins := map[string]interface{}{
		"field1": "value1",
		"field2": 123,
	}

	mockClient.On("Put", mock.Anything, expectedKey, mock.Anything).Return(nil)

	found, err := cache.Set(&SetRequest{
		Key:    "test-key",
		Fields: bins,
		TTL:    60,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if found != true {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeSet_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Put", mock.Anything, expectedKey, mock.Anything).Return(&aerov8.AerospikeError{})

	found, err := cache.Set(&SetRequest{
		Key:    "test-key",
		Fields: map[string]interface{}{"field1": "value1"},
		TTL:    60,
	})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if found != false {
		t.Fatalf("Expected found = false, got %v", found)
	}

	cache.namespace = ""
	_, err = cache.Set(&SetRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeDelete_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Delete", mock.Anything, expectedKey).Return(true, nil)

	deleted, err := cache.Delete(&DeleteRequest{Key: "test-key"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !deleted {
		t.Fatalf("Expected deleted = true, got false")
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeDelete_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Delete", mock.Anything, expectedKey).Return(false, &aerov8.AerospikeError{})

	deleted, err := cache.Delete(&DeleteRequest{Key: "test-key"})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if deleted {
		t.Fatalf("Expected deleted = false, got true")
	}

	cache.namespace = ""
	_, err = cache.Delete(&DeleteRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeMultiGet_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	keys := []string{"key1", "key2"}
	aeroKeys := make([]*aerov8.Key, 0, len(keys))
	for _, k := range keys {
		ak, _ := aerospike.NewKey("test-ns", "test-collection", k)
		aeroKeys = append(aeroKeys, ak)
	}

	records := []*aerov8.Record{
		{Bins: map[string]interface{}{"field1": "value1"}},
		{Bins: map[string]interface{}{"field2": 123}},
	}

	mockClient.On("BatchGet", mock.Anything, aeroKeys, mock.Anything).Return(records, nil)

	data, err := cache.MultiGet(&MultiGetRequest{Keys: keys})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(data) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(data))
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeMultiGet_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	keys := []string{"key1"}
	aeroKeys := make([]*aerov8.Key, 0, len(keys))
	for _, k := range keys {
		ak, _ := aerospike.NewKey("test-ns", "test-collection", k)
		aeroKeys = append(aeroKeys, ak)
	}

	mockClient.On("BatchGet", mock.Anything, aeroKeys, mock.Anything).Return([]*aerov8.Record{}, &aerov8.AerospikeError{})

	data, err := cache.MultiGet(&MultiGetRequest{Keys: keys})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if data != nil {
		t.Fatalf("Expected nil data, got %v", data)
	}

	cache.namespace = ""
	_, err = cache.MultiGet(&MultiGetRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeMultiSet_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	items := map[string]map[string]any{
		"key1": map[string]interface{}{"field1": "value1"},
		"key2": map[string]interface{}{"field2": 123},
	}

	mockClient.On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	_, err := cache.MultiSet(&MultiSetRequest{FieldsMap: items, TTL: 60})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeMultiSet_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	items := map[string]map[string]any{
		"key1": map[string]interface{}{"field1": "value1"},
		"key2": map[string]interface{}{"field2": 123},
	}

	mockClient.On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&aerov8.AerospikeError{})

	_, err := cache.MultiSet(&MultiSetRequest{FieldsMap: items, TTL: 60})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	cache.namespace = ""
	_, err = cache.MultiSet(&MultiSetRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeMultiDelete_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	keys := []string{"key1", "key2"}

	mockClient.On("BatchDelete", mock.Anything, mock.Anything, mock.Anything).Return([]*aerov8.BatchRecord{
		{ResultCode: aero8type.OK},
		{ResultCode: aero8type.OK},
	}, nil)

	results, err := cache.MultiDelete(&MultiDeleteRequest{Keys: keys})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !results["key1"] || !results["key2"] {
		t.Fatalf("Expected true return value for key1 and key2, got false")
	}

	cache.namespace = ""
	_, err = cache.MultiDelete(&MultiDeleteRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeMultiDelete_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	keys := []string{"key1"}

	mockClient.On("BatchDelete", mock.Anything, mock.Anything, mock.Anything).Return([]*aerov8.BatchRecord{}, &aerov8.AerospikeError{})

	results, err := cache.MultiDelete(&MultiDeleteRequest{Keys: keys})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if results["key"] {
		t.Fatalf("Expected false for key, got %v", results["key"])
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeIncrement_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	ops := []*aerov8.Operation{
		aerov8.AddOp(aerov8.NewBin("counter", int64(10))),
	}

	mockClient.On("Operate", mock.Anything, expectedKey, ops).Return(&aerov8.Record{}, nil)

	err := cache.Increment(&IncrementRequest{
		Key:    "test-key",
		Fields: map[string]int64{"counter": int64(10)},
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeIncrement_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	ops := []*aerov8.Operation{
		aerov8.AddOp(aerov8.NewBin("counter", int64(10))),
	}

	mockClient.On("Operate", mock.Anything, expectedKey, ops).Return(&aerov8.Record{}, &aerov8.AerospikeError{})

	err := cache.Increment(&IncrementRequest{
		Key:    "test-key",
		Fields: map[string]int64{"counter": int64(10)},
	})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	cache.namespace = ""
	err = cache.Increment(&IncrementRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeDecrement_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	ops := []*aerov8.Operation{
		aerov8.AddOp(aerov8.NewBin("counter", int64(1))),
	}

	mockClient.On("Operate", mock.Anything, expectedKey, ops).Return(&aerov8.Record{}, nil)

	err := cache.Decrement(&DecrementRequest{
		Key:    "test-key",
		Fields: map[string]int64{"counter": -1 * int64(1)},
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeDecrement_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	ops := []*aerov8.Operation{
		aerov8.AddOp(aerov8.NewBin("counter", int64(10))),
	}

	mockClient.On("Operate", mock.Anything, expectedKey, ops).Return(&aerov8.Record{}, &aerov8.AerospikeError{})

	err := cache.Decrement(&DecrementRequest{
		Key:    "test-key",
		Fields: map[string]int64{"counter": -1 * int64(10)},
	})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	cache.namespace = ""
	err = cache.Decrement(&DecrementRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeAppend_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	ops := []*aerov8.Operation{
		aerov8.AppendOp(aerov8.NewBin("myfield", "vv")),
	}

	mockClient.On("Operate", mock.Anything, expectedKey, ops).Return(&aerov8.Record{}, nil)

	err := cache.Append(&AppendRequest{
		Key:    "test-key",
		Fields: map[string]string{"myfield": "vv"},
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeAppend_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	ops := []*aerov8.Operation{
		aerov8.AppendOp(aerov8.NewBin("myfield", "vv")),
	}

	mockClient.On("Operate", mock.Anything, expectedKey, ops).Return(&aerov8.Record{}, &aerov8.AerospikeError{})

	err := cache.Append(&AppendRequest{
		Key:    "test-key",
		Fields: map[string]string{"myfield": "vv"},
	})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	cache.namespace = ""
	err = cache.Append(&AppendRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeGetTTL_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")
	mockRecord := &aerov8.Record{
		Expiration: 120, // expiration time in seconds
	}

	mockClient.On("GetHeader", mock.Anything, expectedKey).Return(mockRecord, nil)

	ttl, err := cache.GetTTL(&GetTTLRequest{
		Key: "test-key",
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if ttl != 120 {
		t.Fatalf("Expected TTL = 120, got %v", ttl)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeGetTTL_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("GetHeader", mock.Anything, expectedKey).Return(&aerov8.Record{}, &aerov8.AerospikeError{})

	ttl, err := cache.GetTTL(&GetTTLRequest{
		Key: "test-key",
	})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if ttl != 0 {
		t.Fatalf("Expected TTL = 0 on error, got %v", ttl)
	}

	cache.namespace = ""
	_, err = cache.GetTTL(&GetTTLRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeSetTTL_Success(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Touch", mock.Anything, expectedKey).Return(nil)

	err := cache.SetTTL(&SetTTLRequest{
		Key: "test-key",
		TTL: 3600,
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	mockClient.AssertExpectations(t)
}

func TestAerospikeSetTTL_Error(t *testing.T) {
	cache, mockClient := getMockedAerospikeCache(t)

	expectedKey, _ := aerospike.NewKey("test-ns", "test-collection", "test-key")

	mockClient.On("Touch", mock.Anything, expectedKey).Return(&aerov8.AerospikeError{})

	err := cache.SetTTL(&SetTTLRequest{
		Key: "test-key",
		TTL: 3600,
	})

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	cache.namespace = ""
	err = cache.SetTTL(&SetTTLRequest{})
	assert.NotNil(t, err)

	mockClient.AssertExpectations(t)
}

func TestAerospikeVerifyRequest(t *testing.T) {
	cache, _ := getMockedAerospikeCache(t)

	t.Run("Success - fills namespace and collection if missing", func(t *testing.T) {
		req := &GetRequest{
			Key: "test-key",
		}

		verifiedReq, err := cache.verifyRequest(req)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if verifiedReq.GetNamespace() != "test-ns" {
			t.Errorf("Expected namespace 'test-ns', got %v", verifiedReq.GetNamespace())
		}

		if verifiedReq.GetCollection() != "test-collection" {
			t.Errorf("Expected collection 'test-collection', got %v", verifiedReq.GetCollection())
		}
	})

	t.Run("Error - request is nil", func(t *testing.T) {
		_, err := cache.verifyRequest(nil)
		if err == nil || err.Error() != "error: input cache request struct is nil" {
			t.Fatalf("Expected nil request error, got %v", err)
		}
	})

	t.Run("Error - namespace still missing", func(t *testing.T) {
		cacheNoNamespace, _ := getMockedAerospikeCache(t)
		cacheNoNamespace.namespace = ""

		req := &GetRequest{
			Key: "test-key",
		}
		req.SetCollection("some-collection")

		_, err := cacheNoNamespace.verifyRequest(req)
		if err == nil || err.Error() != "error: namespace field is empty" {
			t.Fatalf("Expected namespace empty error, got %v", err)
		}
	})

	t.Run("Error - collection still missing", func(t *testing.T) {
		cacheNoCollection, _ := getMockedAerospikeCache(t)
		cacheNoCollection.collection = ""

		req := &GetRequest{
			Key: "test-key",
		}
		req.SetNamespace("some-namespace")

		_, err := cacheNoCollection.verifyRequest(req)
		if err == nil || err.Error() != "error: collection field is empty" {
			t.Fatalf("Expected collection empty error, got %v", err)
		}
	})
}
