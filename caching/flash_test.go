package caching

import (
	"errors"
	"testing"
	"time"

	"github.com/cshekharsharma/photon/storage/flashdb"
)

// Fake flashdb.Store implementation for testing
type fakeStore struct {
	data map[string]any
	ttls map[string]time.Duration

	// error injections
	existsErr   error
	getErr      error
	setErr      error
	deleteErr   error
	multiGetErr error
	multiDelErr error
	incrErr     error
	decrErr     error
	appendErr   error
	getTTLErr   error
	setTTLErr   error

	// call tracking
	lastSetKey   string
	lastSetValue map[string]any
	lastSetTTL   time.Duration

	lastIncKey    string
	lastIncFields map[string]int64

	lastDecKey    string
	lastDecFields map[string]int64

	lastAppendKey    string
	lastAppendFields map[string]string

	closed bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		data: make(map[string]any),
		ttls: make(map[string]time.Duration),
	}
}

func (f *fakeStore) Close() {
	f.closed = true
}

func (f *fakeStore) Exists(key string) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	_, ok := f.data[key]
	return ok, nil
}

func (f *fakeStore) Get(key string) (any, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	v, ok := f.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (f *fakeStore) Set(key string, value map[string]any, ttl time.Duration) (bool, error) {
	f.lastSetKey = key
	f.lastSetValue = value
	f.lastSetTTL = ttl

	if f.setErr != nil {
		return false, f.setErr
	}
	f.data[key] = value
	f.ttls[key] = ttl
	return true, nil
}

func (f *fakeStore) Delete(key string) (bool, error) {
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	_, ok := f.data[key]
	delete(f.data, key)
	delete(f.ttls, key)
	return ok, nil
}

func (f *fakeStore) MultiGet(keys []string) (map[string]any, error) {
	if f.multiGetErr != nil {
		return nil, f.multiGetErr
	}
	out := make(map[string]any, len(keys))
	for _, k := range keys {
		if v, ok := f.data[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}

func (f *fakeStore) MultiDelete(keys []string) (map[string]bool, error) {
	if f.multiDelErr != nil {
		return nil, f.multiDelErr
	}
	out := make(map[string]bool, len(keys))
	for _, k := range keys {
		_, ok := f.data[k]
		delete(f.data, k)
		delete(f.ttls, k)
		out[k] = ok
	}
	return out, nil
}

func (f *fakeStore) Increment(key string, fields map[string]int64) error {
	if f.incrErr != nil {
		return f.incrErr
	}
	f.lastIncKey = key
	f.lastIncFields = fields
	return nil
}

func (f *fakeStore) Decrement(key string, fields map[string]int64) error {
	if f.decrErr != nil {
		return f.decrErr
	}
	f.lastDecKey = key
	f.lastDecFields = fields
	return nil
}

func (f *fakeStore) Append(key string, fields map[string]string) error {
	if f.appendErr != nil {
		return f.appendErr
	}
	f.lastAppendKey = key
	f.lastAppendFields = fields
	return nil
}

func (f *fakeStore) GetTTL(key string) (int64, error) {
	if f.getTTLErr != nil {
		return 0, f.getTTLErr
	}
	if ttl, ok := f.ttls[key]; ok {
		return int64(ttl.Seconds()), nil
	}
	return 0, nil
}

func (f *fakeStore) SetTTL(key string, ttl time.Duration) error {
	if f.setTTLErr != nil {
		return f.setTTLErr
	}
	f.ttls[key] = ttl
	return nil
}

// helper to create a cache with fake store
func newTestCache(fs *fakeStore, defaultTTL time.Duration) *FlashDBCache {
	return &FlashDBCache{
		store:      fs,
		defaultTTL: defaultTTL,
	}
}

func TestSecondsToDuration(t *testing.T) {
	if got := secondsToDuration(-5); got != 0 {
		t.Fatalf("expected 0 for negative seconds, got %v", got)
	}
	if got := secondsToDuration(0); got != 0 {
		t.Fatalf("expected 0 for zero seconds, got %v", got)
	}
	if got := secondsToDuration(5); got != 5*time.Second {
		t.Fatalf("expected 5s, got %v", got)
	}
}

func TestComposeKeyEmptyKey(t *testing.T) {
	if _, err := composeKey("ns", "coll", ""); err == nil {
		t.Fatalf("expected error for empty key")
	}
}

func TestComposeKeyVariants(t *testing.T) {
	tests := []struct {
		name string
		ns   string
		coll string
		key  string
		want string
	}{
		{"NoNsNoColl", "", "", "k", "k"},
		{"OnlyColl", "", "c", "k", "c::k"},
		{"OnlyNs", "n", "", "k", "n::k"},
		{"NsAndColl", "n", "c", "k", "n::c::k"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := composeKey(tt.ns, tt.coll, tt.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("composeKey(%q,%q,%q) = %q, want %q", tt.ns, tt.coll, tt.key, got, tt.want)
			}
		})
	}
}

func TestExistsNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.Exists(nil); err == nil {
		t.Fatalf("expected error for nil ExistsRequest")
	}
}

func TestExistsHappyPath(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k"] = 123
	c := newTestCache(fs, 0)

	req := &ExistsRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ok, err := c.Exists(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected key to exist")
	}
}

func TestExistsComposeKeyError(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(string, string, string) (string, error) {
		return "", errors.New("bad key")
	}

	c := newTestCache(newFakeStore(), 0)
	req := &ExistsRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")
	_, err := c.Exists(req)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestExistsStoreError(t *testing.T) {
	fs := newFakeStore()
	fs.existsErr = errors.New("boom")
	c := newTestCache(fs, 0)

	req := &ExistsRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.Exists(req)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestGetNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.Get(nil); err == nil {
		t.Fatalf("expected error for nil GetRequest")
	}
}

func TestGetComposeKeyError(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)

	req := &GetRequest{Key: ""}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.Get(req)
	if err == nil {
		t.Fatalf("expected error for empty key")
	}
}

func TestGetStoreError(t *testing.T) {
	fs := newFakeStore()
	fs.getErr = errors.New("boom")
	c := newTestCache(fs, 0)

	req := &GetRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.Get(req)
	if err == nil {
		t.Fatalf("expected store error")
	}
}

func TestGetNoFieldFiltering(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k"] = map[string]any{"a": 1}
	c := newTestCache(fs, 0)

	req := &GetRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	val, err := c.Get(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := val.(map[string]any)
	if !ok || m["a"] != 1 {
		t.Fatalf("unexpected value: %#v", val)
	}
}

func TestGetWithFieldFilteringMapValue(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k"] = map[string]any{"a": 1, "b": 2}
	c := newTestCache(fs, 0)

	req := &GetRequest{
		Key:    "k",
		Fields: []string{"a"},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	val, err := c.Get(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	if len(m) != 1 || m["a"] != 1 {
		t.Fatalf("unexpected filtered map: %#v", m)
	}
}

func TestGetWithFieldFilteringNonMapValue(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k"] = "plain"
	c := newTestCache(fs, 0)

	req := &GetRequest{
		Key:    "k",
		Fields: []string{"a"},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	val, err := c.Get(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "plain" {
		t.Fatalf("expected original value, got %#v", val)
	}
}

func TestSetNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.Set(nil); err == nil {
		t.Fatalf("expected error for nil SetRequest")
	}
}

func TestSetComposeKeyError(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)

	req := &SetRequest{Key: ""}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.Set(req)
	if err == nil {
		t.Fatalf("expected error for empty key")
	}
}

func TestSetWithFieldsPreferred(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 10*time.Second)

	req := &SetRequest{
		Key:    "k",
		Fields: map[string]any{"a": 1},
		TTL:    5, // should override defaultTTL
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ok, err := c.Set(req)
	if err != nil || !ok {
		t.Fatalf("unexpected error/ok: %v %v", ok, err)
	}
	if fs.lastSetKey != "ns::coll::k" {
		t.Fatalf("unexpected key: %s", fs.lastSetKey)
	}
	if fs.lastSetValue["a"] != 1 {
		t.Fatalf("unexpected payload: %#v", fs.lastSetValue)
	}
	if fs.lastSetTTL != 5*time.Second {
		t.Fatalf("unexpected TTL: %v", fs.lastSetTTL)
	}
}

func TestSetWithValueMap(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	valMap := map[string]any{"x": 42}
	req := &SetRequest{
		Key:   "k",
		Value: valMap,
		TTL:   0, // should fall back to defaultTTL (0 here, so pass 0)
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ok, err := c.Set(req)
	if err != nil || !ok {
		t.Fatalf("unexpected error/ok: %v %v", ok, err)
	}
	if fs.lastSetValue["x"] != 42 {
		t.Fatalf("unexpected payload: %#v", fs.lastSetValue)
	}
}

func TestSetWithValueNonMap(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	req := &SetRequest{
		Key:   "k",
		Value: "hello",
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ok, err := c.Set(req)
	if err != nil || !ok {
		t.Fatalf("unexpected error/ok: %v %v", ok, err)
	}
	if fs.lastSetValue["bin"] != "hello" {
		t.Fatalf("expected wrapped bin value, got %#v", fs.lastSetValue)
	}
}

func TestSetWithNoFieldsNoValue(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 3*time.Second)

	req := &SetRequest{
		Key: "k",
		TTL: 0, // uses defaultTTL
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ok, err := c.Set(req)
	if err != nil || !ok {
		t.Fatalf("unexpected error/ok: %v %v", ok, err)
	}
	if len(fs.lastSetValue) != 0 {
		t.Fatalf("expected empty map, got %#v", fs.lastSetValue)
	}
	if fs.lastSetTTL != 3*time.Second {
		t.Fatalf("expected default TTL, got %v", fs.lastSetTTL)
	}
}

func TestDeleteNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.Delete(nil); err == nil {
		t.Fatalf("expected error for nil DeleteRequest")
	}
}

func TestDeleteComposeKeyError(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)

	req := &DeleteRequest{Key: ""}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.Delete(req)
	if err == nil {
		t.Fatalf("expected error for empty key")
	}
}

func TestDeleteHappyPath(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k"] = 1
	c := newTestCache(fs, 0)

	req := &DeleteRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ok, err := c.Delete(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected delete to return true")
	}
	if _, exists := fs.data["ns::coll::k"]; exists {
		t.Fatalf("expected key to be deleted")
	}
}

func TestMultiGetNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.MultiGet(nil); err == nil {
		t.Fatalf("expected error for nil MultiGetRequest")
	}
}

func TestMultiGetEmptyKeys(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiGet(&MultiGetRequest{Keys: []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %#v", out)
	}
}

func TestMultiGetAllInvalidKeys(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiGet(&MultiGetRequest{Keys: []string{""}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %#v", out)
	}
}

func TestMultiGetComposeKeyErrorSkipped(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(ns, coll, key string) (string, error) {
		if key == "bad" {
			return "", errors.New("bad key")
		}
		return composeKey(ns, coll, key)
	}

	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiGet(&MultiGetRequest{Keys: []string{"bad"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %#v", out)
	}
}

func TestMultiGetStoreError(t *testing.T) {
	fs := newFakeStore()
	fs.multiGetErr = errors.New("boom")
	c := newTestCache(fs, 0)

	req := &MultiGetRequest{Keys: []string{"k"}}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.MultiGet(req)
	if err == nil {
		t.Fatalf("expected error from MultiGet")
	}
}

func TestMultiGetHappyPath(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k1"] = 1
	fs.data["ns::coll::k2"] = 2
	c := newTestCache(fs, 0)

	req := &MultiGetRequest{
		Keys: []string{"k1", "k2"},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiGet(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 || out["k1"] != 1 || out["k2"] != 2 {
		t.Fatalf("unexpected output: %#v", out)
	}
}

func TestMultiSetNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.MultiSet(nil); err == nil {
		t.Fatalf("expected error for nil MultiSetRequest")
	}
}

func TestMultiSetNothingToDo(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiSet(&MultiSetRequest{
		FieldsMap: nil,
		ValueMap:  nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty map, got %#v", out)
	}
}

func TestMultiSetFieldsAndValues(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 5*time.Second)

	req := &MultiSetRequest{
		TTL: 0,
		FieldsMap: map[string]map[string]any{
			"key1": {"a": 1},
			"":     {"b": 2}, // invalid key
		},
		ValueMap: map[string]any{
			"key1": "ignored", // overshadowed by FieldsMap
			"key2": map[string]any{"x": 10},
			"key3": "scalar",
			"":     "invalid", // invalid key
		},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiSet(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !out["key1"] {
		t.Fatalf("expected key1 to be true")
	}
	if v, ok := out[""]; ok && v {
		t.Fatalf("expected empty key to be false")
	}

	if v, ok := fs.data["ns::coll::key2"].(map[string]any); !ok || v["x"] != 10 {
		t.Fatalf("unexpected stored value for key2: %#v", fs.data["ns::coll::key2"])
	}

	if v, ok := fs.data["ns::coll::key3"].(map[string]any); !ok || v["bin"] != "scalar" {
		t.Fatalf("unexpected stored value for key3: %#v", fs.data["ns::coll::key3"])
	}

	if fs.ttls["ns::coll::key1"] != 5*time.Second {
		t.Fatalf("expected default TTL, got %v", fs.ttls["ns::coll::key1"])
	}
}

func TestMultiSetNilPayload(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	req := &MultiSetRequest{
		FieldsMap: map[string]map[string]any{
			"key1": nil,
		},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiSet(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out["key1"] {
		t.Fatalf("expected key1 to be true")
	}
	if len(fs.data["ns::coll::key1"].(map[string]any)) != 0 {
		t.Fatalf("expected empty payload")
	}
}

func TestMultiSetComposeKeyError(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(ns, coll, key string) (string, error) {
		return "", errors.New("bad key")
	}

	fs := newFakeStore()
	c := newTestCache(fs, 0)

	req := &MultiSetRequest{
		FieldsMap: map[string]map[string]any{
			"key1": {"a": 1},
		},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.MultiSet(req)
	if err == nil {
		t.Fatalf("expected error from MultiSet")
	}
}

func TestMultiSetValueMapEmptyKey(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	req := &MultiSetRequest{
		ValueMap: map[string]any{
			"": "bad",
		},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiSet(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out[""] {
		t.Fatalf("expected empty key to be false")
	}
}

func TestMultiSetValueMapComposeKeyError(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(ns, coll, key string) (string, error) {
		if key == "bad" {
			return "", errors.New("bad key")
		}
		return composeKey(ns, coll, key)
	}

	fs := newFakeStore()
	c := newTestCache(fs, 0)

	req := &MultiSetRequest{
		ValueMap: map[string]any{
			"bad": "v",
		},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiSet(req)
	if err == nil {
		t.Fatalf("expected error")
	}
	if out["bad"] {
		t.Fatalf("expected bad key to be false")
	}
}

func TestMultiSetWithStoreErrorSetsFirstErr(t *testing.T) {
	fs := newFakeStore()
	fs.setErr = errors.New("set failed")
	c := newTestCache(fs, 0)

	req := &MultiSetRequest{
		TTL: 1,
		FieldsMap: map[string]map[string]any{
			"key1": {"a": 1},
		},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiSet(req)
	if err == nil {
		t.Fatalf("expected error from MultiSet")
	}
	if ok := out["key1"]; ok {
		t.Fatalf("expected key1 result to be false due to error")
	}
}

func TestMultiDeleteNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.MultiDelete(nil); err == nil {
		t.Fatalf("expected error for nil MultiDeleteRequest")
	}
}

func TestMultiDeleteEmptyKeys(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiDelete(&MultiDeleteRequest{Keys: []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty result, got %#v", out)
	}
}

func TestMultiDeleteAllInvalidKeys(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiDelete(&MultiDeleteRequest{Keys: []string{""}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty result, got %#v", out)
	}
}

func TestMultiDeleteComposeKeyErrorSkipped(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(ns, coll, key string) (string, error) {
		if key == "bad" {
			return "", errors.New("bad key")
		}
		return composeKey(ns, coll, key)
	}

	c := newTestCache(newFakeStore(), 0)
	out, err := c.MultiDelete(&MultiDeleteRequest{Keys: []string{"bad"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected empty result, got %#v", out)
	}
}

func TestMultiDeleteStoreError(t *testing.T) {
	fs := newFakeStore()
	fs.multiDelErr = errors.New("boom")
	c := newTestCache(fs, 0)

	req := &MultiDeleteRequest{Keys: []string{"k"}}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.MultiDelete(req)
	if err == nil {
		t.Fatalf("expected error from MultiDelete")
	}
}

func TestMultiDeleteHappyPath(t *testing.T) {
	fs := newFakeStore()
	fs.data["ns::coll::k1"] = 1
	fs.data["ns::coll::k2"] = 2
	c := newTestCache(fs, 0)

	req := &MultiDeleteRequest{
		Keys: []string{"k1", "k2"},
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	out, err := c.MultiDelete(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 || !out["k1"] || !out["k2"] {
		t.Fatalf("unexpected delete map: %#v", out)
	}
	if _, ok := fs.data["ns::coll::k1"]; ok {
		t.Fatalf("expected k1 to be deleted")
	}
}

func TestIncrementNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if err := c.Increment(nil); err == nil {
		t.Fatalf("expected error for nil IncrementRequest")
	}
}

func TestIncrementWithFields(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	incReq := &IncrementRequest{
		Key:    "k",
		Fields: map[string]int64{"a": 1},
	}
	incReq.SetNamespace("ns")
	incReq.SetCollection("coll")

	err := c.Increment(incReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.lastIncKey != "ns::coll::k" {
		t.Fatalf("unexpected key: %s", fs.lastIncKey)
	}
	if fs.lastIncFields["a"] != 1 {
		t.Fatalf("unexpected fields: %#v", fs.lastIncFields)
	}
}

func TestIncrementComposeKeyError(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(string, string, string) (string, error) {
		return "", errors.New("bad key")
	}

	c := newTestCache(newFakeStore(), 0)
	req := &IncrementRequest{Key: "k", Fields: map[string]int64{"a": 1}}
	req.SetNamespace("ns")
	req.SetCollection("coll")
	if err := c.Increment(req); err == nil {
		t.Fatalf("expected error")
	}
}

func TestIncrementWithValue(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	incReq := &IncrementRequest{
		Key:   "k",
		Value: 5,
	}
	incReq.SetNamespace("ns")
	incReq.SetCollection("coll")

	err := c.Increment(incReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.lastIncFields["bin"] != 5 {
		t.Fatalf("expected bin=5, got %#v", fs.lastIncFields)
	}
}

func TestIncrementNoOp(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	incReq := &IncrementRequest{
		Key: "k",
		// no Fields, Value == 0
	}
	incReq.SetNamespace("ns")
	incReq.SetCollection("coll")

	err := c.Increment(incReq)
	if err != nil {
		t.Fatalf("expected nil error for no-op increment")
	}
	if fs.lastIncKey != "" {
		t.Fatalf("expected no call to Increment, got key %s", fs.lastIncKey)
	}
}

func TestDecrementNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if err := c.Decrement(nil); err == nil {
		t.Fatalf("expected error for nil DecrementRequest")
	}
}

func TestDecrementWithFields(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	decReq := &DecrementRequest{
		Key:    "k",
		Fields: map[string]int64{"a": 1},
	}
	decReq.SetNamespace("ns")
	decReq.SetCollection("coll")

	err := c.Decrement(decReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.lastDecKey != "ns::coll::k" {
		t.Fatalf("unexpected key: %s", fs.lastDecKey)
	}
}

func TestDecrementComposeKeyError(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(string, string, string) (string, error) {
		return "", errors.New("bad key")
	}

	c := newTestCache(newFakeStore(), 0)
	req := &DecrementRequest{Key: "k", Fields: map[string]int64{"a": 1}}
	req.SetNamespace("ns")
	req.SetCollection("coll")
	if err := c.Decrement(req); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDecrementWithValue(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	decReq := &DecrementRequest{
		Key:   "k",
		Value: 3,
	}
	decReq.SetNamespace("ns")
	decReq.SetCollection("coll")

	err := c.Decrement(decReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.lastDecFields["bin"] != 3 {
		t.Fatalf("expected bin=3, got %#v", fs.lastDecFields)
	}
}

func TestDecrementNoOp(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	decReq := &DecrementRequest{
		Key: "k",
	}
	decReq.SetNamespace("ns")
	decReq.SetCollection("coll")

	err := c.Decrement(decReq)
	if err != nil {
		t.Fatalf("expected nil error for no-op decrement")
	}
	if fs.lastDecKey != "" {
		t.Fatalf("expected no call to Decrement, got key %s", fs.lastDecKey)
	}
}

func TestAppendNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if err := c.Append(nil); err == nil {
		t.Fatalf("expected error for nil AppendRequest")
	}
}

func TestAppendWithFields(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	appReq := &AppendRequest{
		Key:    "k",
		Fields: map[string]string{"a": "x"},
	}
	appReq.SetNamespace("ns")
	appReq.SetCollection("coll")

	err := c.Append(appReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.lastAppendKey != "ns::coll::k" {
		t.Fatalf("unexpected key: %s", fs.lastAppendKey)
	}
	if fs.lastAppendFields["a"] != "x" {
		t.Fatalf("unexpected fields: %#v", fs.lastAppendFields)
	}
}

func TestAppendComposeKeyError(t *testing.T) {
	orig := composeKeyFn
	t.Cleanup(func() { composeKeyFn = orig })
	composeKeyFn = func(string, string, string) (string, error) {
		return "", errors.New("bad key")
	}

	c := newTestCache(newFakeStore(), 0)
	req := &AppendRequest{Key: "k", Fields: map[string]string{"a": "b"}}
	req.SetNamespace("ns")
	req.SetCollection("coll")
	if err := c.Append(req); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAppendWithValue(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	appReq := &AppendRequest{
		Key:   "k",
		Value: "v",
	}
	appReq.SetNamespace("ns")
	appReq.SetCollection("coll")

	err := c.Append(appReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.lastAppendFields["bin"] != "v" {
		t.Fatalf("expected bin=v, got %#v", fs.lastAppendFields)
	}
}

func TestAppendNoOp(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	appReq := &AppendRequest{
		Key: "k",
	}
	appReq.SetNamespace("ns")
	appReq.SetCollection("coll")

	err := c.Append(appReq)
	if err != nil {
		t.Fatalf("expected nil error for no-op append")
	}
	if fs.lastAppendKey != "" {
		t.Fatalf("expected no call to Append, got key %s", fs.lastAppendKey)
	}
}

func TestGetTTLNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if _, err := c.GetTTL(nil); err == nil {
		t.Fatalf("expected error for nil GetTTLRequest")
	}
}

func TestGetTTLComposeKeyError(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)

	req := &GetTTLRequest{Key: ""}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	_, err := c.GetTTL(req)
	if err == nil {
		t.Fatalf("expected error for empty key")
	}
}

func TestGetTTLHappyPath(t *testing.T) {
	fs := newFakeStore()
	fs.ttls["ns::coll::k"] = 7 * time.Second
	c := newTestCache(fs, 0)

	req := &GetTTLRequest{Key: "k"}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	ttl, err := c.GetTTL(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl != 7 {
		t.Fatalf("expected TTL 7, got %d", ttl)
	}
}

func TestSetTTLNilRequest(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)
	if err := c.SetTTL(nil); err == nil {
		t.Fatalf("expected error for nil SetTTLRequest")
	}
}

func TestSetTTLComposeKeyError(t *testing.T) {
	c := newTestCache(newFakeStore(), 0)

	req := &SetTTLRequest{Key: ""}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	err := c.SetTTL(req)
	if err == nil {
		t.Fatalf("expected error for empty key")
	}
}

func TestSetTTLWithExplicitTTL(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)

	req := &SetTTLRequest{
		Key: "k",
		TTL: 10,
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	err := c.SetTTL(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.ttls["ns::coll::k"] != 10*time.Second {
		t.Fatalf("expected 10s TTL, got %v", fs.ttls["ns::coll::k"])
	}
}

func TestSetTTLWithDefaultTTL(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 4*time.Second)

	req := &SetTTLRequest{
		Key: "k",
		TTL: 0, // should use default TTL
	}
	req.SetNamespace("ns")
	req.SetCollection("coll")

	err := c.SetTTL(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fs.ttls["ns::coll::k"] != 4*time.Second {
		t.Fatalf("expected default TTL, got %v", fs.ttls["ns::coll::k"])
	}
}

func TestClose(t *testing.T) {
	fs := newFakeStore()
	c := newTestCache(fs, 0)
	c.Close()
	if !fs.closed {
		t.Fatalf("expected Close to call store.Close()")
	}
}

func TestNewFlashDBCacheHappyPath(t *testing.T) {
	cache, err := NewFlashDBCache(Options{DefaultTTL: 15})
	if err != nil {
		t.Fatalf("unexpected error from NewFlashDBCache: %v", err)
	}
	if cache == nil {
		t.Fatalf("expected non-nil cache")
	}
	if cache.defaultTTL != 15*time.Second {
		t.Fatalf("expected defaultTTL=15s, got %v", cache.defaultTTL)
	}
	cache.Close()
}

func TestNewFlashDBCache_Error(t *testing.T) {
	orig := flashdbNew
	t.Cleanup(func() { flashdbNew = orig })
	flashdbNew = func(opts flashdb.Options) (*flashdb.Store, error) {
		return nil, errors.New("boom")
	}

	_, err := NewFlashDBCache(Options{DefaultTTL: 1})
	if err == nil {
		t.Fatalf("expected error from NewFlashDBCache")
	}
}
