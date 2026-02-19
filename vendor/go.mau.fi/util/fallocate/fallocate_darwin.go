// Copyright (C) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build darwin

package fallocate

import (
	"os"

	"golang.org/x/sys/unix"
)

var ErrOutOfSpace error = unix.ENOSPC

func Fallocate(file *os.File, size int) error {
	if size <= 0 {
		return nil
	}
	return unix.FcntlFstore(uintptr(file.Fd()), unix.F_PREALLOCATE, &unix.Fstore_t{
		Flags:   unix.F_ALLOCATEALL,
		Posmode: unix.F_PEOFPOSMODE,
		Offset:  0,
		Length:  int64(size),
	})
}
