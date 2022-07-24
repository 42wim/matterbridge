// mauLogger - A logger for Go programs
// Copyright (c) 2016-2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maulogger

import (
	"os"
)

// DefaultLogger ...
var DefaultLogger = Create().(*BasicLogger)

// SetWriter formats the given parts with fmt.Sprint and logs the result with the SetWriter level
func SetWriter(w *os.File) {
	DefaultLogger.SetWriter(w)
}

// OpenFile formats the given parts with fmt.Sprint and logs the result with the OpenFile level
func OpenFile() error {
	return DefaultLogger.OpenFile()
}

// Close formats the given parts with fmt.Sprint and logs the result with the Close level
func Close() error {
	return DefaultLogger.Close()
}

// Sub creates a Sublogger
func Sub(module string) Logger {
	return DefaultLogger.Sub(module)
}

// Raw formats the given parts with fmt.Sprint and logs the result with the Raw level
func Rawm(level Level, metadata map[string]interface{}, module, message string) {
	DefaultLogger.Raw(level, metadata, module, message)
}

func Raw(level Level, module, message string) {
	DefaultLogger.Raw(level, map[string]interface{}{}, module, message)
}

// Log formats the given parts with fmt.Sprint and logs the result with the given level
func Log(level Level, parts ...interface{}) {
	DefaultLogger.DefaultSub.Log(level, parts...)
}

// Logln formats the given parts with fmt.Sprintln and logs the result with the given level
func Logln(level Level, parts ...interface{}) {
	DefaultLogger.DefaultSub.Logln(level, parts...)
}

// Logf formats the given message and args with fmt.Sprintf and logs the result with the given level
func Logf(level Level, message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Logf(level, message, args...)
}

// Logfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the given level
func Logfln(level Level, message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Logfln(level, message, args...)
}

// Debug formats the given parts with fmt.Sprint and logs the result with the Debug level
func Debug(parts ...interface{}) {
	DefaultLogger.DefaultSub.Debug(parts...)
}

// Debugln formats the given parts with fmt.Sprintln and logs the result with the Debug level
func Debugln(parts ...interface{}) {
	DefaultLogger.DefaultSub.Debugln(parts...)
}

// Debugf formats the given message and args with fmt.Sprintf and logs the result with the Debug level
func Debugf(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Debugf(message, args...)
}

// Debugfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Debug level
func Debugfln(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Debugfln(message, args...)
}

// Info formats the given parts with fmt.Sprint and logs the result with the Info level
func Info(parts ...interface{}) {
	DefaultLogger.DefaultSub.Info(parts...)
}

// Infoln formats the given parts with fmt.Sprintln and logs the result with the Info level
func Infoln(parts ...interface{}) {
	DefaultLogger.DefaultSub.Infoln(parts...)
}

// Infof formats the given message and args with fmt.Sprintf and logs the result with the Info level
func Infof(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Infof(message, args...)
}

// Infofln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Info level
func Infofln(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Infofln(message, args...)
}

// Warn formats the given parts with fmt.Sprint and logs the result with the Warn level
func Warn(parts ...interface{}) {
	DefaultLogger.DefaultSub.Warn(parts...)
}

// Warnln formats the given parts with fmt.Sprintln and logs the result with the Warn level
func Warnln(parts ...interface{}) {
	DefaultLogger.DefaultSub.Warnln(parts...)
}

// Warnf formats the given message and args with fmt.Sprintf and logs the result with the Warn level
func Warnf(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Warnf(message, args...)
}

// Warnfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Warn level
func Warnfln(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Warnfln(message, args...)
}

// Error formats the given parts with fmt.Sprint and logs the result with the Error level
func Error(parts ...interface{}) {
	DefaultLogger.DefaultSub.Error(parts...)
}

// Errorln formats the given parts with fmt.Sprintln and logs the result with the Error level
func Errorln(parts ...interface{}) {
	DefaultLogger.DefaultSub.Errorln(parts...)
}

// Errorf formats the given message and args with fmt.Sprintf and logs the result with the Error level
func Errorf(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Errorf(message, args...)
}

// Errorfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Error level
func Errorfln(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Errorfln(message, args...)
}

// Fatal formats the given parts with fmt.Sprint and logs the result with the Fatal level
func Fatal(parts ...interface{}) {
	DefaultLogger.DefaultSub.Fatal(parts...)
}

// Fatalln formats the given parts with fmt.Sprintln and logs the result with the Fatal level
func Fatalln(parts ...interface{}) {
	DefaultLogger.DefaultSub.Fatalln(parts...)
}

// Fatalf formats the given message and args with fmt.Sprintf and logs the result with the Fatal level
func Fatalf(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Fatalf(message, args...)
}

// Fatalfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Fatal level
func Fatalfln(message string, args ...interface{}) {
	DefaultLogger.DefaultSub.Fatalfln(message, args...)
}

// Log formats the given parts with fmt.Sprint and logs the result with the given level
func (log *BasicLogger) Log(level Level, parts ...interface{}) {
	log.DefaultSub.Log(level, parts...)
}

// Logln formats the given parts with fmt.Sprintln and logs the result with the given level
func (log *BasicLogger) Logln(level Level, parts ...interface{}) {
	log.DefaultSub.Logln(level, parts...)
}

// Logf formats the given message and args with fmt.Sprintf and logs the result with the given level
func (log *BasicLogger) Logf(level Level, message string, args ...interface{}) {
	log.DefaultSub.Logf(level, message, args...)
}

// Logfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the given level
func (log *BasicLogger) Logfln(level Level, message string, args ...interface{}) {
	log.DefaultSub.Logfln(level, message, args...)
}

// Debug formats the given parts with fmt.Sprint and logs the result with the Debug level
func (log *BasicLogger) Debug(parts ...interface{}) {
	log.DefaultSub.Debug(parts...)
}

// Debugln formats the given parts with fmt.Sprintln and logs the result with the Debug level
func (log *BasicLogger) Debugln(parts ...interface{}) {
	log.DefaultSub.Debugln(parts...)
}

// Debugf formats the given message and args with fmt.Sprintf and logs the result with the Debug level
func (log *BasicLogger) Debugf(message string, args ...interface{}) {
	log.DefaultSub.Debugf(message, args...)
}

// Debugfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Debug level
func (log *BasicLogger) Debugfln(message string, args ...interface{}) {
	log.DefaultSub.Debugfln(message, args...)
}

// Info formats the given parts with fmt.Sprint and logs the result with the Info level
func (log *BasicLogger) Info(parts ...interface{}) {
	log.DefaultSub.Info(parts...)
}

// Infoln formats the given parts with fmt.Sprintln and logs the result with the Info level
func (log *BasicLogger) Infoln(parts ...interface{}) {
	log.DefaultSub.Infoln(parts...)
}

// Infofln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Info level
func (log *BasicLogger) Infofln(message string, args ...interface{}) {
	log.DefaultSub.Infofln(message, args...)
}

// Infof formats the given message and args with fmt.Sprintf and logs the result with the Info level
func (log *BasicLogger) Infof(message string, args ...interface{}) {
	log.DefaultSub.Infof(message, args...)
}

// Warn formats the given parts with fmt.Sprint and logs the result with the Warn level
func (log *BasicLogger) Warn(parts ...interface{}) {
	log.DefaultSub.Warn(parts...)
}

// Warnln formats the given parts with fmt.Sprintln and logs the result with the Warn level
func (log *BasicLogger) Warnln(parts ...interface{}) {
	log.DefaultSub.Warnln(parts...)
}

// Warnfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Warn level
func (log *BasicLogger) Warnfln(message string, args ...interface{}) {
	log.DefaultSub.Warnfln(message, args...)
}

// Warnf formats the given message and args with fmt.Sprintf and logs the result with the Warn level
func (log *BasicLogger) Warnf(message string, args ...interface{}) {
	log.DefaultSub.Warnf(message, args...)
}

// Error formats the given parts with fmt.Sprint and logs the result with the Error level
func (log *BasicLogger) Error(parts ...interface{}) {
	log.DefaultSub.Error(parts...)
}

// Errorln formats the given parts with fmt.Sprintln and logs the result with the Error level
func (log *BasicLogger) Errorln(parts ...interface{}) {
	log.DefaultSub.Errorln(parts...)
}

// Errorf formats the given message and args with fmt.Sprintf and logs the result with the Error level
func (log *BasicLogger) Errorf(message string, args ...interface{}) {
	log.DefaultSub.Errorf(message, args...)
}

// Errorfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Error level
func (log *BasicLogger) Errorfln(message string, args ...interface{}) {
	log.DefaultSub.Errorfln(message, args...)
}

// Fatal formats the given parts with fmt.Sprint and logs the result with the Fatal level
func (log *BasicLogger) Fatal(parts ...interface{}) {
	log.DefaultSub.Fatal(parts...)
}

// Fatalln formats the given parts with fmt.Sprintln and logs the result with the Fatal level
func (log *BasicLogger) Fatalln(parts ...interface{}) {
	log.DefaultSub.Fatalln(parts...)
}

// Fatalf formats the given message and args with fmt.Sprintf and logs the result with the Fatal level
func (log *BasicLogger) Fatalf(message string, args ...interface{}) {
	log.DefaultSub.Fatalf(message, args...)
}

// Fatalfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Fatal level
func (log *BasicLogger) Fatalfln(message string, args ...interface{}) {
	log.DefaultSub.Fatalfln(message, args...)
}
