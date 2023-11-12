// Copyright 2021 Ross Light
// SPDX-License-Identifier: ISC

// +build !go1.16

package fs

import (
	"os"
)

// FS is a copy of Go 1.16's io/fs.FS interface.
type FS interface {
	Open(name string) (File, error)
}

// File is a copy of Go 1.16's io/fs.File interface.
type File interface {
	Stat() (os.FileInfo, error)
	Read([]byte) (int, error)
	Close() error
}
