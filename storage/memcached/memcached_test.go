package memcached

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/cshekharsharma/photon/utils/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MemcachedConnector mock
type mockConnector struct {
	mock.Mock
}

func (m *mockConnector) New(server ...string) MemcachedInterface {
	args := m.Called(server)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(MemcachedInterface)
}

// --------- Tests ---------

func resetMemcachedRegistry() {
	mutex.Lock()
	defer mutex.Unlock()
	instances = nil
	connectionConfigMap = nil
}

func TestMemcachedConnector_New(t *testing.T) {
	connector := &MemcachedConnector{}
	result := connector.New("127.0.0.1:11211")

	assert.NotNil(t, result)
	assert.NotNil(t, result.GetClient())
}

func TestMemcached_GetClient_SetClient(t *testing.T) {
	m := &Memcached{}
	client := &mocks.MockMemcachedClient{}

	m.SetClient(client)

	assert.Equal(t, client, m.GetClient())
}

func TestMemcached_Close(t *testing.T) {
	client := &mocks.MockMemcachedClient{}
	client.On("Close").Return(nil)

	m := &Memcached{client: client}

	err := m.Close()
	assert.NoError(t, err)

	client.AssertExpectations(t)
}

func TestSetConnectionConfig(t *testing.T) {
	SetConnectionConfig("testCluster", &ConnectionConfig{
		Addresses:   []string{"localhost:11211"},
		Timeout:     5 * time.Second,
		MaxIdleConn: 10,
	})

	assert.NotNil(t, connectionConfigMap["testCluster"])
}

func TestConnect_NewConnection_Success(t *testing.T) {
	resetMemcachedRegistry()
	mockedClient := &mocks.MockMemcachedClient{}
	memcached := &Memcached{client: mockedClient}
	connector := &mockConnector{}

	connector.On("New", []string{"localhost:11211"}).Return(memcached)

	SetConnectionConfig("testCluster", &ConnectionConfig{
		Addresses:   []string{"localhost:11211"},
		Timeout:     5 * time.Second,
		MaxIdleConn: 10,
	})

	client, err := Connect(connector, "testCluster")
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, memcached, client)

	connector.AssertExpectations(t)
}

func TestConnect_AlreadyExistingConnection(t *testing.T) {
	resetMemcachedRegistry()
	mutex.Lock()
	instances = make(map[string]MemcachedInterface)
	existing := &Memcached{}
	instances["testExisting"] = existing
	mutex.Unlock()

	client, err := Connect(&mockConnector{}, "testExisting")
	assert.NoError(t, err)
	assert.Equal(t, existing, client)
}

func TestConnect_MissingConfig(t *testing.T) {
	resetMemcachedRegistry()
	_, err := Connect(&mockConnector{}, "nonexistent")
	assert.Error(t, err)
}

func TestConnect_NewConnection_Error(t *testing.T) {
	resetMemcachedRegistry()
	connector := &mockConnector{}
	connector.On("New", []string{"localhost:11211"}).Return(nil)

	SetConnectionConfig("badCluster", &ConnectionConfig{
		Addresses:   []string{"localhost:11211"},
		Timeout:     1 * time.Second,
		MaxIdleConn: 1,
	})

	_, err := Connect(connector, "badCluster")
	assert.Error(t, err)
}

func TestNewInstance_Success(t *testing.T) {
	connector := &mockConnector{}
	mocked := &Memcached{}
	connector.On("New", []string{"localhost:11211"}).Return(mocked)

	config := &ConnectionConfig{
		Addresses:   []string{"localhost:11211"},
		Timeout:     2 * time.Second,
		MaxIdleConn: 5,
	}

	mc, err := newInstance(connector, config)
	assert.NoError(t, err)
	assert.NotNil(t, mc)
}

func TestNewInstance_ApplyClientConfig(t *testing.T) {
	connector := &mockConnector{}
	mem := &Memcached{client: memcache.New("localhost:11211")}
	connector.On("New", []string{"localhost:11211"}).Return(mem)

	config := &ConnectionConfig{
		Addresses:   []string{"localhost:11211"},
		Timeout:     3 * time.Second,
		MaxIdleConn: 7,
	}

	mc, err := newInstance(connector, config)
	assert.NoError(t, err)
	assert.NotNil(t, mc)

	raw := mem.GetClient().(*memcache.Client)
	assert.Equal(t, config.Timeout, raw.Timeout)
	assert.Equal(t, int(config.MaxIdleConn), raw.MaxIdleConns)
}

