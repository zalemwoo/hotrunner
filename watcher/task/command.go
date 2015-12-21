package task

import "os"

const (
	BUILTIN_CMD_GO_RUN string = "builtin.go.run"
	CUSTOM_CMD                = "custom"
)

type Command interface {
	Reset()
	Run() (<-chan *os.ProcessState, error)
	Kill() error
	Status() Status
	name() string
}
