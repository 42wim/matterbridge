package torrent

import (
	rbm "github.com/RoaringBitmap/roaring"
	roaring "github.com/RoaringBitmap/roaring/BitSliceIndexing"
)

type pendingRequests struct {
	m *roaring.BSI
}

func (p *pendingRequests) Dec(r RequestIndex) {
	_r := uint64(r)
	prev, _ := p.m.GetValue(_r)
	if prev <= 0 {
		panic(prev)
	}
	p.m.SetValue(_r, prev-1)
}

func (p *pendingRequests) Inc(r RequestIndex) {
	_r := uint64(r)
	prev, _ := p.m.GetValue(_r)
	p.m.SetValue(_r, prev+1)
}

func (p *pendingRequests) Init(maxIndex RequestIndex) {
	p.m = roaring.NewDefaultBSI()
}

var allBits rbm.Bitmap

func init() {
	allBits.AddRange(0, rbm.MaxRange)
}

func (p *pendingRequests) AssertEmpty() {
	if p.m == nil {
		panic(p.m)
	}
	sum, _ := p.m.Sum(&allBits)
	if sum != 0 {
		panic(sum)
	}
}

func (p *pendingRequests) Get(r RequestIndex) int {
	count, _ := p.m.GetValue(uint64(r))
	return int(count)
}
