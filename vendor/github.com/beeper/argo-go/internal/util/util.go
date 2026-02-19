// Package util provides miscellaneous utility functions used across the argo-go library.
package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/vektah/gqlparser/v2/ast"
)

// GroupBy groups elements of a slice into a map based on a key extracted from each element.
func GroupBy[T any, K comparable](array []T, extract func(T) K) map[K][]T {
	grouped := make(map[K][]T)
	for _, element := range array {
		key := extract(element)
		grouped[key] = append(grouped[key], element)
	}
	return grouped
}

// AddPathIndex appends an integer index to an ast.Path.
func AddPathIndex(p ast.Path, i int) ast.Path {
	path := make(ast.Path, len(p), len(p)+1)
	copy(path, p)
	path = append(path, ast.PathIndex(i))
	return path
}

// AddPathName appends a string name to an ast.Path.
func AddPathName(p ast.Path, s string) ast.Path {
	path := make(ast.Path, len(p), len(p)+1)
	copy(path, p)
	path = append(path, ast.PathName(s))
	return path
}

func FormatPath(p ast.Path) string {
	if p == nil {
		return "<nil>"
	}
	var parts []string
	for _, el := range p {
		switch v := el.(type) {
		case ast.PathName:
			parts = append(parts, string(v))
		case ast.PathIndex:
			parts = append(parts, strconv.Itoa(int(v)))
		default:
			parts = append(parts, fmt.Sprintf("%v", v))
		}
	}
	return strings.Join(parts, ".")
}

// NewOrderedMapJSON creates a new OrderedMapJSON wrapper.
func NewOrderedMapJSON[K comparable, V any](om *orderedmap.OrderedMap[K, V]) *OrderedMapJSON[K, V] {
	return &OrderedMapJSON[K, V]{om}
}

// OrderedMapJSON is a wrapper around orderedmap.OrderedMap to allow custom JSON marshalling.
// It needs to be generic to match the orderedmap.OrderedMap it wraps.
type OrderedMapJSON[K comparable, V any] struct {
	*orderedmap.OrderedMap[K, V]
}

// MarshalJSON implements the json.Marshaler interface.
// This method will be called by json.Marshal and json.MarshalIndent.
func (omj OrderedMapJSON[K, V]) MarshalJSON() ([]byte, error) {
	if omj.OrderedMap == nil {
		return []byte("null"), nil
	}

	var buf bytes.Buffer
	buf.WriteString("{")

	first := true
	for key := range omj.Keys() { // Keys() returns keys in order
		value, _ := omj.Get(key)

		if !first {
			buf.WriteString(",")
		}
		first = false

		// Marshal key
		keyBytes, keyErr := json.Marshal(key)
		if keyErr != nil {
			return nil, fmt.Errorf("failed to marshal ordered map key (%v): %w", key, keyErr)
		}
		buf.Write(keyBytes)

		buf.WriteString(":")

		// Marshal value
		var valBytes []byte
		var valErr error

		// Convert value (type V) to interface{} to perform dynamic type assertion
		valAsInterface := interface{}(value)

		// Check if the value is specifically an *orderedmap.OrderedMap[string, any]
		if subMap, ok := valAsInterface.(*orderedmap.OrderedMap[string, any]); ok {
			// For this specific project structure where maps are *orderedmap.OrderedMap[string, any],
			// we can create a new wrapper for the sub-map.
			wrappedSubMap := NewOrderedMapJSON[string, any](subMap)
			valBytes, valErr = wrappedSubMap.MarshalJSON() // Recursive call to our MarshalJSON
		} else {
			// Standard marshalling for other types
			valBytes, valErr = json.Marshal(valAsInterface)
		}

		if valErr != nil {
			return nil, fmt.Errorf("failed to marshal ordered map value for key (%v): %w", key, valErr)
		}
		buf.Write(valBytes)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

// MustMarshalJSON calls MarshalJSON and panics if an error occurs.
func (omj OrderedMapJSON[K, V]) MustMarshalJSON() []byte {
	data, err := omj.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("failed to marshal OrderedMapJSON to JSON: %w", err))
	}
	return data
}
