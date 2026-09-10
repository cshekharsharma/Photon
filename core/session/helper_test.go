package session

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSessionEncoding(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Encoding
	}{
		{
			name:  "json lowercase",
			input: "json",
			want:  EncodingJSON,
		},
		{
			name:  "gob lowercase",
			input: "gob",
			want:  EncodingGob,
		},
		{
			name:  "json uppercase with spaces",
			input: "  JSON  ",
			want:  EncodingJSON,
		},
		{
			name:  "gob mixed case with spaces",
			input: " GoB ",
			want:  EncodingGob,
		},
		{
			name:  "invalid value",
			input: "xml",
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseSessionEncoding(tt.input))
		})
	}
}

func TestParseSameSite(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  http.SameSite
	}{
		{
			name:  "lax lowercase",
			input: "lax",
			want:  http.SameSiteLaxMode,
		},
		{
			name:  "strict lowercase",
			input: "strict",
			want:  http.SameSiteStrictMode,
		},
		{
			name:  "none lowercase",
			input: "none",
			want:  http.SameSiteNoneMode,
		},
		{
			name:  "strict uppercase with spaces",
			input: "  STRICT  ",
			want:  http.SameSiteStrictMode,
		},
		{
			name:  "none mixed case with spaces",
			input: " NoNe ",
			want:  http.SameSiteNoneMode,
		},
		{
			name:  "invalid value",
			input: "default",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseSameSite(tt.input))
		})
	}
}
