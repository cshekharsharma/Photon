package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/cshekharsharma/photon/utils/rest"
	"github.com/cshekharsharma/photon/utils/rest/minifier"
)

func TestApiMinifier_SetsHeader_WhenEnabledAndRouteMatches(t *testing.T) {
	minifier.SetupApiKeyMinifierConfig(true, &sync.Map{})
	t.Cleanup(func() {
		minifier.SetupApiKeyMinifierConfig(false, &sync.Map{})
	})

	routes := map[string]bool{
		"/v1/orders": true,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := ApiMinifier(routes)(next)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/v1//orders", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got := rr.Header().Get(rest.HeaderXApiMinifier); got != rest.XAPIMinifierValue {
		t.Fatalf("expected %q header=%q, got %q", rest.HeaderXApiMinifier, rest.XAPIMinifierValue, got)
	}
}

func TestApiMinifier_DoesNotSetHeader_WhenDisabledEvenIfRouteMatches(t *testing.T) {
	minifier.SetupApiKeyMinifierConfig(false, &sync.Map{})

	routes := map[string]bool{
		"/v1/orders": true,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := ApiMinifier(routes)(next)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/v1/orders", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got := rr.Header().Get(rest.HeaderXApiMinifier); got != "" {
		t.Fatalf("expected %q header to be absent when disabled, got %q", rest.HeaderXApiMinifier, got)
	}
}

func TestApiMinifier_DoesNotSetHeader_WhenRouteNotEnabled(t *testing.T) {
	minifier.SetupApiKeyMinifierConfig(true, &sync.Map{})
	t.Cleanup(func() {
		minifier.SetupApiKeyMinifierConfig(false, &sync.Map{})
	})

	routes := map[string]bool{
		"/v1/orders": false,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := ApiMinifier(routes)(next)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/v1/orders", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got := rr.Header().Get(rest.HeaderXApiMinifier); got != "" {
		t.Fatalf("expected %q header to be absent when route disabled, got %q", rest.HeaderXApiMinifier, got)
	}
}

func TestApiMinifier_UsesCleanedPathForLookup(t *testing.T) {
	minifier.SetupApiKeyMinifierConfig(true, &sync.Map{})
	t.Cleanup(func() {
		minifier.SetupApiKeyMinifierConfig(false, &sync.Map{})
	})

	routes := map[string]bool{
		"/v1/orders": true,
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := ApiMinifier(routes)(next)

	// Clean("/v1/../v1/orders") == "/v1/orders"
	req := httptest.NewRequest(http.MethodGet, "http://example.com/v1/../v1/orders", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got := rr.Header().Get(rest.HeaderXApiMinifier); got != rest.XAPIMinifierValue {
		t.Fatalf("expected %q header=%q, got %q", rest.HeaderXApiMinifier, rest.XAPIMinifierValue, got)
	}
}
