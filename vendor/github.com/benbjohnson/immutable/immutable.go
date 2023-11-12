// Package immutable provides immutable collection types.
//
// Introduction
//
// Immutable collections provide an efficient, safe way to share collections
// of data while minimizing locks. The collections in this package provide
// List, Map, and SortedMap implementations. These act similarly to slices
// and maps, respectively, except that altering a collection returns a new
// copy of the collection with that change.
//
// Because collections are unable to change, they are safe for multiple
// goroutines to read from at the same time without a mutex. However, these
// types of collections come with increased CPU & memory usage as compared
// with Go's built-in collection types so please evaluate for your specific
// use.
//
// Collection Types
//
// The List type provides an API similar to Go slices. They allow appending,
// prepending, and updating of elements. Elements can also be fetched by index
// or iterated over using a ListIterator.
//
// The Map & SortedMap types provide an API similar to Go maps. They allow
// values to be assigned to unique keys and allow for the deletion of keys.
// Values can be fetched by key and key/value pairs can be iterated over using
// the appropriate iterator type. Both map types provide the same API. The
// SortedMap, however, provides iteration over sorted keys while the Map
// provides iteration over unsorted keys. Maps improved performance and memory
// usage as compared to SortedMaps.
//
// Hashing and Sorting
//
// Map types require the use of a Hasher implementation to calculate hashes for
// their keys and check for key equality. SortedMaps require the use of a
// Comparer implementation to sort keys in the map.
//
// These collection types automatically provide built-in hasher and comparers
// for int, string, and byte slice keys. If you are using one of these key types
// then simply pass a nil into the constructor. Otherwise you will need to
// implement a custom Hasher or Comparer type. Please see the provided
// implementations for reference.
package immutable

import (
	"bytes"
	"fmt"
	"math/bits"
	"reflect"
	"sort"
	"strings"
)

// List is a dense, ordered, indexed collections. They are analogous to slices
// in Go. They can be updated by appending to the end of the list, prepending
// values to the beginning of the list, or updating existing indexes in the
// list.
type List struct {
	root   listNode // root node
	origin int      // offset to zero index element
	size   int      // total number of elements in use
}

// NewList returns a new empty instance of List.
func NewList() *List {
	return &List{
		root: &listLeafNode{},
	}
}

// clone returns a copy of the list.
func (l *List) clone() *List {
	other := *l
	return &other
}

// Len returns the number of elements in the list.
func (l *List) Len() int {
	return l.size
}

// cap returns the total number of possible elements for the current depth.
func (l *List) cap() int {
	return 1 << (l.root.depth() * listNodeBits)
}

// Get returns the value at the given index. Similar to slices, this method will
// panic if index is below zero or is greater than or equal to the list size.
func (l *List) Get(index int) interface{} {
	if index < 0 || index >= l.size {
		panic(fmt.Sprintf("immutable.List.Get: index %d out of bounds", index))
	}
	return l.root.get(l.origin + index)
}

// Set returns a new list with value set at index. Similar to slices, this
// method will panic if index is below zero or if the index is greater than
// or equal to the list size.
func (l *List) Set(index int, value interface{}) *List {
	return l.set(index, value, false)
}

func (l *List) set(index int, value interface{}, mutable bool) *List {
	if index < 0 || index >= l.size {
		panic(fmt.Sprintf("immutable.List.Set: index %d out of bounds", index))
	}
	other := l
	if !mutable {
		other = l.clone()
	}
	other.root = other.root.set(l.origin+index, value, mutable)
	return other
}

// Append returns a new list with value added to the end of the list.
func (l *List) Append(value interface{}) *List {
	return l.append(value, false)
}

func (l *List) append(value interface{}, mutable bool) *List {
	other := l
	if !mutable {
		other = l.clone()
	}

	// Expand list to the right if no slots remain.
	if other.size+other.origin >= l.cap() {
		newRoot := &listBranchNode{d: other.root.depth() + 1}
		newRoot.children[0] = other.root
		other.root = newRoot
	}

	// Increase size and set the last element to the new value.
	other.size++
	other.root = other.root.set(other.origin+other.size-1, value, mutable)
	return other
}

// Prepend returns a new list with value added to the beginning of the list.
func (l *List) Prepend(value interface{}) *List {
	return l.prepend(value, false)
}

func (l *List) prepend(value interface{}, mutable bool) *List {
	other := l
	if !mutable {
		other = l.clone()
	}

	// Expand list to the left if no slots remain.
	if other.origin == 0 {
		newRoot := &listBranchNode{d: other.root.depth() + 1}
		newRoot.children[listNodeSize-1] = other.root
		other.root = newRoot
		other.origin += (listNodeSize - 1) << (other.root.depth() * listNodeBits)
	}

	// Increase size and move origin back. Update first element to value.
	other.size++
	other.origin--
	other.root = other.root.set(other.origin, value, mutable)
	return other
}

// Slice returns a new list of elements between start index and end index.
// Similar to slices, this method will panic if start or end are below zero or
// greater than the list size. A panic will also occur if start is greater than
// end.
//
// Unlike Go slices, references to inaccessible elements will be automatically
// removed so they can be garbage collected.
func (l *List) Slice(start, end int) *List {
	return l.slice(start, end, false)
}

func (l *List) slice(start, end int, mutable bool) *List {
	// Panics similar to Go slices.
	if start < 0 || start > l.size {
		panic(fmt.Sprintf("immutable.List.Slice: start index %d out of bounds", start))
	} else if end < 0 || end > l.size {
		panic(fmt.Sprintf("immutable.List.Slice: end index %d out of bounds", end))
	} else if start > end {
		panic(fmt.Sprintf("immutable.List.Slice: invalid slice index: [%d:%d]", start, end))
	}

	// Return the same list if the start and end are the entire range.
	if start == 0 && end == l.size {
		return l
	}

	// Create copy, if immutable.
	other := l
	if !mutable {
		other = l.clone()
	}

	// Update origin/size.
	other.origin = l.origin + start
	other.size = end - start

	// Contract tree while the start & end are in the same child node.
	for other.root.depth() > 1 {
		i := (other.origin >> (other.root.depth() * listNodeBits)) & listNodeMask
		j := ((other.origin + other.size - 1) >> (other.root.depth() * listNodeBits)) & listNodeMask
		if i != j {
			break // branch contains at least two nodes, exit
		}

		// Replace the current root with the single child & update origin offset.
		other.origin -= i << (other.root.depth() * listNodeBits)
		other.root = other.root.(*listBranchNode).children[i]
	}

	// Ensure all references are removed before start & after end.
	other.root = other.root.deleteBefore(other.origin, mutable)
	other.root = other.root.deleteAfter(other.origin+other.size-1, mutable)

	return other
}

// Iterator returns a new iterator for this list positioned at the first index.
func (l *List) Iterator() *ListIterator {
	itr := &ListIterator{list: l}
	itr.First()
	return itr
}

// ListBuilder represents an efficient builder for creating new Lists.
type ListBuilder struct {
	list *List // current state
}

// NewListBuilder returns a new instance of ListBuilder.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: NewList()}
}

// List returns the current copy of the list.
// The builder should not be used again after the list after this call.
func (b *ListBuilder) List() *List {
	assert(b.list != nil, "immutable.ListBuilder.List(): duplicate call to fetch list")
	list := b.list
	b.list = nil
	return list
}

// Len returns the number of elements in the underlying list.
func (b *ListBuilder) Len() int {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	return b.list.Len()
}

// Get returns the value at the given index. Similar to slices, this method will
// panic if index is below zero or is greater than or equal to the list size.
func (b *ListBuilder) Get(index int) interface{} {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	return b.list.Get(index)
}

// Set updates the value at the given index. Similar to slices, this method will
// panic if index is below zero or if the index is greater than or equal to the
// list size.
func (b *ListBuilder) Set(index int, value interface{}) {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	b.list = b.list.set(index, value, true)
}

// Append adds value to the end of the list.
func (b *ListBuilder) Append(value interface{}) {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	b.list = b.list.append(value, true)
}

// Prepend adds value to the beginning of the list.
func (b *ListBuilder) Prepend(value interface{}) {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	b.list = b.list.prepend(value, true)
}

// Slice updates the list with a sublist of elements between start and end index.
// See List.Slice() for more details.
func (b *ListBuilder) Slice(start, end int) {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	b.list = b.list.slice(start, end, true)
}

// Iterator returns a new iterator for the underlying list.
func (b *ListBuilder) Iterator() *ListIterator {
	assert(b.list != nil, "immutable.ListBuilder: builder invalid after List() invocation")
	return b.list.Iterator()
}

// Constants for bit shifts used for levels in the List trie.
const (
	listNodeBits = 5
	listNodeSize = 1 << listNodeBits
	listNodeMask = listNodeSize - 1
)

// listNode represents either a branch or leaf node in a List.
type listNode interface {
	depth() uint
	get(index int) interface{}
	set(index int, v interface{}, mutable bool) listNode

	containsBefore(index int) bool
	containsAfter(index int) bool

	deleteBefore(index int, mutable bool) listNode
	deleteAfter(index int, mutable bool) listNode
}

// newListNode returns a leaf node for depth zero, otherwise returns a branch node.
func newListNode(depth uint) listNode {
	if depth == 0 {
		return &listLeafNode{}
	}
	return &listBranchNode{d: depth}
}

// listBranchNode represents a branch of a List tree at a given depth.
type listBranchNode struct {
	d        uint // depth
	children [listNodeSize]listNode
}

// depth returns the depth of this branch node from the leaf.
func (n *listBranchNode) depth() uint { return n.d }

// get returns the child node at the segment of the index for this depth.
func (n *listBranchNode) get(index int) interface{} {
	idx := (index >> (n.d * listNodeBits)) & listNodeMask
	return n.children[idx].get(index)
}

