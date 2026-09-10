// Package providers contains functionality for managing and interfacing with various AWS services.
// It abstracts the complexities of AWS SDK configuration and provides simplified access to AWS resources like S3, SQS, and SNS.
// The package defines structures and methods for initializing AWS clients based on different authentication modes,
// and it handles client lifecycle management including configuration caching and service client instantiation.
//
// Structures:
//   - AwsCloud: Central structure in the package, providing methods to retrieve configured service clients for AWS services.
//     It supports different authentication mechanisms, including IAM roles and access key based authentication.
//
// Key Functionalities:
//   - Configuring AWS clients: AwsCloud uses the AWS SDK's config and credentials packages to configure and initialize
//     service clients with appropriate authentication credentials.
//   - Service Client Initialization: AwsCloud provides methods to get instances of service clients (for S3, SQS, and SNS),
//     ensuring thread safety and reusability using sync.Mutex for each service client type.
//   - Authentication Handling: Supports IAM role-based and access key-based authentication by handling AWS credentials
//     dynamically based on the configured authentication mode.
package providers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	awsservice "github.com/cshekharsharma/photon/cloud/service/aws"

	"github.com/cshekharsharma/photon/cloud/entity"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var loadDefaultConfigFn = config.LoadDefaultConfig

type awsConfigLoader func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)

// AwsCloud represents a structure to interact with various AWS services.
// It contains AWS configuration, synchronization primitives, and service clients.
type AwsCloud struct {
	AuthArguments *entity.CloudAuthArguments

	configLoader awsConfigLoader
	cfgMutex     sync.Mutex
	cfg          aws.Config // AWS configuration.
	cfgLoaded    bool
	osMutex      sync.Mutex // Mutex for object storage
	mqMutex      sync.Mutex // Mutex for message queue
	psMutex      sync.Mutex // Mutex for pubsub service
	esMutex      sync.Mutex // Mutex for email service
	acMutex      sync.Mutex // Mutex for app config service
	vaMutex      sync.Mutex // Mutex for visual analysis service

	objectstorage *awsservice.ObjectStorage    // Service client for AWS S3.
	messagequeue  *awsservice.MessageQueue     // Service client for AWS SQS.
	pubsub        *awsservice.PublishSubscribe // Service client for AWS SNS
	emailservice  *awsservice.EmailService     // Service client for AWS SES
	appconfig     *awsservice.AppConfig        // Service client for AWS AppConfig
	visual        *awsservice.VisualAnalysis   // Service client for AWS Rekognition
}

func NewAwsCloud(authArguments *entity.CloudAuthArguments) (*AwsCloud, error) {
	auth, err := cloneAndValidateAuthArguments(authArguments)
	if err != nil {
		return nil, err
	}

	return &AwsCloud{
		AuthArguments: auth,
		configLoader:  loadDefaultConfigFn,
	}, nil
}

// getClientArguments initializes and returns the AWS configuration.
// The configuration is set up once and reused for subsequent calls.
func (awscloud *AwsCloud) getClientArguments(ctx context.Context) (aws.Config, error) {
	awscloud.cfgMutex.Lock()
	defer awscloud.cfgMutex.Unlock()

	if awscloud.cfgLoaded {
		return awscloud.cfg, nil
	}

	if awscloud.configLoader == nil {
		awscloud.configLoader = loadDefaultConfigFn
	}

	auth, err := cloneAndValidateAuthArguments(awscloud.AuthArguments)
	if err != nil {
		return aws.Config{}, err
	}
	awscloud.AuthArguments = auth

	authMode := awscloud.sanitiseAuthMode(auth.AuthMode)
	awsRegion := auth.Region

	var cfg aws.Config

	if authMode == entity.AuthTypeIAMRole {
		cfg, err = awscloud.getAwsConfigForIAMAuthMode(ctx, awsRegion)
	} else {
		cfg, err = awscloud.getAwsConfigForKeyBasedAuthMode(ctx, awsRegion)
	}

	if err != nil {
		return aws.Config{}, err
	}
	if cfg.Region == "" {
		return aws.Config{}, errors.New("loaded AWS config is missing region")
	}

	awscloud.cfg = cfg
	awscloud.cfgLoaded = true
	return awscloud.cfg, nil
}

func cloneAndValidateAuthArguments(args *entity.CloudAuthArguments) (*entity.CloudAuthArguments, error) {
	if args == nil {
		return nil, errors.New("cloud auth arguments cannot be nil")
	}
	cloned := *args
	if cloned.Region == "" {
		return nil, errors.New("cloud auth region is required")
	}
	if cloned.AuthMode == entity.AuthTypeAccessKey && (cloned.AccessKey == "" || cloned.SecretKey == "") {
		return nil, errors.New("access key and secret key are required for access-key auth")
	}
	return &cloned, nil
}

// getAwsConfigForIAMAuthMode loads the AWS configuration for a specific IAM authentication mode.
// It initializes the AWS configuration with a specified region and sets up credentials
// using an IAM role ARN from the application configuration.
func (awscloud *AwsCloud) getAwsConfigForIAMAuthMode(ctx context.Context, awsRegion string) (aws.Config, error) {
	cfg, err := awscloud.loadConfig(ctx,
		config.WithRegion(awsRegion),
	)

	if err != nil {
		return cfg, err
	}

	if awscloud.AuthArguments.IAMRole == "" {
		return cfg, nil
	}

	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, awscloud.AuthArguments.IAMRole)
	cfg.Credentials = aws.NewCredentialsCache(provider)

	return cfg, nil
}

