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
	"reflect"
)

// GetMapElement pulls a value out of a map (represented by
// the m value). This function differs from the reflect.Value
// MapIndex in that the type and kind of the returned value
// reflect the type and kind of the element, not what was
// declared in the map.
func GetMapElement(m reflect.Value, key reflect.Value) reflect.Value {
	element := m.MapIndex(key)
	return reflect.ValueOf(element.Interface())
}

// GetElement pulls a value out of a array (represented by
// the array value). This function differs from the reflect.Value
// Index in that the type and kind of the returned value
// reflect the type and kind of the element, not what was
// declared in the array.
func GetElement(array reflect.Value, i int) reflect.Value {
	element := array.Index(i)
	return reflect.ValueOf(element.Interface())
}

// KeyValues returns the Values for the keys in a map.
func KeyValues(m reflect.Value) []reflect.Value {
	keys := m.MapKeys()
	result := make([]reflect.Value, len(keys))

	for i, v := range keys {
		result[i] = reflect.ValueOf(v.Interface())
	}
	return result
}

func IsGenericArray(x interface{}) bool {
	switch x.(type) {
	case []interface{}:
		return true
	default:
		return false
	}
}
