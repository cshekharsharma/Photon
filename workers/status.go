package workers

import "time"

type WorkerStatus struct {
	Name          string
	ID            string
	Running       bool
	RestartCount  int
	LastError     string
	LastStartedAt time.Time
	LastStoppedAt time.Time
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
