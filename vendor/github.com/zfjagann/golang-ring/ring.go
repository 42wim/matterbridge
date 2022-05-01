/*
Package ring provides a simple implementation of a ring buffer.
*/
package ring

import "sync"

/*
The DefaultCapacity of an uninitialized Ring buffer.

Changing this value only affects ring buffers created after it is changed.
*/
var DefaultCapacity int = 10

/*
Type Ring implements a Circular Buffer.
The default value of the Ring struct is a valid (empty) Ring buffer with capacity DefaultCapacify.
*/
type Ring struct {
	sync.Mutex
	head int // the most recent value written
	tail int // the least recent value written
	buff []interface{}
}

/*
Set the maximum size of the ring buffer.
*/
func (r *Ring) SetCapacity(size int) {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	r.extend(size)
}

/*
Capacity returns the current capacity of the ring buffer.
*/
func (r *Ring) Capacity() int {
	r.Lock()
	defer r.Unlock()

	return r.capacity()
}

/*
ContentSize returns the current number of elements inside the ring buffer.
*/
func (r *Ring) ContentSize() int {
	r.Lock()
	defer r.Unlock()

	if r.head == -1 {
		return 0
	} else {
		difference := (r.head - r.tail)
		if difference < 0 {
			difference += r.capacity()
		}
		return difference + 1
	}
}

/*
Enqueue a value into the Ring buffer.
*/
func (r *Ring) Enqueue(i interface{}) {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	r.set(r.head+1, i)
	old := r.head
	r.head = r.mod(r.head + 1)
	if old != -1 && r.head == r.tail {
		r.tail = r.mod(r.tail + 1)
	}
}

/*
Dequeue a value from the Ring buffer.

Returns nil if the ring buffer is empty.
*/
func (r *Ring) Dequeue() interface{} {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	if r.head == -1 {
		return nil
	}
	v := r.get(r.tail)
	r.set(r.tail, nil)
	if r.tail == r.head {
		r.head = -1
		r.tail = 0
	} else {
		r.tail = r.mod(r.tail + 1)
	}
	return v
}

/*
Read the value that Dequeue would have dequeued without actually dequeuing it.

Returns nil if the ring buffer is empty.
*/
func (r *Ring) Peek() interface{} {
	r.Lock()
	defer r.Unlock()

	r.checkInit()
	if r.head == -1 {
		return nil
	}
	return r.get(r.tail)
}

/*
Values returns a slice of all the values in the circular buffer without modifying them at all.
The returned slice can be modified independently of the circular buffer. However, the values inside the slice
are shared between the slice and circular buffer.
*/
func (r *Ring) Values() []interface{} {
	r.Lock()
	defer r.Unlock()

	if r.head == -1 {
		return []interface{}{}
	}
	arr := make([]interface{}, 0, r.capacity())
	for i := 0; i < r.capacity(); i++ {
		idx := r.mod(i + r.tail)
		arr = append(arr, r.get(idx))
		if idx == r.head {
			break
		}
	}
	return arr
}

/**
*** Unexported methods beyond this point.
**/

func (r *Ring) capacity() int {
	return len(r.buff)
}

// sets a value at the given unmodified index and returns the modified index of the value
func (r *Ring) set(p int, v interface{}) {
	r.buff[r.mod(p)] = v
}

// gets a value based at a given unmodified index
func (r *Ring) get(p int) interface{} {
	return r.buff[r.mod(p)]
}

// returns the modified index of an unmodified index
func (r *Ring) mod(p int) int {
	v := p % len(r.buff)
	for v < 0 { // this bit fixes negative indices
		v += len(r.buff)
	}
	return v
}

func (r *Ring) checkInit() {
	if r.buff != nil {
		return
	}

	r.buff = make([]interface{}, DefaultCapacity)
	for i := range r.buff {
		r.buff[i] = nil
	}
	r.head, r.tail = -1, 0
}

func (r *Ring) extend(size int) {
	if size == len(r.buff) {
		return
	}

	if size < len(r.buff) {
		// shrink the buffer
		if r.head == -1 {
			// nothing in the buffer, so just shrink it directly
			r.buff = r.buff[0:size]
		} else {
			newb := make([]interface{}, 0, size)
			// buffer has stuff in it, so save the most recent stuff...
			// start at HEAD-SIZE-1 and walk forwards
			for i := size - 1; i >= 0; i-- {
				idx := r.mod(r.head - i)
				newb = append(newb, r.buff[idx])
			}
			// reset head and tail to proper values
			r.head = len(newb) - 1
			r.tail = 0
			r.buff = newb
		}
		return
	}

	// grow the buffer
	newb := make([]interface{}, size-len(r.buff))
	for i := range newb {
		newb[i] = nil
	}
	if r.head == -1 {
		// nothing in the buffer
		r.buff = append(r.buff, newb...)
	} else if r.head >= r.tail {
		// growing at the end is safe
		r.buff = append(r.buff, newb...)
	} else {
		// buffer has stuff that wraps around the end
		// have to rearrange the buffer so the contents are still in order
		part1 := make([]interface{}, len(r.buff[:r.head+1]))
		copy(part1, r.buff[:r.head+1])
		part2 := make([]interface{}, len(r.buff[r.tail:]))
		copy(part2, r.buff[r.tail:])
		r.buff = append(r.buff, newb...)
		newTail := r.mod(r.tail + len(newb))
		r.tail = newTail
		copy(r.buff[:r.head+1], part1)
		copy(r.buff[r.head+1:r.tail], newb)
		copy(r.buff[r.tail:], part2)

	}
}
