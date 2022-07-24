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

// Level is the severity level of a log entry.
type Level struct {
	Name            string
	Severity, Color int
}

var (
	// LevelDebug is the level for debug messages.
	LevelDebug = Level{Name: "DEBUG", Color: -1, Severity: 0}
	// LevelInfo is the level for basic log messages.
	LevelInfo = Level{Name: "INFO", Color: 36, Severity: 10}
	// LevelWarn is the level saying that something went wrong, but the program will continue operating mostly normally.
	LevelWarn = Level{Name: "WARN", Color: 33, Severity: 50}
	// LevelError is the level saying that something went wrong and the program may not operate as expected, but will still continue.
	LevelError = Level{Name: "ERROR", Color: 31, Severity: 100}
	// LevelFatal is the level saying that something went wrong and the program will not operate normally.
	LevelFatal = Level{Name: "FATAL", Color: 35, Severity: 9001}
)

// GetColor gets the ANSI escape color code for the log level.
func (lvl Level) GetColor() string {
	if lvl.Color < 0 {
		return "\x1b[0m"
	}
	return fmt.Sprintf("\x1b[%dm", lvl.Color)
}

// GetReset gets the ANSI escape reset code.
func (lvl Level) GetReset() string {
	if lvl.Color < 0 {
		return ""
	}
	return "\x1b[0m"
}
