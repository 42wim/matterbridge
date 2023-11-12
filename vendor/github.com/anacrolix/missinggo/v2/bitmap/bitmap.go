// Package bitmap provides a []bool/bitmap implementation with standardized
// iteration. Bitmaps are the equivalent of []bool, with improved compression
// for runs of similar values, and faster operations on ranges and the like.
package bitmap

import (
	"github.com/RoaringBitmap/roaring"

	"github.com/anacrolix/missinggo/iter"
)

type (
	BitIndex = uint32
	BitRange = uint64
)

type Interface interface {
	Len() int
}

// Bitmaps store the existence of values in [0,math.MaxUint32] more
// efficiently than []bool. The empty value starts with no bits set.
type Bitmap struct {
	RB *roaring.Bitmap
}

const (
	MaxInt BitIndex = roaring.MaxUint32
	ToEnd  BitRange = roaring.MaxRange
)

// The number of set bits in the bitmap. Also known as cardinality.
func (me Bitmap) Len() BitRange {
	if me.RB == nil {
		return 0
	}
	return me.RB.GetCardinality()
}

func (me Bitmap) ToSortedSlice() []BitIndex {
	if me.RB == nil {
		return nil
	}
	return me.RB.ToArray()
}

func (me *Bitmap) lazyRB() *roaring.Bitmap {
	if me.RB == nil {
		me.RB = roaring.NewBitmap()
	}
	return me.RB
}

func (me Bitmap) Iter(cb iter.Callback) {
	me.IterTyped(func(i int) bool {
		return cb(i)
	})
}

// Returns true if all values were traversed without early termination.
func (me Bitmap) IterTyped(f func(int) bool) bool {
	if me.RB == nil {
		return true
	}
	it := me.RB.Iterator()
	for it.HasNext() {
		if !f(int(it.Next())) {
			return false
		}
	}
	return true
}

func checkInt(i BitIndex) {
	// Nothing to do if BitIndex is uint32, as this matches what roaring can handle.
}

func (me *Bitmap) Add(is ...BitIndex) {
	rb := me.lazyRB()
	for _, i := range is {
		checkInt(i)
		rb.Add(i)
	}
}

func (me *Bitmap) AddRange(begin, end BitRange) {
	// Filter here so we don't prematurely create a bitmap before having roaring do this check
	// anyway.
	if begin >= end {
		return
	}
	me.lazyRB().AddRange(begin, end)
}

func (me *Bitmap) Remove(i BitIndex) bool {
	if me.RB == nil {
		return false
	}
	return me.RB.CheckedRemove(uint32(i))
}

func (me *Bitmap) Union(other Bitmap) {
	me.lazyRB().Or(other.lazyRB())
}

func (me Bitmap) Contains(i BitIndex) bool {
	if me.RB == nil {
		return false
	}
	return me.RB.Contains(i)
}

func (me *Bitmap) Sub(other Bitmap) {
	if other.RB == nil {
		return
	}
	if me.RB == nil {
		return
	}
	me.RB.AndNot(other.RB)
}

func (me *Bitmap) Clear() {
	if me.RB == nil {
		return
	}
	me.RB.Clear()
}

func (me Bitmap) Copy() (ret Bitmap) {
	ret = me
	if ret.RB != nil {
		ret.RB = ret.RB.Clone()
	}
	return
}

func (me *Bitmap) FlipRange(begin, end BitRange) {
	me.lazyRB().Flip(begin, end)
}

func (me Bitmap) Get(bit BitIndex) bool {
	return me.RB != nil && me.RB.Contains(bit)
}

func (me *Bitmap) Set(bit BitIndex, value bool) {
	if value {
		me.lazyRB().Add(bit)
	} else {
		if me.RB != nil {
			me.RB.Remove(bit)
		}
	}
}

func (me *Bitmap) RemoveRange(begin, end BitRange) *Bitmap {
	if me.RB == nil {
		return me
	}
	me.RB.RemoveRange(begin, end)
	return me
}

func (me Bitmap) IsEmpty() bool {
	return me.RB == nil || me.RB.IsEmpty()
}
