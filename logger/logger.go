package logger

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/fatih/color"
)

const (
	VERBOSE int = iota
	DEBUG
	INFO
	WARNING
	ERROR
	PANIC
	FATAL
)

const (
	FormatString = "%s %s %s:\n%s %s \n"
)

var (
	Red     func(...interface{}) string
	Green   func(...interface{}) string
	Yellow  func(...interface{}) string
	Blue    func(...interface{}) string
	Magenta func(...interface{}) string
	Cyan    func(...interface{}) string
	White   func(...interface{}) string

	RedBold     func(...interface{}) string
	GreenBold   func(...interface{}) string
	YellowBold  func(...interface{}) string
	BlueBold    func(...interface{}) string
	MagentaBold func(...interface{}) string
	CyanBold    func(...interface{}) string
	WhiteBold   func(...interface{}) string
)

type colorFunc struct {
	identify   string
	boldFunc   func(...interface{}) string
	normalFunc func(...interface{}) string
}
type colorFuncMap map[int]colorFunc

type logger struct {
	prefix    string
	level     int
	verbose   bool
	format    string
	skipFrame int // the frame to skip when get file:line info from call stack frame
}

var (
	levelColorFunc colorFuncMap
	l              logger
)

func init() {
	Red = color.New(color.FgRed).SprintFunc()
	Green = color.New(color.FgGreen).SprintFunc()
	Yellow = color.New(color.FgYellow).SprintFunc()
	Blue = color.New(color.FgBlue).SprintFunc()
	Magenta = color.New(color.FgMagenta).SprintFunc()
	Cyan = color.New(color.FgCyan).SprintFunc()
	White = color.New(color.FgWhite).SprintFunc()

	RedBold = color.New(color.Bold, color.FgRed).SprintFunc()
	GreenBold = color.New(color.Bold, color.FgGreen).SprintFunc()
	YellowBold = color.New(color.Bold, color.FgYellow).SprintFunc()
	BlueBold = color.New(color.Bold, color.FgBlue).SprintFunc()
	MagentaBold = color.New(color.Bold, color.FgMagenta).SprintFunc()
	CyanBold = color.New(color.Bold, color.FgCyan).SprintFunc()
	WhiteBold = color.New(color.Bold, color.FgWhite).SprintFunc()

	levelColorFunc = make(colorFuncMap)
	levelColorFunc[VERBOSE] = colorFunc{CyanBold("[V]"), CyanBold, Cyan}
	levelColorFunc[DEBUG] = colorFunc{BlueBold("[D]"), BlueBold, Blue}
	levelColorFunc[INFO] = colorFunc{GreenBold("[I]"), GreenBold, Green}
	levelColorFunc[WARNING] = colorFunc{YellowBold("[W]"), YellowBold, Yellow}
	levelColorFunc[ERROR] = colorFunc{RedBold("[E]"), RedBold, RedBold}
	levelColorFunc[PANIC] = colorFunc{RedBold("[P]"), RedBold, RedBold}
	levelColorFunc[FATAL] = colorFunc{RedBold("[F]"), RedBold, RedBold}

	l = New("", VERBOSE, true)
	l.skipFrame = 3
}

func New(prefix string, level int, verbose bool) logger {
	return logger{
		Prefix(prefix),
		level,
		verbose,
		FormatString,
		2,
	}
}

func Prefix(prefix string) string {
	return BlueBold("[" + prefix + "]")
}

// pacakge functions
func SetLevel(level int) {
	l.SetLevel(level)
}
func SetPrefix(prefix string) {
	l.SetPrefix(prefix)
}

func Verbose(formatString string, a ...interface{}) {
	l.Verbose(formatString, a...)
}

func Debug(formatString string, a ...interface{}) {
	l.Debug(formatString, a...)
}

func Info(formatString string, a ...interface{}) {
	l.Info(formatString, a...)
}
func Warning(formatString string, a ...interface{}) {
	l.Warning(formatString, a...)
}

func Error(formatString string, a ...interface{}) {
	l.Error(formatString, a...)
}

func Panic(formatString string, a ...interface{}) {
	l.Panic(formatString, a...)
}

func Fatal(formatString string, a ...interface{}) {
	l.Fatal(formatString, a...)
}

// instance functions
func (this *logger) SetLevel(level int) {
	if level > FATAL {
		return
	}
	this.level = level
	if level <= DEBUG {
		this.SetVerbose(true)
	}
}

func (this *logger) SetPrefix(prefix string) {
	this.prefix = Prefix(prefix)
}

func (this *logger) SetVerbose(verbose bool) {
	this.verbose = verbose
}

func (this *logger) Verbose(formatString string, a ...interface{}) {
	this.printf(VERBOSE, formatString, a...)
}

func (this *logger) Debug(formatString string, a ...interface{}) {
	this.printf(DEBUG, formatString, a...)
}

func (this *logger) Info(formatString string, a ...interface{}) {
	this.printf(INFO, formatString, a...)
}

func (this *logger) Warning(formatString string, a ...interface{}) {
	this.printf(WARNING, formatString, a...)
}

func (this *logger) Error(formatString string, a ...interface{}) {
	this.printf(ERROR, formatString, a...)
}

func (this *logger) Panic(formatString string, a ...interface{}) {
	this.printf(PANIC, formatString, a...)
	panic(errors.New(""))
}

func (this *logger) Fatal(formatString string, a ...interface{}) {
	this.printf(FATAL, formatString, a...)
	os.Exit(1)
}

func (this *logger) printf(level int, formatString string, a ...interface{}) {
	if level < this.level {
		return
	}
	timeString := time.Now().Format("2006-01-02 15:04:05.00000")
	f, ok := levelColorFunc[level]
	if !ok {
		panic("log level not found")
	}

	_, file, line, ok := runtime.Caller(this.skipFrame)
	if !ok {
		file = "<unknown>"
		line = -1
	} else {
		file = filepath.Base(file)
	}
	lineInfo := WhiteBold(fmt.Sprintf("%s:%d", file, line))

	str := fmt.Sprintf(formatString, a...)
	fmt.Fprintf(os.Stderr, this.format, f.normalFunc(timeString), this.prefix, lineInfo, f.identify, f.normalFunc(str))

	if level >= ERROR {
		callStack := debug.Stack()
		fmt.Fprint(os.Stderr, RedBold("=== CALL STACK [START] ===\n"))
		fmt.Fprint(os.Stderr, Red(string(callStack)))
		fmt.Fprint(os.Stderr, RedBold("*** CALL STACK [ END ] ***\n"))
	}
}
