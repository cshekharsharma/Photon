package rest

import (
	"reflect"
	"testing"
)

func TestEncodePathSegment(t *testing.T) {
	input := "a/b"
	expected := "a%2Fb"
	if out := EncodePathSegment(input); out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestDecodePathSegment(t *testing.T) {
	input := "a%2Fb"
	expected := "a/b"
	out, err := DecodePathSegment(input)
	if err != nil || out != expected {
		t.Errorf("expected %q, got %q (err: %v)", expected, out, err)
	}
}

func TestBuildQuery(t *testing.T) {
	input := map[string]string{"q": "golang", "page": "1"}
	out := BuildQuery(input)
	if out != "page=1&q=golang" && out != "q=golang&page=1" { // map iteration is unordered
		t.Errorf("unexpected query string: %q", out)
	}
}

func TestParseQuery(t *testing.T) {
	input := "q=golang&page=1"
	expected := map[string]string{"q": "golang", "page": "1"}
	out, err := ParseQuery(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(out, expected) {
		t.Errorf("expected %v, got %v", expected, out)
	}
}

func TestParseQuery_Invalid(t *testing.T) {
	_, err := ParseQuery("%%invalid%%")
	if err == nil {
		t.Error("expected error for malformed query string, got nil")
	}
}

func TestNormalizeURL(t *testing.T) {
	input := "https://example.com/page#section"
	expected := "https://example.com/page"
	out, err := NormalizeURL(input)
	if err != nil || out != expected {
		t.Errorf("expected %q, got %q (err: %v)", expected, out, err)
	}
}

func TestNormalizeURL_Invalid(t *testing.T) {
	_, err := NormalizeURL("://invalid-url")
	if err == nil {
		t.Error("expected error for malformed URL, got nil")
	}
}

func TestIsValidURL(t *testing.T) {
	valid := "https://example.com"
	invalid := "not a url"
	if !IsValidURL(valid) {
		t.Errorf("expected valid URL for %q", valid)
	}
	if IsValidURL(invalid) {
		t.Errorf("expected invalid URL for %q", invalid)
	}
}
