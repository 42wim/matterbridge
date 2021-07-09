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
