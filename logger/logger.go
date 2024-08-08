package logger

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	info   *log.Logger
	debug  *log.Logger
	errorl *log.Logger
)

var (
	debug_enabled bool = false
)

func init() {
	info = log.New(os.Stdout, "", 0)
	debug = log.New(os.Stdout, "", 0)
	errorl = log.New(os.Stdout, "", 0)

	debugflag := flag.String("debug", "false", "Enable Debug Logging")
	flag.Parse()

	if *debugflag == "true" {
		debug_enabled = true
	}
}

func InfoLog(format string, v ...any) {
	_, file, line, ok := runtime.Caller(2)

	if ok {
		file = filepath.Base(file)
		info.SetPrefix("\033[1m\033[32m [INFO] \033[0m" + file + ":" + strconv.Itoa(line) + ": ")
		info.Printf(format, v...)
	}

}

func DebugLog(format string, v ...any) {
	if debug_enabled {
		_, file, line, ok := runtime.Caller(2)

		if ok {
			file = filepath.Base(file)
			debug.SetPrefix("\033[1m\033[36m [DEBUG] \033[0m" + file + ":" + strconv.Itoa(line) + ": ")
			debug.Printf(format, v...)
		}
	}
}

func ErrorLog(format string, v ...any) {
	_, file, line, ok := runtime.Caller(2)

	if ok {
		file = filepath.Base(file)
		errorl.SetPrefix("\033[1m\033[31m [ERROR] " + file + ":" + strconv.Itoa(line) + ": ")
		errorl.Printf(format, v...)
	}
}

func DisableDebugLogging() {
	debug_enabled = false
}

func EnableDebugLogging() {
	debug_enabled = false
}