// set recursively updates the value at index for each lower depth from the node.
func (n *listBranchNode) set(index int, v interface{}, mutable bool) listNode {
	idx := (index >> (n.d * listNodeBits)) & listNodeMask

	// Find child for the given value in the branch. Create new if it doesn't exist.
	child := n.children[idx]
	if child == nil {
		child = newListNode(n.depth() - 1)
	}

	// Return a copy of this branch with the new child.
	var other *listBranchNode
	if mutable {
		other = n
	} else {
		tmp := *n
		other = &tmp
	}
	other.children[idx] = child.set(index, v, mutable)
	return other
}

// containsBefore returns true if non-nil values exists between [0,index).
func (n *listBranchNode) containsBefore(index int) bool {
	idx := (index >> (n.d * listNodeBits)) & listNodeMask

	// Quickly check if any direct children exist before this segment of the index.
	for i := 0; i < idx; i++ {
		if n.children[i] != nil {
			return true
		}
	}

	// Recursively check for children directly at the given index at this segment.
	if n.children[idx] != nil && n.children[idx].containsBefore(index) {
		return true
	}
	return false
}

// containsAfter returns true if non-nil values exists between (index,listNodeSize).
func (n *listBranchNode) containsAfter(index int) bool {
	idx := (index >> (n.d * listNodeBits)) & listNodeMask

	// Quickly check if any direct children exist after this segment of the index.
	for i := idx + 1; i < len(n.children); i++ {
		if n.children[i] != nil {
			return true
		}
	}

	// Recursively check for children directly at the given index at this segment.
	if n.children[idx] != nil && n.children[idx].containsAfter(index) {
		return true
	}
	return false
}

// deleteBefore returns a new node with all elements before index removed.
func (n *listBranchNode) deleteBefore(index int, mutable bool) listNode {
	// Ignore if no nodes exist before the given index.
	if !n.containsBefore(index) {
		return n
	}

	// Return a copy with any nodes prior to the index removed.
	idx := (index >> (n.d * listNodeBits)) & listNodeMask

	var other *listBranchNode
	if mutable {
		other = n
		for i := 0; i < idx; i++ {
			n.children[i] = nil
		}
	} else {
		other = &listBranchNode{d: n.d}
		copy(other.children[idx:][:], n.children[idx:][:])
	}

	if other.children[idx] != nil {
		other.children[idx] = other.children[idx].deleteBefore(index, mutable)
	}
	return other
}

// deleteBefore returns a new node with all elements before index removed.
func (n *listBranchNode) deleteAfter(index int, mutable bool) listNode {
	// Ignore if no nodes exist after the given index.
	if !n.containsAfter(index) {
		return n
	}

	// Return a copy with any nodes after the index removed.
	idx := (index >> (n.d * listNodeBits)) & listNodeMask

	var other *listBranchNode
	if mutable {
		other = n
		for i := idx + 1; i < len(n.children); i++ {
			n.children[i] = nil
		}
	} else {
		other = &listBranchNode{d: n.d}
		copy(other.children[:idx+1], n.children[:idx+1])
	}

	if other.children[idx] != nil {
		other.children[idx] = other.children[idx].deleteAfter(index, mutable)
	}
	return other
}

// listLeafNode represents a leaf node in a List.
type listLeafNode struct {
	children [listNodeSize]interface{}
}

// depth always returns 0 for leaf nodes.
func (n *listLeafNode) depth() uint { return 0 }

// get returns the value at the given index.
func (n *listLeafNode) get(index int) interface{} {
	return n.children[index&listNodeMask]
}

// set returns a copy of the node with the value at the index updated to v.
func (n *listLeafNode) set(index int, v interface{}, mutable bool) listNode {
	idx := index & listNodeMask
	var other *listLeafNode
	if mutable {
		other = n
	} else {
		tmp := *n
		other = &tmp
	}
	other.children[idx] = v
	return other
}

// containsBefore returns true if non-nil values exists between [0,index).
func (n *listLeafNode) containsBefore(index int) bool {
	idx := index & listNodeMask
	for i := 0; i < idx; i++ {
		if n.children[i] != nil {
			return true
		}
	}
	return false
}

// containsAfter returns true if non-nil values exists between (index,listNodeSize).
func (n *listLeafNode) containsAfter(index int) bool {
	idx := index & listNodeMask
	for i := idx + 1; i < len(n.children); i++ {
		if n.children[i] != nil {
			return true
		}
	}
	return false
}

// deleteBefore returns a new node with all elements before index removed.
func (n *listLeafNode) deleteBefore(index int, mutable bool) listNode {
	if !n.containsBefore(index) {
		return n
	}

	idx := index & listNodeMask
	var other *listLeafNode
	if mutable {
		other = n
		for i := 0; i < idx; i++ {
			other.children[i] = nil
		}
	} else {
		other = &listLeafNode{}
		copy(other.children[idx:][:], n.children[idx:][:])
	}
	return other
}

// deleteBefore returns a new node with all elements before index removed.
func (n *listLeafNode) deleteAfter(index int, mutable bool) listNode {
	if !n.containsAfter(index) {
		return n
	}

	idx := index & listNodeMask
	var other *listLeafNode
	if mutable {
		other = n
		for i := idx + 1; i < len(n.children); i++ {
			other.children[i] = nil
		}
	} else {
		other = &listLeafNode{}
		copy(other.children[:idx+1][:], n.children[:idx+1][:])
	}
	return other
}

// ListIterator represents an ordered iterator over a list.
type ListIterator struct {
	list  *List // source list
	index int   // current index position

	stack [32]listIteratorElem // search stack
	depth int                  // stack depth
}

// Done returns true if no more elements remain in the iterator.
func (itr *ListIterator) Done() bool {
	return itr.index < 0 || itr.index >= itr.list.Len()
}

// First positions the iterator on the first index.
// If source list is empty then no change is made.
func (itr *ListIterator) First() {
	if itr.list.Len() != 0 {
		itr.Seek(0)
	}
}

// Last positions the iterator on the last index.
// If source list is empty then no change is made.
func (itr *ListIterator) Last() {
	if n := itr.list.Len(); n != 0 {
		itr.Seek(n - 1)
	}
}

// Seek moves the iterator position to the given index in the list.
// Similar to Go slices, this method will panic if index is below zero or if
// the index is greater than or equal to the list size.
func (itr *ListIterator) Seek(index int) {
	// Panic similar to Go slices.
	if index < 0 || index >= itr.list.Len() {
		panic(fmt.Sprintf("immutable.ListIterator.Seek: index %d out of bounds", index))
	}
	itr.index = index

	// Reset to the bottom of the stack at seek to the correct position.
	itr.stack[0] = listIteratorElem{node: itr.list.root}
	itr.depth = 0
	itr.seek(index)
}

// Next returns the current index and its value & moves the iterator forward.
// Returns an index of -1 if the there are no more elements to return.
func (itr *ListIterator) Next() (index int, value interface{}) {
	// Exit immediately if there are no elements remaining.
	if itr.Done() {
		return -1, nil
	}

	// Retrieve current index & value.
	elem := &itr.stack[itr.depth]
	index, value = itr.index, elem.node.(*listLeafNode).children[elem.index]

	// Increase index. If index is at the end then return immediately.
	itr.index++
	if itr.Done() {
		return index, value
	}

	// Move up stack until we find a node that has remaining position ahead.
	for ; itr.depth > 0 && itr.stack[itr.depth].index >= listNodeSize-1; itr.depth-- {
	}

	// Seek to correct position from current depth.
	itr.seek(itr.index)

	return index, value
}

// Prev returns the current index and value and moves the iterator backward.
// Returns an index of -1 if the there are no more elements to return.
func (itr *ListIterator) Prev() (index int, value interface{}) {
	// Exit immediately if there are no elements remaining.
	if itr.Done() {
		return -1, nil
	}

	// Retrieve current index & value.
	elem := &itr.stack[itr.depth]
	index, value = itr.index, elem.node.(*listLeafNode).children[elem.index]

	// Decrease index. If index is past the beginning then return immediately.
	itr.index--
	if itr.Done() {
		return index, value
	}

	// Move up stack until we find a node that has remaining position behind.
	for ; itr.depth > 0 && itr.stack[itr.depth].index == 0; itr.depth-- {
	}

	// Seek to correct position from current depth.
	itr.seek(itr.index)

	return index, value
}

// seek positions the stack to the given index from the current depth.
// Elements and indexes below the current depth are assumed to be correct.
func (itr *ListIterator) seek(index int) {
	// Iterate over each level until we reach a leaf node.
	for {
		elem := &itr.stack[itr.depth]
		elem.index = ((itr.list.origin + index) >> (elem.node.depth() * listNodeBits)) & listNodeMask

		switch node := elem.node.(type) {
		case *listBranchNode:
			child := node.children[elem.index]
			itr.stack[itr.depth+1] = listIteratorElem{node: child}
			itr.depth++
		case *listLeafNode:
			return
		}
	}
}

// listIteratorElem represents the node and it's child index within the stack.
type listIteratorElem struct {
	node  listNode
	index int
}

// Size thresholds for each type of branch node.
const (
	maxArrayMapSize      = 8
	maxBitmapIndexedSize = 16
)

// Segment bit shifts within the map tree.
const (
	mapNodeBits = 5
	mapNodeSize = 1 << mapNodeBits
	mapNodeMask = mapNodeSize - 1
)

// Map represents an immutable hash map implementation. The map uses a Hasher
// to generate hashes and check for equality of key values.
//
// It is implemented as an Hash Array Mapped Trie.
type Map struct {
	size   int     // total number of key/value pairs
	root   mapNode // root node of trie
	hasher Hasher  // hasher implementation
}

// NewMap returns a new instance of Map. If hasher is nil, a default hasher
// implementation will automatically be chosen based on the first key added.
// Default hasher implementations only exist for int, string, and byte slice types.
func NewMap(hasher Hasher) *Map {
	return &Map{
		hasher: hasher,
	}
}

