package encoding

import (
	"html"
	"testing"
)

func TestHTMLEscapeAndUnescape(t *testing.T) {
	original := `<div class="test">O'Reilly & Co.</div>`
	escaped := HTMLEscape(original)
	if escaped != htmlEscapeGo(original) { // html.EscapeString
		t.Errorf("HTMLEscape failed. Got %q", escaped)
	}

	unescaped := HTMLUnescape(escaped)
	if unescaped != original {
		t.Errorf("HTMLUnescape failed. Expected %q, got %q", original, unescaped)
	}
}

func TestStripHTMLTags(t *testing.T) {
	input := `<div>Hello <b>world</b>!</div>`
	expected := "Hello world!"
	out := StripHTMLTags(input)
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestIsHTMLEscaped(t *testing.T) {
	escaped := "&lt;p&gt;Example &amp; test&lt;/p&gt;"
	plain := "<p>Example & test</p>"
	if !IsHTMLEscaped(escaped) {
		t.Error("expected true for HTML-escaped string")
	}
	if IsHTMLEscaped(plain) {
		t.Error("expected false for non-escaped string")
	}
}

func TestContainsHTMLTags(t *testing.T) {
	inputWithTags := "<html><body>Test</body></html>"
	inputWithoutTags := "Just some text"
	if !ContainsHTMLTags(inputWithTags) {
		t.Error("expected true for HTML tag content")
	}
	if ContainsHTMLTags(inputWithoutTags) {
		t.Error("expected false for plain text")
	}
}

func TestEncodeSpecialChars(t *testing.T) {
	input := `"Hello" <world> & 'others'`
	expected := `&quot;Hello&quot; &lt;world&gt; &amp; &#39;others&#39;`
	out := EncodeSpecialChars(input)
	if out != expected {
		t.Errorf("EncodeSpecialChars failed. Expected %q, got %q", expected, out)
	}
}

func TestDecodeSpecialChars(t *testing.T) {
	input := `&quot;Hello&quot; &lt;world&gt; &amp; &#39;others&#39;`
	expected := `"Hello" <world> & 'others'`
	out := DecodeSpecialChars(input)
	if out != expected {
		t.Errorf("DecodeSpecialChars failed. Expected %q, got %q", expected, out)
	}
}

// Helper for validating against Go's html.EscapeString
func htmlEscapeGo(s string) string {
	return html.EscapeString(s)
}
