package logger

import (
	"flag"
	"log"
	"os"
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
	info = log.New(os.Stdout, "\033[1m\033[32m [INFO] \033[0m", log.Ltime|log.Lshortfile)
	debug = log.New(os.Stdout, "\033[1m\033[36m [DEBUG] \033[0m", log.Ltime|log.Lshortfile)
	errorl = log.New(os.Stdout, "\033[1m\033[31m [ERROR] ", log.Ltime|log.Lshortfile)

	debugflag := flag.String("debug", "false", "Enable Debug Logging")
	flag.Parse()

	if *debugflag == "true" {
		debug_enabled = true
	}
}

func InfoLog(format string, v ...any) {
	info.Printf(format, v...)
}

func DebugLog(format string, v ...any) {
	if debug_enabled {
		debug.Printf(format, v...)
	}
}

func ErrorLog(format string, v ...any) {
	errorl.Printf(format, v...)
}

func DisableDebugLogging() {
	debug_enabled = false
}

func EnableDebugLogging() {
	debug_enabled = false
}
