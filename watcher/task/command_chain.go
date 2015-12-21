package task

import (
	"errors"
	"fmt"
	"time"

	"logger"
)

type ChainFunc func(chain *CommandChain, resultCh chan<- error) (directiveCh chan<- TaskDirective)

type CommandChain struct {
	commands []Command
	statusAware
	chainFunc ChainFunc
}

func NewChain(len int) CommandChain {
	return CommandChain{
		commands:  make([]Command, 0, len),
		chainFunc: defaultChainFunc,
	}
}

func (this *CommandChain) RegisterCommand(cmd Command) {
	this.commands = append(this.commands, cmd)
}

func (this *CommandChain) SetChainFunc(chainFunc ChainFunc) {
	this.chainFunc = chainFunc
}

func (this *CommandChain) Run(c chan TaskDirective) <-chan error {
	resultCh := make(chan error, 2)

	go func(resultCh chan error) {
		defer func() {
			<-resultCh
			<-resultCh
			close(resultCh)
		}()
		directiveCh := this.chainFunc(this, resultCh)
		for this.Status() != STOPPING {
			select {
			case directive := <-c:
				logger.Verbose("[this: %p], CommandChain Run. directive= %s, status= %s",
					this, directive, this.Status())
				switch directive {
				case TaskStart:
					if this.Status() == RUNNING {
						resultCh <- new(BusyError)
						break
					}
					directiveCh <- TaskStart
				case TaskRestart:
					if this.Status() == RUNNING {
						directiveCh <- TaskRestart
					} else {
						directiveCh <- TaskStart
					}
				case TaskExit:
					this.setStatus(STOPPING)
					fallthrough
				case TaskStop:
					if this.Status() == RUNNING {
						directiveCh <- TaskStop
					}
				}
			}

		}
	}(resultCh)

	return resultCh
}

func defaultChainFunc(chain *CommandChain, resultCh chan<- error) (directiveCh chan<- TaskDirective) {
	dCh := make(chan TaskDirective)
	go func(directiveCh chan TaskDirective, resultCh chan<- error) {
		defer func() {
			close(directiveCh)
			chain.setStatus(WAITING)
		}()

	PENDING:
		chain.setStatus(PENDING)
		canceled := false
		success := false

		switch directive := <-directiveCh; directive {
		case TaskStart:
			break
		default:
			resultCh <- errors.New(
				fmt.Sprintf("directive error.(must be | TaskStart |, recived: | %s |)", directive))
			return
		}

		chain.setStatus(RUNNING)

	RESTART:
		for _, cmd := range chain.commands {
			completeErr := &CompleteError{
				Name: cmd.name(),
			}
			logger.Debug("will run command:[%s], current status: [%v]", cmd.name(), cmd.Status())
			if cmd.Status() == RUNNING {
				cmd.Kill()
			}
			time.Sleep(200 * time.Millisecond)
			ch, err := cmd.Run()
			if err != nil {
				resultCh <- err
			}
		S:
			select {
			case directive := <-directiveCh:
				switch directive {
				case TaskStop:
					cmd.Kill()
					canceled = true
					goto S
				case TaskRestart:
					goto RESTART
				}
			case processState, ok := <-ch:
				if !ok {
					break
				}
				if processState != nil {
					completeErr.Pid = processState.Pid()
					completeErr.Success = processState.Success()
					completeErr.Interrupt = canceled
				}
			}

			success = completeErr.Success
			resultCh <- completeErr

			if canceled {
				break
			}
			if !success {
				goto PENDING
			}
		}

		resultCh <- &ChainCompleteError{
			Name:      "CommandChain",
			Interrupt: canceled,
			Success:   success,
		}
	}(dCh, resultCh)

	return dCh
}