// Len returns the number of elements in the map.
func (m *Map) Len() int {
	return m.size
}

// clone returns a shallow copy of m.
func (m *Map) clone() *Map {
	other := *m
	return &other
}

// Get returns the value for a given key and a flag indicating whether the
// key exists. This flag distinguishes a nil value set on a key versus a
// non-existent key in the map.
func (m *Map) Get(key interface{}) (value interface{}, ok bool) {
	if m.root == nil {
		return nil, false
	}
	keyHash := m.hasher.Hash(key)
	return m.root.get(key, 0, keyHash, m.hasher)
}

// Set returns a map with the key set to the new value. A nil value is allowed.
//
// This function will return a new map even if the updated value is the same as
// the existing value because Map does not track value equality.
func (m *Map) Set(key, value interface{}) *Map {
	return m.set(key, value, false)
}

func (m *Map) set(key, value interface{}, mutable bool) *Map {
	// Set a hasher on the first value if one does not already exist.
	hasher := m.hasher
	if hasher == nil {
		hasher = NewHasher(key)
	}

	// Generate copy if necessary.
	other := m
	if !mutable {
		other = m.clone()
	}
	other.hasher = hasher

	// If the map is empty, initialize with a simple array node.
	if m.root == nil {
		other.size = 1
		other.root = &mapArrayNode{entries: []mapEntry{{key: key, value: value}}}
		return other
	}

	// Otherwise copy the map and delegate insertion to the root.
	// Resized will return true if the key does not currently exist.
	var resized bool
	other.root = m.root.set(key, value, 0, hasher.Hash(key), hasher, mutable, &resized)
	if resized {
		other.size++
	}
	return other
}

// Delete returns a map with the given key removed.
// Removing a non-existent key will cause this method to return the same map.
func (m *Map) Delete(key interface{}) *Map {
	return m.delete(key, false)
}

func (m *Map) delete(key interface{}, mutable bool) *Map {
	// Return original map if no keys exist.
	if m.root == nil {
		return m
	}

	// If the delete did not change the node then return the original map.
	var resized bool
	newRoot := m.root.delete(key, 0, m.hasher.Hash(key), m.hasher, mutable, &resized)
	if !resized {
		return m
	}

	// Generate copy if necessary.
	other := m
	if !mutable {
		other = m.clone()
	}

	// Return copy of map with new root and decreased size.
	other.size = m.size - 1
	other.root = newRoot
	return other
}

// Iterator returns a new iterator for the map.
func (m *Map) Iterator() *MapIterator {
	itr := &MapIterator{m: m}
	itr.First()
	return itr
}

// MapBuilder represents an efficient builder for creating Maps.
type MapBuilder struct {
	m *Map // current state
}

// NewMapBuilder returns a new instance of MapBuilder.
func NewMapBuilder(hasher Hasher) *MapBuilder {
	return &MapBuilder{m: NewMap(hasher)}
}

// Map returns the underlying map. Only call once.
// Builder is invalid after call. Will panic on second invocation.
func (b *MapBuilder) Map() *Map {
	assert(b.m != nil, "immutable.SortedMapBuilder.Map(): duplicate call to fetch map")
	m := b.m
	b.m = nil
	return m
}

// Len returns the number of elements in the underlying map.
func (b *MapBuilder) Len() int {
	assert(b.m != nil, "immutable.MapBuilder: builder invalid after Map() invocation")
	return b.m.Len()
}

// Get returns the value for the given key.
func (b *MapBuilder) Get(key interface{}) (value interface{}, ok bool) {
	assert(b.m != nil, "immutable.MapBuilder: builder invalid after Map() invocation")
	return b.m.Get(key)
}

// Set sets the value of the given key. See Map.Set() for additional details.
func (b *MapBuilder) Set(key, value interface{}) {
	assert(b.m != nil, "immutable.MapBuilder: builder invalid after Map() invocation")
	b.m = b.m.set(key, value, true)
}

// Delete removes the given key. See Map.Delete() for additional details.
func (b *MapBuilder) Delete(key interface{}) {
	assert(b.m != nil, "immutable.MapBuilder: builder invalid after Map() invocation")
	b.m = b.m.delete(key, true)
}

// Iterator returns a new iterator for the underlying map.
func (b *MapBuilder) Iterator() *MapIterator {
	assert(b.m != nil, "immutable.MapBuilder: builder invalid after Map() invocation")
	return b.m.Iterator()
}

