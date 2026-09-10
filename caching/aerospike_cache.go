package caching

import (
	"errors"
	"fmt"

	aerov8 "github.com/aerospike/aerospike-client-go/v8"
	"github.com/aerospike/aerospike-client-go/v8/types"
	"github.com/cshekharsharma/photon/storage/aerospike"
)

// AerospikeCache provides a high-level caching interface over the Aerospike key-value store.
//
// This struct abstracts operations such as Get, Set, Exists, Delete, Increment, Decrement,
// and TTL management by internally using an Aerospike client that adheres to the AerospikeInterface.
//
// Fields:
//   - aero: An implementation of the AerospikeInterface that handles low-level interactions
//     with the Aerospike database, including connection management and operations.
//   - namespace: A logical grouping of data in Aerospike. Similar to a database in relational systems.
//   - collection: Represents the 'set' within the namespace, used for organizing related records,
//     similar to a table in relational databases.
//
// Typical usage includes initializing this struct via NewAerospikeCache, and then using its
// methods to interact with the underlying Aerospike store in a type-safe, request-driven manner.
type AerospikeCache struct {
	aero       aerospike.AerospikeInterface
	namespace  string
	collection string
}

var (
	aerospikeConnect = aerospike.Connect
	aerospikeNewKey  = aerospike.NewKey
	aerospikeNewKeys = aerospike.NewMultipleKey
)

// NewAerospikeCache creates a new AerospikeCache instance using the provided options.
// It establishes a connection with the Aerospike server based on the specified cluster.
func NewAerospikeCache(opts *Options, connector aerospike.AerospikeConnectorInterface) (*AerospikeCache, error) {
	sanitizedTTL := uint32(opts.DefaultTTL)
	if opts.DefaultTTL < 0 {
		sanitizedTTL = aerospike.DefaultRecordTTL
	}

	aerospike.SetConnectionConfig(opts.Cluster, &aerospike.ConnectionConfig{
		Hosts:             opts.Hosts,
		ConnectionTimeout: opts.ConnTimeout,
		DefaultTTL:        sanitizedTTL,
	})

	if connector == nil {
		connector = &aerospike.AerospikeConnector{}
	}

	aero, err := aerospikeConnect(connector, opts.Cluster)

	if err != nil {
		return nil, err
	}

	aerocache := &AerospikeCache{
		aero:       aero,
		namespace:  opts.Namespace,
		collection: opts.Collection,
	}

	return aerocache, nil
}

// Exists checks if a given key exists in the Aerospike store using the provided ExistsRequest.
// Returns true if the key exists, otherwise false with an error if any occurs.
func (a *AerospikeCache) Exists(request *ExistsRequest) (bool, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return false, err
	}

	request, _ = verifiedRequest.(*ExistsRequest)

	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)
	if err != nil {
		return false, fmt.Errorf("failed to create aerospike key: %w", err)
	}

	exists, err := a.aero.GetClient().Exists(nil, aerospikeKey)
	if err != nil {
		return false, fmt.Errorf("aerospike::exists() failed: %w", err)
	}

	return exists, nil
}

// Get retrieves the value for a given key and optional fields from Aerospike using the provided GetRequest.
// Returns the bin map associated with the key, or an error if the retrieval fails.
func (a *AerospikeCache) Get(request *GetRequest) (any, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return false, err
	}

	request, _ = verifiedRequest.(*GetRequest)

	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)
	if err != nil {
		return false, fmt.Errorf("failed to create aerospike key: %w", err)
	}

	record, err := a.aero.GetClient().Get(nil, aerospikeKey, request.Fields...)
	if err != nil {
		return nil, fmt.Errorf("aerospike::get() failed: %w", err)
	}

	return map[string]any(record.Bins), nil
}

