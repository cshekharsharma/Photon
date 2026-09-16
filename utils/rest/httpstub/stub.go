// Package httpstub provides a mechanism to create and manage HTTP stubs for testing, development,
// and other scenarios where controlling outgoing HTTP responses is beneficial. It allows you to
// define stubbed responses identified by IDs, simulate latency, enforce maximum hit limits, and
// optionally return negative responses based on probabilities.
//
// The package maintains an in-memory registry of stubs represented by the StubEntry struct. Each
// stub defines how to respond to a given request scenario, including response codes, headers, and
// bodies. Consumers can add, remove, list, and retrieve these stubs. A global enabled/disabled flag
// allows switching the entire stubbing mechanism on or off.
//
// Typical usage includes:
//   - Initializing the stub system with InitStubConfig.
//   - Adding stubs via AddStub.
//   - Calling stubs by ID using CallStub, which returns an *http.Response or an error.
//   - Clearing or listing stubs as needed for test setup and teardown.
//
// This abstraction is particularly useful in testing or local development environments where
// external services might not be available or to simulate various response conditions without
// changing real external services.
package httpstub

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/utils/types"
)

const (
	HeaderContentType string = "Content-Type"
	ContentTypeJSON   string = "application/json"
)

// StubEntry represents a single HTTP stub configuration. Each stub determines how to respond
// when called by its unique ID via CallStub. Fields control the response code, headers, latency,
// probability of returning positive or negative responses, and other behaviors.
//
// A StubEntry can be partially specified. The framework applies sensible defaults (e.g., method=GET,
// response code=200) when AddStub is called. Users can also specify constraints like MaxHits to limit
// how many times a stub can be called before it starts returning errors.
type StubEntry struct {
	ID             string            // Unique identifier for this stub. If empty, generated automatically.
	Endpoint       string            // Informational endpoint field for reference. Does not affect CallStub logic.
	Method         string            // HTTP method associated with this stub (defaults to GET).
	QueryParams    map[string]string // Optional query parameters (informational).
	RequestHeaders map[string]string // Optional request headers (informational).
	BodyRegex      string            // Regex to match request body (informational).

	ResponseCode     int               // HTTP status code for the positive response.
	ResponseHeaders  map[string]string // Headers to return with the response.
	PositiveResponse string            // Response body when returning a positive response.
	ContentType      string            // Content-Type of the response (defaults to application/json).
	Latency          time.Duration     // Artificial latency to simulate slow responses.

	ErrorResponseCode int    // HTTP status code for the negative response scenario.
	NegativeResponse  string // Response body for negative response scenario.

	Probability float64 // Probability of returning a positive response (1.0 = always positive).
	MaxHits     int     // Maximum number of times this stub can be successfully hit. After that, errors occur.
	HitCount    int     // Internal counter of how many times this stub has been called.

	Labels map[string]string // Arbitrary key-value pairs for organizational or labeling purposes.
}

var (
	mutex       sync.RWMutex
	enableStubs bool                  = true
	httpStubs   map[string]*StubEntry = make(map[string]*StubEntry)
)

var cryptoRandRead = cryptorand.Read

// InitStubConfig initializes or resets the stub configuration. It clears all existing stubs and sets
// whether stubbing is enabled or disabled. If enabled is false, CallStub will return nil responses.
//
// Params:
//   - enable: If false, all calls to CallStub return nil, effectively disabling the stubbing mechanism.
func InitStubConfig(enable bool) {
	mutex.Lock()
	defer mutex.Unlock()

	httpStubs = make(map[string]*StubEntry)
	enableStubs = enable
}

// IsStubbingEnabled returns the current global stubbing status.
// If false, all calls to CallStub will return nil responses.
//
// Returns:
//   - true if stubbing is enabled, false otherwise.
//   - false if stubbing is disabled.
func IsStubbingEnabled() bool {
	return enableStubs
}

// AddStub adds a new stub entry to the registry, applying default values to missing fields. If no ID
// is provided, a random one is generated. The Endpoint field must not be empty.
//
// Returns:
//   - A string containing the stub ID.
//   - An error if the endpoint is empty.
//
// Example:
//
//	entry := &StubEntry{Endpoint: "/api/test", PositiveResponse: "OK"}
//	id, err := AddStub(entry)
func AddStub(entry *StubEntry) (string, error) {
	mutex.Lock()
	defer mutex.Unlock()

	entry = addDefaultValues(entry)

	if entry.Endpoint == "" {
		return "", errors.New("endpoint cannot be empty")
	}
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("stub-%s", types.GetRandomString(16))
	}

	httpStubs[entry.ID] = entry
	return entry.ID, nil
}

