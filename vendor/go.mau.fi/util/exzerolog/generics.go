// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package exzerolog

import (
	"fmt"

	"github.com/rs/zerolog"
)

func ArrayOf[T any](slice []T, fn func(arr *zerolog.Array, item T)) *zerolog.Array {
	arr := zerolog.Arr()
	for _, item := range slice {
		fn(arr, item)
	}
	return arr
}

func AddObject[T zerolog.LogObjectMarshaler](arr *zerolog.Array, obj T) {
	arr.Object(obj)
}

func AddStringer[T fmt.Stringer](arr *zerolog.Array, str T) {
	arr.Str(str.String())
}

func AddStr[T ~string](arr *zerolog.Array, str T) {
	arr.Str(string(str))
}

func ArrayOfObjs[T zerolog.LogObjectMarshaler](slice []T) *zerolog.Array {
	return ArrayOf(slice, AddObject[T])
}

func ArrayOfStringers[T fmt.Stringer](slice []T) *zerolog.Array {
	return ArrayOf(slice, AddStringer[T])
}

func ArrayOfStrs[T ~string](slice []T) *zerolog.Array {
	return ArrayOf(slice, AddStr[T])
}
