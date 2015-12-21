package task

type BusyError struct {
}

func (e *BusyError) Error() string {
	return "is busy"
}

type CompleteError struct {
	Name      string
	Success   bool
	Interrupt bool
	Pid       int
}

func (e *CompleteError) Error() string {
	return "execute complete"
}

type ChainCompleteError struct {
	Name      string
	Success   bool
	Interrupt bool
}

func (e *ChainCompleteError) Error() string {
	return "command chain execute complete"
}
