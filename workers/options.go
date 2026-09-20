package workers

import (
	"reflect"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
)

type RestartPolicy struct {
	Limit      int
	Window     time.Duration
	MinBackoff time.Duration
	MaxBackoff time.Duration
}

type WorkerEvent struct {
	Name         string
	ID           string
	Err          error
	RestartCount int
	At           time.Time
}

type WorkerHooks struct {
	OnStart   func(WorkerEvent)
	OnStop    func(WorkerEvent)
	OnError   func(WorkerEvent)
	OnRestart func(WorkerEvent)
}

type OverseerOptions struct {
	RestartPolicy RestartPolicy
	Hooks         WorkerHooks
	Watchdog      WatchdogOptions
}

type WatchdogOptions struct {
	Interval         time.Duration
	HeartbeatTimeout time.Duration
}

// SetLogger allows the logger to be replaced or injected externally,
// particularly useful in testing or for custom output destinations.
func SetLogger(l logger.Logger) {
	workerlogger = resolveWorkerLogger(l)
}

func SetOverseerSleepTimeout(timeout time.Duration) {
	if timeout <= 0 {
		timeout = 1 * time.Second
	}
	if timeout < minOverseerSleep {
		timeout = minOverseerSleep
	}

	overseerConfigMu.Lock()
	defer overseerConfigMu.Unlock()
	overseerSleepTimeout = timeout
}

func GetOverseerSleepTimeout() time.Duration {
	overseerConfigMu.RLock()
	defer overseerConfigMu.RUnlock()
	return overseerSleepTimeout
}

func SetWorkerWatchdogInterval(interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if interval < minOverseerSleep {
		interval = minOverseerSleep
	}

	overseerConfigMu.Lock()
	defer overseerConfigMu.Unlock()
	workerWatchdogEvery = interval
}

func GetWorkerWatchdogInterval() time.Duration {
	overseerConfigMu.RLock()
	defer overseerConfigMu.RUnlock()
	return workerWatchdogEvery
}

func SetWorkerHeartbeatTimeout(timeout time.Duration) {
	if timeout < 0 {
		timeout = 0
	}

	overseerConfigMu.Lock()
	defer overseerConfigMu.Unlock()
	workerHeartbeatLimit = timeout
}

func GetWorkerHeartbeatTimeout() time.Duration {
	overseerConfigMu.RLock()
	defer overseerConfigMu.RUnlock()
	return workerHeartbeatLimit
}

func resolveWorkerLogger(l logger.Logger) logger.Logger {
	if isNilLogger(l) {
		return logger.Init(&logger.LoggerConfig{
			Name:     "photon_worker_default_logger",
			Provider: logger.LoggerProviderZerolog,
			Type:     logger.LoggerTypeStdout,
			Level:    logger.LogLevelInfo,
		})
	}
	return l
}

func isNilLogger(l logger.Logger) bool {
	if l == nil {
		return true
	}

	v := reflect.ValueOf(l)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func cloneWorkerConfigs(workers []*WorkerConfig) []*WorkerConfig {
	if workers == nil {
		return nil
	}
	cloned := make([]*WorkerConfig, len(workers))
	for i, cfg := range workers {
		if cfg == nil {
			continue
		}
		cfgCopy := *cfg
		cloned[i] = &cfgCopy
	}
	return cloned
}

func applyOverseerOptions(options *OverseerOptions) {
	if options == nil {
		setOverseerOptions(OverseerOptions{})
		return
	}

	if options.RestartPolicy.Limit > 0 {
		workerRestartLimit = options.RestartPolicy.Limit
	}
	if options.RestartPolicy.Window > 0 {
		workerRestartWindow = options.RestartPolicy.Window
	}
	if options.RestartPolicy.MinBackoff > 0 {
		workerRestartBackoff = options.RestartPolicy.MinBackoff
	}
	if options.RestartPolicy.MaxBackoff > 0 {
		workerRestartBackoffM = options.RestartPolicy.MaxBackoff
	}
	if options.Watchdog.Interval > 0 {
		SetWorkerWatchdogInterval(options.Watchdog.Interval)
	}
	if options.Watchdog.HeartbeatTimeout > 0 {
		SetWorkerHeartbeatTimeout(options.Watchdog.HeartbeatTimeout)
	}
	setOverseerOptions(*options)
}

func setOverseerOptions(options OverseerOptions) {
	overseerOptionsMu.Lock()
	defer overseerOptionsMu.Unlock()
	overseerOptions = options
}

func callWorkerHook(selector func(WorkerHooks) func(WorkerEvent), event WorkerEvent) {
	overseerOptionsMu.RLock()
	hook := selector(overseerOptions.Hooks)
	overseerOptionsMu.RUnlock()
	if hook != nil {
		hook(event)
	}
}

func notifyWorkerChanInitialized(ch chan WorkerInterface) {
	overseerTestHookMu.RLock()
	hook := workerChanInitializedHook
	overseerTestHookMu.RUnlock()
	if hook != nil {
		hook(ch)
	}
}

func notifyExecuteOverseerDone() {
	overseerTestHookMu.RLock()
	hook := executeOverseerDoneHook
	overseerTestHookMu.RUnlock()
	if hook != nil {
		hook()
	}
}
