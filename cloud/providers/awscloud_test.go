package providers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/cshekharsharma/photon/cloud/entity"
	"github.com/stretchr/testify/require"
)

type contextKey string

func validAuthArguments(authType entity.CloudAuthType) *entity.CloudAuthArguments {
	return &entity.CloudAuthArguments{
		AuthMode:  authType,
		AccessKey: "access-key",
		SecretKey: "secret-key",
		SessionId: "session-id",
		IAMRole:   "arn:aws:iam::123456789012:role/test-role",
		Region:    "us-east-1",
	}
}

func stubLoadDefaultConfig(t *testing.T, fn awsConfigLoader) {
	t.Helper()
	orig := loadDefaultConfigFn
	loadDefaultConfigFn = fn
	t.Cleanup(func() { loadDefaultConfigFn = orig })
}

func successLoader(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return aws.Config{Region: "us-east-1"}, nil
}

func TestNewAwsCloudValidatesAndClonesAuth(t *testing.T) {
	auth := validAuthArguments(entity.AuthTypeAccessKey)
	cloud, err := NewAwsCloud(auth)
	require.NoError(t, err)
	require.NotSame(t, auth, cloud.AuthArguments)

	auth.Region = "mutated"
	require.Equal(t, "us-east-1", cloud.AuthArguments.Region)
	require.Equal(t, entity.AuthTypeIAMRole, cloud.sanitiseAuthMode(entity.CloudAuthType(999)))
	require.Equal(t, entity.AuthTypeAccessKey, cloud.sanitiseAuthMode(entity.AuthTypeAccessKey))
}

func TestNewAwsCloudRejectsInvalidAuth(t *testing.T) {
	for _, tc := range []struct {
		name string
		auth *entity.CloudAuthArguments
		want string
	}{
		{name: "nil", auth: nil, want: "cloud auth arguments cannot be nil"},
		{name: "missing region", auth: &entity.CloudAuthArguments{AuthMode: entity.AuthTypeIAMRole}, want: "cloud auth region is required"},
		{name: "missing access key", auth: &entity.CloudAuthArguments{AuthMode: entity.AuthTypeAccessKey, Region: "us-east-1", SecretKey: "secret"}, want: "access key and secret key are required"},
		{name: "missing secret key", auth: &entity.CloudAuthArguments{AuthMode: entity.AuthTypeAccessKey, Region: "us-east-1", AccessKey: "access"}, want: "access key and secret key are required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cloud, err := NewAwsCloud(tc.auth)
			require.Nil(t, cloud)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestGetClientArguments(t *testing.T) {
	t.Run("access key propagates context and caches config", func(t *testing.T) {
		expectedCtx := context.WithValue(context.Background(), contextKey("ctx"), "access-key")
		var calls atomic.Int32

		stubLoadDefaultConfig(t, func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			require.Same(t, expectedCtx, ctx)
			calls.Add(1)
			return aws.Config{Region: "us-west-2"}, nil
		})

		cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeAccessKey))
		require.NoError(t, err)

		cfg, err := cloud.getClientArguments(expectedCtx)
		require.NoError(t, err)
		require.Equal(t, "us-west-2", cfg.Region)

		cfg, err = cloud.getClientArguments(context.Background())
		require.NoError(t, err)
		require.Equal(t, "us-west-2", cfg.Region)
		require.Equal(t, int32(1), calls.Load())
	})

	t.Run("iam auth", func(t *testing.T) {
		stubLoadDefaultConfig(t, successLoader)
		cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeIAMRole))
		require.NoError(t, err)

		cfg, err := cloud.getClientArguments(context.Background())
		require.NoError(t, err)
		require.Equal(t, "us-east-1", cfg.Region)
		require.NotNil(t, cfg.Credentials)
	})

	t.Run("load failure surfaces", func(t *testing.T) {
		stubLoadDefaultConfig(t, func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, errors.New("load config failed")
		})
		cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeAccessKey))
		require.NoError(t, err)

		cfg, err := cloud.getClientArguments(context.Background())
		require.Equal(t, aws.Config{}, cfg)
		require.EqualError(t, err, "load config failed")
	})

	t.Run("empty loaded config is rejected", func(t *testing.T) {
		stubLoadDefaultConfig(t, func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			return aws.Config{}, nil
		})
		cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeAccessKey))
		require.NoError(t, err)

		cfg, err := cloud.getClientArguments(context.Background())
		require.Equal(t, aws.Config{}, cfg)
		require.EqualError(t, err, "loaded AWS config is missing region")
	})

	t.Run("manual construction validates auth", func(t *testing.T) {
		cloud := &AwsCloud{}
		cfg, err := cloud.getClientArguments(context.Background())
		require.Equal(t, aws.Config{}, cfg)
		require.ErrorContains(t, err, "cloud auth arguments cannot be nil")
	})
}

