// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maulogadapt

import (
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"

	"maunium.net/go/maulogger/v2"
)

type MauZeroLog struct {
	*zerolog.Logger
	orig *zerolog.Logger
	mod  string
}

func ZeroAsMau(log *zerolog.Logger) maulogger.Logger {
	return MauZeroLog{log, log, ""}
}

var _ maulogger.Logger = (*MauZeroLog)(nil)

func (m MauZeroLog) Sub(module string) maulogger.Logger {
	return m.Subm(module, map[string]interface{}{})
}

func (m MauZeroLog) Subm(module string, metadata map[string]interface{}) maulogger.Logger {
	if m.mod != "" {
		module = fmt.Sprintf("%s/%s", m.mod, module)
	}
	var orig zerolog.Logger
	if m.orig != nil {
		orig = *m.orig
	} else {
		orig = *m.Logger
	}
	if len(metadata) > 0 {
		with := m.orig.With()
		for key, value := range metadata {
			with = with.Interface(key, value)
		}
		orig = with.Logger()
	}
	log := orig.With().Str("module", module).Logger()
	return MauZeroLog{&log, &orig, module}
}

func (m MauZeroLog) WithDefaultLevel(_ maulogger.Level) maulogger.Logger {
	return m
}

func (m MauZeroLog) GetParent() maulogger.Logger {
	return nil
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func (m MauZeroLog) Writer(level maulogger.Level) io.WriteCloser {
	return nopWriteCloser{m.Logger.With().Str(zerolog.LevelFieldName, zerolog.LevelFieldMarshalFunc(mauToZeroLevel(level))).Logger()}
}

func mauToZeroLevel(level maulogger.Level) zerolog.Level {
	switch level {
	case maulogger.LevelDebug:
		return zerolog.DebugLevel
	case maulogger.LevelInfo:
		return zerolog.InfoLevel
	case maulogger.LevelWarn:
		return zerolog.WarnLevel
	case maulogger.LevelError:
		return zerolog.ErrorLevel
	case maulogger.LevelFatal:
		return zerolog.FatalLevel
	default:
		return zerolog.TraceLevel
	}
}

func (m MauZeroLog) Log(level maulogger.Level, parts ...interface{}) {
	m.Logger.WithLevel(mauToZeroLevel(level)).Msg(fmt.Sprint(parts...))
}

func (m MauZeroLog) Logln(level maulogger.Level, parts ...interface{}) {
	m.Logger.WithLevel(mauToZeroLevel(level)).Msg(strings.TrimSuffix(fmt.Sprintln(parts...), "\n"))
}

func (m MauZeroLog) Logf(level maulogger.Level, message string, args ...interface{}) {
	m.Logger.WithLevel(mauToZeroLevel(level)).Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Logfln(level maulogger.Level, message string, args ...interface{}) {
	m.Logger.WithLevel(mauToZeroLevel(level)).Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Debug(parts ...interface{}) {
	m.Logger.Debug().Msg(fmt.Sprint(parts...))
}

func (m MauZeroLog) Debugln(parts ...interface{}) {
	m.Logger.Debug().Msg(strings.TrimSuffix(fmt.Sprintln(parts...), "\n"))
}

func (m MauZeroLog) Debugf(message string, args ...interface{}) {
	m.Logger.Debug().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Debugfln(message string, args ...interface{}) {
	m.Logger.Debug().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Info(parts ...interface{}) {
	m.Logger.Info().Msg(fmt.Sprint(parts...))
}

func (m MauZeroLog) Infoln(parts ...interface{}) {
	m.Logger.Info().Msg(strings.TrimSuffix(fmt.Sprintln(parts...), "\n"))
}

func (m MauZeroLog) Infof(message string, args ...interface{}) {
	m.Logger.Info().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Infofln(message string, args ...interface{}) {
	m.Logger.Info().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Warn(parts ...interface{}) {
	m.Logger.Warn().Msg(fmt.Sprint(parts...))
}

func (m MauZeroLog) Warnln(parts ...interface{}) {
	m.Logger.Warn().Msg(strings.TrimSuffix(fmt.Sprintln(parts...), "\n"))
}

func (m MauZeroLog) Warnf(message string, args ...interface{}) {
	m.Logger.Warn().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Warnfln(message string, args ...interface{}) {
	m.Logger.Warn().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Error(parts ...interface{}) {
	m.Logger.Error().Msg(fmt.Sprint(parts...))
}

func (m MauZeroLog) Errorln(parts ...interface{}) {
	m.Logger.Error().Msg(strings.TrimSuffix(fmt.Sprintln(parts...), "\n"))
}

func (m MauZeroLog) Errorf(message string, args ...interface{}) {
	m.Logger.Error().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Errorfln(message string, args ...interface{}) {
	m.Logger.Error().Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Fatal(parts ...interface{}) {
	m.Logger.WithLevel(zerolog.FatalLevel).Msg(fmt.Sprint(parts...))
}

func (m MauZeroLog) Fatalln(parts ...interface{}) {
	m.Logger.WithLevel(zerolog.FatalLevel).Msg(strings.TrimSuffix(fmt.Sprintln(parts...), "\n"))
}

func (m MauZeroLog) Fatalf(message string, args ...interface{}) {
	m.Logger.WithLevel(zerolog.FatalLevel).Msg(fmt.Sprintf(message, args...))
}

func (m MauZeroLog) Fatalfln(message string, args ...interface{}) {
	m.Logger.WithLevel(zerolog.FatalLevel).Msg(fmt.Sprintf(message, args...))
}
