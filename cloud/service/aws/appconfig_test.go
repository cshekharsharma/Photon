package aws

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfigdata"
	"github.com/cshekharsharma/photon/cloud/entity/appconfig"
	"github.com/cshekharsharma/photon/core/logger"
)

type mockAcClient struct {
	startConfigurationSessionFunc func(context.Context, *appconfigdata.StartConfigurationSessionInput, ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error)
	getLatestConfigurationFunc    func(context.Context, *appconfigdata.GetLatestConfigurationInput, ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error)
}

func (m *mockAcClient) StartConfigurationSession(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, opts ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
	return m.startConfigurationSessionFunc(ctx, in, opts...)
}

func (m *mockAcClient) GetLatestConfiguration(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, opts ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
	return m.getLatestConfigurationFunc(ctx, in, opts...)
}

func newAppConfigWithMockClient(client awsAppConfigClientInterface) *AppConfig {
	return &AppConfig{
		AcClient: client,
		watchers: make(map[string]context.CancelFunc),
	}
}

func TestWatchConfig_InitializesWatchersMap(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return &appconfigdata.GetLatestConfigurationOutput{
				Configuration: []byte(`{"a":1}`),
				ContentType:   aws.String("application/json"),
			}, nil
		},
	}

	appCfg := &AppConfig{
		AcClient: mockClient,
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "init-watchers",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "init-watchers",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	err := appCfg.WatchConfig(context.Background(), input, func(result *appconfig.FetchConfigResult) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if appCfg.watchers == nil {
		t.Fatalf("expected watchers map to be initialized")
	}
}
func TestFetchConfig_Success(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("test-token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			if in.ConfigurationToken == nil || *in.ConfigurationToken == "" {
				return nil, errors.New("missing token")
			}
			return &appconfigdata.GetLatestConfigurationOutput{
				NextPollConfigurationToken: aws.String("next-token"),
				Configuration:              []byte(`{"some":"config"}`),
				ContentType:                aws.String("application/json"),
			}, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	input := &appconfig.FetchConfigInput{
		ApplicationName: "app",
		ConfigProfile:   "profile",
		EnvironmentName: "env",
	}

	ctx := context.Background()
	result, err := appCfg.FetchConfig(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ConfigProfile != "profile" {
		t.Errorf("Expected ConfigID == 'profile', got %q", result.ConfigID)
	}
	if result.Content != `{"some":"config"}` {
		t.Errorf("Unexpected Content: got %q", result.Content)
	}
	if result.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got %q", result.ContentType)
	}
}

func TestFetchConfig_NilToken(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: nil,
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return nil, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	_, err := appCfg.FetchConfig(context.Background(), &appconfig.FetchConfigInput{
		ApplicationName: "app",
		ConfigProfile:   "profile",
		EnvironmentName: "env",
	})
	if err == nil {
		t.Fatalf("expected error for nil token")
	}
}

func TestFetchConfig_GetLatestError(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return nil, errors.New("latest error")
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	_, err := appCfg.FetchConfig(context.Background(), &appconfig.FetchConfigInput{
		ApplicationName: "app",
		ConfigProfile:   "profile",
		EnvironmentName: "env",
	})
	if err == nil {
		t.Fatalf("expected error for latest configuration")
	}
}
func TestFetchConfig_StartSessionError(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return nil, errors.New("session error")
		},
		getLatestConfigurationFunc: nil, // won't be called
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	ctx := context.Background()
	_, err := appCfg.FetchConfig(ctx, &appconfig.FetchConfigInput{
		ApplicationName: "app",
		ConfigProfile:   "profile",
		EnvironmentName: "env",
	})

	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
	if err.Error() != "failed to start configuration session: session error" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestWatchConfig_AlreadyWatching(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("initial-token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return nil, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "myID",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   time.Second,
	}

	callback := func(result *appconfig.FetchConfigResult) {}

	ctx := context.Background()
	err := appCfg.WatchConfig(ctx, input, callback)
	if err != nil {
		t.Fatalf("Expected no error on first watch, got %v", err)
	}

	err2 := appCfg.WatchConfig(ctx, input, callback)
	if err2 == nil {
		t.Fatal("Expected error on second watch, got nil")
	}
	wantMsg := "watcher for config profile[myID] already exists"
	if err2.Error() != wantMsg {
		t.Errorf("Expected error %q, got %q", wantMsg, err2.Error())
	}
}

func TestGetCurrentConfiguration_Success(t *testing.T) {
	appConfig := &AppConfig{
		watchedConfigs: map[string]*appconfig.FetchConfigResult{
			"config-123": {
				ConfigID:      "config-123",
				ConfigProfile: "profile-a",
				Content:       "example-config-data",
				ContentType:   "application/json",
			},
		},
	}

	result, err := appConfig.GetCurrentConfiguration("config-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if result.ConfigID != "config-123" {
		t.Errorf("expected ConfigID 'config-123', got '%s'", result.ConfigID)
	}
	if result.Content != "example-config-data" {
		t.Errorf("expected Content 'example-config-data', got '%s'", result.Content)
	}
	if result.ContentType != "application/json" {
		t.Errorf("expected ContentType 'application/json', got '%s'", result.ContentType)
	}
}

func TestGetCurrentConfiguration_NotFound(t *testing.T) {
	appConfig := &AppConfig{
		watchedConfigs: map[string]*appconfig.FetchConfigResult{},
	}

	result, err := appConfig.GetCurrentConfiguration("missing-id")
	if err == nil {
		t.Fatal("expected error for missing configuration, got nil")
	}
	expectedErr := fmt.Sprintf("no configuration found for %s", "missing-id")
	if err.Error() != expectedErr {
		t.Errorf("expected error '%s', got '%s'", expectedErr, err.Error())
	}
	if result != nil {
		t.Errorf("expected nil result, got %+v", result)
	}
}

func TestWatchConfig_StartSessionError(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return nil, errors.New("cannot start session")
		},
		getLatestConfigurationFunc: nil,
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "myID",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   time.Second,
	}

	ctx := context.Background()
	err := appCfg.WatchConfig(ctx, input, func(result *appconfig.FetchConfigResult) {})
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	if err.Error() != "failed to start configuration session for myID: cannot start session" {
		t.Errorf("Unexpected error: %v", err)
	}

	if _, exists := appCfg.watchers["myID"]; exists {
		t.Errorf("Expected watcher not to be registered on failure")
	}
}

