package rest

import "net/url"

// EncodePathSegment safely escapes a path segment so it can be used
// in a URL without affecting its structure.
// For example, "a/b" becomes "a%2Fb".
func EncodePathSegment(segment string) string {
	return url.PathEscape(segment)
}

// DecodePathSegment unescapes a previously encoded path segment string.
// It reverses the effect of EncodePathSegment.
func DecodePathSegment(segment string) (string, error) {
	return url.PathUnescape(segment)
}

// BuildQuery takes a map of key-value pairs and constructs a URL-encoded
// query string. Keys and values are automatically escaped.
// Example: map{"q": "golang", "page": "1"} → "page=1&q=golang"
func BuildQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// ParseQuery parses a URL-encoded query string into a map of key-value pairs.
// If a key has multiple values, only the first one is retained.
func ParseQuery(query string) (map[string]string, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result, nil
}

// NormalizeURL parses and returns a normalized version of the input URL,
// stripping any fragment (anchor) part like "#section".
func NormalizeURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	u.Fragment = ""
	return u.String(), nil
}

// IsValidURL checks whether a string is a valid absolute URL,
// containing a scheme (e.g., http) and a host.
func IsValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme != "" && u.Host != ""
}
