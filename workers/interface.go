// Package contract keeps all the interfaces required for background workers.
package workers

import "context"

// WorkerRuntime is the per-run context handed to a worker.
// Workers must call Beat while making healthy progress; the overseer cancels
// this context when the worker is stale or the overseer is shutting down.
type WorkerRuntime interface {
	context.Context
	Beat()
	WorkerID() string
	WorkerName() string
}

// Worker interface provides a common contract for all kinds of
// background workers. These workers are monitored and managed
// by an worker overseer, that invokes methods of worker interface's
// implementations from outside.
type WorkerInterface interface {

	// Get human readable name for the worker. This human readable name
	// can be useful in more contextualised logging, and tagging purposes.
	GetWorkerName() string

	// Get unique alphanumeric worker id for the current instance
	// of running worker (go-routine)
	GetWorkerId() string

	// Set unique alphanumeric worker id for the current instance
	// of running worker (go-routine)
	SetWorkerId(id string)

	// Get errors that have occurred during the execution
	// of current execution cycle of the worker loop.
	// Usually these errors are caught and recovered through
	// panic-recover workflow.
	GetWorkerExecutionErr() error

	// Set errors to the current instance of running worker
	// that have been occurred during the execution.
	SetWorkerExecutionErr(err error)

	// The main execution method of the worker implementation.
	// This method is called from outside by worker overseer and
	// is responsible for processing all the queued messages through
	// an always running loop. It must return when the runtime context is
	// canceled, and call runtime.Beat while making healthy progress.
	Run(runtime WorkerRuntime) error
}