func TestWatchConfig_NilNextPollConfigurationToken(t *testing.T) {
	var getConfigCalls int32

	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("initial-token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			atomic.AddInt32(&getConfigCalls, 1)
			return &appconfigdata.GetLatestConfigurationOutput{
				Configuration: []byte(`{"update":"true"}`),
				ContentType:   aws.String("application/json"),
			}, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "nilTokenTest",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   10 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "test-nil-token",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := appCfg.WatchConfig(ctx, input, func(result *appconfig.FetchConfigResult) {})
	if err != nil {
		t.Fatalf("Expected no error from WatchConfig, got: %v", err)
	}

	<-ctx.Done()

	if atomic.LoadInt32(&getConfigCalls) == 0 {
		t.Error("Expected GetLatestConfiguration to be called at least once")
	}

	result, err := appCfg.GetCurrentConfiguration("nilTokenTest")
	if err != nil {
		t.Fatalf("Expected config to be stored, got error: %v", err)
	}
	if result.Content != `{"update":"true"}` {
		t.Errorf("Expected updated content, got %q", result.Content)
	}
}

func TestWatchConfig_NilInitialToken(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: nil, // Simulate missing token
			}, nil
		},
		getLatestConfigurationFunc: nil,
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "nil-token-test",
		Application:    "test-app",
		Environment:    "test-env",
		ConfigProfile:  "test-profile",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "testlogger",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	ctx := context.Background()
	err := appCfg.WatchConfig(ctx, input, func(result *appconfig.FetchConfigResult) {})

	// Expected: error due to nil InitialConfigurationToken
	if err == nil {
		t.Fatal("Expected error due to nil InitialConfigurationToken, got nil")
	}

	expected := "received nil InitialConfigurationToken for nil-token-test"
	if err.Error() != expected {
		t.Errorf("Expected error: %q, got: %q", expected, err.Error())
	}

	// Ensure the watcher entry was not retained
	appCfg.mutex.RLock()
	_, exists := appCfg.watchers[input.UniqueConfigID]
	appCfg.mutex.RUnlock()
	if exists {
		t.Errorf("Watcher for %s should have been removed after error", input.UniqueConfigID)
	}
}

