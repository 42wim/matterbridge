package orderedmap

// Element is an element of a null terminated (non circular) intrusive doubly linked list that contains the key of the correspondent element in the ordered map too.
type Element[K comparable, V any] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[K, V]

	// The key that corresponds to this element in the ordered map.
	Key K

	// The value stored with this element.
	Value V
}

// Next returns the next list element or nil.
func (e *Element[K, V]) Next() *Element[K, V] {
	return e.next
}

// Prev returns the previous list element or nil.
func (e *Element[K, V]) Prev() *Element[K, V] {
	return e.prev
}

// list represents a null terminated (non circular) intrusive doubly linked list.
// The list is immediately usable after instantiation without the need of a dedicated initialization.
type list[K comparable, V any] struct {
	root Element[K, V] // list head and tail
}

func (l *list[K, V]) IsEmpty() bool {
	return l.root.next == nil
}

// Front returns the first element of list l or nil if the list is empty.
func (l *list[K, V]) Front() *Element[K, V] {
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *list[K, V]) Back() *Element[K, V] {
	return l.root.prev
}

// Remove removes e from its list
func (l *list[K, V]) Remove(e *Element[K, V]) {
	if e.prev == nil {
		l.root.next = e.next
	} else {
		e.prev.next = e.next
	}
	if e.next == nil {
		l.root.prev = e.prev
	} else {
		e.next.prev = e.prev
	}
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *list[K, V]) PushFront(key K, value V) *Element[K, V] {
	e := &Element[K, V]{Key: key, Value: value}
	if l.root.next == nil {
		// It's the first element
		l.root.next = e
		l.root.prev = e
		return e
	}

	e.next = l.root.next
	l.root.next.prev = e
	l.root.next = e
	return e
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *list[K, V]) PushBack(key K, value V) *Element[K, V] {
	e := &Element[K, V]{Key: key, Value: value}
	if l.root.prev == nil {
		// It's the first element
		l.root.next = e
		l.root.prev = e
		return e
	}

	e.prev = l.root.prev
	l.root.prev.next = e
	l.root.prev = e
	return e
}
