// Copyright (C) 2024 Sumner Evans
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

//go:build !linux && !android && !darwin

package fallocate

import "os"

var ErrOutOfSpace error = nil

func Fallocate(file *os.File, size int) error {
	return nil
}
