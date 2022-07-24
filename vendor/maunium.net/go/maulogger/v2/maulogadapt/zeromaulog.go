// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maulogadapt

import (
	"bytes"

	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"maunium.net/go/maulogger/v2"
)

// ZeroMauLog is a simple wrapper for a maulogger that can be set as the output writer for zerolog.
type ZeroMauLog struct {
	maulogger.Logger
}

func MauAsZero(log maulogger.Logger) *zerolog.Logger {
	zero := zerolog.New(&ZeroMauLog{log})
	return &zero
}

var _ zerolog.LevelWriter = (*ZeroMauLog)(nil)

func (z *ZeroMauLog) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (z *ZeroMauLog) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	var mauLevel maulogger.Level
	switch level {
	case zerolog.DebugLevel:
		mauLevel = maulogger.LevelDebug
	case zerolog.InfoLevel, zerolog.NoLevel:
		mauLevel = maulogger.LevelInfo
	case zerolog.WarnLevel:
		mauLevel = maulogger.LevelWarn
	case zerolog.ErrorLevel:
		mauLevel = maulogger.LevelError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		mauLevel = maulogger.LevelFatal
	case zerolog.Disabled, zerolog.TraceLevel:
		fallthrough
	default:
		return 0, nil
	}
	p = bytes.TrimSuffix(p, []byte{'\n'})
	msg := gjson.GetBytes(p, zerolog.MessageFieldName).Str

	p, err = sjson.DeleteBytes(p, zerolog.MessageFieldName)
	if err != nil {
		return
	}
	p, err = sjson.DeleteBytes(p, zerolog.LevelFieldName)
	if err != nil {
		return
	}
	p, err = sjson.DeleteBytes(p, zerolog.TimestampFieldName)
	if err != nil {
		return
	}
	if len(p) > 2 {
		msg += " " + string(p)
	}
	z.Log(mauLevel, msg)
	return len(p), nil
}
