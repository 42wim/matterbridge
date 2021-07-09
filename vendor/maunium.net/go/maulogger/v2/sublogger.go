// mauLogger - A logger for Go programs
// Copyright (C) 2016-2018 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package maulogger

import (
	"fmt"
)

type Sublogger struct {
	topLevel     *BasicLogger
	parent       Logger
	Module       string
	DefaultLevel Level
}

// Sub creates a Sublogger
func (log *BasicLogger) Sub(module string) Logger {
	return &Sublogger{
		topLevel:     log,
		parent:       log,
		Module:       module,
		DefaultLevel: LevelInfo,
	}
}

// WithDefaultLevel creates a Sublogger with the same Module but different DefaultLevel
func (log *BasicLogger) WithDefaultLevel(lvl Level) Logger {
	return log.DefaultSub.WithDefaultLevel(lvl)
}

func (log *Sublogger) GetParent() Logger {
	return log.parent
}

// Sub creates a Sublogger
func (log *Sublogger) Sub(module string) Logger {
	return &Sublogger{
		topLevel:     log.topLevel,
		parent:       log,
		Module:       fmt.Sprintf("%s/%s", log.Module, module),
		DefaultLevel: log.DefaultLevel,
	}
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
	log.topLevel.Raw(log.DefaultLevel, log.Module, string(p))
	return len(p), nil
}

// Log formats the given parts with fmt.Sprint and logs the result with the given level
func (log *Sublogger) Log(level Level, parts ...interface{}) {
	log.topLevel.Raw(level, "", fmt.Sprint(parts...))
}

// Logln formats the given parts with fmt.Sprintln and logs the result with the given level
func (log *Sublogger) Logln(level Level, parts ...interface{}) {
	log.topLevel.Raw(level, "", fmt.Sprintln(parts...))
}

// Logf formats the given message and args with fmt.Sprintf and logs the result with the given level
func (log *Sublogger) Logf(level Level, message string, args ...interface{}) {
	log.topLevel.Raw(level, "", fmt.Sprintf(message, args...))
}

// Logfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the given level
func (log *Sublogger) Logfln(level Level, message string, args ...interface{}) {
	log.topLevel.Raw(level, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Debug formats the given parts with fmt.Sprint and logs the result with the Debug level
func (log *Sublogger) Debug(parts ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.Module, fmt.Sprint(parts...))
}

// Debugln formats the given parts with fmt.Sprintln and logs the result with the Debug level
func (log *Sublogger) Debugln(parts ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.Module, fmt.Sprintln(parts...))
}

// Debugf formats the given message and args with fmt.Sprintf and logs the result with the Debug level
func (log *Sublogger) Debugf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.Module, fmt.Sprintf(message, args...))
}

// Debugfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Debug level
func (log *Sublogger) Debugfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelDebug, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Info formats the given parts with fmt.Sprint and logs the result with the Info level
func (log *Sublogger) Info(parts ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.Module, fmt.Sprint(parts...))
}

// Infoln formats the given parts with fmt.Sprintln and logs the result with the Info level
func (log *Sublogger) Infoln(parts ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.Module, fmt.Sprintln(parts...))
}

// Infof formats the given message and args with fmt.Sprintf and logs the result with the Info level
func (log *Sublogger) Infof(message string, args ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.Module, fmt.Sprintf(message, args...))
}

// Infofln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Info level
func (log *Sublogger) Infofln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelInfo, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Warn formats the given parts with fmt.Sprint and logs the result with the Warn level
func (log *Sublogger) Warn(parts ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.Module, fmt.Sprint(parts...))
}

// Warnln formats the given parts with fmt.Sprintln and logs the result with the Warn level
func (log *Sublogger) Warnln(parts ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.Module, fmt.Sprintln(parts...))
}

// Warnf formats the given message and args with fmt.Sprintf and logs the result with the Warn level
func (log *Sublogger) Warnf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.Module, fmt.Sprintf(message, args...))
}

// Warnfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Warn level
func (log *Sublogger) Warnfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelWarn, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Error formats the given parts with fmt.Sprint and logs the result with the Error level
func (log *Sublogger) Error(parts ...interface{}) {
	log.topLevel.Raw(LevelError, log.Module, fmt.Sprint(parts...))
}

// Errorln formats the given parts with fmt.Sprintln and logs the result with the Error level
func (log *Sublogger) Errorln(parts ...interface{}) {
	log.topLevel.Raw(LevelError, log.Module, fmt.Sprintln(parts...))
}

// Errorf formats the given message and args with fmt.Sprintf and logs the result with the Error level
func (log *Sublogger) Errorf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelError, log.Module, fmt.Sprintf(message, args...))
}

// Errorfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Error level
func (log *Sublogger) Errorfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelError, log.Module, fmt.Sprintf(message+"\n", args...))
}

// Fatal formats the given parts with fmt.Sprint and logs the result with the Fatal level
func (log *Sublogger) Fatal(parts ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.Module, fmt.Sprint(parts...))
}

// Fatalln formats the given parts with fmt.Sprintln and logs the result with the Fatal level
func (log *Sublogger) Fatalln(parts ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.Module, fmt.Sprintln(parts...))
}

// Fatalf formats the given message and args with fmt.Sprintf and logs the result with the Fatal level
func (log *Sublogger) Fatalf(message string, args ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.Module, fmt.Sprintf(message, args...))
}

// Fatalfln formats the given message and args with fmt.Sprintf, appends a newline and logs the result with the Fatal level
func (log *Sublogger) Fatalfln(message string, args ...interface{}) {
	log.topLevel.Raw(LevelFatal, log.Module, fmt.Sprintf(message+"\n", args...))
}
