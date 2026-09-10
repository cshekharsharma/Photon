package encoding

import "net/url"

// URLEncode encodes a string for use in a URL query parameter.
// It replaces special characters with their percent-encoded equivalents.
func URLEncode(input string) string {
	return url.QueryEscape(input)
}

// URLDecode decodes a percent-encoded string back to its original form.
// It reverses the effect of URLEncode.
// If the input is not valid percent-encoded data, it returns an error.
func URLDecode(input string) (string, error) {
	return url.QueryUnescape(input)
}
