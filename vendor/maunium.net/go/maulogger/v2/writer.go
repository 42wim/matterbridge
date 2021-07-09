// mauLogger - A logger for Go programs
// Copyright (C) 2021 Tulir Asokan
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
		lw.writeLine(data[:1])
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