// Set stores the given fields in Aerospike under the specified key using the SetRequest.
// Returns true on successful set, or false and an error if the operation fails.
func (a *AerospikeCache) Set(request *SetRequest) (bool, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return false, err
	}

	request, _ = verifiedRequest.(*SetRequest)

	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)
	if err != nil {
		return false, fmt.Errorf("failed to create aerospike key: %w", err)
	}

	writePolicy := aerospike.GetClusterDefaultWritePolicy(request.Namespace, request.TTL)

	err = a.aero.GetClient().Put(writePolicy, aerospikeKey, request.Fields)
	if err != nil {
		return false, fmt.Errorf("aerospike::put() failed: %w", err)
	}

	return true, nil
}

// Delete removes a record from Aerospike for the specified key provided in the DeleteRequest.
// Returns true if the key was deleted, false otherwise along with an error.
func (a *AerospikeCache) Delete(request *DeleteRequest) (bool, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return false, err
	}

	request, _ = verifiedRequest.(*DeleteRequest)

	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)
	if err != nil {
		return false, fmt.Errorf("failed to create aerospike key: %w", err)
	}

	deleted, err := a.aero.GetClient().Delete(nil, aerospikeKey)
	if err != nil {
		return false, fmt.Errorf("aerospike::delete() failed: %w", err)
	}

	return deleted, nil
}

// MultiGet retrieves multiple records from Aerospike for a list of keys provided in the MultiGetRequest.
// Returns a map of key to bin map for each found key, or an error if the operation fails.
func (a *AerospikeCache) MultiGet(request *MultiGetRequest) (map[string]any, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return nil, err
	}

	request, _ = verifiedRequest.(*MultiGetRequest)

	keys, err := aerospikeNewKeys(request.Namespace, request.Collection, request.Keys)
	if err != nil {
		return nil, fmt.Errorf("failed to create aerospike keys: %w", err)
	}

	records, err := a.aero.GetClient().BatchGet(nil, keys)
	if err != nil {
		return nil, fmt.Errorf("aerospike::batchGet() failed: %w", err)
	}

	result := make(map[string]any)
	for i, record := range records {
		if record != nil {
			result[request.Keys[i]] = map[string]any(record.Bins)
		}
	}

	return result, nil
}

// MultiSet sets multiple records in Aerospike using the provided MultiSetRequest.
// Returns a map of key to boolean indicating whether each record was successfully written.
func (a *AerospikeCache) MultiSet(request *MultiSetRequest) (map[string]bool, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return nil, err
	}

	request, _ = verifiedRequest.(*MultiSetRequest)
	var responseMap = make(map[string]bool)

	var lastError error
	for k, fieldEntry := range request.FieldsMap {
		key, err := aerospikeNewKey(request.Namespace, request.Collection, k)
		if err != nil {
			return nil, fmt.Errorf("failed to create aerospike key: %w", err)
		}

		err = a.aero.GetClient().Put(nil, key, fieldEntry)
		if err == nil {
			responseMap[k] = true
		} else {
			lastError = err
			responseMap[k] = false
		}
	}

	if lastError != nil {
		return responseMap, fmt.Errorf("error in one or more write op, last error: %w", lastError)
	} else {
		return responseMap, nil
	}
}

// MultiDelete deletes multiple records in Aerospike based on the keys in the MultiDeleteRequest.
// Returns a map of key to boolean indicating whether each record was successfully deleted.
func (a *AerospikeCache) MultiDelete(request *MultiDeleteRequest) (map[string]bool, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return nil, err
	}

	request, _ = verifiedRequest.(*MultiDeleteRequest)
	var responseMap = make(map[string]bool)

	keys, err := aerospikeNewKeys(request.Namespace, request.Collection, request.Keys)
	if err != nil {
		return nil, fmt.Errorf("failed to create aerospike keys: %w", err)
	}

	records, err := a.aero.GetClient().BatchDelete(nil, nil, keys)
	if err != nil {
		return nil, err
	}

	for i, record := range records {
		if record != nil {
			responseMap[request.Keys[i]] = (record.ResultCode == types.OK)
		}
	}

	return responseMap, nil
}

