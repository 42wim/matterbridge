// Copyright (c) 2023 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package util

import "sync"

// SyncMap is a simple map with a built-in mutex.
type SyncMap[Key comparable, Value any] struct {
	data map[Key]Value
	lock sync.RWMutex
}

func NewSyncMap[Key comparable, Value any]() *SyncMap[Key, Value] {
	return &SyncMap[Key, Value]{
		data: make(map[Key]Value),
	}
}

// Set stores a value in the map.
func (sm *SyncMap[Key, Value]) Set(key Key, value Value) {
	sm.Swap(key, value)
}

// Swap sets a value in the map and returns the old value.
//
// The boolean return parameter is true if the value already existed, false if not.
func (sm *SyncMap[Key, Value]) Swap(key Key, value Value) (oldValue Value, wasReplaced bool) {
	sm.lock.Lock()
	oldValue, wasReplaced = sm.data[key]
	sm.data[key] = value
	sm.lock.Unlock()
	return
}

// Delete removes a key from the map.
func (sm *SyncMap[Key, Value]) Delete(key Key) {
	sm.Pop(key)
}

// Pop removes a key from the map and returns the old value.
//
// The boolean return parameter is the same as with normal Go map access (true if the key exists, false if not).
func (sm *SyncMap[Key, Value]) Pop(key Key) (value Value, ok bool) {
	sm.lock.Lock()
	value, ok = sm.data[key]
	delete(sm.data, key)
	sm.lock.Unlock()
	return
}

// Get gets a value in the map.
//
// The boolean return parameter is the same as with normal Go map access (true if the key exists, false if not).
func (sm *SyncMap[Key, Value]) Get(key Key) (value Value, ok bool) {
	sm.lock.RLock()
	value, ok = sm.data[key]
	sm.lock.RUnlock()
	return
}

// GetOrSet gets a value in the map if the key already exists, otherwise inserts the given value and returns it.
//
// The boolean return parameter is true if the key already exists, and false if the given value was inserted.
func (sm *SyncMap[Key, Value]) GetOrSet(key Key, value Value) (actual Value, wasGet bool) {
	sm.lock.Lock()
	defer sm.lock.Unlock()
	actual, wasGet = sm.data[key]
	if wasGet {
		return
	}
	sm.data[key] = value
	actual = value
	return
}

// Clone returns a copy of the map.
func (sm *SyncMap[Key, Value]) Clone() *SyncMap[Key, Value] {
	return &SyncMap[Key, Value]{data: sm.CopyData()}
}

// CopyData returns a copy of the data in the map as a normal (non-atomic) map.
func (sm *SyncMap[Key, Value]) CopyData() map[Key]Value {
	sm.lock.RLock()
	copied := make(map[Key]Value, len(sm.data))
	for key, value := range sm.data {
		copied[key] = value
	}
	sm.lock.RUnlock()
	return copied
}
