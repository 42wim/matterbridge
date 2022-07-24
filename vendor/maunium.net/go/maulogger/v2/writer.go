// mauLogger - A logger for Go programs
// Copyright (c) 2021 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maulogger

import (
	"bytes"
	"io"
	"sync"
)

// LogWriter is a buffered io.Writer that writes lines to a Logger.
type LogWriter struct {
	log   Logger
	lock  sync.Mutex
	level Level
	buf   bytes.Buffer
}

func (log *BasicLogger) Writer(level Level) io.WriteCloser {
	return &LogWriter{
		log:   log,
		level: level,
	}
}

func (log *Sublogger) Writer(level Level) io.WriteCloser {
	return &LogWriter{
		log:   log,
		level: level,
	}
}

func (lw *LogWriter) writeLine(data []byte) {
	if lw.buf.Len() == 0 {
		if len(data) == 0 {
			return
		}
		lw.log.Logln(lw.level, string(data))
	} else {
		lw.buf.Write(data)
		lw.log.Logln(lw.level, lw.buf.String())
		lw.buf.Reset()
	}
}

// Write will write lines from the given data to the buffer. If the data doesn't end with a line break,
// everything after the last line break will be buffered until the next Write or Close call.
func (lw *LogWriter) Write(data []byte) (int, error) {
	lw.lock.Lock()
	newline := bytes.IndexByte(data, '\n')
	if newline == len(data)-1 {
		lw.writeLine(data[:len(data)-1])
	} else if newline < 0 {
		lw.buf.Write(data)
	} else {
		lines := bytes.Split(data, []byte("\n"))
		for _, line := range lines[:len(lines)-1] {
			lw.writeLine(line)
		}
		lw.buf.Write(lines[len(lines)-1])
	}
	lw.lock.Unlock()
	return len(data), nil
}

// Close will flush remaining data in the buffer into the logger.
func (lw *LogWriter) Close() error {
	lw.lock.Lock()
	lw.log.Logln(lw.level, lw.buf.String())
	lw.buf.Reset()
	lw.lock.Unlock()
	return nil
}
