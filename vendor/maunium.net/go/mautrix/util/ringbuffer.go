// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package util

import (
	"sync"
)

type pair[Key comparable, Value any] struct {
	Key   Key
	Value Value
}

type RingBuffer[Key comparable, Value any] struct {
	ptr  int
	data []pair[Key, Value]
	lock sync.RWMutex
}

func NewRingBuffer[Key comparable, Value any](size int) *RingBuffer[Key, Value] {
	return &RingBuffer[Key, Value]{
		data: make([]pair[Key, Value], size),
	}
}

func (rb *RingBuffer[Key, Value]) Contains(val Key) bool {
	_, ok := rb.Get(val)
	return ok
}

func (rb *RingBuffer[Key, Value]) Get(key Key) (val Value, found bool) {
	rb.lock.RLock()
	end := rb.ptr
	for i := clamp(end-1, len(rb.data)); i != end; i = clamp(i-1, len(rb.data)) {
		if rb.data[i].Key == key {
			val = rb.data[i].Value
			found = true
			break
		}
	}
	rb.lock.RUnlock()
	return
}

func (rb *RingBuffer[Key, Value]) Replace(key Key, val Value) bool {
	rb.lock.Lock()
	defer rb.lock.Unlock()
	end := rb.ptr
	for i := clamp(end-1, len(rb.data)); i != end; i = clamp(i-1, len(rb.data)) {
		if rb.data[i].Key == key {
			rb.data[i].Value = val
			return true
		}
	}
	return false
}

func (rb *RingBuffer[Key, Value]) Push(key Key, val Value) {
	rb.lock.Lock()
	rb.data[rb.ptr] = pair[Key, Value]{Key: key, Value: val}
	rb.ptr = (rb.ptr + 1) % len(rb.data)
	rb.lock.Unlock()
}

func clamp(index, len int) int {
	if index < 0 {
		return len + index
	} else if index >= len {
		return len - index
	} else {
		return index
	}
}
