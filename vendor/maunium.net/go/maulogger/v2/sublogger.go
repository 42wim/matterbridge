// mauLogger - A logger for Go programs
// Copyright (c) 2016-2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maulogger

import (
	"fmt"
)

type Sublogger struct {
	topLevel     *BasicLogger
	parent       Logger
	Module       string
	DefaultLevel Level
	metadata     map[string]interface{}
}

// Subm creates a Sublogger
func (log *BasicLogger) Subm(module string, metadata map[string]interface{}) Logger {
	return &Sublogger{
		topLevel:     log,
		parent:       log,
		Module:       module,
		DefaultLevel: LevelInfo,
		metadata:     metadata,
	}
}

func (log *BasicLogger) Sub(module string) Logger {
	return log.Subm(module, map[string]interface{}{})
}

// WithDefaultLevel creates a Sublogger with the same Module but different DefaultLevel
func (log *BasicLogger) WithDefaultLevel(lvl Level) Logger {
	return log.DefaultSub.WithDefaultLevel(lvl)
}

func (log *Sublogger) GetParent() Logger {
	return log.parent
}

// Sub creates a Sublogger
func (log *Sublogger) Subm(module string, metadata map[string]interface{}) Logger {
	if len(module) > 0 {
		module = fmt.Sprintf("%s/%s", log.Module, module)
	} else {
		module = log.Module
	}

	return &Sublogger{
		topLevel:     log.topLevel,
		parent:       log,
		Module:       module,
		DefaultLevel: log.DefaultLevel,
		metadata:     metadata,
	}
}

func (log *Sublogger) Sub(module string) Logger {
	return log.Subm(module, map[string]interface{}{})
}

// WithDefaultLevel creates a Sublogger with the same Module but different DefaultLevel
func (log *Sublogger) WithDefaultLevel(lvl Level) Logger {
	return &Sublogger{
		topLevel:     log.topLevel,
		parent:       log.parent,
		Module:       log.Module,
		DefaultLevel: lvl,
	}
}

// SetModule changes the module name of this Sublogger
func (log *Sublogger) SetModule(mod string) {
	log.Module = mod
}

// SetDefaultLevel changes the default logging level of this Sublogger
func (log *Sublogger) SetDefaultLevel(lvl Level) {
	log.DefaultLevel = lvl
}

// SetParent changes the parent of this Sublogger
func (log *Sublogger) SetParent(parent *BasicLogger) {
	log.topLevel = parent
}

//Write ...
func (log *Sublogger) Write(p []byte) (n int, err error) {
	log.topLevel.Raw(log.DefaultLevel, log.metadata, log.Module, string(p))
	return len(p), nil
}

// Log formats the given parts with fmt.Sprint and logs the result with the given level
func (log *Sublogger) Log(level Level, parts ...interface{}) {
	log.topLevel.Raw(level, log.metadata, log.Module, fmt.Sprint(parts...))
}

// Logln formats the given parts with fmt.Sprintln and logs the result with the given level
func (log *Sublogger) Logln(level Level, parts ...interface{}) {
	log.topLevel.Raw(level, log.metadata, log.Module, fmt.Sprintln(parts...))
}

// Logf formats the given message and args with fmt.Sprintf and logs the result with the given level
func (log *Sublogger) Logf(level Level, message string, args ...interface{}) {
	log.topLevel.Raw(level, log.metadata, log.Module, fmt.Sprintf(message, args...))
}

// Logfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the given level
func (log *Sublogger) Logfln(level Level, message string, args ...interface{}) {
	log.topLevel.Raw(level, log.metadata, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Debug formats the given parts with fmt.Sprint and logs the result with the Debug level
func (log *Sublogger) Debug(parts ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.metadata, log.Module, fmt.Sprint(parts...))
}

// Debugln formats the given parts with fmt.Sprintln and logs the result with the Debug level
func (log *Sublogger) Debugln(parts ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.metadata, log.Module, fmt.Sprintln(parts...))
}

// Debugf formats the given message and args with fmt.Sprintf and logs the result with the Debug level
func (log *Sublogger) Debugf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.metadata, log.Module, fmt.Sprintf(message, args...))
}

// Debugfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Debug level
func (log *Sublogger) Debugfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.metadata, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Info formats the given parts with fmt.Sprint and logs the result with the Info level
func (log *Sublogger) Info(parts ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.metadata, log.Module, fmt.Sprint(parts...))
}

// Infoln formats the given parts with fmt.Sprintln and logs the result with the Info level
func (log *Sublogger) Infoln(parts ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.metadata, log.Module, fmt.Sprintln(parts...))
}

// Infof formats the given message and args with fmt.Sprintf and logs the result with the Info level
func (log *Sublogger) Infof(message string, args ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.metadata, log.Module, fmt.Sprintf(message, args...))
}

// Infofln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Info level
func (log *Sublogger) Infofln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.metadata, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Warn formats the given parts with fmt.Sprint and logs the result with the Warn level
func (log *Sublogger) Warn(parts ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.metadata, log.Module, fmt.Sprint(parts...))
}

// Warnln formats the given parts with fmt.Sprintln and logs the result with the Warn level
func (log *Sublogger) Warnln(parts ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.metadata, log.Module, fmt.Sprintln(parts...))
}

// Warnf formats the given message and args with fmt.Sprintf and logs the result with the Warn level
func (log *Sublogger) Warnf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.metadata, log.Module, fmt.Sprintf(message, args...))
}

// Warnfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Warn level
func (log *Sublogger) Warnfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.metadata, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Error formats the given parts with fmt.Sprint and logs the result with the Error level
func (log *Sublogger) Error(parts ...interface{}) {
	log.topLevel.Raw(LevelError, log.metadata, log.Module, fmt.Sprint(parts...))
}

// Errorln formats the given parts with fmt.Sprintln and logs the result with the Error level
func (log *Sublogger) Errorln(parts ...interface{}) {
	log.topLevel.Raw(LevelError, log.metadata, log.Module, fmt.Sprintln(parts...))
}

// Errorf formats the given message and args with fmt.Sprintf and logs the result with the Error level
func (log *Sublogger) Errorf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelError, log.metadata, log.Module, fmt.Sprintf(message, args...))
}

// Errorfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Error level
func (log *Sublogger) Errorfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelError, log.metadata, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Fatal formats the given parts with fmt.Sprint and logs the result with the Fatal level
func (log *Sublogger) Fatal(parts ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.metadata, log.Module, fmt.Sprint(parts...))
}

// Fatalln formats the given parts with fmt.Sprintln and logs the result with the Fatal level
func (log *Sublogger) Fatalln(parts ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.metadata, log.Module, fmt.Sprintln(parts...))
}

// Fatalf formats the given message and args with fmt.Sprintf and logs the result with the Fatal level
func (log *Sublogger) Fatalf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.metadata, log.Module, fmt.Sprintf(message, args...))
}

// Fatalfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Fatal level
func (log *Sublogger) Fatalfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.metadata, log.Module, fmt.Sprintf(message+"\n", args...))
}