func TestWatchConfig_PollingError(t *testing.T) {
	var getConfigCalls int32
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("init-token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			atomic.AddInt32(&getConfigCalls, 1)
			return nil, errors.New("poll error")
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "testPoll",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   10 * time.Millisecond, // short to speed the test
		Logger: logger.Init(&logger.LoggerConfig{
			Provider: logger.LoggerProviderZerolog,
			Name:     "test",
			Type:     logger.LoggerTypeStdout,
			Level:    logger.LogLevelDebug,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := appCfg.WatchConfig(ctx, input, func(result *appconfig.FetchConfigResult) {})
	if err != nil {
		t.Fatalf("Expected no immediate error from WatchConfig, got: %v", err)
	}

	<-ctx.Done()

	calls := atomic.LoadInt32(&getConfigCalls)
	if calls < 1 {
		t.Errorf("Expected getLatestConfigurationFunc to be called multiple times, got %d", calls)
	}
}

func TestStopWatching_NotFound(t *testing.T) {
	mockClient := &mockAcClient{}
	appCfg := newAppConfigWithMockClient(mockClient)

	err := appCfg.StopWatching("doesNotExist")
	if err == nil {
		t.Fatal("Expected error for non-existent config, got nil")
	}
	expected := "no watcher found for config doesNotExist"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestStopWatching_Success(t *testing.T) {
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, in *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return nil, nil
		},
	}
	appCfg := newAppConfigWithMockClient(mockClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wInput := &appconfig.WatchConfigInput{
		UniqueConfigID: "myID",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   time.Second,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "testlogger",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}
	err := appCfg.WatchConfig(ctx, wInput, func(result *appconfig.FetchConfigResult) {})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	stopErr := appCfg.StopWatching("myID")
	if stopErr != nil {
		t.Errorf("Unexpected error stopping watcher: %v", stopErr)
	}

	appCfg.mutex.Lock()
	_, stillExists := appCfg.watchers["myID"]
	appCfg.mutex.Unlock()
	if stillExists {
		t.Error("Expected watcher to be removed after StopWatching")
	}
}

func TestStopAllWatching(t *testing.T) {
	mockClient := &mockAcClient{}
	appCfg := newAppConfigWithMockClient(mockClient)

	_, cancel1 := context.WithCancel(context.Background())
	_, cancel2 := context.WithCancel(context.Background())
	appCfg.watchers["id1"] = cancel1
	appCfg.watchers["id2"] = cancel2

	appCfg.StopAllWatching()

	if len(appCfg.watchers) != 0 {
		t.Errorf("Expected watchers map to be empty, got size %d", len(appCfg.watchers))
	}
}

func TestStartConfigWatcherLoop_PanicRecoveryAndRestart(t *testing.T) {
	var attempt int32
	var wg sync.WaitGroup
	wg.Add(2) // expecting 2 attempts: 1 original + 1 retry

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	appCfg := &AppConfig{
		restartDelay: 20 * time.Millisecond,
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			current := atomic.AddInt32(&attempt, 1)
			wg.Done()
			if current == 1 {
				panic("simulated panic")
			}
			return nil, context.Canceled
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "panic-retry",
		PollInterval:   10 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "testlogger",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("token"), func(result *appconfig.FetchConfigResult) {})

	wg.Wait() // Wait for both attempts to complete

	got := atomic.LoadInt32(&attempt)
	if got != 2 {
		t.Errorf("Expected 2 attempts (1 initial + 1 retry), got %d", got)
	}
}

func TestStartConfigWatcherLoop_PanicWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	appCfg := &AppConfig{
		restartDelay: 10 * time.Millisecond,
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			cancel()
			panic("boom")
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "panic-canceled",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "panic-canceled",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("token"), func(result *appconfig.FetchConfigResult) {})
}

func TestWatchConfigOnce_WithBeforeLockHookAndWatchedConfigsInit(t *testing.T) {
	var hookCalled int32
	mockClient := &mockAcClient{
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return &appconfigdata.GetLatestConfigurationOutput{
				NextPollConfigurationToken: aws.String("next"),
				Configuration:              []byte(`{"a":1}`),
				ContentType:                aws.String("application/json"),
			}, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	appCfg.watchedConfigs = nil
	appCfg.beforeWatchConfigOnceLock = func(a *AppConfig) {
		atomic.AddInt32(&hookCalled, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "hook-test",
		PollInterval:   5 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "hook-test",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	_, _ = appCfg.watchConfigOnce(ctx, input, aws.String("token"), func(result *appconfig.FetchConfigResult) {})

	if atomic.LoadInt32(&hookCalled) == 0 {
		t.Fatalf("expected beforeWatchConfigOnceLock to be called")
	}
	appCfg.mutex.RLock()
	_, ok := appCfg.watchedConfigs[input.UniqueConfigID]
	appCfg.mutex.RUnlock()
	if !ok {
		t.Fatalf("expected watchedConfigs to be initialized and updated")
	}
}
func TestStartConfigWatcherLoop_ContextCanceledImmediately(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	appCfg := &AppConfig{
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			t.Error("watchConfigOnce should not have been called")
			return nil, nil
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "cancel-now",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "testlogger",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("init"), func(result *appconfig.FetchConfigResult) {})
}

func TestStartConfigWatcherLoop_ErrorAndRetry(t *testing.T) {
	var count int32

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	appCfg := &AppConfig{
		restartDelay: 30 * time.Millisecond,
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			atomic.AddInt32(&count, 1)
			return aws.String("new-token"), errors.New("simulated error")
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "error-case",
		PollInterval:   10 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "error-case",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("initial"), func(result *appconfig.FetchConfigResult) {})

	if count < 2 {
		t.Errorf("Expected multiple retry attempts, got %d", count)
	}
}

func TestStartConfigWatcherLoop_GracefulExit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	appCfg := &AppConfig{
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			return nil, nil // no error, so loop ends
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "graceful-exit",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "exit-logger",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("init"), func(result *appconfig.FetchConfigResult) {})
}

func TestStartConfigWatcherLoop_CanceledInsideWatch(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	appCfg := &AppConfig{
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			time.Sleep(50 * time.Millisecond)
			return nil, context.Canceled
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "canceled-inside",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "cancel-case",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("init"), func(result *appconfig.FetchConfigResult) {})
}

func TestWatchConfig_DuplicateContentSkipped(t *testing.T) {
	var callbackCalls int32
	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, _ *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("init-token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			return &appconfigdata.GetLatestConfigurationOutput{
				NextPollConfigurationToken: aws.String("next-token"),
				Configuration:              []byte(`{"same":"content"}`), // identical content every time
				ContentType:                aws.String("application/json"),
			}, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	appCfg.restartDelay = 20 * time.Millisecond

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "dedup-test",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   10 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "dedup-test",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := appCfg.WatchConfig(ctx, input, func(cfg *appconfig.FetchConfigResult) {
		atomic.AddInt32(&callbackCalls, 1)
	})

	if err != nil {
		t.Fatalf("WatchConfig returned error: %v", err)
	}

	<-ctx.Done()

	if atomic.LoadInt32(&callbackCalls) != 1 {
		t.Errorf("Expected callback to be called once, got %d", callbackCalls)
	}
}

func TestWatchConfig_CallsOnContentChange(t *testing.T) {
	var callCount int32
	contentVariants := [][]byte{
		[]byte(`{"key":"v1"}`),
		[]byte(`{"key":"v2"}`),
	}

	var currentIndex int32

	mockClient := &mockAcClient{
		startConfigurationSessionFunc: func(ctx context.Context, _ *appconfigdata.StartConfigurationSessionInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.StartConfigurationSessionOutput, error) {
			return &appconfigdata.StartConfigurationSessionOutput{
				InitialConfigurationToken: aws.String("init-token"),
			}, nil
		},
		getLatestConfigurationFunc: func(ctx context.Context, _ *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			i := atomic.AddInt32(&currentIndex, 1)
			return &appconfigdata.GetLatestConfigurationOutput{
				NextPollConfigurationToken: aws.String(fmt.Sprintf("token-%d", i)),
				Configuration:              contentVariants[(i-1)%int32(len(contentVariants))],
				ContentType:                aws.String("application/json"),
			}, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "multi-update",
		Application:    "app",
		Environment:    "env",
		ConfigProfile:  "profile",
		PollInterval:   15 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "multi-update",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := appCfg.WatchConfig(ctx, input, func(cfg *appconfig.FetchConfigResult) {
		atomic.AddInt32(&callCount, 1)
	})
	if err != nil {
		t.Fatalf("unexpected error from WatchConfig: %v", err)
	}

	<-ctx.Done()

	if atomic.LoadInt32(&callCount) < 2 {
		t.Errorf("expected callback to be triggered on content change, got %d", callCount)
	}
}

func TestStartConfigWatcherLoop_DefaultDelayAndCancelDuringRestart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	appCfg := &AppConfig{
		restartDelay: 0,
		watchConfigOnceFn: func(ctx context.Context, input *appconfig.WatchConfigInput, token *string, callback func(result *appconfig.FetchConfigResult)) (*string, error) {
			cancel()
			return aws.String("new-token"), errors.New("retry")
		},
	}

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "default-delay",
		PollInterval:   time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "default-delay",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	appCfg.startConfigWatcherLoop(ctx, input, aws.String("token"), func(result *appconfig.FetchConfigResult) {})
}

func TestWatchConfigOnce_EmptyConfigAndCancel(t *testing.T) {
	var calls int32
	mockClient := &mockAcClient{
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			atomic.AddInt32(&calls, 1)
			return &appconfigdata.GetLatestConfigurationOutput{
				NextPollConfigurationToken: aws.String("next"),
				Configuration:              []byte{},
				ContentType:                aws.String("application/json"),
			}, nil
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "empty-config",
		PollInterval:   5 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "empty-config",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	_, err := appCfg.watchConfigOnce(ctx, input, aws.String("token"), func(result *appconfig.FetchConfigResult) {})
	if err == nil {
		t.Fatalf("expected context error")
	}
	if atomic.LoadInt32(&calls) == 0 {
		t.Fatalf("expected at least one call")
	}
}

func TestWatchConfigOnce_PanicRecovery(t *testing.T) {
	mockClient := &mockAcClient{
		getLatestConfigurationFunc: func(ctx context.Context, in *appconfigdata.GetLatestConfigurationInput, _ ...func(*appconfigdata.Options)) (*appconfigdata.GetLatestConfigurationOutput, error) {
			panic("boom")
		},
	}

	appCfg := newAppConfigWithMockClient(mockClient)
	input := &appconfig.WatchConfigInput{
		UniqueConfigID: "panic-once",
		PollInterval:   5 * time.Millisecond,
		Logger: logger.Init(&logger.LoggerConfig{
			Name:     "panic-once",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := appCfg.watchConfigOnce(ctx, input, aws.String("token"), func(result *appconfig.FetchConfigResult) {})
	if err == nil {
		t.Fatalf("expected panic recovery error")
	}
}
