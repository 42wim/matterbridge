// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exzerolog

import (
	"fmt"
	"runtime"
	"strings"
)

// CallerWithFunctionName is an implementation for zerolog.CallerMarshalFunc that includes the caller function name
// in addition to the file and line number.
//
// Use as
//
//	zerolog.CallerMarshalFunc = exzerolog.CallerWithFunctionName
func CallerWithFunctionName(pc uintptr, file string, line int) string {
	files := strings.Split(file, "/")
	file = files[len(files)-1]
	name := runtime.FuncForPC(pc).Name()
	fns := strings.Split(name, ".")
	name = fns[len(fns)-1]
	return fmt.Sprintf("%s:%d:%s()", file, line, name)
}
