// Copyright (C) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build linux

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
	return unix.Fallocate(int(file.Fd()), 0, 0, int64(size))
}
