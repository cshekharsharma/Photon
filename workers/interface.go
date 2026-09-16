// Package contract keeps all the interfaces required for background workers.
package workers

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
	// an always running loop. In case of any exception/error or panic
	// situation, this method should be able to recover from that and
	// emit relevant message to the callee, so while current running
	// instance of worker goes down, but the callee is able to respawn
	// another similar instance to carry on the queue processing flow.
	Run(workerChan chan<- WorkerInterface) error
}
