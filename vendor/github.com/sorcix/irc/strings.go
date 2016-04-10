// Copyright 2014 Vic Demuzere
//
// Use of this source code is governed by the MIT license.

// +build go1.2

// Documented in strings_legacy.go

package irc

import (
	"strings"
)

func indexByte(s string, c byte) int {
	return strings.IndexByte(s, c)
}
