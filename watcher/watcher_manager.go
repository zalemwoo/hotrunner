package watcher

import (
	"config"
	"errors"
	"os"
	"os/signal"
	"reflect"
	"time"

	"logger"
	"watcher/task"
)

type WatcherManager struct {
	Watchers []Watcher
	args     []string
}

var watcherManager WatcherManager

var globalExcludePatterns []string
var hasGlobalExcludePatterns bool

func NewManager(configFilename string, flagArgs []string) (*WatcherManager, error) {
	err := config.ReadConfigFile(configFilename)
	if err != nil {
		return nil, err
	}

	watchersConf, err := config.GetNodeList("watchers")
	if err != nil {
		return nil, err
	}

	globalExcludePatterns, _ = config.GetStringList("excludes")
	hasGlobalExcludePatterns = len(globalExcludePatterns) > 0

	watcherManager = WatcherManager{
		Watchers: make([]Watcher, len(watchersConf)),
		args:     flagArgs,
	}

	for idx, item := range watchersConf {
		cmd, err := item.GetString("command:type")
		exec, err := item.GetString("command:exec")
		params, err := item.GetString("command:params")
		args, err := item.GetString("command:args")
		command := task.ExecCommand{
			Name:        cmd,
			Exec:        exec,
			ParamString: params,
			ArgString:   args,
		}

		if typeInfo, ok := registeredWatcherType[command.Name]; ok {
			watcherManager.Watchers[idx] = reflect.New(typeInfo.Type).Interface().(Watcher)
		} else {
			return nil, errors.New("unknown watcher type")
		}

		err = watcherManager.Watchers[idx].loadMeta(item)
		if err != nil {
			logger.Fatal("config error: ", err)
			continue
		}
		err = watcherManager.Watchers[idx].prepare()
		if err != nil {
			logger.Fatal("config error: ", err)
			continue
		}

		watcherManager.Watchers[idx].RegisterCommand(&command)
	}

	return &watcherManager, nil
}

func (this *WatcherManager) Run() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, os.Kill)
	stopRunCh := make([]chan bool, len(this.Watchers))
	errchs := []<-chan error{}
	errch := make(chan error)

	for idx, watcher := range this.Watchers {
		stopRunCh[idx] = make(chan bool)
		tc := watcher.Run(stopRunCh[idx])
		errchs = append(errchs, tc)
	}

	go func(errch chan<- error) {
		cases := make([]reflect.SelectCase, len(this.Watchers))
		for idx, ch := range errchs {
			cases[idx] = reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ch),
			}
		}
		remaining := len(cases)
		for remaining > 0 {
			chosen, value, ok := reflect.Select(cases)
			logger.Debug("select watcher channel cases, ok= %v\n", ok)
			if !ok {
				cases[chosen].Chan = reflect.ValueOf(nil)
				remaining -= 1
				continue
			}
			errch <- value.Interface().(error)
		}
	}(errch)

MANAGER_RUN:
	for {
		select {
		case sig := <-sigch:
			logger.Warning("signal trigger, will exit. signal: %v\n", sig)
			for idx, _ := range this.Watchers {
				stopRunCh[idx] <- false
				close(stopRunCh[idx])
			}
			stopRunCh = nil
			logger.Verbose("exit manager running.")
			break MANAGER_RUN
		case err := <-errch:
			logger.Error("WatcherManager error found. err: %+v.", err)
		}
	}
	logger.Debug("WatcherManager sleep 1 second.")
	time.Sleep(1 * time.Second)
}
