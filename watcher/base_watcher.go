package watcher

import (
	"os"
	"path"
	"time"
	"watcher/task"

	"github.com/bmatcuk/doublestar"
	"github.com/go-fsnotify/fsnotify"

	"config"
	"logger"
)

type pathMeta struct {
	path         string
	includePaths []string
	excludePaths []string
	recursive    bool
}

type watcherMeta struct {
	name         string
	duration     time.Duration
	excludePaths []string
	pathMeta     []pathMeta
	targetFiles  []string
}

type BaseWatcher struct {
	meta         watcherMeta
	Name         string
	fsWatcher    *fsnotify.Watcher
	watchingList map[string]bool
	commandChain task.CommandChain
}

func (this *BaseWatcher) loadMeta(c config.ConfigNode) error {
	name, err := c.GetString("name")
	if err != nil {
		this.meta.name = "UNKNOWN"
	} else {
		this.meta.name = name
	}
	this.meta.duration, err = c.GetDuration("duration")
	this.meta.excludePaths, err = c.GetStringList("excludes")
	directories, err := c.GetNodeList("directories")
	if err != nil {
		logger.Fatal("config file error", err)
		return err
	}

	this.meta.pathMeta = make([]pathMeta, len(directories))

	for idx, directory := range directories {
		pathMeta := &this.meta.pathMeta[idx]
		pathMeta.path, err = directory.GetString("path")
		if err != nil {
			logger.Fatal("config file error", err)
			return err
		}
		pathMeta.includePaths, _ = directory.GetStringList("includes")
		excludes, _ := directory.GetStringList("excludes")
		if len(this.meta.excludePaths) > 0 {
			excludes = append(excludes, this.meta.excludePaths...)
		}
		if hasGlobalExcludePatterns {
			excludes = append(excludes, globalExcludePatterns...)
		}
		excludes = sliceRemoveDuplicates(excludes)

		pathMeta.excludePaths = excludes
		pathMeta.recursive, err = directory.GetBool("recursive")
		if err != nil {
			pathMeta.recursive, err = config.GetBool("params:recursive")
		}
	}

	return nil
}

func (this *BaseWatcher) prepare() error {
	includes := make([]string, 0, 10)
	excludes := make([]string, 0, 10)
	for _, pathMeta := range this.meta.pathMeta {
		for _, include := range pathMeta.includePaths {
			pattern := path.Join(pathMeta.path, include)
			matchs, err := doublestar.Glob(pattern)
			if err != nil {
				logger.Warning("expand path error. err= %v", err)
				return err
			}
			if len(matchs) > 0 {
				includes = append(includes, matchs...)
			}
		}
		for _, exclude := range pathMeta.excludePaths {
			pattern := path.Join(pathMeta.path, exclude)
			matchs, err := doublestar.Glob(pattern)
			if err != nil {
				logger.Warning("expand path error. err= %v", err)
				return err
			}
			if len(matchs) > 0 {
				excludes = append(excludes, matchs...)
			}
		}
	}

	includes = sliceRemoveDuplicates(includes)
	excludes = sliceRemoveDuplicates(excludes)
	expandDirectory(&excludes)
	this.meta.targetFiles = sliceDifference(includes, excludes)

	this.Name = this.meta.name
	this.watchingList = make(map[string]bool, len(this.meta.targetFiles))
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Warning("instance fsWatcher error. err= %v", err)
		return err
	}
	this.fsWatcher = fsWatcher
	this.commandChain = task.NewChain(1)
	return nil
}

func (this *BaseWatcher) RegisterCommand(cmd task.Command) {
	this.commandChain.RegisterCommand(cmd)
}

func (this *BaseWatcher) StartWatch() {
	for _, file := range this.meta.targetFiles {
		err := this.AddWatchFile(file)
		if err != nil {
			logger.Warning("add watch file error. err= %v", err)
		}
		continue
	}
}

func (this *BaseWatcher) AddWatchFile(filepath string) error {
	added, ok := this.watchingList[filepath]
	if !ok || added == false {
		err := this.fsWatcher.Add(filepath)
		if err != nil {
			this.watchingList[filepath] = false
			return err
		}
		this.watchingList[filepath] = true
	}
	return nil
}

func (this *BaseWatcher) RemoveWatchFile(filepath string) {
	added, ok := this.watchingList[filepath]
	if !ok {
		goto L
	}
	if added {
		this.fsWatcher.Remove(filepath)
	}
L:
	this.watchingList[filepath] = false
}

func (this *BaseWatcher) Run(exitCh <-chan bool) <-chan error {
	resultCh := make(chan error)
	go func() {
		defer func() {
			close(resultCh)
			this.fsWatcher.Close()
		}()
		runner, _ := NewRunner(&this.commandChain)
		runner.SetMinimalDuration(this.meta.duration)
		taskResultCh := runner.ResultChan()
	WATCHER_RUN:
		for {
			select {
			case event := <-this.fsWatcher.Events:
				logger.Info("file changed. event= %+v", event)
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					go this.rewatch(event.Name)
				}
				runner.Schedule()
			case err := <-this.fsWatcher.Errors:
				resultCh <- err
			case err, ok := <-taskResultCh:
				if !ok {
					break WATCHER_RUN
				}
				switch e := err.(type) {
				case *task.BusyError:
					logger.Warning("watcher is busy. err:  %+v", err)
				case *task.CompleteError:
					logger.Info("command finished: name= %s, pid= %d, Success= %v, Interrupt:= %v", e.Name, e.Pid, e.Success, e.Interrupt)
				case *task.ChainCompleteError:
					logger.Info("command chain finished: name= %s, Success= %v, Interrupt:= %v", e.Name, e.Success, e.Interrupt)
				default:
					resultCh <- err
				}
			case <-exitCh:
				runner.Exit()
				break WATCHER_RUN
			}
		}
	}()
	this.StartWatch()
	return resultCh
}

func (this *BaseWatcher) rewatch(filepath string) {
	this.RemoveWatchFile(filepath)
	time.Sleep(500 * time.Millisecond)
	_, err := os.Stat(filepath)
	if err != nil {
		return
	}
	this.AddWatchFile(filepath)
	return
}

func expandDirectory(dir *[]string) {
	found := []string{}
	for _, item := range *dir {
		file, err := os.Open(item)
		if err != nil {
			continue
		}
		stat, err := file.Stat()
		if err != nil || stat.IsDir() == false {
			continue
		}
		pattern := path.Join(item, "**")
		matchs, err := doublestar.Glob(pattern)
		if len(matchs) > 0 {
			found = append(found, matchs...)
		}
	}

	if len(found) > 0 {
		*dir = append(*dir, found...)
	}
}

func sliceRemoveDuplicates(a []string) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, val := range a {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = true
		}
	}
	return result
}

func sliceDifference(fst []string, snd []string) []string {
	result := []string{}
	for _, fi := range fst {
		found := false
		for _, si := range snd {
			if fi == si {
				found = true
				break
			}
		}
		if found == false {
			result = append(result, fi)
		}
	}
	return result
}
