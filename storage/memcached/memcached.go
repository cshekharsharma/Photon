// Package memcached provides functionality for establishing connections to Memcached servers
// using the gomemcache package "github.com/bradfitz/gomemcache/memcache". It allows you to
// connect to specific Memcached servers, configure connection parameters, and reuse existing
// connections. The package also handles configurations such as server addresses derived from
// the application configuration.
//
// Usage:
//  1. Use Connect(serverName) to connect to a specific Memcached server, ensuring
//     thread-safe and singleton connection instances.
//
// The Connect() method uses a mutex to manage concurrent access and returns a pointer to the
// established Memcached client connection or an error if the connection process fails.
//
// Example:
//
//	client, err := memcached.Connect("my_server")
//	if err != nil {
//	    log.Fatalf("Failed to connect to Memcached: %v", err)
//	}
//
//	// Use the Memcached client for further operations, e.g., get or set data.
package memcached

import (
	"fmt"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
)

var (
	mutex               sync.RWMutex
	instances           map[string]MemcachedInterface
	connectionConfigMap map[string]*ConnectionConfig
)

// MemcachedConnectorInterface defines methods to establish a connection
// to a Memcached cluster and retrieve a client interface.
type MemcachedConnectorInterface interface {

	// New establishes a connection to the memcached server
	New(server ...string) MemcachedInterface
}

// MemcachedConnector represents a connector to a Memcached instance
type MemcachedConnector struct {
	Memcached *Memcached
}

// New attmpets to establish a connection to a Memcached instance.
//
// Parameters:
// - server: List of host addresses
//
// Returns:
// - An instance of MemcachedInterface, representing the connection
func (mc *MemcachedConnector) New(server ...string) MemcachedInterface {
	return &Memcached{
		client: memcache.New(server...),
	}
}

// MemcachedInterface defines methods to interact with an memcached server.
type MemcachedInterface interface {
	GetClient() MemcachedClientInterface
	SetClient(client MemcachedClientInterface)
	Close() error
}

// Memcached encapsulates the client instance for an memcached server.
type Memcached struct {
	client MemcachedClientInterface
}

// GetClient returns the underlying client for the memcached instance.
func (m *Memcached) GetClient() MemcachedClientInterface {
	return m.client
}

func (m *Memcached) SetClient(client MemcachedClientInterface) {
	m.client = client
}

// Close shuts down the connection to Memcached.
func (m *Memcached) Close() error {
	return m.client.Close()
}

// Connect establishes a connection to a Memcached server specified by the
// given serverName. If a connection to the server already exists, the
// existing connection is reused. This method uses a singleton pattern to
// ensure only one connection i nstance per serverName.
//
// Parameters:
//   - serverName: The name of the Memcached server to connect to.
//
// Returns:
//   - *memcache.Client: A pointer to the established Memcached client connection.
//     If a connection to the given serverName already exists, the existing connection will be returned.
//   - error: An error object that describes the reason for any connection
//     failures. It returns nil if the connection was successful.
//
// Note: This method is thread-safe and uses mutexes to handle concurrent access.
func Connect(connector MemcachedConnectorInterface, serverName string) (MemcachedInterface, error) {
	if connector == nil {
		return nil, fmt.Errorf("memcached: connector is required")
	}

	mutex.Lock()
	defer mutex.Unlock()
	if instances == nil {
		instances = make(map[string]MemcachedInterface)
	}

	if instance := instances[serverName]; instance != nil {
		return instance, nil
	}

	config, ok := connectionConfigMap[serverName]
	if !ok || config == nil {
		return nil, fmt.Errorf("connection config for server '%s' not set", serverName)
	}

	newConn, err := newInstance(connector, config)
	if err != nil {
		return nil, err
	}
	instances[serverName] = newConn

	return newConn, nil
}

// newInstance creates a new Memcached client connection instance for a
// specified server using configurations derived from the server name. The
// configurations such as server address are fetched using the given serverName.
func newInstance(connector MemcachedConnectorInterface, config *ConnectionConfig) (MemcachedInterface, error) {
	if connector == nil {
		return nil, fmt.Errorf("memcached: connector is required")
	}
	if config == nil {
		return nil, fmt.Errorf("memcached: connection config is required")
	}

	client := connector.New(config.Addresses...)

	if client == nil {
		return nil, fmt.Errorf("failed to create memcache client for server '%v'", config.Addresses)
	}

	if rawclient, ok := client.GetClient().(*memcache.Client); ok {
		rawclient.Timeout = config.Timeout
		rawclient.MaxIdleConns = int(config.MaxIdleConn)

		client.SetClient(rawclient)
	}

	return client, nil
}

func SetConnectionConfig(clusterName string, config *ConnectionConfig) {
	mutex.Lock()
	defer mutex.Unlock()

	if connectionConfigMap == nil {
		connectionConfigMap = make(map[string]*ConnectionConfig)
	}

	connectionConfigMap[clusterName] = cloneConnectionConfig(config)
}

func GetConnectionConfig(clusterName string) *ConnectionConfig {
	mutex.RLock()
	defer mutex.RUnlock()

	if connectionConfigMap == nil {
		return nil
	}
	return cloneConnectionConfig(connectionConfigMap[clusterName])
}

func CloseCluster(serverName string) error {
	mutex.Lock()
	defer mutex.Unlock()

	if instances == nil || instances[serverName] == nil {
		return nil
	}
	err := instances[serverName].Close()
	delete(instances, serverName)
	return err
}

func CloseAll() error {
	mutex.Lock()
	defer mutex.Unlock()

	var firstErr error
	for serverName, instance := range instances {
		if instance != nil {
			if err := instance.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		delete(instances, serverName)
	}
	return firstErr
}

func cloneConnectionConfig(config *ConnectionConfig) *ConnectionConfig {
	if config == nil {
		return nil
	}
	copied := *config
	copied.Addresses = append([]string(nil), config.Addresses...)
	return &copied
}