// getAwsConfigForKeyBasedAuthMode creates an AWS configuration for the specified region using key-based authentication.
// This function initializes an AWS configuration with the specified region and sets up
// key-based authentication using the StaticCredentialsProvider. It loads the AWS access key, secret key,
// and session ID from the application's configuration.
func (awscloud *AwsCloud) getAwsConfigForKeyBasedAuthMode(ctx context.Context, awsRegion string) (aws.Config, error) {
	accessKey := awscloud.AuthArguments.AccessKey
	secretKey := awscloud.AuthArguments.SecretKey
	sessionId := awscloud.AuthArguments.SessionId

	cfg, err := awscloud.loadConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionId),
		),
	)

	return cfg, err
}

// Sanitizes the provided auth mode if that is valid or not.
// If invalid mode is provided then it returns the default auth mode.
func (awscloud *AwsCloud) sanitiseAuthMode(mode entity.CloudAuthType) entity.CloudAuthType {
	if mode == entity.AuthTypeAccessKey || mode == entity.AuthTypeIAMRole {
		return mode
	}

	return entity.AuthTypeIAMRole
}

func (awscloud *AwsCloud) loadConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	loader := awscloud.configLoader
	if loader == nil {
		loader = loadDefaultConfigFn
	}
	return loader(ctx, optFns...)
}

// GetObjectStorage initializes and returns the ObjectStorage service client for AWS S3.
// The service client is set up once and reused for subsequent calls.
func (awscloud *AwsCloud) GetObjectStorage(ctx context.Context) (*awsservice.ObjectStorage, error) {
	awscloud.osMutex.Lock()
	defer awscloud.osMutex.Unlock()

	if awscloud.objectstorage == nil {
		awsCfg, err := awscloud.getClientArguments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS object storage: %w", err)
		}

		objectstorage := new(awsservice.ObjectStorage)
		objectstorage.S3Client = s3.NewFromConfig(awsCfg)
		objectstorage.S3PresignerClient = s3.NewPresignClient(objectstorage.S3Client.(*s3.Client))

		awscloud.objectstorage = objectstorage
	}

	return awscloud.objectstorage, nil
}

// GetMessageQueue initializes and returns the MessageQueue service client for AWS SQS.
// The service client is set up once and reused for subsequent calls.
func (awscloud *AwsCloud) GetMessageQueue(ctx context.Context) (*awsservice.MessageQueue, error) {
	awscloud.mqMutex.Lock()
	defer awscloud.mqMutex.Unlock()

	if awscloud.messagequeue == nil {
		awsCfg, err := awscloud.getClientArguments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS message queue: %w", err)
		}

		messageQueue := new(awsservice.MessageQueue)
		messageQueue.SqsClient = sqs.NewFromConfig(awsCfg)

		awscloud.messagequeue = messageQueue
	}

	return awscloud.messagequeue, nil
}

// GetPublishSubscribe initializes and returns the PublishSubscribe service client for AWS SNS.
// The service client is set up once and reused for subsequent calls.
func (awscloud *AwsCloud) GetPublishSubscribe(ctx context.Context) (*awsservice.PublishSubscribe, error) {
	awscloud.psMutex.Lock()
	defer awscloud.psMutex.Unlock()

	if awscloud.pubsub == nil {
		awsCfg, err := awscloud.getClientArguments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS publish-subscribe: %w", err)
		}

		pubsub := new(awsservice.PublishSubscribe)
		pubsub.SNSClient = sns.NewFromConfig(awsCfg)

		awscloud.pubsub = pubsub
	}

	return awscloud.pubsub, nil
}

func (awscloud *AwsCloud) GetEmailService(ctx context.Context) (*awsservice.EmailService, error) {
	awscloud.esMutex.Lock()
	defer awscloud.esMutex.Unlock()

	if awscloud.emailservice == nil {
		awsCfg, err := awscloud.getClientArguments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS email service: %w", err)
		}

		emailservice := new(awsservice.EmailService)
		emailservice.SesClient = sesv2.NewFromConfig(awsCfg)

		awscloud.emailservice = emailservice
	}

	return awscloud.emailservice, nil
}

func (awscloud *AwsCloud) GetAppConfigService(ctx context.Context) (*awsservice.AppConfig, error) {
	awscloud.acMutex.Lock()
	defer awscloud.acMutex.Unlock()

	if awscloud.appconfig == nil {
		awsCfg, err := awscloud.getClientArguments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS app config: %w", err)
		}

		appConfigService := new(awsservice.AppConfig)
		appConfigService.AcClient = appconfigdata.NewFromConfig(awsCfg)

		awscloud.appconfig = appConfigService
	}

	return awscloud.appconfig, nil
}

func (awscloud *AwsCloud) GetVisualAnalysisService(ctx context.Context) (*awsservice.VisualAnalysis, error) {
	awscloud.vaMutex.Lock()
	defer awscloud.vaMutex.Unlock()

	if awscloud.visual == nil {
		awsCfg, err := awscloud.getClientArguments(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS visual analysis: %w", err)
		}

		visual := new(awsservice.VisualAnalysis)
		visual.RekognitionClient = rekognition.NewFromConfig(awsCfg)

		awscloud.visual = visual
	}

	return awscloud.visual, nil
}