// Increment performs atomic addition on one or more bins in a record specified in the IncrementRequest.
// Returns an error if any part of the operation fails.
func (a *AerospikeCache) Increment(request *IncrementRequest) error {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return err
	}

	request, _ = verifiedRequest.(*IncrementRequest)
	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)

	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}

	operations := []*aerov8.Operation{}
	for fieldKey, fieldVal := range request.Fields {
		operations = append(operations, aerov8.AddOp(aerospike.NewBin(fieldKey, fieldVal)))
	}

	_, err = a.aero.GetClient().Operate(nil, aerospikeKey, operations...)
	if err != nil {
		return fmt.Errorf("aerospike::Add() failed: %w", err)
	}

	return nil
}

// Decrement performs atomic subtraction by using AppendOp with negative values from the DecrementRequest.
// Returns an error if the operation fails.
func (a *AerospikeCache) Decrement(request *DecrementRequest) error {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return err
	}

	request, _ = verifiedRequest.(*DecrementRequest)
	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)

	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}

	operations := []*aerov8.Operation{}
	for fieldKey, fieldVal := range request.Fields {
		operations = append(operations, aerov8.AddOp(aerospike.NewBin(fieldKey, -1*fieldVal)))
	}

	_, err = a.aero.GetClient().Operate(nil, aerospikeKey, operations...)
	if err != nil {
		return fmt.Errorf("aerospike::Add() failed in decrement: %w", err)
	}

	return nil
}

// Append appends the given values to their corresponding bins in a record using the AppendRequest.
// Returns an error if the operation fails.
func (a *AerospikeCache) Append(request *AppendRequest) error {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return err
	}

	request, _ = verifiedRequest.(*AppendRequest)
	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)

	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}

	operations := []*aerov8.Operation{}
	for fieldKey, fieldVal := range request.Fields {
		operations = append(operations, aerov8.AppendOp(aerospike.NewBin(fieldKey, fieldVal)))
	}

	_, err = a.aero.GetClient().Operate(nil, aerospikeKey, operations...)
	if err != nil {
		return fmt.Errorf("aerospike::Append(): %w", err)
	}

	return nil
}

// GetTTL retrieves the TTL (time to live) of a record for a given key using the GetTTLRequest.
// Returns the TTL in seconds or an error if the retrieval fails.
func (a *AerospikeCache) GetTTL(request *GetTTLRequest) (int64, error) {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return 0, err
	}

	request, _ = verifiedRequest.(*GetTTLRequest)

	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)
	if err != nil {
		return 0, fmt.Errorf("failed to create aerospike key: %w", err)
	}

	record, err := a.aero.GetClient().GetHeader(nil, aerospikeKey)
	if err != nil {
		return 0, fmt.Errorf("aerospike::getHeader() failed: %w", err)
	}

	return int64(record.Expiration), nil
}

// SetTTL updates the TTL for a given key using the SetTTLRequest.
// Returns an error if the TTL update operation fails.
func (a *AerospikeCache) SetTTL(request *SetTTLRequest) error {
	verifiedRequest, err := a.verifyRequest(request)
	if err != nil {
		return err
	}

	request, _ = verifiedRequest.(*SetTTLRequest)

	aerospikeKey, err := aerospikeNewKey(request.Namespace, request.Collection, request.Key)
	if err != nil {
		return fmt.Errorf("failed to create aerospike key: %w", err)
	}

	policy := &aerov8.WritePolicy{
		Expiration: uint32(request.TTL),
	}

	err = a.aero.GetClient().Touch(policy, aerospikeKey)
	if err != nil {
		return fmt.Errorf("aerospike::touch() failed: %w", err)
	}

	return nil
}

// verifyRequest validates and enriches the provided CacheRequest with default namespace and collection.
// Returns the verified request or an error if required fields are missing.
func (a *AerospikeCache) verifyRequest(request CacheRequest) (CacheRequest, error) {
	if request == nil {
		return request, errors.New("error: input cache request struct is nil")
	}

	if request.GetNamespace() == "" {
		request.SetNamespace(a.namespace)

		if request.GetNamespace() == "" {
			return nil, errors.New("error: namespace field is empty")
		}
	}

	if request.GetCollection() == "" {
		request.SetCollection(a.collection)

		if request.GetCollection() == "" {
			return nil, errors.New("error: collection field is empty")
		}
	}

	return request, nil
}
