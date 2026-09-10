// Package elastic provides functionality for establishing connections to Elasticsearch v8 clusters
// using the official Go client library "github.com/elastic/go-elasticsearch/v8". It allows you to
// connect to specific Elasticsearch clusters, configure connection parameters, and reuse existing
// connections. The package also handles configurations such as cluster addresses, usernames,
// passwords, and max retries, which are derived from the application configuration.
//
// Usage:
//  2. Use Connect(clusterName) to connect to a specific Elasticsearch cluster, ensuring
//     thread-safe and singleton connection instances.
//
// The Connect() method uses a mutex to manage concurrent access and returns a pointer to the
// established Elasticsearch client connection or an error if the connection process fails.
//
// Example:
//
//	client, err := elastic.Connect("my_cluster")
//	if err != nil {
//	    log.Fatalf("Failed to connect to Elasticsearch: %v", err)
//	}
//
//	// Use the Elasticsearch client for further operations.
//	resp, err := client.Info()
//	if err != nil {
//	    log.Fatalf("Elasticsearch info request failed: %v", err)
//	}
package elasticsearch

import (
	"fmt"
	"sync"

	es8 "github.com/elastic/go-elasticsearch/v8"
)

var (
	mutex               sync.RWMutex
	instances           map[string]*es8.TypedClient
	connectionConfigMap map[string]*ConnectionConfig
	newTypedClientHook  = es8.NewTypedClient
)

const EsMaxFetchSize int64 = 10000
const EsScrollApiDefaultTimeout = 10

func SetConnectionConfig(clusterName string, config *ConnectionConfig) {
	mutex.Lock()
	defer mutex.Unlock()

	if connectionConfigMap == nil {
		connectionConfigMap = make(map[string]*ConnectionConfig)
	}

	connectionConfigMap[clusterName] = cloneConnectionConfig(config)
}

// Connect establishes a connection to an Elasticsearch v8 cluster specified by the
// given clusterName. If a connection to the cluster already exists, the
// existing connection is reused. This method uses a singleton pattern to
// ensure only one connection instance per clusterName.
//
// Parameters:
//   - clusterName: The name of the Elasticsearch v8 cluster to connect to.
//
// Returns:
//   - *es8.TypedClient: A pointer to the established Elasticsearch client connection.
//     If a connection to the given clusterName already exists,
//     the existing connection will be returned.
//   - error: An error object that describes the reason for any connection
//     failures. It returns nil if the connection was successful.
//
// Note: This method is thread-safe and uses mutexes to handle concurrent
func Connect(clusterName string) (*es8.TypedClient, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if instances == nil {
		instances = make(map[string]*es8.TypedClient)
	}
	if instances[clusterName] != nil {
		return instances[clusterName], nil
	}

	config, ok := connectionConfigMap[clusterName]
	if !ok || config == nil {
		return nil, fmt.Errorf("initialise db config using elasticsearch.SetConnectionConfig() before use")
	}

	newConn, err := newInstance(config)
	if err != nil {
		return nil, err
	}
	instances[clusterName] = newConn
	return newConn, nil
}

// newInstance creates a new Elasticsearch v8 client connection instance for a
// specified cluster using configurations derived from the cluster name. The
// configurations such as addresses, username, password, and max retries are
// fetched using the given clusterName.
func newInstance(config *ConnectionConfig) (*es8.TypedClient, error) {
	if config == nil {
		return nil, fmt.Errorf("elasticsearch: connection config is required")
	}

	esCfg := es8.Config{
		Addresses:  config.Addresses,
		Username:   config.Username,
		Password:   config.Password,
		MaxRetries: int(config.MaxRetries),
		Transport:  config.Transport,
	}

	// Create a new client and connect to the server
	return newTypedClientHook(esCfg)
}

func GetConnectionConfig(clusterName string) *ConnectionConfig {
	mutex.RLock()
	defer mutex.RUnlock()

	if connectionConfigMap == nil {
		return nil
	}
	return cloneConnectionConfig(connectionConfigMap[clusterName])
}

func cloneConnectionConfig(config *ConnectionConfig) *ConnectionConfig {
	if config == nil {
		return nil
	}
	copied := *config
	copied.Addresses = append([]string(nil), config.Addresses...)
	return &copied
}
