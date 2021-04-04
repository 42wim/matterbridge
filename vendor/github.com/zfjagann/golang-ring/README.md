# ring
--
    import "github.com/zealws/golang-ring"

Package ring provides a simple implementation of a ring buffer.

## Usage

```go
var DefaultCapacity int = 10
```
The DefaultCapacity of an uninitialized Ring buffer.

Changing this value only affects ring buffers created after it is changed.

#### type Ring

```go
type Ring struct {
	sync.Mutex
}
```

Type Ring implements a Circular Buffer. The default value of the Ring struct is
a valid (empty) Ring buffer with capacity DefaultCapacify.

#### func (*Ring) Capacity

```go
func (r *Ring) Capacity() int
```
Capacity returns the current capacity of the ring buffer.

#### func (*Ring) ContentSize

```go
func (r *Ring) ContentSize() int
```
ContentSize returns the current number of elements inside the ring buffer.

#### func (*Ring) Dequeue

```go
func (r *Ring) Dequeue() interface{}
```
Dequeue a value from the Ring buffer.

Returns nil if the ring buffer is empty.

#### func (*Ring) Enqueue

```go
func (r *Ring) Enqueue(i interface{})
```
Enqueue a value into the Ring buffer.

#### func (*Ring) Peek

```go
func (r *Ring) Peek() interface{}
```
Read the value that Dequeue would have dequeued without actually dequeuing it.

Returns nil if the ring buffer is empty.

#### func (*Ring) SetCapacity

```go
func (r *Ring) SetCapacity(size int)
```
Set the maximum size of the ring buffer.

#### func (*Ring) Values

```go
func (r *Ring) Values() []interface{}
```
Values returns a slice of all the values in the circular buffer without
modifying them at all. The returned slice can be modified independently of the
circular buffer. However, the values inside the slice are shared between the
slice and circular buffer.
