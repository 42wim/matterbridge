// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package mlog

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"

	"github.com/mattermost/logr/v2"
	"github.com/mattermost/logr/v2/formatters"
	"github.com/mattermost/logr/v2/targets"
)

// AddWriterTarget adds a simple io.Writer target to an existing Logger.
// The `io.Writer` can be a buffer which is useful for testing.
// When adding a buffer to collect logs make sure to use `mlog.Buffer` which is
// a thread safe version of `bytes.Buffer`.
func AddWriterTarget(logger *Logger, w io.Writer, useJSON bool, levels ...Level) error {
	filter := logr.NewCustomFilter(levels...)

	var formatter logr.Formatter
	if useJSON {
		formatter = &formatters.JSON{EnableCaller: true}
	} else {
		formatter = &formatters.Plain{EnableCaller: true}
	}

	target := targets.NewWriterTarget(w)
	return logger.log.Logr().AddTarget(target, "_testWriter", filter, formatter, 1000)
}

// CreateConsole createa a logger that outputs to [os.Stdout].
// It's useful in places where no log configuration is accessible.
func CreateConsoleLogger() *Logger {
	logger, err := NewLogger()
	if err != nil {
		panic("failed create logger " + err.Error())
	}

	filter := logr.StdFilter{
		Lvl:        LvlTrace,
		Stacktrace: LvlPanic,
	}
	formatter := &formatters.Plain{
		EnableCaller: true,
		EnableColor:  true,
	}

	target := targets.NewWriterTarget(os.Stdout)
	if err := logger.log.Logr().AddTarget(target, "_testcon", filter, formatter, 1000); err != nil {
		panic("failed to add target " + err.Error())
	}

	return logger
}

// CreateConsoleTestLogger creates a logger for unit tests. Log records are output to `os.Stdout`.
// All log messages with level trace or lower are logged.
// The returned logger get Shutdown() when the tests completes. The caller should not shut it down.
func CreateConsoleTestLogger(tb testing.TB) *Logger {
	tb.Helper()

	logger, err := NewLogger()
	if err != nil {
		tb.Fatalf("failed create logger %v", err)
	}

	filter := logr.StdFilter{
		Lvl:        LvlTrace,
		Stacktrace: LvlPanic,
	}
	formatter := &formatters.Plain{EnableCaller: true}

	target := targets.NewWriterTarget(os.Stdout)
	if err := logger.log.Logr().AddTarget(target, "_testcon", filter, formatter, 1000); err != nil {
		tb.Fatalf("failed to add target %v", err)
	}

	tb.Cleanup(func() {
		err := logger.Shutdown()
		if err != nil {
			tb.Fatalf("failed to shut down test logger %v", err)
		}
	})

	return logger
}

// CreateTestLogger creates a logger for unit tests. Log records are output via `t.Log`.
// All log messages with level trace or lower are logged.
// The returned logger get Shutdown() when the tests completes. The caller should not shut it down.
func CreateTestLogger(t *testing.T) *Logger {
	t.Helper()

	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("failed create logger %v", err)
	}

	filter := logr.StdFilter{
		Lvl:        LvlTrace,
		Stacktrace: LvlPanic,
	}
	formatter := &formatters.Plain{EnableCaller: true}
	target := targets.NewTestingTarget(t)

	if err := logger.log.Logr().AddTarget(target, "test", filter, formatter, 1000); err != nil {
		t.Fatalf("failed to add target %v", err)
	}

	t.Cleanup(func() {
		err := logger.Shutdown()
		if err != nil {
			t.Errorf("failed to shut down test logger %v", err)
		}
	})

	return logger
}

// Buffer provides a thread-safe buffer useful for logging to memory in unit tests.
type Buffer struct {
	buf bytes.Buffer
	mux sync.Mutex
}

func (b *Buffer) Read(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.buf.Read(p)
}
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.buf.Write(p)
}
func (b *Buffer) String() string {
	b.mux.Lock()
	defer b.mux.Unlock()
	return b.buf.String()
}
