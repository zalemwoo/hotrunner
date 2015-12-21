package watcher

import (
	"config"
	"errors"
	"reflect"

	"watcher/task"
)

type Watcher interface {
	Run(stopRunCh <-chan bool) (resultCh <-chan error)
	AddWatchFile(filepath string) error
	RemoveWatchFile(filepath string)
	RegisterCommand(cmd task.Command)
	loadMeta(c config.ConfigNode) error
	prepare() error
}

type WatcherTypeInfo struct {
	Name string
	Type reflect.Type
}

func NewWatcherTypeInfo(t reflect.Type) (*WatcherTypeInfo, error) {
	if reflect.PtrTo(t).Implements(reflect.TypeOf((*Watcher)(nil)).Elem()) {
		return &WatcherTypeInfo{
			Name: t.Name(),
			Type: t,
		}, nil
	}
	return nil, errors.New("type MUST implements interface |Watcher|")
}

var registeredWatcherType map[string]*WatcherTypeInfo = make(map[string]*WatcherTypeInfo)

func RegisterWatcherType(name string, watcherType reflect.Type) error {
	t, err := NewWatcherTypeInfo(watcherType)
	if err != nil {
		return err
	}
	registeredWatcherType[name] = t
	return nil
}
