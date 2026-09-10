// html_utils.go
package encoding

import (
	"html"
	"strings"
)

// HTMLEscape escapes special characters in a string so it can be safely used in HTML.
// For example, '<' becomes '&lt;', '>' becomes '&gt;', etc.
func HTMLEscape(s string) string {
	return html.EscapeString(s)
}

// HTMLUnescape unescapes a string that contains HTML character entities back to its original form.
// For example, '&lt;' becomes '<', '&gt;' becomes '>', etc.
func HTMLUnescape(s string) string {
	return html.UnescapeString(s)
}

// StripHTMLTags removes all HTML tags from the input string.
// It does not parse or sanitize malformed HTML, but removes everything between '<' and '>'.
// This is a simple heuristic and may not be 100% accurate.
func StripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				result.WriteRune(r)
			}
		}
	}
	return result.String()
}

// IsHTMLEscaped checks if the string contains any HTML-escaped sequences.
// This is a simple heuristic and may not be 100% accurate.
func IsHTMLEscaped(s string) bool {
	return strings.Contains(s, "&lt;") || strings.Contains(s, "&gt;") ||
		strings.Contains(s, "&amp;") || strings.Contains(s, "&quot;") || strings.Contains(s, "&#39;")
}

// ContainsHTMLTags checks whether a string likely contains HTML tags.
// This is a simple heuristic and may not be 100% accurate.
func ContainsHTMLTags(s string) bool {
	return strings.Contains(s, "<") && strings.Contains(s, ">")
}

// EncodeSpecialChars replaces certain ASCII characters with their corresponding HTML entities.
// This is useful for escaping characters that have special meaning in HTML attributes.
func EncodeSpecialChars(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// DecodeSpecialChars replaces basic HTML entities back to their original characters.
// EscapeHTMLForAttribute escapes special characters in a string for use in HTML attributes.
func DecodeSpecialChars(s string) string {
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}
