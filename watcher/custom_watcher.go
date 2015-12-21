package watcher

import (
	"reflect"
	"watcher/task"
)

func init() {
	RegisterWatcherType(task.CUSTOM_CMD, reflect.TypeOf((*CustomWatcher)(nil)).Elem())
}

type CustomWatcher struct {
	BaseWatcher
}