func TestConfigHelpers(t *testing.T) {
	stubLoadDefaultConfig(t, successLoader)

	iamCloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeIAMRole))
	require.NoError(t, err)
	iamCloud.AuthArguments.IAMRole = ""
	cfg, err := iamCloud.getAwsConfigForIAMAuthMode(context.Background(), "us-east-1")
	require.NoError(t, err)
	require.Equal(t, "us-east-1", cfg.Region)

	stubLoadDefaultConfig(t, func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("iam load failed")
	})
	iamCloud.configLoader = loadDefaultConfigFn
	_, err = iamCloud.getAwsConfigForIAMAuthMode(context.Background(), "us-east-1")
	require.EqualError(t, err, "iam load failed")

	_, err = iamCloud.getAwsConfigForKeyBasedAuthMode(context.Background(), "us-east-1")
	require.EqualError(t, err, "iam load failed")

	loadDefaultConfigFn = successLoader
	manualCloud := &AwsCloud{
		AuthArguments: validAuthArguments(entity.AuthTypeAccessKey),
	}
	cfg, err = manualCloud.getAwsConfigForKeyBasedAuthMode(context.Background(), "us-east-1")
	require.NoError(t, err)
	require.Equal(t, "us-east-1", cfg.Region)
}

func TestServiceGetters(t *testing.T) {
	stubLoadDefaultConfig(t, successLoader)
	cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeAccessKey))
	require.NoError(t, err)
	ctx := context.Background()

	objectStorage, err := cloud.GetObjectStorage(ctx)
	require.NoError(t, err)
	require.NotNil(t, objectStorage.S3Client)
	require.NotNil(t, objectStorage.S3PresignerClient)

	messageQueue, err := cloud.GetMessageQueue(ctx)
	require.NoError(t, err)
	require.NotNil(t, messageQueue.SqsClient)

	pubSub, err := cloud.GetPublishSubscribe(ctx)
	require.NoError(t, err)
	require.NotNil(t, pubSub.SNSClient)

	emailService, err := cloud.GetEmailService(ctx)
	require.NoError(t, err)
	require.NotNil(t, emailService.SesClient)

	appConfig, err := cloud.GetAppConfigService(ctx)
	require.NoError(t, err)
	require.NotNil(t, appConfig.AcClient)

	visualAnalysis, err := cloud.GetVisualAnalysisService(ctx)
	require.NoError(t, err)
	require.NotNil(t, visualAnalysis.RekognitionClient)
}

func TestServiceGettersReturnConfigErrors(t *testing.T) {
	stubLoadDefaultConfig(t, func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("load failed")
	})
	cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeAccessKey))
	require.NoError(t, err)
	ctx := context.Background()

	_, err = cloud.GetObjectStorage(ctx)
	require.ErrorContains(t, err, "failed to initialize AWS object storage: load failed")
	_, err = cloud.GetMessageQueue(ctx)
	require.ErrorContains(t, err, "failed to initialize AWS message queue: load failed")
	_, err = cloud.GetPublishSubscribe(ctx)
	require.ErrorContains(t, err, "failed to initialize AWS publish-subscribe: load failed")
	_, err = cloud.GetEmailService(ctx)
	require.ErrorContains(t, err, "failed to initialize AWS email service: load failed")
	_, err = cloud.GetAppConfigService(ctx)
	require.ErrorContains(t, err, "failed to initialize AWS app config: load failed")
	_, err = cloud.GetVisualAnalysisService(ctx)
	require.ErrorContains(t, err, "failed to initialize AWS visual analysis: load failed")
}

func TestConcurrentServiceGetterReuse(t *testing.T) {
	var calls atomic.Int32
	stubLoadDefaultConfig(t, func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		calls.Add(1)
		return aws.Config{Region: "us-east-1"}, nil
	})
	cloud, err := NewAwsCloud(validAuthArguments(entity.AuthTypeAccessKey))
	require.NoError(t, err)

	const goroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			_, err := cloud.GetObjectStorage(context.Background())
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	require.Equal(t, int32(1), calls.Load())
}
