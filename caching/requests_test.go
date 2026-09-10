package caching

import "testing"

func TestCacheRequest_GetNamespace(t *testing.T) {
	req := &cacheRequest{Namespace: "test-ns"}
	if got := req.GetNamespace(); got != "test-ns" {
		t.Errorf("GetNamespace() = %v, want %v", got, "test-ns")
	}
}

func TestCacheRequest_GetCollection(t *testing.T) {
	req := &cacheRequest{Collection: "test-coll"}
	if got := req.GetCollection(); got != "test-coll" {
		t.Errorf("GetCollection() = %v, want %v", got, "test-coll")
	}
}

func TestCacheRequest_SetNamespace(t *testing.T) {
	req := &cacheRequest{}
	req.SetNamespace("new-ns")
	if req.Namespace != "new-ns" {
		t.Errorf("SetNamespace() failed, got: %v, want: %v", req.Namespace, "new-ns")
	}
}

func TestCacheRequest_SetCollection(t *testing.T) {
	req := &cacheRequest{}
	req.SetCollection("new-coll")
	if req.Collection != "new-coll" {
		t.Errorf("SetCollection() failed, got: %v, want: %v", req.Collection, "new-coll")
	}
}
