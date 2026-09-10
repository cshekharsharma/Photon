package workers

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"sync"
	"time"

	"github.com/cshekharsharma/photon/core/logger"
	"github.com/google/uuid"
)

var (
	startOnce sync.Once // Ensure that the overseer is started only once

	workerList   []*WorkerConfig
	workerChan   chan WorkerInterface
	workerlogger logger.Logger

	restartTimestamps []time.Time
	restartLimit      = 5
	restartWindow     = 60 * time.Second

	workerRestartMu       sync.Mutex
	workerRestartHistory  map[string][]time.Time
	workerRestartLimit    = 10
	workerRestartWindow   = 60 * time.Second
	workerRestartBackoff  = 500 * time.Millisecond
	workerRestartBackoffM = 30 * time.Second

	workerByIDMu     sync.Mutex
	workerConfigByID map[string]*WorkerConfig

	overseerConfigMu     sync.RWMutex
	overseerSleepTimeout = 1 * time.Second
	minOverseerSleep     = 100 * time.Millisecond

	workerFailureSignalMu   sync.Mutex
	workerFailureSignalSeen map[string]time.Time
	workerFailureSignalTTL  = 2 * time.Minute

	workerStatusMu sync.RWMutex
	workerStatuses map[string]WorkerStatus

	overseerOptionsMu sync.RWMutex
	overseerOptions   OverseerOptions

	overseerTestHookMu        sync.RWMutex
	workerChanInitializedHook func(chan WorkerInterface)
	executeOverseerDoneHook   func()
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
}

type WorkerStatus struct {
	Name          string
	ID            string
	Running       bool
	RestartCount  int
	LastError     string
	LastStartedAt time.Time
	LastStoppedAt time.Time
}

// StartOverseer initializes the overseer with a list of workers and a logger,
// ensuring it only starts once. It launches the overseer loop in a separate goroutine.
// The overseer manages lifecycle of workers and restarts them upon failure.
func StartOverseer(workers []*WorkerConfig, logger logger.Logger) {
	StartOverseerWithContext(context.Background(), workers, logger, nil)
}

func StartOverseerWithContext(ctx context.Context, workers []*WorkerConfig, logger logger.Logger, options *OverseerOptions) {
	ctx = normalizeOverseerContext(ctx)
	startOnce.Do(func() {
		workerList = cloneWorkerConfigs(workers)
		workerlogger = resolveWorkerLogger(logger)
		applyOverseerOptions(options)
		workerRestartHistory = make(map[string][]time.Time, len(workerList))
		workerConfigByID = make(map[string]*WorkerConfig)
		workerFailureSignalSeen = make(map[string]time.Time)
		workerStatuses = make(map[string]WorkerStatus)
		startWorkers := true

		workerlogger.Info("[Overseer] Starting overseer job with %d workers: %v", len(workerList), workerList)
		go executeOverseerWithContext(ctx, startWorkers)
	})
}

func normalizeOverseerContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

// executeOverseer is the main loop that initializes the worker pool and monitors it.
// It includes recovery logic to restart itself in case of internal panic, and enforces
// a restart rate limit to avoid infinite crash loops.
func executeOverseer(startWorkers bool) {
	executeOverseerWithContext(context.Background(), startWorkers)
}

func executeOverseerWithContext(ctx context.Context, startWorkers bool) {
	defer notifyExecuteOverseerDone()
	defer func() {
		if r := recover(); r != nil {
			workerlogger.Warn("[WorkerOverseer] Recovered from panic: %+v. Restarting overseer...\n", r)

			now := time.Now()

			pruned := make([]time.Time, 0, len(restartTimestamps))
			for _, ts := range restartTimestamps {
				if now.Sub(ts) <= restartWindow {
					pruned = append(pruned, ts)
				}
			}
			restartTimestamps = append(pruned, now)

			if len(restartTimestamps) > restartLimit {
				workerlogger.Error("[WorkerOverseer] Exceeded %d restarts in %v. "+
					"Shutting down for safety.", restartLimit, restartWindow)

				return // in case of too many restarts, we don't want to restart again
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(GetOverseerSleepTimeout()):
			}
			go executeOverseerWithContext(ctx, false) // restart self
		}
	}()

	// Allocate exact buffer size to avoid blocking on signal.
	totalWorkerCount := 0
	for _, cfg := range workerList {
		if cfg == nil || !cfg.IsEnabled {
			continue
		}
		if err := cfg.Validate(); err != nil {
			workerlogger.Error("[WorkerOverseer] Invalid worker config name=%s: %v", cfg.Name, err)
			continue
		}

		totalWorkerCount += int(cfg.MaxCount)

		workerlogger.Info("[WorkerOverseer] Total %d threads to monitor for worker: %s",
			totalWorkerCount, cfg.Name)
	}
	if totalWorkerCount < 1 {
		totalWorkerCount = 1
	}
	workerChan = make(chan WorkerInterface, totalWorkerCount)
	notifyWorkerChanInitialized(workerChan)

	if startWorkers {
		startAllWorkers()
	}

	monitorWorkersWithContext(ctx)
}

