package discovery

import (
	"errors"
	"testing"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/stretchr/testify/assert"
)

func TestOptions_Validate(t *testing.T) {
	validLogger := logger.Init(&logger.LoggerConfig{
		Name:     "discovery-options-test",
		Provider: logger.LoggerProviderZerolog,
		Type:     logger.LoggerTypeStdout,
	})

	tests := []struct {
		name    string
		opts    *Options
		wantErr error
	}{
		{
			name:    "missing provider",
			opts:    &Options{Address: "localhost:8500", Logger: validLogger},
			wantErr: errors.New("provider type is required"),
		},
		{
			name:    "missing address",
			opts:    &Options{Provider: ProviderConsul, Logger: validLogger},
			wantErr: errors.New("address is required"),
		},
		{
			name:    "missing logger",
			opts:    &Options{Provider: ProviderConsul, Address: "localhost:8500"},
			wantErr: errors.New("logger is required"),
		},
		{
			name:    "all valid",
			opts:    &Options{Provider: ProviderConsul, Address: "localhost:8500", Logger: validLogger},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()
			if tc.wantErr != nil {
				assert.EqualError(t, err, tc.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