// mapNode represents any node in the map tree.
type mapNode interface {
	get(key interface{}, shift uint, keyHash uint32, h Hasher) (value interface{}, ok bool)
	set(key, value interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode
	delete(key interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode
}

var _ mapNode = (*mapArrayNode)(nil)
var _ mapNode = (*mapBitmapIndexedNode)(nil)
var _ mapNode = (*mapHashArrayNode)(nil)
var _ mapNode = (*mapValueNode)(nil)
var _ mapNode = (*mapHashCollisionNode)(nil)

// mapLeafNode represents a node that stores a single key hash at the leaf of the map tree.
type mapLeafNode interface {
	mapNode
	keyHashValue() uint32
}

var _ mapLeafNode = (*mapValueNode)(nil)
var _ mapLeafNode = (*mapHashCollisionNode)(nil)

// mapArrayNode is a map node that stores key/value pairs in a slice.
// Entries are stored in insertion order. An array node expands into a bitmap
// indexed node once a given threshold size is crossed.
type mapArrayNode struct {
	entries []mapEntry
}

// indexOf returns the entry index of the given key. Returns -1 if key not found.
func (n *mapArrayNode) indexOf(key interface{}, h Hasher) int {
	for i := range n.entries {
		if h.Equal(n.entries[i].key, key) {
			return i
		}
	}
	return -1
}

// get returns the value for the given key.
func (n *mapArrayNode) get(key interface{}, shift uint, keyHash uint32, h Hasher) (value interface{}, ok bool) {
	i := n.indexOf(key, h)
	if i == -1 {
		return nil, false
	}
	return n.entries[i].value, true
}

// set inserts or updates the value for a given key. If the key is inserted and
// the new size crosses the max size threshold, a bitmap indexed node is returned.
func (n *mapArrayNode) set(key, value interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	idx := n.indexOf(key, h)

	// Mark as resized if the key doesn't exist.
	if idx == -1 {
		*resized = true
	}

	// If we are adding and it crosses the max size threshold, expand the node.
	// We do this by continually setting the entries to a value node and expanding.
	if idx == -1 && len(n.entries) >= maxArrayMapSize {
		var node mapNode = newMapValueNode(h.Hash(key), key, value)
		for _, entry := range n.entries {
			node = node.set(entry.key, entry.value, 0, h.Hash(entry.key), h, false, resized)
		}
		return node
	}

	// Update in-place if mutable.
	if mutable {
		if idx != -1 {
			n.entries[idx] = mapEntry{key, value}
		} else {
			n.entries = append(n.entries, mapEntry{key, value})
		}
		return n
	}

	// Update existing entry if a match is found.
	// Otherwise append to the end of the element list if it doesn't exist.
	var other mapArrayNode
	if idx != -1 {
		other.entries = make([]mapEntry, len(n.entries))
		copy(other.entries, n.entries)
		other.entries[idx] = mapEntry{key, value}
	} else {
		other.entries = make([]mapEntry, len(n.entries)+1)
		copy(other.entries, n.entries)
		other.entries[len(other.entries)-1] = mapEntry{key, value}
	}
	return &other
}

// delete removes the given key from the node. Returns the same node if key does
// not exist. Returns a nil node when removing the last entry.
func (n *mapArrayNode) delete(key interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	idx := n.indexOf(key, h)

	// Return original node if key does not exist.
	if idx == -1 {
		return n
	}
	*resized = true

	// Return nil if this node will contain no nodes.
	if len(n.entries) == 1 {
		return nil
	}

	// Update in-place, if mutable.
	if mutable {
		copy(n.entries[idx:], n.entries[idx+1:])
		n.entries[len(n.entries)-1] = mapEntry{}
		n.entries = n.entries[:len(n.entries)-1]
		return n
	}

	// Otherwise create a copy with the given entry removed.
	other := &mapArrayNode{entries: make([]mapEntry, len(n.entries)-1)}
	copy(other.entries[:idx], n.entries[:idx])
	copy(other.entries[idx:], n.entries[idx+1:])
	return other
}

// mapBitmapIndexedNode represents a map branch node with a variable number of
// node slots and indexed using a bitmap. Indexes for the node slots are
// calculated by counting the number of set bits before the target bit using popcount.
type mapBitmapIndexedNode struct {
	bitmap uint32
	nodes  []mapNode
}

// get returns the value for the given key.
func (n *mapBitmapIndexedNode) get(key interface{}, shift uint, keyHash uint32, h Hasher) (value interface{}, ok bool) {
	bit := uint32(1) << ((keyHash >> shift) & mapNodeMask)
	if (n.bitmap & bit) == 0 {
		return nil, false
	}
	child := n.nodes[bits.OnesCount32(n.bitmap&(bit-1))]
	return child.get(key, shift+mapNodeBits, keyHash, h)
}

// set inserts or updates the value for the given key. If a new key is inserted
// and the size crosses the max size threshold then a hash array node is returned.
func (n *mapBitmapIndexedNode) set(key, value interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	// Extract the index for the bit segment of the key hash.
	keyHashFrag := (keyHash >> shift) & mapNodeMask

	// Determine the bit based on the hash index.
	bit := uint32(1) << keyHashFrag
	exists := (n.bitmap & bit) != 0

	// Mark as resized if the key doesn't exist.
	if !exists {
		*resized = true
	}

	// Find index of node based on popcount of bits before it.
	idx := bits.OnesCount32(n.bitmap & (bit - 1))

	// If the node already exists, delegate set operation to it.
	// If the node doesn't exist then create a simple value leaf node.
	var newNode mapNode
	if exists {
		newNode = n.nodes[idx].set(key, value, shift+mapNodeBits, keyHash, h, mutable, resized)
	} else {
		newNode = newMapValueNode(keyHash, key, value)
	}

	// Convert to a hash-array node once we exceed the max bitmap size.
	// Copy each node based on their bit position within the bitmap.
	if !exists && len(n.nodes) > maxBitmapIndexedSize {
		var other mapHashArrayNode
		for i := uint(0); i < uint(len(other.nodes)); i++ {
			if n.bitmap&(uint32(1)<<i) != 0 {
				other.nodes[i] = n.nodes[other.count]
				other.count++
			}
		}
		other.nodes[keyHashFrag] = newNode
		other.count++
		return &other
	}

	// Update in-place if mutable.
	if mutable {
		if exists {
			n.nodes[idx] = newNode
		} else {
			n.bitmap |= bit
			n.nodes = append(n.nodes, nil)
			copy(n.nodes[idx+1:], n.nodes[idx:])
			n.nodes[idx] = newNode
		}
		return n
	}

	// If node exists at given slot then overwrite it with new node.
	// Otherwise expand the node list and insert new node into appropriate position.
	other := &mapBitmapIndexedNode{bitmap: n.bitmap | bit}
	if exists {
		other.nodes = make([]mapNode, len(n.nodes))
		copy(other.nodes, n.nodes)
		other.nodes[idx] = newNode
	} else {
		other.nodes = make([]mapNode, len(n.nodes)+1)
		copy(other.nodes, n.nodes[:idx])
		other.nodes[idx] = newNode
		copy(other.nodes[idx+1:], n.nodes[idx:])
	}
	return other
}

// delete removes the key from the tree. If the key does not exist then the
// original node is returned. If removing the last child node then a nil is
// returned. Note that shrinking the node will not convert it to an array node.
func (n *mapBitmapIndexedNode) delete(key interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	bit := uint32(1) << ((keyHash >> shift) & mapNodeMask)

	// Return original node if key does not exist.
	if (n.bitmap & bit) == 0 {
		return n
	}

	// Find index of node based on popcount of bits before it.
	idx := bits.OnesCount32(n.bitmap & (bit - 1))

	// Delegate delete to child node.
	child := n.nodes[idx]
	newChild := child.delete(key, shift+mapNodeBits, keyHash, h, mutable, resized)

	// Return original node if key doesn't exist in child.
	if !*resized {
		return n
	}

	// Remove if returned child has been deleted.
	if newChild == nil {
		// If we won't have any children then return nil.
		if len(n.nodes) == 1 {
			return nil
		}

		// Update in-place if mutable.
		if mutable {
			n.bitmap ^= bit
			copy(n.nodes[idx:], n.nodes[idx+1:])
			n.nodes[len(n.nodes)-1] = nil
			n.nodes = n.nodes[:len(n.nodes)-1]
			return n
		}

		// Return copy with bit removed from bitmap and node removed from node list.
		other := &mapBitmapIndexedNode{bitmap: n.bitmap ^ bit, nodes: make([]mapNode, len(n.nodes)-1)}
		copy(other.nodes[:idx], n.nodes[:idx])
		copy(other.nodes[idx:], n.nodes[idx+1:])
		return other
	}

	// Generate copy, if necessary.
	other := n
	if !mutable {
		other = &mapBitmapIndexedNode{bitmap: n.bitmap, nodes: make([]mapNode, len(n.nodes))}
		copy(other.nodes, n.nodes)
	}

	// Update child.
	other.nodes[idx] = newChild
	return other
}

// mapHashArrayNode is a map branch node that stores nodes in a fixed length
// array. Child nodes are indexed by their index bit segment for the current depth.
type mapHashArrayNode struct {
	count uint                 // number of set nodes
	nodes [mapNodeSize]mapNode // child node slots, may contain empties
}

// clone returns a shallow copy of n.
func (n *mapHashArrayNode) clone() *mapHashArrayNode {
	other := *n
	return &other
}

// get returns the value for the given key.
func (n *mapHashArrayNode) get(key interface{}, shift uint, keyHash uint32, h Hasher) (value interface{}, ok bool) {
	node := n.nodes[(keyHash>>shift)&mapNodeMask]
	if node == nil {
		return nil, false
	}
	return node.get(key, shift+mapNodeBits, keyHash, h)
}

// set returns a node with the value set for the given key.
func (n *mapHashArrayNode) set(key, value interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	idx := (keyHash >> shift) & mapNodeMask
	node := n.nodes[idx]

	// If node at index doesn't exist, create a simple value leaf node.
	// Otherwise delegate set to child node.
	var newNode mapNode
	if node == nil {
		*resized = true
		newNode = newMapValueNode(keyHash, key, value)
	} else {
		newNode = node.set(key, value, shift+mapNodeBits, keyHash, h, mutable, resized)
	}

	// Generate copy, if necessary.
	other := n
	if !mutable {
		other = n.clone()
	}

	// Update child node (and update size, if new).
	if node == nil {
		other.count++
	}
	other.nodes[idx] = newNode
	return other
}

// delete returns a node with the given key removed. Returns the same node if
// the key does not exist. If node shrinks to within bitmap-indexed size then
// converts to a bitmap-indexed node.
func (n *mapHashArrayNode) delete(key interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	idx := (keyHash >> shift) & mapNodeMask
	node := n.nodes[idx]

	// Return original node if child is not found.
	if node == nil {
		return n
	}

	// Return original node if child is unchanged.
	newNode := node.delete(key, shift+mapNodeBits, keyHash, h, mutable, resized)
	if !*resized {
		return n
	}

	// If we remove a node and drop below a threshold, convert back to bitmap indexed node.
	if newNode == nil && n.count <= maxBitmapIndexedSize {
		other := &mapBitmapIndexedNode{nodes: make([]mapNode, 0, n.count-1)}
		for i, child := range n.nodes {
			if child != nil && uint32(i) != idx {
				other.bitmap |= 1 << uint(i)
				other.nodes = append(other.nodes, child)
			}
		}
		return other
	}

	// Generate copy, if necessary.
	other := n
	if !mutable {
		other = n.clone()
	}

	// Return copy of node with child updated.
	other.nodes[idx] = newNode
	if newNode == nil {
		other.count--
	}
	return other
}

// mapValueNode represents a leaf node with a single key/value pair.
// A value node can be converted to a hash collision leaf node if a different
// key with the same keyHash is inserted.
type mapValueNode struct {
	keyHash uint32
	key     interface{}
	value   interface{}
}

// newMapValueNode returns a new instance of mapValueNode.
func newMapValueNode(keyHash uint32, key, value interface{}) *mapValueNode {
	return &mapValueNode{
		keyHash: keyHash,
		key:     key,
		value:   value,
	}
}

// keyHashValue returns the key hash for this node.
func (n *mapValueNode) keyHashValue() uint32 {
	return n.keyHash
}

// get returns the value for the given key.
func (n *mapValueNode) get(key interface{}, shift uint, keyHash uint32, h Hasher) (value interface{}, ok bool) {
	if !h.Equal(n.key, key) {
		return nil, false
	}
	return n.value, true
}

// set returns a new node with the new value set for the key. If the key equals
// the node's key then a new value node is returned. If key is not equal to the
// node's key but has the same hash then a hash collision node is returned.
// Otherwise the nodes are merged into a branch node.
func (n *mapValueNode) set(key, value interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	// If the keys match then return a new value node overwriting the value.
	if h.Equal(n.key, key) {
		// Update in-place if mutable.
		if mutable {
			n.value = value
			return n
		}
		// Otherwise return a new copy.
		return newMapValueNode(n.keyHash, key, value)
	}

	*resized = true

	// Recursively merge nodes together if key hashes are different.
	if n.keyHash != keyHash {
		return mergeIntoNode(n, shift, keyHash, key, value)
	}

	// Merge into collision node if hash matches.
	return &mapHashCollisionNode{keyHash: keyHash, entries: []mapEntry{
		{key: n.key, value: n.value},
		{key: key, value: value},
	}}
}

// delete returns nil if the key matches the node's key. Otherwise returns the original node.
func (n *mapValueNode) delete(key interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	// Return original node if the keys do not match.
	if !h.Equal(n.key, key) {
		return n
	}

	// Otherwise remove the node if keys do match.
	*resized = true
	return nil
}

// mapHashCollisionNode represents a leaf node that contains two or more key/value
// pairs with the same key hash. Single pairs for a hash are stored as value nodes.
type mapHashCollisionNode struct {
	keyHash uint32 // key hash for all entries
	entries []mapEntry
}

// keyHashValue returns the key hash for all entries on the node.
func (n *mapHashCollisionNode) keyHashValue() uint32 {
	return n.keyHash
}

// indexOf returns the index of the entry for the given key.
// Returns -1 if the key does not exist in the node.
func (n *mapHashCollisionNode) indexOf(key interface{}, h Hasher) int {
	for i := range n.entries {
		if h.Equal(n.entries[i].key, key) {
			return i
		}
	}
	return -1
}

// get returns the value for the given key.
func (n *mapHashCollisionNode) get(key interface{}, shift uint, keyHash uint32, h Hasher) (value interface{}, ok bool) {
	for i := range n.entries {
		if h.Equal(n.entries[i].key, key) {
			return n.entries[i].value, true
		}
	}
	return nil, false
}

// set returns a copy of the node with key set to the given value.
func (n *mapHashCollisionNode) set(key, value interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	// Merge node with key/value pair if this is not a hash collision.
	if n.keyHash != keyHash {
		*resized = true
		return mergeIntoNode(n, shift, keyHash, key, value)
	}

	// Update in-place if mutable.
	if mutable {
		if idx := n.indexOf(key, h); idx == -1 {
			*resized = true
			n.entries = append(n.entries, mapEntry{key, value})
		} else {
			n.entries[idx] = mapEntry{key, value}
		}
		return n
	}

	// Append to end of node if key doesn't exist & mark resized.
	// Otherwise copy nodes and overwrite at matching key index.
	other := &mapHashCollisionNode{keyHash: n.keyHash}
	if idx := n.indexOf(key, h); idx == -1 {
		*resized = true
		other.entries = make([]mapEntry, len(n.entries)+1)
		copy(other.entries, n.entries)
		other.entries[len(other.entries)-1] = mapEntry{key, value}
	} else {
		other.entries = make([]mapEntry, len(n.entries))
		copy(other.entries, n.entries)
		other.entries[idx] = mapEntry{key, value}
	}
	return other
}

// delete returns a node with the given key deleted. Returns the same node if
// the key does not exist. If removing the key would shrink the node to a single
// entry then a value node is returned.
func (n *mapHashCollisionNode) delete(key interface{}, shift uint, keyHash uint32, h Hasher, mutable bool, resized *bool) mapNode {
	idx := n.indexOf(key, h)

	// Return original node if key is not found.
	if idx == -1 {
		return n
	}

	// Mark as resized if key exists.
	*resized = true

	// Convert to value node if we move to one entry.
	if len(n.entries) == 2 {
		return &mapValueNode{
			keyHash: n.keyHash,
			key:     n.entries[idx^1].key,
			value:   n.entries[idx^1].value,
		}
	}

	// Remove entry in-place if mutable.
	if mutable {
		copy(n.entries[idx:], n.entries[idx+1:])
		n.entries[len(n.entries)-1] = mapEntry{}
		n.entries = n.entries[:len(n.entries)-1]
		return n
	}

	// Return copy without entry if immutable.
	other := &mapHashCollisionNode{keyHash: n.keyHash, entries: make([]mapEntry, len(n.entries)-1)}
	copy(other.entries[:idx], n.entries[:idx])
	copy(other.entries[idx:], n.entries[idx+1:])
	return other
}

// mergeIntoNode merges a key/value pair into an existing node.
// Caller must verify that node's keyHash is not equal to keyHash.
func mergeIntoNode(node mapLeafNode, shift uint, keyHash uint32, key, value interface{}) mapNode {
	idx1 := (node.keyHashValue() >> shift) & mapNodeMask
	idx2 := (keyHash >> shift) & mapNodeMask

	// Recursively build branch nodes to combine the node and its key.
	other := &mapBitmapIndexedNode{bitmap: (1 << idx1) | (1 << idx2)}
	if idx1 == idx2 {
		other.nodes = []mapNode{mergeIntoNode(node, shift+mapNodeBits, keyHash, key, value)}
	} else {
		if newNode := newMapValueNode(keyHash, key, value); idx1 < idx2 {
			other.nodes = []mapNode{node, newNode}
		} else {
			other.nodes = []mapNode{newNode, node}
		}
	}
	return other
}

// mapEntry represents a single key/value pair.
type mapEntry struct {
	key   interface{}
	value interface{}
}

// MapIterator represents an iterator over a map's key/value pairs. Although
// map keys are not sorted, the iterator's order is deterministic.
type MapIterator struct {
	m *Map // source map

	stack [32]mapIteratorElem // search stack
	depth int                 // stack depth
}

// Done returns true if no more elements remain in the iterator.
func (itr *MapIterator) Done() bool {
	return itr.depth == -1
}

// First resets the iterator to the first key/value pair.
func (itr *MapIterator) First() {
	// Exit immediately if the map is empty.
	if itr.m.root == nil {
		itr.depth = -1
		return
	}

	// Initialize the stack to the left most element.
	itr.stack[0] = mapIteratorElem{node: itr.m.root}
	itr.depth = 0
	itr.first()
}

// Next returns the next key/value pair. Returns a nil key when no elements remain.
func (itr *MapIterator) Next() (key, value interface{}) {
	// Return nil key if iteration is done.
	if itr.Done() {
		return nil, nil
	}

	// Retrieve current index & value. Current node is always a leaf.
	elem := &itr.stack[itr.depth]
	switch node := elem.node.(type) {
	case *mapArrayNode:
		entry := &node.entries[elem.index]
		key, value = entry.key, entry.value
	case *mapValueNode:
		key, value = node.key, node.value
	case *mapHashCollisionNode:
		entry := &node.entries[elem.index]
		key, value = entry.key, entry.value
	}

	// Move up stack until we find a node that has remaining position ahead
	// and move that element forward by one.
	itr.next()
	return key, value
}

// next moves to the next available key.
func (itr *MapIterator) next() {
	for ; itr.depth >= 0; itr.depth-- {
		elem := &itr.stack[itr.depth]

		switch node := elem.node.(type) {
		case *mapArrayNode:
			if elem.index < len(node.entries)-1 {
				elem.index++
				return
			}

		case *mapBitmapIndexedNode:
			if elem.index < len(node.nodes)-1 {
				elem.index++
				itr.stack[itr.depth+1].node = node.nodes[elem.index]
				itr.depth++
				itr.first()
				return
			}

		case *mapHashArrayNode:
			for i := elem.index + 1; i < len(node.nodes); i++ {
				if node.nodes[i] != nil {
					elem.index = i
					itr.stack[itr.depth+1].node = node.nodes[elem.index]
					itr.depth++
					itr.first()
					return
				}
			}

		case *mapValueNode:
			continue // always the last value, traverse up

		case *mapHashCollisionNode:
			if elem.index < len(node.entries)-1 {
				elem.index++
				return
			}
		}
	}
}

// first positions the stack left most index.
// Elements and indexes at and below the current depth are assumed to be correct.
func (itr *MapIterator) first() {
	for ; ; itr.depth++ {
		elem := &itr.stack[itr.depth]

		switch node := elem.node.(type) {
		case *mapBitmapIndexedNode:
			elem.index = 0
			itr.stack[itr.depth+1].node = node.nodes[0]

		case *mapHashArrayNode:
			for i := 0; i < len(node.nodes); i++ {
				if node.nodes[i] != nil { // find first node
					elem.index = i
					itr.stack[itr.depth+1].node = node.nodes[i]
					break
				}
			}

		default: // *mapArrayNode, mapLeafNode
			elem.index = 0
			return
		}
	}
}

// mapIteratorElem represents a node/index pair in the MapIterator stack.
type mapIteratorElem struct {
	node  mapNode
	index int
}

// Sorted map child node limit size.
const (
	sortedMapNodeSize = 32
)

// SortedMap represents a map of key/value pairs sorted by key. The sort order
// is determined by the Comparer used by the map.
//
// This map is implemented as a B+tree.
type SortedMap struct {
	size     int           // total number of key/value pairs
	root     sortedMapNode // root of b+tree
	comparer Comparer
}

// NewSortedMap returns a new instance of SortedMap. If comparer is nil then
// a default comparer is set after the first key is inserted. Default comparers
// exist for int, string, and byte slice keys.
func NewSortedMap(comparer Comparer) *SortedMap {
	return &SortedMap{
		comparer: comparer,
	}
}

// Len returns the number of elements in the sorted map.
func (m *SortedMap) Len() int {
	return m.size
}

// Get returns the value for a given key and a flag indicating if the key is set.
// The flag can be used to distinguish between a nil-set key versus an unset key.
func (m *SortedMap) Get(key interface{}) (interface{}, bool) {
	if m.root == nil {
		return nil, false
	}
	return m.root.get(key, m.comparer)
}

// Set returns a copy of the map with the key set to the given value.
func (m *SortedMap) Set(key, value interface{}) *SortedMap {
	return m.set(key, value, false)
}

func (m *SortedMap) set(key, value interface{}, mutable bool) *SortedMap {
	// Set a comparer on the first value if one does not already exist.
	comparer := m.comparer
	if comparer == nil {
		comparer = NewComparer(key)
	}

	// Create copy, if necessary.
	other := m
	if !mutable {
		other = m.clone()
	}
	other.comparer = comparer

	// If no values are set then initialize with a leaf node.
	if m.root == nil {
		other.size = 1
		other.root = &sortedMapLeafNode{entries: []mapEntry{{key: key, value: value}}}
		return other
	}

	// Otherwise delegate to root node.
	// If a split occurs then grow the tree from the root.
	var resized bool
	newRoot, splitNode := m.root.set(key, value, comparer, mutable, &resized)
	if splitNode != nil {
		newRoot = newSortedMapBranchNode(newRoot, splitNode)
	}

	// Update root and size (if resized).
	other.size = m.size
	other.root = newRoot
	if resized {
		other.size++
	}
	return other
}

// Delete returns a copy of the map with the key removed.
// Returns the original map if key does not exist.
func (m *SortedMap) Delete(key interface{}) *SortedMap {
	return m.delete(key, false)
}

func (m *SortedMap) delete(key interface{}, mutable bool) *SortedMap {
	// Return original map if no keys exist.
	if m.root == nil {
		return m
	}

	// If the delete did not change the node then return the original map.
	var resized bool
	newRoot := m.root.delete(key, m.comparer, mutable, &resized)
	if !resized {
		return m
	}

	// Create copy, if necessary.
	other := m
	if !mutable {
		other = m.clone()
	}

	// Update root and size.
	other.size = m.size - 1
	other.root = newRoot
	return other
}

// clone returns a shallow copy of m.
func (m *SortedMap) clone() *SortedMap {
	other := *m
	return &other
}

// Iterator returns a new iterator for this map positioned at the first key.
func (m *SortedMap) Iterator() *SortedMapIterator {
	itr := &SortedMapIterator{m: m}
	itr.First()
	return itr
}

// SortedMapBuilder represents an efficient builder for creating sorted maps.
type SortedMapBuilder struct {
	m *SortedMap // current state
}

// NewSortedMapBuilder returns a new instance of SortedMapBuilder.
func NewSortedMapBuilder(comparer Comparer) *SortedMapBuilder {
	return &SortedMapBuilder{m: NewSortedMap(comparer)}
}

// SortedMap returns the current copy of the map.
// The returned map is safe to use even if after the builder continues to be used.
func (b *SortedMapBuilder) Map() *SortedMap {
	assert(b.m != nil, "immutable.SortedMapBuilder.Map(): duplicate call to fetch map")
	m := b.m
	b.m = nil
	return m
}

// Len returns the number of elements in the underlying map.
func (b *SortedMapBuilder) Len() int {
	assert(b.m != nil, "immutable.SortedMapBuilder: builder invalid after Map() invocation")
	return b.m.Len()
}

// Get returns the value for the given key.
func (b *SortedMapBuilder) Get(key interface{}) (value interface{}, ok bool) {
	assert(b.m != nil, "immutable.SortedMapBuilder: builder invalid after Map() invocation")
	return b.m.Get(key)
}

// Set sets the value of the given key. See SortedMap.Set() for additional details.
func (b *SortedMapBuilder) Set(key, value interface{}) {
	assert(b.m != nil, "immutable.SortedMapBuilder: builder invalid after Map() invocation")
	b.m = b.m.set(key, value, true)
}

// Delete removes the given key. See SortedMap.Delete() for additional details.
func (b *SortedMapBuilder) Delete(key interface{}) {
	assert(b.m != nil, "immutable.SortedMapBuilder: builder invalid after Map() invocation")
	b.m = b.m.delete(key, true)
}

// Iterator returns a new iterator for the underlying map positioned at the first key.
func (b *SortedMapBuilder) Iterator() *SortedMapIterator {
	assert(b.m != nil, "immutable.SortedMapBuilder: builder invalid after Map() invocation")
	return b.m.Iterator()
}

// sortedMapNode represents a branch or leaf node in the sorted map.
type sortedMapNode interface {
	minKey() interface{}
	indexOf(key interface{}, c Comparer) int
	get(key interface{}, c Comparer) (value interface{}, ok bool)
	set(key, value interface{}, c Comparer, mutable bool, resized *bool) (sortedMapNode, sortedMapNode)
	delete(key interface{}, c Comparer, mutable bool, resized *bool) sortedMapNode
}

var _ sortedMapNode = (*sortedMapBranchNode)(nil)
var _ sortedMapNode = (*sortedMapLeafNode)(nil)

// sortedMapBranchNode represents a branch in the sorted map.
type sortedMapBranchNode struct {
	elems []sortedMapBranchElem
}

// newSortedMapBranchNode returns a new branch node with the given child nodes.
func newSortedMapBranchNode(children ...sortedMapNode) *sortedMapBranchNode {
	// Fetch min keys for every child.
	elems := make([]sortedMapBranchElem, len(children))
	for i, child := range children {
		elems[i] = sortedMapBranchElem{
			key:  child.minKey(),
			node: child,
		}
	}

	return &sortedMapBranchNode{elems: elems}
}

// minKey returns the lowest key stored in this node's tree.
func (n *sortedMapBranchNode) minKey() interface{} {
	return n.elems[0].node.minKey()
}

// indexOf returns the index of the key within the child nodes.
func (n *sortedMapBranchNode) indexOf(key interface{}, c Comparer) int {
	if idx := sort.Search(len(n.elems), func(i int) bool { return c.Compare(n.elems[i].key, key) == 1 }); idx > 0 {
		return idx - 1
	}
	return 0
}

// get returns the value for the given key.
func (n *sortedMapBranchNode) get(key interface{}, c Comparer) (value interface{}, ok bool) {
	idx := n.indexOf(key, c)
	return n.elems[idx].node.get(key, c)
}

// set returns a copy of the node with the key set to the given value.
func (n *sortedMapBranchNode) set(key, value interface{}, c Comparer, mutable bool, resized *bool) (sortedMapNode, sortedMapNode) {
	idx := n.indexOf(key, c)

	// Delegate insert to child node.
	newNode, splitNode := n.elems[idx].node.set(key, value, c, mutable, resized)

	// Update in-place, if mutable.
	if mutable {
		n.elems[idx] = sortedMapBranchElem{key: newNode.minKey(), node: newNode}
		if splitNode != nil {
			n.elems = append(n.elems, sortedMapBranchElem{})
			copy(n.elems[idx+1:], n.elems[idx:])
			n.elems[idx+1] = sortedMapBranchElem{key: splitNode.minKey(), node: splitNode}
		}

		// If the child splits and we have no more room then we split too.
		if len(n.elems) > sortedMapNodeSize {
			splitIdx := len(n.elems) / 2
			newNode := &sortedMapBranchNode{elems: n.elems[:splitIdx:splitIdx]}
			splitNode := &sortedMapBranchNode{elems: n.elems[splitIdx:]}
			return newNode, splitNode
		}
		return n, nil
	}

	// If no split occurs, copy branch and update keys.
	// If the child splits, insert new key/child into copy of branch.
	var other sortedMapBranchNode
	if splitNode == nil {
		other.elems = make([]sortedMapBranchElem, len(n.elems))
		copy(other.elems, n.elems)
		other.elems[idx] = sortedMapBranchElem{
			key:  newNode.minKey(),
			node: newNode,
		}
	} else {
		other.elems = make([]sortedMapBranchElem, len(n.elems)+1)
		copy(other.elems[:idx], n.elems[:idx])
		copy(other.elems[idx+1:], n.elems[idx:])
		other.elems[idx] = sortedMapBranchElem{
			key:  newNode.minKey(),
			node: newNode,
		}
		other.elems[idx+1] = sortedMapBranchElem{
			key:  splitNode.minKey(),
			node: splitNode,
		}
	}

	// If the child splits and we have no more room then we split too.
	if len(other.elems) > sortedMapNodeSize {
		splitIdx := len(other.elems) / 2
		newNode := &sortedMapBranchNode{elems: other.elems[:splitIdx:splitIdx]}
		splitNode := &sortedMapBranchNode{elems: other.elems[splitIdx:]}
		return newNode, splitNode
	}

	// Otherwise return the new branch node with the updated entry.
	return &other, nil
}

// delete returns a node with the key removed. Returns the same node if the key
// does not exist. Returns nil if all child nodes are removed.
func (n *sortedMapBranchNode) delete(key interface{}, c Comparer, mutable bool, resized *bool) sortedMapNode {
	idx := n.indexOf(key, c)

	// Return original node if child has not changed.
	newNode := n.elems[idx].node.delete(key, c, mutable, resized)
	if !*resized {
		return n
	}

	// Remove child if it is now nil.
	if newNode == nil {
		// If this node will become empty then simply return nil.
		if len(n.elems) == 1 {
			return nil
		}

		// If mutable, update in-place.
		if mutable {
			copy(n.elems[idx:], n.elems[idx+1:])
			n.elems[len(n.elems)-1] = sortedMapBranchElem{}
			n.elems = n.elems[:len(n.elems)-1]
			return n
		}

		// Return a copy without the given node.
		other := &sortedMapBranchNode{elems: make([]sortedMapBranchElem, len(n.elems)-1)}
		copy(other.elems[:idx], n.elems[:idx])
		copy(other.elems[idx:], n.elems[idx+1:])
		return other
	}

	// If mutable, update in-place.
	if mutable {
		n.elems[idx] = sortedMapBranchElem{key: newNode.minKey(), node: newNode}
		return n
	}

	// Return a copy with the updated node.
	other := &sortedMapBranchNode{elems: make([]sortedMapBranchElem, len(n.elems))}
	copy(other.elems, n.elems)
	other.elems[idx] = sortedMapBranchElem{
		key:  newNode.minKey(),
		node: newNode,
	}
	return other
}

type sortedMapBranchElem struct {
	key  interface{}
	node sortedMapNode
}

// sortedMapLeafNode represents a leaf node in the sorted map.
type sortedMapLeafNode struct {
	entries []mapEntry
}

// minKey returns the first key stored in this node.
func (n *sortedMapLeafNode) minKey() interface{} {
	return n.entries[0].key
}

// indexOf returns the index of the given key.
func (n *sortedMapLeafNode) indexOf(key interface{}, c Comparer) int {
	return sort.Search(len(n.entries), func(i int) bool {
		return c.Compare(n.entries[i].key, key) != -1 // GTE
	})
}

// get returns the value of the given key.
func (n *sortedMapLeafNode) get(key interface{}, c Comparer) (value interface{}, ok bool) {
	idx := n.indexOf(key, c)

	// If the index is beyond the entry count or the key is not equal then return 'not found'.
	if idx == len(n.entries) || c.Compare(n.entries[idx].key, key) != 0 {
		return nil, false
	}

	// If the key matches then return its value.
	return n.entries[idx].value, true
}

// set returns a copy of node with the key set to the given value. If the update
// causes the node to grow beyond the maximum size then it is split in two.
func (n *sortedMapLeafNode) set(key, value interface{}, c Comparer, mutable bool, resized *bool) (sortedMapNode, sortedMapNode) {
	// Find the insertion index for the key.
	idx := n.indexOf(key, c)
	exists := idx < len(n.entries) && c.Compare(n.entries[idx].key, key) == 0

	// Update in-place, if mutable.
	if mutable {
		if !exists {
			*resized = true
			n.entries = append(n.entries, mapEntry{})
			copy(n.entries[idx+1:], n.entries[idx:])
		}
		n.entries[idx] = mapEntry{key: key, value: value}

		// If the key doesn't exist and we exceed our max allowed values then split.
		if len(n.entries) > sortedMapNodeSize {
			splitIdx := len(n.entries) / 2
			newNode := &sortedMapLeafNode{entries: n.entries[:splitIdx:splitIdx]}
			splitNode := &sortedMapLeafNode{entries: n.entries[splitIdx:]}
			return newNode, splitNode
		}
		return n, nil
	}

	// If the key matches then simply return a copy with the entry overridden.
	// If there is no match then insert new entry and mark as resized.
	var newEntries []mapEntry
	if exists {
		newEntries = make([]mapEntry, len(n.entries))
		copy(newEntries, n.entries)
		newEntries[idx] = mapEntry{key: key, value: value}
	} else {
		*resized = true
		newEntries = make([]mapEntry, len(n.entries)+1)
		copy(newEntries[:idx], n.entries[:idx])
		newEntries[idx] = mapEntry{key: key, value: value}
		copy(newEntries[idx+1:], n.entries[idx:])
	}

	// If the key doesn't exist and we exceed our max allowed values then split.
	if len(newEntries) > sortedMapNodeSize {
		splitIdx := len(newEntries) / 2
		newNode := &sortedMapLeafNode{entries: newEntries[:splitIdx:splitIdx]}
		splitNode := &sortedMapLeafNode{entries: newEntries[splitIdx:]}
		return newNode, splitNode
	}

	// Otherwise return the new leaf node with the updated entry.
	return &sortedMapLeafNode{entries: newEntries}, nil
}

// delete returns a copy of node with key removed. Returns the original node if
// the key does not exist. Returns nil if the removed key is the last remaining key.
func (n *sortedMapLeafNode) delete(key interface{}, c Comparer, mutable bool, resized *bool) sortedMapNode {
	idx := n.indexOf(key, c)

	// Return original node if key is not found.
	if idx >= len(n.entries) || c.Compare(n.entries[idx].key, key) != 0 {
		return n
	}
	*resized = true

	// If this is the last entry then return nil.
	if len(n.entries) == 1 {
		return nil
	}

	// Update in-place, if mutable.
	if mutable {
		copy(n.entries[idx:], n.entries[idx+1:])
		n.entries[len(n.entries)-1] = mapEntry{}
		n.entries = n.entries[:len(n.entries)-1]
		return n
	}

	// Return copy of node with entry removed.
	other := &sortedMapLeafNode{entries: make([]mapEntry, len(n.entries)-1)}
	copy(other.entries[:idx], n.entries[:idx])
	copy(other.entries[idx:], n.entries[idx+1:])
	return other
}

// SortedMapIterator represents an iterator over a sorted map.
// Iteration can occur in natural or reverse order based on use of Next() or Prev().
type SortedMapIterator struct {
	m *SortedMap // source map

	stack [32]sortedMapIteratorElem // search stack
	depth int                       // stack depth
}

// Done returns true if no more key/value pairs remain in the iterator.
func (itr *SortedMapIterator) Done() bool {
	return itr.depth == -1
}

// First moves the iterator to the first key/value pair.
func (itr *SortedMapIterator) First() {
	if itr.m.root == nil {
		itr.depth = -1
		return
	}
	itr.stack[0] = sortedMapIteratorElem{node: itr.m.root}
	itr.depth = 0
	itr.first()
}

// Last moves the iterator to the last key/value pair.
func (itr *SortedMapIterator) Last() {
	if itr.m.root == nil {
		itr.depth = -1
		return
	}
	itr.stack[0] = sortedMapIteratorElem{node: itr.m.root}
	itr.depth = 0
	itr.last()
}

// Seek moves the iterator position to the given key in the map.
// If the key does not exist then the next key is used. If no more keys exist
// then the iteartor is marked as done.
func (itr *SortedMapIterator) Seek(key interface{}) {
	if itr.m.root == nil {
		itr.depth = -1
		return
	}
	itr.stack[0] = sortedMapIteratorElem{node: itr.m.root}
	itr.depth = 0
	itr.seek(key)
}

// Next returns the current key/value pair and moves the iterator forward.
// Returns a nil key if the there are no more elements to return.
func (itr *SortedMapIterator) Next() (key, value interface{}) {
	// Return nil key if iteration is complete.
	if itr.Done() {
		return nil, nil
	}

	// Retrieve current key/value pair.
	leafElem := &itr.stack[itr.depth]
	leafNode := leafElem.node.(*sortedMapLeafNode)
	leafEntry := &leafNode.entries[leafElem.index]
	key, value = leafEntry.key, leafEntry.value

	// Move to the next available key/value pair.
	itr.next()

	// Only occurs when iterator is done.
	return key, value
}

// next moves to the next key. If no keys are after then depth is set to -1.
func (itr *SortedMapIterator) next() {
	for ; itr.depth >= 0; itr.depth-- {
		elem := &itr.stack[itr.depth]

		switch node := elem.node.(type) {
		case *sortedMapLeafNode:
			if elem.index < len(node.entries)-1 {
				elem.index++
				return
			}
		case *sortedMapBranchNode:
			if elem.index < len(node.elems)-1 {
				elem.index++
				itr.stack[itr.depth+1].node = node.elems[elem.index].node
				itr.depth++
				itr.first()
				return
			}
		}
	}
}

// Prev returns the current key/value pair and moves the iterator backward.
// Returns a nil key if the there are no more elements to return.
func (itr *SortedMapIterator) Prev() (key, value interface{}) {
	// Return nil key if iteration is complete.
	if itr.Done() {
		return nil, nil
	}

	// Retrieve current key/value pair.
	leafElem := &itr.stack[itr.depth]
	leafNode := leafElem.node.(*sortedMapLeafNode)
	leafEntry := &leafNode.entries[leafElem.index]
	key, value = leafEntry.key, leafEntry.value

	itr.prev()
	return key, value
}

// prev moves to the previous key. If no keys are before then depth is set to -1.
func (itr *SortedMapIterator) prev() {
	for ; itr.depth >= 0; itr.depth-- {
		elem := &itr.stack[itr.depth]

		switch node := elem.node.(type) {
		case *sortedMapLeafNode:
			if elem.index > 0 {
				elem.index--
				return
			}
		case *sortedMapBranchNode:
			if elem.index > 0 {
				elem.index--
				itr.stack[itr.depth+1].node = node.elems[elem.index].node
				itr.depth++
				itr.last()
				return
			}
		}
	}
}

// first positions the stack to the leftmost key from the current depth.
// Elements and indexes below the current depth are assumed to be correct.
func (itr *SortedMapIterator) first() {
	for {
		elem := &itr.stack[itr.depth]
		elem.index = 0

		switch node := elem.node.(type) {
		case *sortedMapBranchNode:
			itr.stack[itr.depth+1] = sortedMapIteratorElem{node: node.elems[elem.index].node}
			itr.depth++
		case *sortedMapLeafNode:
			return
		}
	}
}

// last positions the stack to the rightmost key from the current depth.
// Elements and indexes below the current depth are assumed to be correct.
func (itr *SortedMapIterator) last() {
	for {
		elem := &itr.stack[itr.depth]

		switch node := elem.node.(type) {
		case *sortedMapBranchNode:
			elem.index = len(node.elems) - 1
			itr.stack[itr.depth+1] = sortedMapIteratorElem{node: node.elems[elem.index].node}
			itr.depth++
		case *sortedMapLeafNode:
			elem.index = len(node.entries) - 1
			return
		}
	}
}

// seek positions the stack to the given key from the current depth.
// Elements and indexes below the current depth are assumed to be correct.
func (itr *SortedMapIterator) seek(key interface{}) {
	for {
		elem := &itr.stack[itr.depth]
		elem.index = elem.node.indexOf(key, itr.m.comparer)

		switch node := elem.node.(type) {
		case *sortedMapBranchNode:
			itr.stack[itr.depth+1] = sortedMapIteratorElem{node: node.elems[elem.index].node}
			itr.depth++
		case *sortedMapLeafNode:
			if elem.index == len(node.entries) {
				itr.next()
			}
			return
		}
	}
}

// sortedMapIteratorElem represents node/index pair in the SortedMapIterator stack.
type sortedMapIteratorElem struct {
	node  sortedMapNode
	index int
}

// Hasher hashes keys and checks them for equality.
type Hasher interface {
	// Computes a 32-bit hash for key.
	Hash(key interface{}) uint32

	// Returns true if a and b are equal.
	Equal(a, b interface{}) bool
}

// NewHasher returns the built-in hasher for a given key type.
func NewHasher(key interface{}) Hasher {
	// Attempt to use non-reflection based hasher first.
	switch key.(type) {
	case int:
		return &intHasher{}
	case int8:
		return &int8Hasher{}
	case int16:
		return &int16Hasher{}
	case int32:
		return &int32Hasher{}
	case int64:
		return &int64Hasher{}
	case uint:
		return &uintHasher{}
	case uint8:
		return &uint8Hasher{}
	case uint16:
		return &uint16Hasher{}
	case uint32:
		return &uint32Hasher{}
	case uint64:
		return &uint64Hasher{}
	case string:
		return &stringHasher{}
	case []byte:
		return &byteSliceHasher{}
	}

	// Fallback to reflection-based hasher otherwise.
	// This is used when caller wraps a type around a primitive type.
	switch reflect.TypeOf(key).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &reflectIntHasher{}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &reflectUintHasher{}
	case reflect.String:
		return &reflectStringHasher{}
	}

	// If no hashers match then panic.
	// This is a compile time issue so it should not return an error.
	panic(fmt.Sprintf("immutable.NewHasher: must set hasher for %T type", key))
}

// intHasher implements Hasher for int keys.
type intHasher struct{}

// Hash returns a hash for key.
func (h *intHasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(int)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not ints.
func (h *intHasher) Equal(a, b interface{}) bool {
	return a.(int) == b.(int)
}

// int8Hasher implements Hasher for int8 keys.
type int8Hasher struct{}

// Hash returns a hash for key.
func (h *int8Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(int8)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not int8s.
func (h *int8Hasher) Equal(a, b interface{}) bool {
	return a.(int8) == b.(int8)
}

// int16Hasher implements Hasher for int16 keys.
type int16Hasher struct{}

// Hash returns a hash for key.
func (h *int16Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(int16)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not int16s.
func (h *int16Hasher) Equal(a, b interface{}) bool {
	return a.(int16) == b.(int16)
}

// int32Hasher implements Hasher for int32 keys.
type int32Hasher struct{}

// Hash returns a hash for key.
func (h *int32Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(int32)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not int32s.
func (h *int32Hasher) Equal(a, b interface{}) bool {
	return a.(int32) == b.(int32)
}

// int64Hasher implements Hasher for int64 keys.
type int64Hasher struct{}

// Hash returns a hash for key.
func (h *int64Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(int64)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not int64s.
func (h *int64Hasher) Equal(a, b interface{}) bool {
	return a.(int64) == b.(int64)
}

// uintHasher implements Hasher for uint keys.
type uintHasher struct{}

// Hash returns a hash for key.
func (h *uintHasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(uint)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not uints.
func (h *uintHasher) Equal(a, b interface{}) bool {
	return a.(uint) == b.(uint)
}

// uint8Hasher implements Hasher for uint8 keys.
type uint8Hasher struct{}

// Hash returns a hash for key.
func (h *uint8Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(uint8)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not uint8s.
func (h *uint8Hasher) Equal(a, b interface{}) bool {
	return a.(uint8) == b.(uint8)
}

// uint16Hasher implements Hasher for uint16 keys.
type uint16Hasher struct{}

// Hash returns a hash for key.
func (h *uint16Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(uint16)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not uint16s.
func (h *uint16Hasher) Equal(a, b interface{}) bool {
	return a.(uint16) == b.(uint16)
}

// uint32Hasher implements Hasher for uint32 keys.
type uint32Hasher struct{}

// Hash returns a hash for key.
func (h *uint32Hasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(key.(uint32)))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not uint32s.
func (h *uint32Hasher) Equal(a, b interface{}) bool {
	return a.(uint32) == b.(uint32)
}

// uint64Hasher implements Hasher for uint64 keys.
type uint64Hasher struct{}

// Hash returns a hash for key.
func (h *uint64Hasher) Hash(key interface{}) uint32 {
	return hashUint64(key.(uint64))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not uint64s.
func (h *uint64Hasher) Equal(a, b interface{}) bool {
	return a.(uint64) == b.(uint64)
}

// stringHasher implements Hasher for string keys.
type stringHasher struct{}

// Hash returns a hash for value.
func (h *stringHasher) Hash(value interface{}) uint32 {
	var hash uint32
	for i, value := 0, value.(string); i < len(value); i++ {
		hash = 31*hash + uint32(value[i])
	}
	return hash
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not strings.
func (h *stringHasher) Equal(a, b interface{}) bool {
	return a.(string) == b.(string)
}

// byteSliceHasher implements Hasher for byte slice keys.
type byteSliceHasher struct{}

// Hash returns a hash for value.
func (h *byteSliceHasher) Hash(value interface{}) uint32 {
	var hash uint32
	for i, value := 0, value.([]byte); i < len(value); i++ {
		hash = 31*hash + uint32(value[i])
	}
	return hash
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not byte slices.
func (h *byteSliceHasher) Equal(a, b interface{}) bool {
	return bytes.Equal(a.([]byte), b.([]byte))
}

// reflectIntHasher implements a reflection-based Hasher for int keys.
type reflectIntHasher struct{}

// Hash returns a hash for key.
func (h *reflectIntHasher) Hash(key interface{}) uint32 {
	return hashUint64(uint64(reflect.ValueOf(key).Int()))
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not ints.
func (h *reflectIntHasher) Equal(a, b interface{}) bool {
	return reflect.ValueOf(a).Int() == reflect.ValueOf(b).Int()
}

// reflectUintHasher implements a reflection-based Hasher for uint keys.
type reflectUintHasher struct{}

// Hash returns a hash for key.
func (h *reflectUintHasher) Hash(key interface{}) uint32 {
	return hashUint64(reflect.ValueOf(key).Uint())
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not ints.
func (h *reflectUintHasher) Equal(a, b interface{}) bool {
	return reflect.ValueOf(a).Uint() == reflect.ValueOf(b).Uint()
}

// reflectStringHasher implements a refletion-based Hasher for string keys.
type reflectStringHasher struct{}

// Hash returns a hash for value.
func (h *reflectStringHasher) Hash(value interface{}) uint32 {
	var hash uint32
	s := reflect.ValueOf(value).String()
	for i := 0; i < len(s); i++ {
		hash = 31*hash + uint32(s[i])
	}
	return hash
}

// Equal returns true if a is equal to b. Otherwise returns false.
// Panics if a and b are not strings.
func (h *reflectStringHasher) Equal(a, b interface{}) bool {
	return reflect.ValueOf(a).String() == reflect.ValueOf(b).String()
}

// hashUint64 returns a 32-bit hash for a 64-bit value.
func hashUint64(value uint64) uint32 {
	hash := value
	for value > 0xffffffff {
		value /= 0xffffffff
		hash ^= value
	}
	return uint32(hash)
}

// Comparer allows the comparison of two keys for the purpose of sorting.
type Comparer interface {
	// Returns -1 if a is less than b, returns 1 if a is greater than b,
	// and returns 0 if a is equal to b.
	Compare(a, b interface{}) int
}

// NewComparer returns the built-in comparer for a given key type.
func NewComparer(key interface{}) Comparer {
	// Attempt to use non-reflection based comparer first.
	switch key.(type) {
	case int:
		return &intComparer{}
	case int8:
		return &int8Comparer{}
	case int16:
		return &int16Comparer{}
	case int32:
		return &int32Comparer{}
	case int64:
		return &int64Comparer{}
	case uint:
		return &uintComparer{}
	case uint8:
		return &uint8Comparer{}
	case uint16:
		return &uint16Comparer{}
	case uint32:
		return &uint32Comparer{}
	case uint64:
		return &uint64Comparer{}
	case string:
		return &stringComparer{}
	case []byte:
		return &byteSliceComparer{}
	}

	// Fallback to reflection-based comparer otherwise.
	// This is used when caller wraps a type around a primitive type.
	switch reflect.TypeOf(key).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &reflectIntComparer{}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &reflectUintComparer{}
	case reflect.String:
		return &reflectStringComparer{}
	}

	// If no comparers match then panic.
	// This is a compile time issue so it should not return an error.
	panic(fmt.Sprintf("immutable.NewComparer: must set comparer for %T type", key))
}

// intComparer compares two integers. Implements Comparer.
type intComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int.
func (c *intComparer) Compare(a, b interface{}) int {
	if i, j := a.(int), b.(int); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// int8Comparer compares two int8 values. Implements Comparer.
type int8Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int8.
func (c *int8Comparer) Compare(a, b interface{}) int {
	if i, j := a.(int8), b.(int8); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// int16Comparer compares two int16 values. Implements Comparer.
type int16Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int16.
func (c *int16Comparer) Compare(a, b interface{}) int {
	if i, j := a.(int16), b.(int16); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// int32Comparer compares two int32 values. Implements Comparer.
type int32Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int32.
func (c *int32Comparer) Compare(a, b interface{}) int {
	if i, j := a.(int32), b.(int32); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// int64Comparer compares two int64 values. Implements Comparer.
type int64Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int64.
func (c *int64Comparer) Compare(a, b interface{}) int {
	if i, j := a.(int64), b.(int64); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// uintComparer compares two uint values. Implements Comparer.
type uintComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an uint.
func (c *uintComparer) Compare(a, b interface{}) int {
	if i, j := a.(uint), b.(uint); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// uint8Comparer compares two uint8 values. Implements Comparer.
type uint8Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an uint8.
func (c *uint8Comparer) Compare(a, b interface{}) int {
	if i, j := a.(uint8), b.(uint8); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// uint16Comparer compares two uint16 values. Implements Comparer.
type uint16Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an uint16.
func (c *uint16Comparer) Compare(a, b interface{}) int {
	if i, j := a.(uint16), b.(uint16); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// uint32Comparer compares two uint32 values. Implements Comparer.
type uint32Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an uint32.
func (c *uint32Comparer) Compare(a, b interface{}) int {
	if i, j := a.(uint32), b.(uint32); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// uint64Comparer compares two uint64 values. Implements Comparer.
type uint64Comparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an uint64.
func (c *uint64Comparer) Compare(a, b interface{}) int {
	if i, j := a.(uint64), b.(uint64); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// stringComparer compares two strings. Implements Comparer.
type stringComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not a string.
func (c *stringComparer) Compare(a, b interface{}) int {
	return strings.Compare(a.(string), b.(string))
}

// byteSliceComparer compares two byte slices. Implements Comparer.
type byteSliceComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not a byte slice.
func (c *byteSliceComparer) Compare(a, b interface{}) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}

// reflectIntComparer compares two int values using reflection. Implements Comparer.
type reflectIntComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int.
func (c *reflectIntComparer) Compare(a, b interface{}) int {
	if i, j := reflect.ValueOf(a).Int(), reflect.ValueOf(b).Int(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// reflectUintComparer compares two uint values using reflection. Implements Comparer.
type reflectUintComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int.
func (c *reflectUintComparer) Compare(a, b interface{}) int {
	if i, j := reflect.ValueOf(a).Uint(), reflect.ValueOf(b).Uint(); i < j {
		return -1
	} else if i > j {
		return 1
	}
	return 0
}

// reflectStringComparer compares two string values using reflection. Implements Comparer.
type reflectStringComparer struct{}

// Compare returns -1 if a is less than b, returns 1 if a is greater than b, and
// returns 0 if a is equal to b. Panic if a or b is not an int.
func (c *reflectStringComparer) Compare(a, b interface{}) int {
	return strings.Compare(reflect.ValueOf(a).String(), reflect.ValueOf(b).String())
}

func assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