// startAllWorkers initializes and launches all workers based on their configuration.
// Each worker is launched in its own goroutine.
func startAllWorkers() {
	for _, cfg := range workerList {
		if cfg == nil || !cfg.IsEnabled {
			continue
		}
		if err := cfg.Validate(); err != nil {
			workerlogger.Error("[WorkerOverseer] Skipping invalid worker config name=%s: %v", cfg.Name, err)
			continue
		}

		for i := 0; i < int(cfg.MaxCount); i++ {
			workerlogger.Info("[WorkerOverseer] Launching worker: %s (%d/%d)", cfg.Name, i+1, cfg.MaxCount)
			launchWorkerFromConfig(cfg)
		}
	}
}

func launchWorkerFromConfig(cfg *WorkerConfig) {
	if cfg == nil {
		workerlogger.Error("[WorkerOverseer] launchWorkerFromConfig called with nil config")
		return
	}

	worker, err := cfg.NewWorker()
	if err != nil {
		workerlogger.Error("[WorkerOverseer] Failed creating worker name=%s: %v", cfg.Name, err)
		scheduleWorkerRestart(cfg, nil)
		return
	}

	launchWorker(worker, cfg)
}

// launchWorker assigns a new ID to the worker, clears any previous error, and starts it
// in a goroutine. The worker must push itself to the channel on exit or panic.
func launchWorker(worker WorkerInterface, cfg *WorkerConfig) {
	worker.SetWorkerExecutionErr(nil)
	workerID := generateNewWorkerId()
	worker.SetWorkerId(workerID)
	registerRunningWorker(workerID, cfg)
	updateWorkerStatusStarted(workerID, cfg.Name)
	callWorkerHook(func(h WorkerHooks) func(WorkerEvent) { return h.OnStart }, WorkerEvent{
		Name: cfg.Name,
		ID:   workerID,
		At:   time.Now(),
	})

	go runWorkerSafely(worker)

	workerlogger.Info("[WorkerOverseer] Starting worker %s with ID %s",
		worker.GetWorkerName(), worker.GetWorkerId())
}

func runWorkerSafely(worker WorkerInterface) {
	defer func() {
		if r := recover(); r != nil {
			worker.SetWorkerExecutionErr(fmt.Errorf("worker panic: %v", r))
			workerlogger.Error("[WorkerOverseer] Worker %s (id=%s) panicked: %v\n%s",
				worker.GetWorkerName(),
				worker.GetWorkerId(),
				r,
				string(debug.Stack()),
			)
			callWorkerHook(func(h WorkerHooks) func(WorkerEvent) { return h.OnError }, WorkerEvent{
				Name: worker.GetWorkerName(),
				ID:   worker.GetWorkerId(),
				Err:  worker.GetWorkerExecutionErr(),
				At:   time.Now(),
			})
			signalWorkerFailure(worker)
		}
	}()

	if err := worker.Run(workerChan); err != nil {
		worker.SetWorkerExecutionErr(err)
		workerlogger.Error("[WorkerOverseer] Worker %s (id=%s) exited with error: %v",
			worker.GetWorkerName(),
			worker.GetWorkerId(),
			err,
		)
		callWorkerHook(func(h WorkerHooks) func(WorkerEvent) { return h.OnError }, WorkerEvent{
			Name: worker.GetWorkerName(),
			ID:   worker.GetWorkerId(),
			Err:  err,
			At:   time.Now(),
		})
		signalWorkerFailure(worker)
		return
	}

	worker.SetWorkerExecutionErr(fmt.Errorf("worker exited unexpectedly without error"))
	workerlogger.Warn("[WorkerOverseer] Worker %s (id=%s) exited unexpectedly without error. Restarting.",
		worker.GetWorkerName(),
		worker.GetWorkerId(),
	)
	callWorkerHook(func(h WorkerHooks) func(WorkerEvent) { return h.OnError }, WorkerEvent{
		Name: worker.GetWorkerName(),
		ID:   worker.GetWorkerId(),
		Err:  worker.GetWorkerExecutionErr(),
		At:   time.Now(),
	})
	signalWorkerFailure(worker)
}

