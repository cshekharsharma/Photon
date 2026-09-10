package caching

import (
	"testing"
)

func TestValidate_AllCases(t *testing.T) {
	tests := []struct {
		name        string
		opts        *Options
		expectedErr string
	}{
		{
			name: "missing provider",
			opts: &Options{
				Provider:    "",
				Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
				ConnTimeout: 5,
				DefaultTTL:  60,
				Cluster:     "cluster",
			},
			expectedErr: "Options.Provider is not set",
		},
		{
			name: "missing hosts",
			opts: &Options{
				Provider:    "aerospike",
				Hosts:       nil,
				ConnTimeout: 10,
				DefaultTTL:  60,
				Cluster:     "cluster",
			},
			expectedErr: "Options.Hosts is not set",
		},
		{
			name: "missing connection timeout",
			opts: &Options{
				Provider:    "aerospike",
				Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
				ConnTimeout: 0,
				DefaultTTL:  60,
				Cluster:     "cluster",
			},
			expectedErr: "Options.ConnTimeout is not set",
		},
		{
			name: "missing default TTL",
			opts: &Options{
				Provider:    "aerospike",
				Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
				ConnTimeout: 5,
				DefaultTTL:  0,
				Cluster:     "cluster",
			},
			expectedErr: "Options.DefaultTTL is not set",
		},
		{
			name: "missing cluster",
			opts: &Options{
				Provider:    "aerospike",
				Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
				ConnTimeout: 5,
				DefaultTTL:  60,
				Cluster:     "",
			},
			expectedErr: "Options.Cluster is not set",
		},
		{
			name: "all valid",
			opts: &Options{
				Provider:    "aerospike",
				Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
				ConnTimeout: 5,
				DefaultTTL:  60,
				Cluster:     "cluster",
			},
			expectedErr: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.opts)
			if tc.expectedErr == "" && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Errorf("expected error '%s', got '%v'", tc.expectedErr, err)
				}
			}
		})
	}
}

func TestGetProvider_InvalidParams(t *testing.T) {
	opts := &Options{
		Provider:    "",
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}

	_, err := GetProvider(opts)
	if err == nil {
		t.Fatalf("Expected error, got no error from GetProvider")
	}
}

func TestGetProvider_Aerospike_Success(t *testing.T) {
	opts := &Options{
		Provider:    ProviderAerospike,
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}

	cache, err := GetProvider(opts)
	if err != nil {
		t.Fatalf("Expected no error from GetProvider, got: %v", err)
	}

	if cache == nil {
		t.Fatal("Expected cache instance, got nil")
	}
}

func TestGetProvider_Redis_Success(t *testing.T) {
	opts := &Options{
		Provider:    ProviderRedis,
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1"},
		ConnTimeout: 5,
		DefaultTTL:  60,
		Username:    "",
		Password:    "",
	}

	cache, err := GetProvider(opts)
	if err != nil {
		t.Fatalf("Expected no error from GetProvider, got: %v", err)
	}

	if cache == nil {
		t.Fatal("Expected cache instance, got nil")
	}
}

func TestGetProvider_Memcached_Success(t *testing.T) {
	opts := &Options{
		Provider:    ProviderMemcached,
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}

	cache, err := GetProvider(opts)
	if err != nil {
		t.Fatalf("Expected no error from GetProvider, got: %v", err)
	}

	if cache == nil {
		t.Fatal("Expected cache instance, got nil")
	}
}

func TestGetProvider_FlashDB_Success(t *testing.T) {
	opts := &Options{
		Provider:    ProviderFlashDB,
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}

	cache, err := GetProvider(opts)
	if err != nil {
		t.Fatalf("Expected no error from GetProvider, got: %v", err)
	}

	if cache == nil {
		t.Fatal("Expected cache instance, got nil")
	}
}

func TestGetProvider_UnsupportedProvider(t *testing.T) {
	opts := &Options{
		Provider:    "unknown", // unsupported in current impl
		Cluster:     "test-cluster",
		Namespace:   "test-ns",
		Collection:  "test-collection",
		Hosts:       []string{"127.0.0.1", "127.0.0.1", "127.0.0.1"},
		ConnTimeout: 5,
		DefaultTTL:  60,
	}

	cache, err := GetProvider(opts)
	if cache != nil {
		t.Errorf("Expected nil cache for unsupported provider, got: %#v", cache)
	}

	if err == nil || err.Error() != "unsupported cache provider: unknown" {
		t.Errorf("Expected unsupported provider error, got: %v", err)
	}
}
