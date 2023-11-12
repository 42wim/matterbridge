// Copyright 2016 Russ Olsen. All Rights Reserved.
//
// This code is a Go port of the Java version created and maintained by Cognitect, therefore:
//
// Copyright 2014 Cognitect. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS-IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transit

type CMapEntry struct {
	Key   interface{}
	Value interface{}
}

func NewCMapEntry(key, value interface{}) *CMapEntry {
	return &CMapEntry{Key: key, Value: value}
}

// CMap is used to hold maps that have composite keys (i.e. #cmap values).
// Since Go arrays and maps cannot be used as map keys, a CMap is represented
// here as a simple array of key/value entrys.

type CMap struct {
	Entries []CMapEntry
}

func NewCMap() *CMap {
	return &CMap{}
}

// FindBy searches thru the map, calling mf on each key in turn
// and returns the first entry for which mf evaluates to true.

func (cm CMap) FindBy(key interface{}, mf MatchF) *CMapEntry {
	for _, entry := range cm.Entries {
		if mf(key, entry.Key) {
			return &entry
		}
	}
	return nil
}

// Find a given key in the map and return it's corresponding value.
// The search thru the keys is done with ==, so the keys in the map
// must be comparable.

func (cm CMap) Index(keyValue interface{}) interface{} {
	entry := cm.FindBy(keyValue, Equals)

	if entry == nil {
		return nil
	}
	return entry.Value
}

func (cm CMap) Put(key, value interface{}, mf MatchF) *CMap {
	entry := cm.FindBy(key, mf)

	if entry != nil {
		entry.Value = value
	} else {
		entry = NewCMapEntry(key, value)
		cm.Entries = append(cm.Entries, *entry)
	}
	return &cm
}

// Append inserts a new key/value pair w/o paying attention to
// duplicate keys.
func (cm *CMap) Append(key, value interface{}) *CMap {
	entry := NewCMapEntry(key, value)
	cm.Entries = append(cm.Entries, *entry)
	return cm
}

// Size returns the number of key/value pairs.
func (cm *CMap) Size() int {
	return len(cm.Entries)
}