// monitorWorkers listens for crashed or failed workers on a channel,
// and restarts them automatically.
func monitorWorkers() {
	monitorWorkersWithContext(context.Background())
}

func monitorWorkersWithContext(ctx context.Context) {
	ticker := time.NewTicker(GetOverseerSleepTimeout())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case failedWorker := <-workerChan:
			if failedWorker == nil {
				workerlogger.Error("[WorkerOverseer] Nil worker failure signal received. Skipping restart.")
				continue
			}

			cfg := unregisterRunningWorker(failedWorker.GetWorkerId())
			updateWorkerStatusStopped(failedWorker.GetWorkerId(), failedWorker.GetWorkerExecutionErr())
			callWorkerHook(func(h WorkerHooks) func(WorkerEvent) { return h.OnStop }, WorkerEvent{
				Name: failedWorker.GetWorkerName(),
				ID:   failedWorker.GetWorkerId(),
				Err:  failedWorker.GetWorkerExecutionErr(),
				At:   time.Now(),
			})
			if cfg == nil {
				workerlogger.Warn(
					"[WorkerOverseer] Missing worker config for worker=%s id=%s. "+
						"Skipping restart (likely duplicate or stale failure signal).",
					failedWorker.GetWorkerName(),
					failedWorker.GetWorkerId(),
				)
				continue
			}

			workerlogger.Warn("[WorkerOverseer] Worker %s (id=%s) reported failure. Restarting...",
				failedWorker.GetWorkerName(), failedWorker.GetWorkerId())

			scheduleWorkerRestartWithContext(ctx, cfg, failedWorker)

		case <-ticker.C:
			// keep loop alive
		}
	}
}

func scheduleWorkerRestart(cfg *WorkerConfig, failedWorker WorkerInterface) {
	scheduleWorkerRestartWithContext(context.Background(), cfg, failedWorker)
}

func scheduleWorkerRestartWithContext(ctx context.Context, cfg *WorkerConfig, failedWorker WorkerInterface) {
	workerName := cfg.Name
	if failedWorker != nil && failedWorker.GetWorkerName() != "" {
		workerName = failedWorker.GetWorkerName()
	}

	restartCount, restartDelay, allowed := trackWorkerRestart(workerName)
	if !allowed {
		workerlogger.Error(
			"[WorkerOverseer] Worker %s exceeded restart budget (%d in %v). Not restarting.",
			workerName,
			workerRestartLimit,
			workerRestartWindow,
		)
		return
	}
	updateWorkerStatusRestart(workerName, restartCount)
	callWorkerHook(func(h WorkerHooks) func(WorkerEvent) { return h.OnRestart }, WorkerEvent{
		Name:         workerName,
		ID:           workerIDFromFailure(failedWorker),
		Err:          errorFromFailure(failedWorker),
		RestartCount: restartCount,
		At:           time.Now(),
	})

	workerlogger.Warn(
		"[WorkerOverseer] Worker %s restart attempt=%d delayed by %s",
		workerName,
		restartCount,
		restartDelay,
	)

	timer := time.NewTimer(restartDelay)
	go func() {
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		launchWorkerFromConfig(cfg)
	}()
}

func signalWorkerFailure(worker WorkerInterface) {
	if worker == nil || workerChan == nil {
		return
	}
	if !shouldEmitWorkerFailureSignal(worker.GetWorkerId()) {
		return
	}

	select {
	case workerChan <- worker:
	default:
		workerlogger.Warn("[WorkerOverseer] Worker failure channel is full. Dropping duplicate/stale signal for worker id=%s", worker.GetWorkerId())
	}
}

