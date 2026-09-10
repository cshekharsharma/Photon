package session

import (
	"net/http"
	"strings"
)

const (
	SameSiteLaxMode    = "lax"
	SameSiteStrictMode = "strict"
	SameSiteNoneMode   = "none"
)

func ParseSessionEncoding(value string) Encoding {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(EncodingJSON):
		return EncodingJSON
	case string(EncodingGob):
		return EncodingGob
	default:
		return ""
	}
}

func ParseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case SameSiteLaxMode:
		return http.SameSiteLaxMode
	case SameSiteStrictMode:
		return http.SameSiteStrictMode
	case SameSiteNoneMode:
		return http.SameSiteNoneMode
	default:
		return 0
	}
}
