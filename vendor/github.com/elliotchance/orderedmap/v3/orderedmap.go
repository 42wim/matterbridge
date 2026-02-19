package orderedmap

import "iter"

type OrderedMap[K comparable, V any] struct {
	kv map[K]*Element[K, V]
	ll list[K, V]
}

func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		kv: make(map[K]*Element[K, V]),
	}
}

// NewOrderedMapWithCapacity creates a map with enough pre-allocated space to
// hold the specified number of elements.
func NewOrderedMapWithCapacity[K comparable, V any](capacity int) *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		kv: make(map[K]*Element[K, V], capacity),
	}
}

func NewOrderedMapWithElements[K comparable, V any](els ...*Element[K, V]) *OrderedMap[K, V] {
	om := NewOrderedMapWithCapacity[K, V](len(els))
	for _, el := range els {
		om.Set(el.Key, el.Value)
	}
	return om
}

// Get returns the value for a key. If the key does not exist, the second return
// parameter will be false and the value will be nil.
func (m *OrderedMap[K, V]) Get(key K) (value V, ok bool) {
	v, ok := m.kv[key]
	if ok {
		value = v.Value
	}

	return
}

// Set will set (or replace) a value for a key. If the key was new, then true
// will be returned. The returned value will be false if the value was replaced
// (even if the value was the same).
func (m *OrderedMap[K, V]) Set(key K, value V) bool {
	_, alreadyExist := m.kv[key]
	if alreadyExist {
		m.kv[key].Value = value
		return false
	}

	element := m.ll.PushBack(key, value)
	m.kv[key] = element
	return true
}

// ReplaceKey replaces an existing key with a new key while preserving order of
// the value. This function will return true if the operation was successful, or
// false if 'originalKey' is not found OR 'newKey' already exists (which would be an overwrite).
func (m *OrderedMap[K, V]) ReplaceKey(originalKey, newKey K) bool {
	element, originalExists := m.kv[originalKey]
	_, newKeyExists := m.kv[newKey]
	if originalExists && !newKeyExists {
		delete(m.kv, originalKey)
		m.kv[newKey] = element
		element.Key = newKey
		return true
	}
	return false
}

// GetOrDefault returns the value for a key. If the key does not exist, returns
// the default value instead.
func (m *OrderedMap[K, V]) GetOrDefault(key K, defaultValue V) V {
	if value, ok := m.kv[key]; ok {
		return value.Value
	}

	return defaultValue
}

// GetElement returns the element for a key. If the key does not exist, the
// pointer will be nil.
func (m *OrderedMap[K, V]) GetElement(key K) *Element[K, V] {
	element, ok := m.kv[key]
	if ok {
		return element
	}

	return nil
}

// Len returns the number of elements in the map.
func (m *OrderedMap[K, V]) Len() int {
	return len(m.kv)
}

// AllFromFront returns an iterator that yields all elements in the map starting
// at the front (oldest Set element).
func (m *OrderedMap[K, V]) AllFromFront() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for el := m.Front(); el != nil; el = el.Next() {
			if !yield(el.Key, el.Value) {
				return
			}
		}
	}
}

// AllFromBack returns an iterator that yields all elements in the map starting
// at the back (most recent Set element).
func (m *OrderedMap[K, V]) AllFromBack() iter.Seq2[K, V] {
	return func(yield func(key K, value V) bool) {
		for el := m.Back(); el != nil; el = el.Prev() {
			if !yield(el.Key, el.Value) {
				return
			}
		}
	}
}

// Keys returns an iterator that yields all the keys in the map starting at the
// front (oldest Set element). To create a slice containing all the map keys,
// use the slices.Collect function on the returned iterator.
func (m *OrderedMap[K, V]) Keys() iter.Seq[K] {
	return func(yield func(key K) bool) {
		for el := m.Front(); el != nil; el = el.Next() {
			if !yield(el.Key) {
				return
			}
		}
	}
}

// Values returns an iterator that yields all the values in the map starting at
// the front (oldest Set element). To create a slice containing all the map
// values, use the slices.Collect function on the returned iterator.
func (m *OrderedMap[K, V]) Values() iter.Seq[V] {
	return func(yield func(value V) bool) {
		for el := m.Front(); el != nil; el = el.Next() {
			if !yield(el.Value) {
				return
			}
		}
	}
}

// Delete will remove a key from the map. It will return true if the key was
// removed (the key did exist).
func (m *OrderedMap[K, V]) Delete(key K) (didDelete bool) {
	element, ok := m.kv[key]
	if ok {
		m.ll.Remove(element)
		delete(m.kv, key)
	}

	return ok
}

// Front will return the element that is the first (oldest Set element). If
// there are no elements this will return nil.
func (m *OrderedMap[K, V]) Front() *Element[K, V] {
	return m.ll.Front()
}

// Back will return the element that is the last (most recent Set element). If
// there are no elements this will return nil.
func (m *OrderedMap[K, V]) Back() *Element[K, V] {
	return m.ll.Back()
}

// Copy returns a new OrderedMap with the same elements.
// Using Copy while there are concurrent writes may mangle the result.
func (m *OrderedMap[K, V]) Copy() *OrderedMap[K, V] {
	m2 := NewOrderedMapWithCapacity[K, V](m.Len())
	for el := m.Front(); el != nil; el = el.Next() {
		m2.Set(el.Key, el.Value)
	}
	return m2
}

// Has checks if a key exists in the map.
func (m *OrderedMap[K, V]) Has(key K) bool {
	_, exists := m.kv[key]
	return exists
}
