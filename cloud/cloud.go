// Package cloud provides abstractions and utilities to interact with cloud service providers,
// allowing easy switching and management of different cloud services based on configuration.
// This package handles initialization and retrieval of various cloud services like object storage,
// message queues, and publish-subscribe systems depending on the configured cloud provider.
package cloud

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cshekharsharma/photon/cloud/contract"
	"github.com/cshekharsharma/photon/cloud/entity"
	"github.com/cshekharsharma/photon/cloud/providers"
)

var (
	defaultConfigMu sync.RWMutex
	defaultConfig   Config

	AppConfigServiceProvider = getAppConfigService
	isValidCloudVendor       = IsValidCloudVendor
)

const (
	CloudVendorDefault string = "aws" // Default cloud vendor name
	CloudVendorAws     string = "aws" // cloud vendorn name for AWS
)

type Config struct {
	Vendor        string
	AuthArguments *entity.CloudAuthArguments
}

type Client struct {
	vendor   string
	provider *providers.AwsCloud
}

func NewClient(config Config) (*Client, error) {
	config = cloneConfig(config)
	if config.Vendor == "" {
		config.Vendor = CloudVendorDefault
	}
	if !isValidCloudVendor(config.Vendor) {
		return nil, fmt.Errorf("invalid cloud vendor `%s` configured", config.Vendor)
	}
	if config.AuthArguments == nil {
		return nil, errors.New("cloud auth arguments cannot be nil")
	}

	switch config.Vendor {
	case CloudVendorAws:
		awsCloud, err := providers.NewAwsCloud(config.AuthArguments)
		if err != nil {
			return nil, err
		}
		return &Client{vendor: config.Vendor, provider: awsCloud}, nil
	default:
		return nil, fmt.Errorf("cannot find provider for configured cloud type %q", config.Vendor)
	}
}

func MustNewClient(config Config) *Client {
	client, err := NewClient(config)
	if err != nil {
		panic(err)
	}
	return client
}

// SetCloudAuthArguments sets the global cloud authentication arguments used to initialize cloud services.
// This function stores the provided CloudAuthArguments into a global variable.
//
// Parameters:
//   - args: A pointer to an entity.CloudAuthArguments struct containing authentication data required by cloud services.
//
// Usage:
//   - This function should be called prior to initializing any cloud service interfaces to ensure that all services
//     are configured with the correct authentication parameters.
func SetCloudAuthArguments(args *entity.CloudAuthArguments) {
	defaultConfigMu.Lock()
	defer defaultConfigMu.Unlock()
	defaultConfig.AuthArguments = cloneAuthArguments(args)
}

// SetCloudVendor sets the identifier for the cloud vendor to be used globally across the cloud services.
// This function assigns the provided vendor string to a global variable.
//
// Parameters:
//   - vendor: A string identifier for the cloud vendor (e.g., "AWS", "Azure", "GCP").
//
// Usage:
//   - Call this function to define which cloud provider's services should be initialized and used in the application.
//   - This is typically set at application startup and used across various cloud service initializations.
func SetCloudVendor(vendor string) {
	defaultConfigMu.Lock()
	defer defaultConfigMu.Unlock()
	defaultConfig.Vendor = vendor
}

// GetObjectStorage provides an instance of the ObjectStorageInterface based on the configured cloud vendor.
//
// Returns:
//   - An implementation of the ObjectStorageInterface based on the cloud vendor specified in the configuration.
//   - An error if the configured cloud vendor is not supported or if there is an issue initializing the object storage.
func GetObjectStorage() (contract.ObjectStorageInterface, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	return client.GetObjectStorage(context.Background())
}

// GetMessageQueue provides an instance of the MessageQueueInterface based on the configured cloud vendor.
//
// Returns:
//   - An implementation of the MessageQueueInterface based on the cloud vendor specified in the configuration.
//   - An error if the configured cloud vendor is not supported or if there is an issue initializing the message queue.
func GetMessageQueue() (contract.MessageQueueInterface, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	return client.GetMessageQueue(context.Background())
}

// GetPublishSubscribe returns an instance of the PublishSubscribeInterface based on the configured cloud vendor.
// It checks for valid configuration and initializes the appropriate publish-subscribe service.
//
// Returns:
//   - An implementation of PublishSubscribeInterface if the initialization is successful.
//   - An error if the cloud vendor is not supported or other initialization issues occur.
//
// Usage:
//   - This function provides access to publish-subscribe services, facilitating communication and notifications in applications.
func GetPublishSubscribe() (contract.PublishSubscribeInterface, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	return client.GetPublishSubscribe(context.Background())
}

