// Copyright (c) 2024 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exzerolog

import (
	"time"

	"github.com/rs/zerolog"
	deflog "github.com/rs/zerolog/log"
)

// SetupDefaults updates zerolog globals with sensible defaults.
//
// * [zerolog.TimeFieldFormat] is set to time.RFC3339Nano instead of time.RFC3339
// * [zerolog.CallerMarshalFunc] is set to [CallerWithFunctionName]
// * [zerolog.DefaultContextLogger] is set to the given logger with default_context_log=true and caller info enabled
// * The global default [log.Logger] is set to the given logger with global_log=true and caller info enabled
// * [zerolog.LevelColors] are updated to swap trace and debug level colors
func SetupDefaults(log *zerolog.Logger) {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.CallerMarshalFunc = CallerWithFunctionName
	defaultCtxLog := log.With().Bool("default_context_log", true).Caller().Logger()
	zerolog.DefaultContextLogger = &defaultCtxLog
	deflog.Logger = log.With().Bool("global_log", true).Caller().Logger()
	// Swap trace and debug level colors so trace pops out the least
	zerolog.LevelColors[zerolog.TraceLevel] = 0
	zerolog.LevelColors[zerolog.DebugLevel] = 34 // blue
}
