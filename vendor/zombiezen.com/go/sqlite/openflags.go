// Copyright 2021 Ross Light
// SPDX-License-Identifier: ISC

package sqlite

import (
	"fmt"
	"strings"

	lib "modernc.org/sqlite/lib"
)

// OpenFlags are flags used when opening a Conn.
//
// https://www.sqlite.org/c3ref/c_open_autoproxy.html
type OpenFlags uint

const (
	OpenReadOnly     OpenFlags = lib.SQLITE_OPEN_READONLY
	OpenReadWrite    OpenFlags = lib.SQLITE_OPEN_READWRITE
	OpenCreate       OpenFlags = lib.SQLITE_OPEN_CREATE
	OpenURI          OpenFlags = lib.SQLITE_OPEN_URI
	OpenMemory       OpenFlags = lib.SQLITE_OPEN_MEMORY
	OpenNoMutex      OpenFlags = lib.SQLITE_OPEN_NOMUTEX
	OpenFullMutex    OpenFlags = lib.SQLITE_OPEN_FULLMUTEX
	OpenSharedCache  OpenFlags = lib.SQLITE_OPEN_SHAREDCACHE
	OpenPrivateCache OpenFlags = lib.SQLITE_OPEN_PRIVATECACHE
	OpenWAL          OpenFlags = lib.SQLITE_OPEN_WAL
)

// String returns a pipe-separated list of the C constant names set in flags.
func (flags OpenFlags) String() string {
	var parts []string
	if flags&OpenReadOnly != 0 {
		parts = append(parts, "SQLITE_OPEN_READONLY")
		flags &^= OpenReadOnly
	}
	if flags&OpenReadWrite != 0 {
		parts = append(parts, "SQLITE_OPEN_READWRITE")
		flags &^= OpenReadWrite
	}
	if flags&OpenCreate != 0 {
		parts = append(parts, "SQLITE_OPEN_CREATE")
		flags &^= OpenCreate
	}
	if flags&OpenURI != 0 {
		parts = append(parts, "SQLITE_OPEN_URI")
		flags &^= OpenURI
	}
	if flags&OpenMemory != 0 {
		parts = append(parts, "SQLITE_OPEN_MEMORY")
		flags &^= OpenMemory
	}
	if flags&OpenNoMutex != 0 {
		parts = append(parts, "SQLITE_OPEN_NOMUTEX")
		flags &^= OpenNoMutex
	}
	if flags&OpenFullMutex != 0 {
		parts = append(parts, "SQLITE_OPEN_FULLMUTEX")
		flags &^= OpenFullMutex
	}
	if flags&OpenSharedCache != 0 {
		parts = append(parts, "SQLITE_OPEN_SHAREDCACHE")
		flags &^= OpenSharedCache
	}
	if flags&OpenPrivateCache != 0 {
		parts = append(parts, "SQLITE_OPEN_PRIVATECACHE")
		flags &^= OpenPrivateCache
	}
	if flags&OpenWAL != 0 {
		parts = append(parts, "SQLITE_OPEN_WAL")
		flags &^= OpenWAL
	}
	if flags != 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%#x", uint(flags)))
	}
	return strings.Join(parts, "|")
}
