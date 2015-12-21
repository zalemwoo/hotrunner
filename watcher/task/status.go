package task

import "sync"

type Status int

const (
	WAITING Status = iota
	RUNNING
	PENDING
	STOPPING
)

func (s Status) String() string {
	switch s {
	case WAITING:
		return "Status.Waiting"
	case RUNNING:
		return "Stats.Running"
	case PENDING:
		return "Status.Pending"
	case STOPPING:
		return "Status.Stopping"
	default:
		return "Status.Unknown"
	}
}

type StatusAware interface {
	Status() Status
	setStatus(status Status)
}

type statusAware struct {
	status       Status
	statusLocker sync.RWMutex
}

func (this *statusAware) Status() Status {
	defer this.statusLocker.RUnlock()
	this.statusLocker.RLock()
	status := this.status
	return status
}

func (this *statusAware) setStatus(status Status) {
	defer this.statusLocker.Unlock()
	this.statusLocker.Lock()
	this.status = status
}
