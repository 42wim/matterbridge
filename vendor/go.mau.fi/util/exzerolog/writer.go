// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exzerolog

import (
	"bytes"
	"sync"

	"github.com/rs/zerolog"
)

// LogWriter wraps a zerolog.Logger and provides an io.Writer with buffering so each written line is logged separately.
type LogWriter struct {
	log   zerolog.Logger
	level zerolog.Level
	field string
	mu    sync.Mutex
	buf   bytes.Buffer
}

func NewLogWriter(log zerolog.Logger) *LogWriter {
	zerolog.Nop()
	return &LogWriter{
		log:   log,
		level: zerolog.DebugLevel,
		field: zerolog.MessageFieldName,
	}
}

func (lw *LogWriter) WithLevel(level zerolog.Level) *LogWriter {
	return &LogWriter{
		log:   lw.log,
		level: level,
		field: lw.field,
	}
}

func (lw *LogWriter) WithField(field string) *LogWriter {
	return &LogWriter{
		log:   lw.log,
		level: lw.level,
		field: field,
	}
}

func (lw *LogWriter) writeLine(data []byte) {
	if len(data) == 0 {
		return
	}
	if data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	if lw.buf.Len() > 0 {
		lw.buf.Write(data)
		data = lw.buf.Bytes()
		lw.buf.Reset()
	}
	lw.log.WithLevel(lw.level).Bytes(lw.field, data).Send()
}

func (lw *LogWriter) Write(data []byte) (n int, err error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	newline := bytes.IndexByte(data, '\n')
	if newline == len(data)-1 {
		lw.writeLine(data)
	} else if newline < 0 {
		lw.buf.Write(data)
	} else {
		lines := bytes.Split(data, []byte{'\n'})
		for _, line := range lines[:len(lines)-1] {
			lw.writeLine(line)
		}
		lw.buf.Write(lines[len(lines)-1])
	}
	return len(data), nil
}
