package watcher

import (
	"errors"
	"time"

	"logger"
	"watcher/task"
)

type Runner interface {
	Schedule()
	Start()
	Exit()
	Stop()
	Restart()
	ResultChan() <-chan error
	SetMinimalDuration(time.Duration)
}

type runner struct {
	minimalDuration time.Duration
	toTaskCh        chan task.TaskDirective
	resultCh        <-chan error
	lastTime        time.Time
	task            task.Task
	timer           *time.Timer
	timerFunc       func()
}

func NewRunner(t task.Task) (Runner, error) {
	if t == nil {
		return nil, errors.New("task can not be nil")
	}
	r := runner{}
	r.task = t
	r.toTaskCh = make(chan task.TaskDirective)
	r.resultCh = r.task.Run(r.toTaskCh)
	r.lastTime = time.Now()
	r.timerFunc = makeTimerFunc(&r)
	r.timer = time.AfterFunc(0, r.timerFunc)

	return &r, nil
}

func (this *runner) ResultChan() <-chan error {
	return this.resultCh
}

func (this *runner) SetMinimalDuration(duration time.Duration) {
	this.minimalDuration = duration
}

func (this *runner) Schedule() {
	if time.Since(this.lastTime) > this.minimalDuration {
		this.timer.Reset(this.minimalDuration)
	}
}

func (this *runner) Start() {
	this.toTaskCh <- task.TaskStart
}

func (this *runner) Exit() {
	this.timer.Stop()
	this.toTaskCh <- task.TaskExit
}

func (this *runner) Stop() {
	this.toTaskCh <- task.TaskStop
}

func (this *runner) Restart() {
	this.toTaskCh <- task.TaskRestart
}

func makeTimerFunc(r *runner) func() {
	return func() {
		logger.Verbose("runner timer started. runner= %+v", r)
		if r.task.Status() == task.RUNNING {
			r.Restart()
		} else {
			r.Start()
		}
		r.lastTime = time.Now()
	}
}