func TestNewInstance_NilClient(t *testing.T) {
	mockConnector := new(mockConnector)
	config := &ConnectionConfig{
		Addresses: []string{"127.0.0.1:11211"},
	}

	mockConnector.On("New", config.Addresses).Return(nil)

	client, err := newInstance(mockConnector, config)

	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create memcache client")

	mockConnector.AssertExpectations(t)
}

func TestRegistryHardening(t *testing.T) {
	t.Run("nil connector and config", func(t *testing.T) {
		resetMemcachedRegistry()
		client, err := Connect(nil, "missing")
		assert.Nil(t, client)
		assert.Error(t, err)

		client, err = newInstance(nil, &ConnectionConfig{})
		assert.Nil(t, client)
		assert.Error(t, err)

		client, err = newInstance(&mockConnector{}, nil)
		assert.Nil(t, client)
		assert.Error(t, err)

		SetConnectionConfig("nil-config", nil)
		client, err = Connect(&mockConnector{}, "nil-config")
		assert.Nil(t, client)
		assert.Error(t, err)
	})

	t.Run("config is cloned on set and get", func(t *testing.T) {
		resetMemcachedRegistry()
		cfg := &ConnectionConfig{
			Addresses:   []string{"localhost:11211"},
			Timeout:     5 * time.Second,
			MaxIdleConn: 3,
		}
		SetConnectionConfig("clone", cfg)
		cfg.Addresses[0] = "mutated:11211"
		cfg.Timeout = time.Second
		cfg.MaxIdleConn = 1

		got := GetConnectionConfig("clone")
		assert.Equal(t, []string{"localhost:11211"}, got.Addresses)
		assert.Equal(t, 5*time.Second, got.Timeout)
		assert.Equal(t, int64(3), got.MaxIdleConn)

		got.Addresses[0] = "changed-again:11211"
		assert.Equal(t, "localhost:11211", GetConnectionConfig("clone").Addresses[0])
		assert.Nil(t, GetConnectionConfig("missing"))
		resetMemcachedRegistry()
		assert.Nil(t, GetConnectionConfig("missing"))
		assert.Nil(t, cloneConnectionConfig(nil))
	})

	t.Run("close cluster and all", func(t *testing.T) {
		resetMemcachedRegistry()
		okClient := &mocks.MockMemcachedClient{}
		okClient.On("Close").Return(nil).Twice()
		errClient := &mocks.MockMemcachedClient{}
		closeErr := errors.New("close failed")
		errClient.On("Close").Return(closeErr).Once()

		mutex.Lock()
		instances = map[string]MemcachedInterface{
			"one": &Memcached{client: okClient},
		}
		mutex.Unlock()

		assert.NoError(t, CloseCluster("missing"))
		assert.NoError(t, CloseCluster("one"))
		assert.NoError(t, CloseCluster("one"))

		mutex.Lock()
		instances = map[string]MemcachedInterface{
			"ok":  &Memcached{client: okClient},
			"err": &Memcached{client: errClient},
			"nil": nil,
		}
		mutex.Unlock()

		assert.ErrorIs(t, CloseAll(), closeErr)
		okClient.AssertExpectations(t)
		errClient.AssertExpectations(t)
	})

	t.Run("concurrent set connect get", func(t *testing.T) {
		resetMemcachedRegistry()
		cluster := "concurrent"
		SetConnectionConfig(cluster, &ConnectionConfig{
			Addresses:   []string{"localhost:11211"},
			Timeout:     time.Second,
			MaxIdleConn: 2,
		})

		memcached := &Memcached{client: &mocks.MockMemcachedClient{}}
		connector := &mockConnector{}
		connector.On("New", []string{"localhost:11211"}).Return(memcached).Once()

		const workers = 8
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				SetConnectionConfig(fmt.Sprintf("cluster-%d", i), &ConnectionConfig{Addresses: []string{"localhost:11211"}})
				_ = GetConnectionConfig(fmt.Sprintf("cluster-%d", i))
				client, err := Connect(connector, cluster)
				assert.NoError(t, err)
				assert.Equal(t, memcached, client)
			}(i)
		}
		wg.Wait()

		connector.AssertExpectations(t)
	})
}
