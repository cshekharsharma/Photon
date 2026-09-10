package elasticsearch

import (
	"errors"
	"sync"
	"testing"

	es8 "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockElasticsearchClient struct {
	mock.Mock
	*es8.TypedClient
}

func (m *MockElasticsearchClient) Info() (*esapi.Response, error) {
	args := m.Called()
	return args.Get(0).(*esapi.Response), args.Error(1)
}

func resetElasticsearchRegistry() {
	mutex.Lock()
	defer mutex.Unlock()
	instances = nil
	connectionConfigMap = nil
}

func TestConnect_NewInstance(t *testing.T) {
	resetElasticsearchRegistry()
	clusterName := "test_cluster"

	connectionConfigMap = map[string]*ConnectionConfig{
		clusterName: {
			Addresses:  []string{"http://localhost:9200"},
			Username:   "user",
			Password:   "pass",
			MaxRetries: 3,
		},
	}

	instances = make(map[string]*es8.TypedClient)

	client, err := Connect(clusterName)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, &es8.TypedClient{}, client)
}

func TestConnect_NoConfigSet(t *testing.T) {
	resetElasticsearchRegistry()
	clusterName := "test_cluster"
	connectionConfigMap = make(map[string]*ConnectionConfig)
	instances = make(map[string]*es8.TypedClient)

	client, err := Connect(clusterName)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestSetConnectionConfig_NewConfig(t *testing.T) {
	resetElasticsearchRegistry()
	clusterName := "new_cluster"
	config := &ConnectionConfig{
		Addresses:  []string{"http://localhost:9200"},
		Username:   "elastic",
		Password:   "password",
		MaxRetries: 5,
	}

	connectionConfigMap = nil

	SetConnectionConfig(clusterName, config)

	assert.NotNil(t, connectionConfigMap, "The connection config map should be initialized")
	assert.Equal(t, 1, len(connectionConfigMap), "There should be exactly one configuration in the map")
	assert.Equal(t, config, connectionConfigMap[clusterName], "The configuration should match what was set")
}

func TestRegistryHardening(t *testing.T) {
	t.Run("nil config and constructor error", func(t *testing.T) {
		resetElasticsearchRegistry()
		client, err := newInstance(nil)
		assert.Nil(t, client)
		assert.Error(t, err)

		orig := newTypedClientHook
		defer func() { newTypedClientHook = orig }()

		wantErr := errors.New("constructor failed")
		newTypedClientHook = func(es8.Config) (*es8.TypedClient, error) {
			return nil, wantErr
		}

		SetConnectionConfig("bad", &ConnectionConfig{Addresses: []string{"http://localhost:9200"}})
		client, err = Connect("bad")
		assert.Nil(t, client)
		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("config is cloned on set and get", func(t *testing.T) {
		resetElasticsearchRegistry()
		cfg := &ConnectionConfig{
			Addresses:  []string{"http://localhost:9200"},
			Username:   "elastic",
			Password:   "password",
			MaxRetries: 5,
		}
		SetConnectionConfig("clone", cfg)
		cfg.Addresses[0] = "http://mutated:9200"
		cfg.Username = "mutated"

		got := GetConnectionConfig("clone")
		assert.Equal(t, []string{"http://localhost:9200"}, got.Addresses)
		assert.Equal(t, "elastic", got.Username)
		assert.Equal(t, "password", got.Password)
		assert.Equal(t, int64(5), got.MaxRetries)

		got.Addresses[0] = "http://changed-again:9200"
		assert.Equal(t, "http://localhost:9200", GetConnectionConfig("clone").Addresses[0])
		assert.Nil(t, GetConnectionConfig("missing"))
		resetElasticsearchRegistry()
		assert.Nil(t, GetConnectionConfig("missing"))
		assert.Nil(t, cloneConnectionConfig(nil))
	})

	t.Run("singleton reuse and concurrent access", func(t *testing.T) {
		resetElasticsearchRegistry()
		orig := newTypedClientHook
		defer func() { newTypedClientHook = orig }()

		created := &es8.TypedClient{}
		newTypedClientHook = func(config es8.Config) (*es8.TypedClient, error) {
			assert.Equal(t, []string{"http://localhost:9200"}, config.Addresses)
			assert.Equal(t, "user", config.Username)
			return created, nil
		}

		SetConnectionConfig("shared", &ConnectionConfig{
			Addresses: []string{"http://localhost:9200"},
			Username:  "user",
		})

		first, err := Connect("shared")
		assert.NoError(t, err)
		assert.Same(t, created, first)
		second, err := Connect("shared")
		assert.NoError(t, err)
		assert.Same(t, first, second)

		const workers = 8
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				SetConnectionConfig("other", &ConnectionConfig{Addresses: []string{"http://localhost:9200"}})
				_ = GetConnectionConfig("other")
				client, err := Connect("shared")
				assert.NoError(t, err)
				assert.Same(t, first, client)
			}(i)
		}
		wg.Wait()
	})

	t.Run("nil stored config", func(t *testing.T) {
		resetElasticsearchRegistry()
		SetConnectionConfig("nil-config", nil)
		client, err := Connect("nil-config")
		assert.Nil(t, client)
		assert.Error(t, err)
	})
}
