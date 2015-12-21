package task

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"logger"
)

type ExecCommand struct {
	Name        string
	Exec        string
	ParamString string
	ArgString   string
	statusAware
	cmd *exec.Cmd
}

func (this *ExecCommand) Reset() {
	if this.Status() == RUNNING {
		this.Kill()
	}
	this.cmd = nil
}

func (this *ExecCommand) Run() (<-chan *os.ProcessState, error) {
	if this.Status() == RUNNING {
		return nil, errors.New("command already running")
	}
	this.cmd = exec.Command(this.Exec, strings.Split(this.ParamString, " ")...)
	this.cmd.Stdin = os.Stdin
	this.cmd.Stdout = os.Stdout
	this.cmd.Stderr = os.Stderr
	err := this.cmd.Start()
	if err != nil {
		return nil, err
	}

	logger.Info("ExecCommand::Run() Start. command: %s, args: %v", this.cmd.Path, this.cmd.Args)

	this.setStatus(RUNNING)
	ch := make(chan *os.ProcessState, 1)
	go func(cmd *exec.Cmd, ch chan<- *os.ProcessState) {
		defer func() {
			ch <- this.cmd.ProcessState
			close(ch)
			this.setStatus(WAITING)
		}()
		err := cmd.Wait()
		if err != nil {
			switch e := err.(type) {
			case *exec.Error:
				logger.Error("ExecCommand::Run() exec.Error. name: %s, err: %s", e.Name, e.Err.Error())
			case *exec.ExitError:
				logger.Error("ExecCommand::Run() exec.ExitError. err: %+v", e)
			default:
				logger.Error("ExecCommand::Run() Error. err: %+v", e)
			}
		}

		logger.Info("ExecCommand::Run() Exit . status: %+v", this.cmd.ProcessState)
		logger.Verbose("ExecCommand::Run() Exit. command: %s, args: %v", this.cmd.Path, this.cmd.Args)

	}(this.cmd, ch)
	return ch, nil
}

func (this *ExecCommand) Kill() error {
	logger.Warning("ExecCommand::Kill() kill Start.")
	if this.cmd == nil || this.Status() != RUNNING {
		return errors.New("command not running")
	}
	logger.Verbose("ExecCommand::Kill() kill. command: %s, args: %v", this.cmd.Path, this.cmd.Args)
	return this.cmd.Process.Kill()
}

func (this *ExecCommand) Pid() int {
	if this.cmd == nil || this.Status() != RUNNING {
		return -1
	}
	return this.cmd.Process.Pid
}

func (this *ExecCommand) name() string {
	return this.Name
}
