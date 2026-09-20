package config

import (
	"context"
	"testing"

	"github.com/cshekharsharma/photon/coordination/network/watcher"
	"github.com/cshekharsharma/photon/utils/filesys"
)

func resetConfigTestState() {
	Close()
}

func newValidOptions(t *testing.T) *Options {
	dir := t.TempDir()
	fullpath := dir + "/config.json"
	if err := filesys.WriteFile(fullpath, []byte(`{"key":"value"}`), 0755); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	return &Options{
		Source:      SourceFile,
		Format:      FormatJson,
		FilePath:    fullpath,
		EnableWatch: true,
		WatcherOptions: &watcher.WatcherOptions{
			WatcherType:  watcher.WatcherTypeAwsAppConfig,
			WatchContext: context.Background(),
		},
	}
}

func TestInit_InvalidOptions(t *testing.T) {
	resetConfigTestState()
	invalidOpt := &Options{
		Source:         0,
		Format:         FormatJson,
		WatcherOptions: &watcher.WatcherOptions{},
	}

	err := Init("koanf", invalidOpt)
	if err == nil {
		t.Error("Expected error for invalid options")
	}
}

func TestInit_SetsValidProvider(t *testing.T) {
	resetConfigTestState()
	opt := newValidOptions(t)
	err := Init("koanf", opt)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestInitDefaultsEmptyProvider(t *testing.T) {
	resetConfigTestState()
	opt := newValidOptions(t)

	if err := Init("", opt); err != nil {
		t.Fatalf("expected empty provider to default, got %v", err)
	}
	if cProvider != ConfigProviderKoanf {
		t.Fatalf("unexpected provider: %s", cProvider)
	}
}

func TestInitRejectsInvalidProviderAndNilOptions(t *testing.T) {
	resetConfigTestState()

	if err := Init("unknown", newValidOptions(t)); err == nil {
		t.Fatal("expected invalid provider error")
	}

	if err := Init(ConfigProviderKoanf, nil); err == nil {
		t.Fatal("expected nil options error")
	}
}

func TestLoad_ReturnsSingletonInstance(t *testing.T) {
	resetConfigTestState()
	opt := newValidOptions(t)
	if err := Init("koanf", opt); err != nil {
		t.Fatalf("expected no init error, got %v", err)
	}

	c := Load()
	if c == nil {
		t.Error("Expected config instance to be created")
	}

	// ensure singleton
	c2 := Load()
	if c != c2 {
		t.Error("Expected singleton config instance")
	}
}

func TestCloseClearsConfigState(t *testing.T) {
	resetConfigTestState()
	opt := newValidOptions(t)
	if err := Init("koanf", opt); err != nil {
		t.Fatalf("expected no init error, got %v", err)
	}
	if cfg, err := LoadE(); err != nil || cfg == nil {
		t.Fatalf("expected loaded config, cfg=%#v err=%v", cfg, err)
	}

	Close()

	if cfg, err := LoadE(); err == nil || cfg != nil {
		t.Fatalf("expected uninitialized error after close, cfg=%#v err=%v", cfg, err)
	}
}

func TestLoadE_ReturnsErrors(t *testing.T) {
	resetConfigTestState()

	if cfg, err := LoadE(); err == nil || cfg != nil {
		t.Fatalf("expected uninitialized error, cfg=%#v err=%v", cfg, err)
	}

	cProvider = "unknown"
	cOptions = &Options{Source: SourceRawBytes, Format: FormatJson, Content: []byte(`{"key":"value"}`)}
	if cfg, err := LoadE(); err == nil || cfg != nil {
		t.Fatalf("expected invalid provider error, cfg=%#v err=%v", cfg, err)
	}

	resetConfigTestState()
	cProvider = ConfigProviderKoanf
	cOptions = &Options{Source: SourceRawBytes, Format: FormatJson, Content: nil}
	if cfg, err := LoadE(); err == nil || cfg != nil {
		t.Fatalf("expected invalid options error, cfg=%#v err=%v", cfg, err)
	}

	resetConfigTestState()
	cProvider = ConfigProviderKoanf
	cOptions = &Options{Source: SourceRawBytes, Format: FormatJson, Content: []byte(`{"bad-json"`), Delimiter: "."}
	if cfg, err := LoadE(); err == nil || cfg != nil {
		t.Fatalf("expected koanf load error, cfg=%#v err=%v", cfg, err)
	}
}

func TestLoadUsesFatalHookOnError(t *testing.T) {
	resetConfigTestState()

	orig := koanfFatalfHook
	defer func() { koanfFatalfHook = orig }()

	called := false
	koanfFatalfHook = func(format string, v ...interface{}) {
		called = true
	}

	if cfg := Load(); cfg != nil {
		t.Fatalf("expected nil config, got %#v", cfg)
	}
	if !called {
		t.Fatal("expected fatal hook")
	}
}

func TestInitWatcher_WhenWatchEnabled(t *testing.T) {
	resetConfigTestState()
	opt := newValidOptions(t)

	if err := Init("koanf", opt); err != nil {
		t.Fatalf("expected no init error, got %v", err)
	}

	instance = nil
	configTest := Load()
	if configTest == nil {
		t.Error("Expected config instance")
	}

	// check default case for watcher
	instance = nil
	opt.WatcherOptions.WatcherType = ""

	configTest = Load()
	if configTest == nil {
		t.Error("Expected config instance")
	}
}

func TestInitWatcher_Disabled(t *testing.T) {
	resetConfigTestState()
	opt := newValidOptions(t)
	opt.EnableWatch = false

	if err := Init("koanf", opt); err != nil {
		t.Fatalf("expected no init error, got %v", err)
	}

	c := Load()
	if c == nil {
		t.Fatal("Expected config to initialize even with watcher disabled")
	}
}
