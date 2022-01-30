// Package logger provides optional debug logging of the Signal library.
package logger

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

// Logger is a shared loggable interface that this library will use for all log messages.
var Logger Loggable

// Loggable is an interface for logging.
type Loggable interface {
	Debug(caller, message string)
	Info(caller, message string)
	Warning(caller, message string)
	Error(caller, message string)
	Configure(settings string)
}

// Setup will configure the shared logger to use the provided logger.
func Setup(logger *Loggable) {
	Logger = *logger
}

// ToString converts an arbitrary number of objects to a string for use in a logger.
func toString(a ...interface{}) string {
	return fmt.Sprint(a...)
}

// EnsureLogger will use the default logger if one was not set up.
func ensureLogger() {
	if Logger == nil {
		// fmt.Println("Error: No logger was configured. Use `logger.Setup` to configure a logger.")
		Logger = &defaultLogger{}
	}
}

// GetCaller gets the go file name and line number that the logger was called from.
func getCaller() string {
	var file string
	_, path, line, _ := runtime.Caller(2)
	paths := strings.Split(path, "/")
	if len(paths) > 0 {
		file = paths[len(paths)-1]
	} else {
		file = "<unkn>"
	}

	return file + ":" + strconv.Itoa(line)
}

/*
 * Go methods used by the library for logging.
 */

// Debug prints debug level logs.
func Debug(msg ...interface{}) {
	ensureLogger()
	Logger.Debug(getCaller(), toString(msg...))
}

// Info prints info level logs.
func Info(msg ...interface{}) {
	ensureLogger()
	Logger.Info(getCaller(), toString(msg...))
}

// Warning prints warning level logs.
func Warning(msg ...interface{}) {
	ensureLogger()
	Logger.Warning(getCaller(), toString(msg...))
}

// Error prints error level logs.
func Error(msg ...interface{}) {
	ensureLogger()
	Logger.Error(getCaller(), toString(msg...))
}

// Configure allows arbitrary logger configuration settings. The
// default logger uses this method to configure what Go files
// are allowed to log.
func Configure(settings string) {
	ensureLogger()
	Logger.Configure(settings)
}