// RemoveStub removes the stub identified by the given ID. If no stub with that ID exists, returns an error.
//
// Params:
//   - id: The ID of the stub to remove.
//
// Returns:
//   - error if the stub does not exist, nil otherwise.
func RemoveStub(id string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := httpStubs[id]; !exists {
		return fmt.Errorf("stub with ID %s does not exist", id)
	}

	delete(httpStubs, id)
	return nil
}

// ListAllStubs returns a slice of pointers to all currently registered stubs. Useful for debugging
// or introspection. The order of stubs in the slice is not guaranteed.
func ListAllStubs() []*StubEntry {
	mutex.RLock()
	defer mutex.RUnlock()

	result := make([]*StubEntry, 0, len(httpStubs))
	for _, s := range httpStubs {
		result = append(result, s)
	}

	return result
}

// ClearAllStubs removes all stubs from the registry, resetting the stub map to empty. Typically used
// in test setups and teardowns.
func ClearAllStubs() {
	mutex.Lock()
	defer mutex.Unlock()
	httpStubs = make(map[string]*StubEntry)
}

// GetHttpStubByID returns the StubEntry pointer corresponding to the provided ID. If the ID is empty
// or does not exist in the registry, it returns an error.
//
// Params:
//   - id: The stub ID to look up.
//
// Returns:
//   - *StubEntry if found
//   - error if not found or if ID is empty
func GetHttpStubByID(id string) (*StubEntry, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	mutex.RLock()
	defer mutex.RUnlock()

	if stub, exists := httpStubs[id]; exists {
		return stub, nil
	}

	return nil, fmt.Errorf("stub with ID %s does not exist", id)
}

// CallStub simulates calling a stub by its ID. It constructs a response according to the stub's fields.
// If stubs are disabled globally (enableStubs = false), it returns nil with no error. If the stub is not
// found, returns an error. If the stub has a max hits limit and it's exceeded, returns an error.
//
// Latency is simulated if specified, and if Probability < 1.0 and NegativeResponse is set, there's a chance
// to return the negative response instead of the positive one.
//
// Params:
//   - ctx: Context used to support cancellation (if latency is specified and the context times out, returns context error).
//   - id: The stub ID to invoke.
//
// Returns:
//   - *http.Response constructed from the stub's configuration.
//   - error if stub not found, max hits exceeded, or context canceled.
func CallStub(ctx context.Context, id string) (*http.Response, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if !enableStubs {
		return nil, nil
	}

	stub, exists := httpStubs[id]
	if !exists {
		return nil, fmt.Errorf("stub with ID %s not found", id)
	}

	// Check MaxHits
	if stub.MaxHits > 0 && stub.HitCount >= stub.MaxHits {
		return nil, fmt.Errorf("stub with ID %s has reached max hits limit", id)
	}

	stub.HitCount++

	// Decide response type based on Probability
	usePositive := cryptoFloat64() <= stub.Probability
	respCode := stub.ResponseCode
	respBody := stub.PositiveResponse

	if !usePositive && stub.NegativeResponse != "" {
		respCode = stub.ErrorResponseCode
		respBody = stub.NegativeResponse
	}

	// Simulate latency
	if stub.Latency > 0 {
		select {
		case <-time.After(stub.Latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Construct response
	resp := &http.Response{
		StatusCode: respCode,
		Header:     make(http.Header),
	}
	for k, v := range stub.ResponseHeaders {
		resp.Header.Set(k, v)
	}

	if stub.ContentType != "" {
		resp.Header.Set(HeaderContentType, stub.ContentType)
	}
	resp.Body = io.NopCloser(bytes.NewBufferString(respBody))
	resp.Status = http.StatusText(respCode)

	return resp, nil
}

func cryptoFloat64() float64 {
	var buf [8]byte
	if _, err := cryptoRandRead(buf[:]); err != nil {
		return 1
	}
	return float64(binary.BigEndian.Uint64(buf[:])) / float64(^uint64(0))
}

// addDefaultValues fills in fields of StubEntry with defaults if they are zero-valued.
//
// Defaults:
//   - Method: GET
//   - ResponseCode: 200
//   - ContentType: application/json
//   - ResponseHeaders: empty map if nil
//   - Labels: empty map if nil
//   - If NegativeResponse is set but ErrorResponseCode is not, ErrorResponseCode defaults to 500.
func addDefaultValues(entry *StubEntry) *StubEntry {
	if entry.Method == "" {
		entry.Method = http.MethodGet
	}
	if entry.ResponseCode == 0 {
		entry.ResponseCode = http.StatusOK
	}
	if entry.ContentType == "" {
		entry.ContentType = ContentTypeJSON
	}
	if entry.ResponseHeaders == nil {
		entry.ResponseHeaders = make(map[string]string)
	}
	if entry.Labels == nil {
		entry.Labels = make(map[string]string)
	}

	if entry.NegativeResponse != "" && entry.ErrorResponseCode == 0 {
		entry.ErrorResponseCode = http.StatusInternalServerError
	}

	return entry
}
