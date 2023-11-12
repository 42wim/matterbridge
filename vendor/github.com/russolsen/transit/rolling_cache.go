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

import (
	"fmt"
)

const cacheCodeDigits = 44
const baseCharIndex = 48
const firstOrd = 48
const cacheSize = (cacheCodeDigits * cacheCodeDigits)
const minSizeCacheable = 4

type StringMap map[string]string

type RollingCache struct {
	keyToValue StringMap
	valueToKey StringMap
}

func NewRollingCache() *RollingCache {
	return &RollingCache{keyToValue: make(StringMap), valueToKey: make(StringMap)}
}

func (rc *RollingCache) String() string {
	return fmt.Sprintf("Cache: %v", rc.keyToValue)
}

func (rc *RollingCache) HasKey(name string) bool {
	_, present := rc.keyToValue[name]
	return present
}

func (rc *RollingCache) Read(name string) string {
	return rc.keyToValue[name]
}

// Enter the name into the cache if it passes the cacheable critieria.
// Returns either the name or the value that was previously cached for
// the name.
func (rc *RollingCache) Write(name string) string {
	existing_key, present := rc.valueToKey[name]

	if present {
		return existing_key
	}

	if rc.isCacheFull() {
		rc.Clear()
	}

	var key = rc.encodeKey(len(rc.keyToValue))
	rc.keyToValue[key] = name
	rc.valueToKey[name] = key

	return name
}

// IsCacheable returns true if the string is long enough to be cached
// and either asKey is true or the string represents a symbol, keyword
// or tag.
func (rc *RollingCache) IsCacheable(s string, asKey bool) bool {
	if len(s) < minSizeCacheable {
		return false
	} else if asKey {
		return true
	} else {
		var firstTwo = s[0:2]
		//return firstTwo == "~#" || firstTwo == "~$" || firstTwo == "~:"
		return firstTwo == startTag || firstTwo == startKW || firstTwo == startSym
	}
}

// IsCacheKey returns true if the string looks like a cache key.
func (rc *RollingCache) IsCacheKey(name string) bool {
	if len(name) == 0 {
		return false
	} else if (name[0:1] == sub) && (name != mapAsArray) {
		return true
	} else {
		return false
	}
}

func (rc *RollingCache) encodeKey(index int) string {
	var hi = index / cacheCodeDigits
	var lo = index % cacheCodeDigits
	if hi == 0 {
		return sub + string(lo+baseCharIndex)
	} else {
		return sub + string(hi+baseCharIndex) + string(lo+baseCharIndex)
	}
}

func (rc *RollingCache) codeToIndex(s string) int {
	var sz = len(s)
	if sz == 2 {
		return int(s[1]) - baseCharIndex
	} else {
		return ((int(s[1]) - baseCharIndex) * cacheCodeDigits) + (int(s[2]) - baseCharIndex)
	}
}

func (rc *RollingCache) isCacheFull() bool {
	return len(rc.keyToValue) >= cacheSize
}

func (rc *RollingCache) Clear() {
	rc.valueToKey = make(StringMap)
	rc.keyToValue = make(StringMap)
}
