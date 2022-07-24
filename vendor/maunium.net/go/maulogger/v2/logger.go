// mauLogger - A logger for Go programs
// Copyright (c) 2016-2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maulogger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LoggerFileFormat ...
type LoggerFileFormat func(now string, i int) string

type BasicLogger struct {
	PrintLevel         int
	FlushLineThreshold int
	FileTimeFormat     string
	FileFormat         LoggerFileFormat
	TimeFormat         string
	FileMode           os.FileMode
	DefaultSub         Logger

	JSONFile   bool
	JSONStdout bool

	stdoutEncoder *json.Encoder
	fileEncoder   *json.Encoder

	writer     *os.File
	writerLock sync.Mutex
	StdoutLock sync.Mutex
	StderrLock sync.Mutex
	lines      int

	metadata map[string]interface{}
}

// Logger contains advanced logging functions
type Logger interface {
	Sub(module string) Logger
	Subm(module string, metadata map[string]interface{}) Logger
	WithDefaultLevel(level Level) Logger
	GetParent() Logger

	Writer(level Level) io.WriteCloser

	Log(level Level, parts ...interface{})
	Logln(level Level, parts ...interface{})
	Logf(level Level, message string, args ...interface{})
	Logfln(level Level, message string, args ...interface{})

	Debug(parts ...interface{})
	Debugln(parts ...interface{})
	Debugf(message string, args ...interface{})
	Debugfln(message string, args ...interface{})
	Info(parts ...interface{})
	Infoln(parts ...interface{})
	Infof(message string, args ...interface{})
	Infofln(message string, args ...interface{})
	Warn(parts ...interface{})
	Warnln(parts ...interface{})
	Warnf(message string, args ...interface{})
	Warnfln(message string, args ...interface{})
	Error(parts ...interface{})
	Errorln(parts ...interface{})
	Errorf(message string, args ...interface{})
	Errorfln(message string, args ...interface{})
	Fatal(parts ...interface{})
	Fatalln(parts ...interface{})
	Fatalf(message string, args ...interface{})
	Fatalfln(message string, args ...interface{})
}

// Create a Logger
func Createm(metadata map[string]interface{}) Logger {
	var log = &BasicLogger{
		PrintLevel:         10,
		FileTimeFormat:     "2006-01-02",
		FileFormat:         func(now string, i int) string { return fmt.Sprintf("%[1]s-%02[2]d.log", now, i) },
		TimeFormat:         "15:04:05 02.01.2006",
		FileMode:           0600,
		FlushLineThreshold: 5,
		lines:              0,
		metadata:           metadata,
	}
	log.DefaultSub = log.Sub("")
	return log
}

func Create() Logger {
	return Createm(map[string]interface{}{})
}

func (log *BasicLogger) EnableJSONStdout() {
	log.JSONStdout = true
	log.stdoutEncoder = json.NewEncoder(os.Stdout)
}

func (log *BasicLogger) GetParent() Logger {
	return nil
}

// SetWriter formats the given parts with fmt.Sprint and logs the result with the SetWriter level
func (log *BasicLogger) SetWriter(w *os.File) {
	log.writer = w
	if log.JSONFile {
		log.fileEncoder = json.NewEncoder(w)
	}
}

// OpenFile formats the given parts with fmt.Sprint and logs the result with the OpenFile level
func (log *BasicLogger) OpenFile() error {
	now := time.Now().Format(log.FileTimeFormat)
	i := 1
	for ; ; i++ {
		if _, err := os.Stat(log.FileFormat(now, i)); os.IsNotExist(err) {
			break
		}
	}
	writer, err := os.OpenFile(log.FileFormat(now, i), os.O_WRONLY|os.O_CREATE|os.O_APPEND, log.FileMode)
	if err != nil {
		return err
	} else if writer == nil {
		return os.ErrInvalid
	}
	log.SetWriter(writer)
	return nil
}

// Close formats the given parts with fmt.Sprint and logs the result with the Close level
func (log *BasicLogger) Close() error {
	if log.writer != nil {
		return log.writer.Close()
	}
	return nil
}

type logLine struct {
	log *BasicLogger

	Command  string                 `json:"command"`
	Time     time.Time              `json:"time"`
	Level    string                 `json:"level"`
	Module   string                 `json:"module"`
	Message  string                 `json:"message"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (ll logLine) String() string {
	if len(ll.Module) == 0 {
		return fmt.Sprintf("[%s] [%s] %s", ll.Time.Format(ll.log.TimeFormat), ll.Level, ll.Message)
	} else {
		return fmt.Sprintf("[%s] [%s/%s] %s", ll.Time.Format(ll.log.TimeFormat), ll.Module, ll.Level, ll.Message)
	}
}

func reduceItem(m1, m2 map[string]interface{}) map[string]interface{} {
	m3 := map[string]interface{}{}

	_merge := func(m map[string]interface{}) {
		for ia, va := range m {
			m3[ia] = va
		}
	}

	_merge(m1)
	_merge(m2)

	return m3
}

// Raw formats the given parts with fmt.Sprint and logs the result with the Raw level
func (log *BasicLogger) Raw(level Level, extraMetadata map[string]interface{}, module, origMessage string) {
	message := logLine{log, "log", time.Now(), level.Name, module, strings.TrimSpace(origMessage), reduceItem(log.metadata, extraMetadata)}

	if log.writer != nil {
		log.writerLock.Lock()
		var err error
		if log.JSONFile {
			err = log.fileEncoder.Encode(&message)
		} else {
			_, err = log.writer.WriteString(message.String())
			_, _ = log.writer.WriteString("\n")
		}
		log.writerLock.Unlock()
		if err != nil {
			log.StderrLock.Lock()
			_, _ = os.Stderr.WriteString("Failed to write to log file:")
			_, _ = os.Stderr.WriteString(err.Error())
			log.StderrLock.Unlock()
		}
	}

	if level.Severity >= log.PrintLevel {
		if log.JSONStdout {
			log.StdoutLock.Lock()
			_ = log.stdoutEncoder.Encode(&message)
			log.StdoutLock.Unlock()
		} else if level.Severity >= LevelError.Severity {
			log.StderrLock.Lock()
			_, _ = os.Stderr.WriteString(level.GetColor())
			_, _ = os.Stderr.WriteString(message.String())
			_, _ = os.Stderr.WriteString(level.GetReset())
			_, _ = os.Stderr.WriteString("\n")
			log.StderrLock.Unlock()
		} else {
			log.StdoutLock.Lock()
			_, _ = os.Stdout.WriteString(level.GetColor())
			_, _ = os.Stdout.WriteString(message.String())
			_, _ = os.Stdout.WriteString(level.GetReset())
			_, _ = os.Stdout.WriteString("\n")
			log.StdoutLock.Unlock()
		}
	}
}
