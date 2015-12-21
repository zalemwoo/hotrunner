package watcher

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"time"

	"watcher/task"
)

func init() {
	RegisterWatcherType(task.BUILTIN_CMD_GO_RUN, reflect.TypeOf((*GoWatcher)(nil)).Elem())
}

type GoWatcher struct {
	BaseWatcher
}

func (this *GoWatcher) RegisterCommand(cmd task.Command) {
	command, _ := cmd.(*task.ExecCommand)
	tmpDir := os.TempDir()
	fileName := fmt.Sprintf("%s%d", command.Exec, time.Now().Unix())
	fileName = path.Join(tmpDir, fileName)
	paramString := fmt.Sprintf("build -o %s %s", fileName, command.ParamString)
	buildCmd := task.ExecCommand{
		Name:        "go.build",
		Exec:        "go",
		ParamString: paramString,
	}
	this.BaseWatcher.RegisterCommand(&buildCmd)

	execCmd := task.ExecCommand{
		Name:        "go.exec",
		Exec:        fileName,
		ParamString: command.ArgString,
	}
	this.BaseWatcher.RegisterCommand(&execCmd)
}
