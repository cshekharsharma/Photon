package cloud

import (
	"context"
	"sync"
	"testing"

	"github.com/cshekharsharma/photon/cloud/contract"
	"github.com/cshekharsharma/photon/cloud/entity"
	"github.com/cshekharsharma/photon/cloud/entity/appconfig"
	"github.com/stretchr/testify/require"
)

func validCloudAuth() *entity.CloudAuthArguments {
	return &entity.CloudAuthArguments{
		AuthMode:  entity.AuthTypeAccessKey,
		AccessKey: "access-key",
		SecretKey: "secret-key",
		SessionId: "session-id",
		IAMRole:   "role",
		Region:    "us-east-1",
	}
}

func resetDefaultCloudConfig(t *testing.T) {
	t.Helper()
	defaultConfigMu.Lock()
	originalConfig := cloneConfig(defaultConfig)
	defaultConfig = Config{}
	defaultConfigMu.Unlock()

	originalProvider := AppConfigServiceProvider
	originalValidator := isValidCloudVendor
	t.Cleanup(func() {
		defaultConfigMu.Lock()
		defaultConfig = originalConfig
		defaultConfigMu.Unlock()
		AppConfigServiceProvider = originalProvider
		isValidCloudVendor = originalValidator
	})
}

func TestNewClient(t *testing.T) {
	t.Run("success defaults vendor and clones auth", func(t *testing.T) {
		auth := validCloudAuth()
		client, err := NewClient(Config{AuthArguments: auth})
		require.NoError(t, err)
		require.NotNil(t, client)
		require.Equal(t, CloudVendorAws, client.vendor)
		require.Equal(t, "us-east-1", client.provider.AuthArguments.Region)

		auth.Region = "mutated"
		require.Equal(t, "us-east-1", client.provider.AuthArguments.Region)
	})

	t.Run("invalid vendor", func(t *testing.T) {
		client, err := NewClient(Config{Vendor: "gcp", AuthArguments: validCloudAuth()})
		require.Nil(t, client)
		require.ErrorContains(t, err, "invalid cloud vendor")
	})

	t.Run("validator accepts unsupported vendor", func(t *testing.T) {
		resetDefaultCloudConfig(t)
		isValidCloudVendor = func(string) bool { return true }

		client, err := NewClient(Config{Vendor: "gcp", AuthArguments: validCloudAuth()})
		require.Nil(t, client)
		require.ErrorContains(t, err, "cannot find provider")
	})

	t.Run("nil auth", func(t *testing.T) {
		client, err := NewClient(Config{Vendor: CloudVendorAws})
		require.Nil(t, client)
		require.ErrorContains(t, err, "cloud auth arguments cannot be nil")
	})

	t.Run("invalid auth from provider", func(t *testing.T) {
		auth := validCloudAuth()
		auth.Region = ""
		client, err := NewClient(Config{Vendor: CloudVendorAws, AuthArguments: auth})
		require.Nil(t, client)
		require.ErrorContains(t, err, "cloud auth region is required")
	})
}

func TestMustNewClientPanicsOnInvalidConfig(t *testing.T) {
	client := MustNewClient(Config{Vendor: CloudVendorAws, AuthArguments: validCloudAuth()})
	require.NotNil(t, client)

	require.Panics(t, func() {
		MustNewClient(Config{Vendor: "invalid", AuthArguments: validCloudAuth()})
	})
}

func TestGlobalConfigHelpersCloneAndValidate(t *testing.T) {
	resetDefaultCloudConfig(t)

	auth := validCloudAuth()
	SetCloudVendor(CloudVendorAws)
	SetCloudAuthArguments(auth)
	auth.Region = "mutated"

	snapshot := defaultConfigSnapshot()
	require.Equal(t, "us-east-1", snapshot.AuthArguments.Region)

	snapshot.AuthArguments.Region = "changed-again"
	require.Equal(t, "us-east-1", defaultConfigSnapshot().AuthArguments.Region)
	require.NoError(t, sanitiseInitStatus())

	SetCloudAuthArguments(nil)
	require.ErrorContains(t, sanitiseInitStatus(), "cloud auth arguments cannot be nil")
}

