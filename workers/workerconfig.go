package workers

import "fmt"

// WorkerFactory builds a fresh worker instance for each goroutine launch.
// Returning a new handle per invocation avoids shared mutable state across workers.
type WorkerFactory func() (WorkerInterface, error)

type WorkerConfig struct {
	Name      string
	New       WorkerFactory
	MaxCount  int64
	IsEnabled bool
}

func (cfg *WorkerConfig) Validate() error {
	if cfg == nil {
		return fmt.Errorf("worker config is nil")
	}
	if cfg.Name == "" {
		return fmt.Errorf("worker name is required")
	}
	if cfg.MaxCount <= 0 {
		return fmt.Errorf("worker %s has invalid MaxCount=%d", cfg.Name, cfg.MaxCount)
	}
	if cfg.New == nil {
		return fmt.Errorf("worker %s has no New factory", cfg.Name)
	}
	return nil
}

func (cfg *WorkerConfig) NewWorker() (WorkerInterface, error) {
	if cfg == nil {
		return nil, fmt.Errorf("worker config is nil")
	}

	worker, err := cfg.New()
	if err != nil {
		return nil, fmt.Errorf("worker %s factory error: %w", cfg.Name, err)
	}
	if worker == nil {
		return nil, fmt.Errorf("worker %s factory returned nil worker", cfg.Name)
	}

	return worker, nil
}
