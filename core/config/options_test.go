package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigOptionIsValid(t *testing.T) {

	tests := []struct {
		name    string
		options Options
		wantErr bool
		errMsg  string
	}{
		{
			name:    "ValidRawSourceWithFormatJson",
			options: Options{Source: SourceRawBytes, Format: FormatJson, Content: []byte(`{"key":"value"}`), Delimiter: ","},
			wantErr: false,
		},
		{
			name:    "InvalidSource",
			options: Options{Source: 3, Format: FormatJson, FilePath: "valid/file/path", Delimiter: ","},
			wantErr: true,
			errMsg:  "invalid config source 3 provided",
		},
		{
			name:    "InvalidFormat",
			options: Options{Source: SourceFile, Format: 3, FilePath: "valid/file/path", Delimiter: ","},
			wantErr: true,
			errMsg:  "invalid config format 3 provided",
		},
		{
			name:    "UnreadableSourceFile",
			options: Options{Source: SourceFile, Format: FormatJson, FilePath: "invalid/file/path", Delimiter: ","},
			wantErr: true,
			errMsg:  "filePath must be a readable file when Source is of type file",
		},
		{
			name:    "EmptyRawBytesWithEnableWatch",
			options: Options{Source: SourceRawBytes, Format: FormatJson, Delimiter: ",", EnableWatch: true},
			wantErr: true,
			errMsg:  "content must be set when Source is of type raw bytes",
		},
		{
			name:    "EmptyRawBytesContent",
			options: Options{Source: SourceRawBytes, Format: FormatJson, Delimiter: ","},
			wantErr: true,
			errMsg:  "content must be set when Source is of type raw bytes",
		},
		{
			name:    "EmptyWatcherWithEnableWatch",
			options: Options{Source: SourceRawBytes, Format: FormatJson, Content: []byte(`{"key":"value"}`), Delimiter: ",", EnableWatch: true},
			wantErr: true,
			errMsg:  "watcher options must be set when EnableWatch is true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.IsValid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigOptionValidateNil(t *testing.T) {
	var opts *Options
	assert.EqualError(t, opts.Validate(), "config options are required")
}

func TestPopulateRequiredOptionsPropertiesDefaultDelimiter(t *testing.T) {
	assert.Nil(t, populateRequiredOptionsProperties(nil))

	opts := populateRequiredOptionsProperties(&Options{Delimiter: ""})
	assert.Equal(t, DefaultConfigPathDelimiter, opts.Delimiter)
}
