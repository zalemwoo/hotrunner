package task

type TaskDirective int

const (
	TaskStart TaskDirective = iota
	TaskStop
	TaskRestart
	TaskExit
)

func (t TaskDirective) String() string {
	switch t {
	case TaskStart:
		return "TaskDirective.Start"
	case TaskStop:
		return "TaskDirective.Stop"
	case TaskRestart:
		return "TaskDirective.Restart"
	case TaskExit:
		return "TaskDirective.Exit"
	default:
		return "TaskDirective.Unknown"
	}
}

type Task interface {
	StatusAware
	Run(c chan TaskDirective) <-chan error
}