func GetEmailService() (contract.EmailInterface, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	return client.GetEmailService(context.Background())
}

func GetAppConfigService() (contract.AppConfigInterface, error) {
	return AppConfigServiceProvider()
}

func GetVisualAnalysisService() (contract.VisualAnalysisInterface, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	return client.GetVisualAnalysisService(context.Background())
}

func getAppConfigService() (contract.AppConfigInterface, error) {
	client, err := defaultClient()
	if err != nil {
		return nil, err
	}
	return client.GetAppConfigService(context.Background())
}

// sanitiseInitStatus checks the validity of the cloud vendor and authentication arguments.
// It ensures that the cloud vendor is supported and that authentication arguments are properly initialized.
//
// Returns:
//   - nil if the cloud vendor and authentication arguments are valid.
//   - An error if the cloud vendor is not recognized or if authentication arguments are not properly initialized.
//
// Usage:
//   - This function is called internally to validate the configuration before initializing any cloud service interfaces.
func sanitiseInitStatus() error {
	_, err := NewClient(defaultConfigSnapshot())
	return err
}

// IsValidCloudVendor checks if the given cloudVendor is a valid and supported cloud provider.
//
// Parameters:
//   - cloudVendor: The identifier string for the cloud vendor to be checked.
//
// Returns:
//   - true if the cloudVendor is valid and supported, false otherwise.
func IsValidCloudVendor(cloudVendor string) bool {
	_, exists := getAllowedCloudVendors()[cloudVendor]
	return exists && cloudVendor != ""
}

// getAllowedCloudVendors returns a map of supported cloud vendors.
//
// Returns:
//   - A map where keys are cloud vendor identifiers and values are booleans
//     indicating if the vendor is allowed.
func getAllowedCloudVendors() map[string]bool {
	return map[string]bool{
		CloudVendorAws: true,
	}
}

func (c *Client) GetObjectStorage(ctx context.Context) (contract.ObjectStorageInterface, error) {
	if c == nil || c.provider == nil {
		return nil, errors.New("cloud client is not initialized")
	}
	return c.provider.GetObjectStorage(ctx)
}

func (c *Client) GetMessageQueue(ctx context.Context) (contract.MessageQueueInterface, error) {
	if c == nil || c.provider == nil {
		return nil, errors.New("cloud client is not initialized")
	}
	return c.provider.GetMessageQueue(ctx)
}

func (c *Client) GetPublishSubscribe(ctx context.Context) (contract.PublishSubscribeInterface, error) {
	if c == nil || c.provider == nil {
		return nil, errors.New("cloud client is not initialized")
	}
	return c.provider.GetPublishSubscribe(ctx)
}

func (c *Client) GetEmailService(ctx context.Context) (contract.EmailInterface, error) {
	if c == nil || c.provider == nil {
		return nil, errors.New("cloud client is not initialized")
	}
	return c.provider.GetEmailService(ctx)
}

func (c *Client) GetAppConfigService(ctx context.Context) (contract.AppConfigInterface, error) {
	if c == nil || c.provider == nil {
		return nil, errors.New("cloud client is not initialized")
	}
	return c.provider.GetAppConfigService(ctx)
}

func (c *Client) GetVisualAnalysisService(ctx context.Context) (contract.VisualAnalysisInterface, error) {
	if c == nil || c.provider == nil {
		return nil, errors.New("cloud client is not initialized")
	}
	return c.provider.GetVisualAnalysisService(ctx)
}

func defaultClient() (*Client, error) {
	return NewClient(defaultConfigSnapshot())
}

func defaultConfigSnapshot() Config {
	defaultConfigMu.RLock()
	defer defaultConfigMu.RUnlock()
	return cloneConfig(defaultConfig)
}

func cloneConfig(config Config) Config {
	return Config{
		Vendor:        config.Vendor,
		AuthArguments: cloneAuthArguments(config.AuthArguments),
	}
}

func cloneAuthArguments(args *entity.CloudAuthArguments) *entity.CloudAuthArguments {
	if args == nil {
		return nil
	}
	cloned := *args
	return &cloned
}
