package qdrant

import "testing"

func TestHTTPError(t *testing.T) {
	err := &HTTPError{StatusCode: 400, Body: "bad"}
	if got := err.Error(); got == "" || got[0] == 0 {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestMapHTTPStatus(t *testing.T) {
	if mapHTTPStatus(401) != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized")
	}
	if mapHTTPStatus(403) != ErrForbidden {
		t.Fatalf("expected ErrForbidden")
	}
	if mapHTTPStatus(404) != ErrNotFound {
		t.Fatalf("expected ErrNotFound")
	}
	if mapHTTPStatus(400) != ErrBadRequest {
		t.Fatalf("expected ErrBadRequest")
	}
	if mapHTTPStatus(409) != ErrConflict {
		t.Fatalf("expected ErrConflict")
	}
	if mapHTTPStatus(429) != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited")
	}
	if mapHTTPStatus(500) != ErrServer {
		t.Fatalf("expected ErrServer")
	}
	if mapHTTPStatus(418) != nil {
		t.Fatalf("expected nil for non-mapped non-5xx status")
	}
}