func shouldEmitWorkerFailureSignal(workerID string) bool {
	if workerID == "" {
		return true
	}

	now := time.Now()
	workerFailureSignalMu.Lock()
	defer workerFailureSignalMu.Unlock()

	if workerFailureSignalSeen == nil {
		workerFailureSignalSeen = make(map[string]time.Time)
	}

	for id, ts := range workerFailureSignalSeen {
		if now.Sub(ts) > workerFailureSignalTTL {
			delete(workerFailureSignalSeen, id)
		}
	}

	if ts, ok := workerFailureSignalSeen[workerID]; ok && now.Sub(ts) <= workerFailureSignalTTL {
		return false
	}

	workerFailureSignalSeen[workerID] = now
	return true
}

func registerRunningWorker(workerID string, cfg *WorkerConfig) {
	if workerID == "" || cfg == nil {
		return
	}
	workerByIDMu.Lock()
	defer workerByIDMu.Unlock()
	if workerConfigByID == nil {
		workerConfigByID = make(map[string]*WorkerConfig)
	}
	workerConfigByID[workerID] = cfg
}

func unregisterRunningWorker(workerID string) *WorkerConfig {
	if workerID == "" {
		return nil
	}
	workerByIDMu.Lock()
	defer workerByIDMu.Unlock()
	if workerConfigByID == nil {
		return nil
	}
	cfg := workerConfigByID[workerID]
	delete(workerConfigByID, workerID)
	return cfg
}

func trackWorkerRestart(workerName string) (int, time.Duration, bool) {
	workerRestartMu.Lock()
	defer workerRestartMu.Unlock()

	if workerRestartHistory == nil {
		workerRestartHistory = make(map[string][]time.Time)
	}

	now := time.Now()
	history := workerRestartHistory[workerName]

	pruned := make([]time.Time, 0, len(history)+1)
	for _, ts := range history {
		if now.Sub(ts) <= workerRestartWindow {
			pruned = append(pruned, ts)
		}
	}

	pruned = append(pruned, now)
	workerRestartHistory[workerName] = pruned

	restartCount := len(pruned)
	if restartCount > workerRestartLimit {
		return restartCount, 0, false
	}

	return restartCount, computeWorkerRestartBackoff(restartCount), true
}

func computeWorkerRestartBackoff(restartCount int) time.Duration {
	if restartCount <= 1 {
		return workerRestartBackoff
	}

	backoff := workerRestartBackoff
	for i := 1; i < restartCount; i++ {
		if backoff >= workerRestartBackoffM/2 {
			return workerRestartBackoffM
		}
		backoff *= 2
	}

	return backoff
}

func SnapshotWorkerStatuses() []WorkerStatus {
	workerStatusMu.RLock()
	defer workerStatusMu.RUnlock()

	statuses := make([]WorkerStatus, 0, len(workerStatuses))
	for _, status := range workerStatuses {
		statuses = append(statuses, status)
	}
	return statuses
}

// generateNewWorkerId returns a new UUID string to assign as worker ID.
func generateNewWorkerId() string {
	return uuid.New().String()
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

func updateWorkerStatusStarted(workerID, workerName string) {
	workerStatusMu.Lock()
	defer workerStatusMu.Unlock()
	if workerStatuses == nil {
		workerStatuses = make(map[string]WorkerStatus)
	}
	status := workerStatuses[workerID]
	status.ID = workerID
	status.Name = workerName
	status.Running = true
	status.LastStartedAt = time.Now()
	workerStatuses[workerID] = status
}

func updateWorkerStatusStopped(workerID string, err error) {
	if workerID == "" {
		return
	}
	workerStatusMu.Lock()
	defer workerStatusMu.Unlock()
	if workerStatuses == nil {
		return
	}
	status := workerStatuses[workerID]
	status.Running = false
	status.LastStoppedAt = time.Now()
	if err != nil {
		status.LastError = err.Error()
	}
	workerStatuses[workerID] = status
}

func updateWorkerStatusRestart(workerName string, restartCount int) {
	workerStatusMu.Lock()
	defer workerStatusMu.Unlock()
	if workerStatuses == nil {
		workerStatuses = make(map[string]WorkerStatus)
	}
	for id, status := range workerStatuses {
		if status.Name == workerName {
			status.RestartCount = restartCount
			workerStatuses[id] = status
		}
	}
}

func workerIDFromFailure(worker WorkerInterface) string {
	if worker == nil {
		return ""
	}
	return worker.GetWorkerId()
}

func errorFromFailure(worker WorkerInterface) error {
	if worker == nil {
		return nil
	}
	return worker.GetWorkerExecutionErr()
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
