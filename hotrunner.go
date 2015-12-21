package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"logger"
	"watcher"
)

var appName string

var configFilename string
var verbose bool
var version bool

var watcherManager *watcher.WatcherManager

const (
	VERSION = "0.0.1"
)

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", appName)
	fmt.Fprintf(os.Stderr, "  %s [options] file[s]\n", appName)
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}

func init() {
	const (
		defaultConfigFilename = "config.yml"
		configFlagUsage       = "config file path"

		verboseFlagUsage = "verbose"
		versionFlagUsage = "show version info"
	)

	parts := strings.Split(os.Args[0], string(os.PathSeparator))
	appName = parts[len(parts)-1]

	flag.Usage = Usage

	flag.StringVar(&configFilename, "c", appName+"_"+defaultConfigFilename, configFlagUsage)
	flag.BoolVar(&verbose, "v", false, verboseFlagUsage)
	flag.BoolVar(&version, "version", false, versionFlagUsage)
}

func main() {
	flag.Parse()

	if version {
		showVersion()
		os.Exit(0)
	}

	logger.SetPrefix(appName)
	if verbose {
		logger.SetLevel(logger.VERBOSE)
	} else {
		logger.SetLevel(logger.DEBUG)
	}
	watcherManager, err := watcher.NewManager(configFilename, flag.Args())

	if err != nil {
		logger.Fatal("Watch Manager Create Error. err= ", err)
	}
	watcherManager.Run()
	logger.Info("Exit.")
}

func showVersion() {
	if verbose {
		fmt.Fprintf(os.Stderr, "Version: %s\n", VERSION)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", VERSION)
	}
}
