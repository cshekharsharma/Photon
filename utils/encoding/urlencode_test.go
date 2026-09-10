package encoding

import (
	"net/url"
	"testing"
)

func TestURLEncode(t *testing.T) {
	input := "hello world@2025"
	expected := url.QueryEscape(input)
	if out := URLEncode(input); out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestURLDecode(t *testing.T) {
	input := "hello%20world%402025"
	expected := "hello world@2025"
	out, err := URLDecode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestURLDecode_Invalid(t *testing.T) {
	input := "%zzinvalid"
	_, err := URLDecode(input)
	if err == nil {
		t.Error("expected error for invalid URL encoding, got nil")
	}
}