func TestPackageLevelGetters(t *testing.T) {
	resetDefaultCloudConfig(t)

	SetCloudVendor(CloudVendorAws)
	SetCloudAuthArguments(validCloudAuth())

	objectStorage, err := GetObjectStorage()
	require.NoError(t, err)
	require.Implements(t, (*contract.ObjectStorageInterface)(nil), objectStorage)

	messageQueue, err := GetMessageQueue()
	require.NoError(t, err)
	require.Implements(t, (*contract.MessageQueueInterface)(nil), messageQueue)

	pubSub, err := GetPublishSubscribe()
	require.NoError(t, err)
	require.Implements(t, (*contract.PublishSubscribeInterface)(nil), pubSub)

	emailService, err := GetEmailService()
	require.NoError(t, err)
	require.Implements(t, (*contract.EmailInterface)(nil), emailService)

	visualService, err := GetVisualAnalysisService()
	require.NoError(t, err)
	require.Implements(t, (*contract.VisualAnalysisInterface)(nil), visualService)

	appConfigService, err := GetAppConfigService()
	require.NoError(t, err)
	require.Implements(t, (*contract.AppConfigInterface)(nil), appConfigService)
}

func TestPackageLevelGettersReturnConfigErrors(t *testing.T) {
	resetDefaultCloudConfig(t)

	SetCloudVendor("invalid")
	SetCloudAuthArguments(validCloudAuth())

	_, err := GetObjectStorage()
	require.Error(t, err)
	_, err = GetMessageQueue()
	require.Error(t, err)
	_, err = GetPublishSubscribe()
	require.Error(t, err)
	_, err = GetEmailService()
	require.Error(t, err)
	_, err = GetVisualAnalysisService()
	require.Error(t, err)
	_, err = getAppConfigService()
	require.Error(t, err)
}

func TestClientGettersNilClient(t *testing.T) {
	var client *Client
	ctx := context.Background()

	_, err := client.GetObjectStorage(ctx)
	require.ErrorContains(t, err, "cloud client is not initialized")
	_, err = client.GetMessageQueue(ctx)
	require.ErrorContains(t, err, "cloud client is not initialized")
	_, err = client.GetPublishSubscribe(ctx)
	require.ErrorContains(t, err, "cloud client is not initialized")
	_, err = client.GetEmailService(ctx)
	require.ErrorContains(t, err, "cloud client is not initialized")
	_, err = client.GetAppConfigService(ctx)
	require.ErrorContains(t, err, "cloud client is not initialized")
	_, err = client.GetVisualAnalysisService(ctx)
	require.ErrorContains(t, err, "cloud client is not initialized")
}

func TestClientGettersReuseProviderSafely(t *testing.T) {
	client, err := NewClient(Config{Vendor: CloudVendorAws, AuthArguments: validCloudAuth()})
	require.NoError(t, err)

	const goroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*2)
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			_, err := client.GetObjectStorage(context.Background())
			errCh <- err
			_, err = client.GetMessageQueue(context.Background())
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

type dummyAppConfig struct{}

func (d dummyAppConfig) FetchConfig(context.Context, *appconfig.FetchConfigInput) (*appconfig.FetchConfigResult, error) {
	return &appconfig.FetchConfigResult{}, nil
}

func (d dummyAppConfig) WatchConfig(context.Context, *appconfig.WatchConfigInput, func(result *appconfig.FetchConfigResult)) error {
	return nil
}

func (d dummyAppConfig) GetCurrentConfiguration(string) (*appconfig.FetchConfigResult, error) {
	return &appconfig.FetchConfigResult{}, nil
}

func (d dummyAppConfig) StopWatching(string) error { return nil }

func (d dummyAppConfig) StopAllWatching() {}

func TestGetAppConfigServiceCustomProvider(t *testing.T) {
	resetDefaultCloudConfig(t)
	AppConfigServiceProvider = func() (contract.AppConfigInterface, error) {
		return dummyAppConfig{}, nil
	}

	appCfg, err := GetAppConfigService()
	require.NoError(t, err)
	require.NotNil(t, appCfg)
}

func TestIsValidCloudVendor(t *testing.T) {
	require.True(t, IsValidCloudVendor(CloudVendorAws))
	require.False(t, IsValidCloudVendor(""))
	require.False(t, IsValidCloudVendor("SomeRandomVendor"))
}
