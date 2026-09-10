package discovery

import "context"

// ServiceInstance represents a single instance of a registered service.
type ServiceInstance struct {
	ID       string
	Name     string
	Address  string
	Port     int
	Tags     []string
	Metadata map[string]string
}

// ServiceDiscovery defines the contract for a service discovery system.
type ServiceDiscovery interface {
	// Register a service instance with the discovery provider.
	Register(ctx context.Context, instance *ServiceInstance) error

	// Deregister removes the service instance from the registry.
	Deregister(ctx context.Context, instanceID string) error

	// Discover returns healthy service instances by name.
	Discover(ctx context.Context, serviceName string) ([]*ServiceInstance, error)

	// Watch monitors changes to a given service and triggers the callback.
	Watch(ctx context.Context, serviceName string, onChange func([]*ServiceInstance)) error

	// Close gracefully shuts down the discovery client (e.g., closes connections, stops watchers).
	Close() error
}
