package libtgsconverter

import "image/color"

type colorAxis uint8

// Color axis constants
const (
	red colorAxis = iota
	green
	blue
)

type colorPriority struct {
	p uint32
	color.RGBA
}

func (c colorPriority) axis(span colorAxis) uint8 {
	switch span {
	case red:
		return c.R
	case green:
		return c.G
	default:
		return c.B
	}
}

type colorBucket []colorPriority

func (cb colorBucket) partition() (colorBucket, colorBucket) {
	mean, span := cb.span()
	left, right := 0, len(cb)-1
	for left < right {
		cb[left], cb[right] = cb[right], cb[left]
		for cb[left].axis(span) < mean && left < right {
			left++
		}
		for cb[right].axis(span) >= mean && left < right {
			right--
		}
	}
	if left == 0 {
		return cb[:1], cb[1:]
	}
	if left == len(cb)-1 {
		return cb[:len(cb)-1], cb[len(cb)-1:]
	}
	return cb[:left], cb[left:]
}

func (cb colorBucket) mean() color.RGBA {
	var r, g, b uint64
	var p uint64
	for _, c := range cb {
		p += uint64(c.p)
		r += uint64(c.R) * uint64(c.p)
		g += uint64(c.G) * uint64(c.p)
		b += uint64(c.B) * uint64(c.p)
	}
	return color.RGBA{uint8(r / p), uint8(g / p), uint8(b / p), 255}
}

type constraint struct {
	min  uint8
	max  uint8
	vals [256]uint64
}

func (c *constraint) update(index uint8, p uint32) {
	if index < c.min {
		c.min = index
	}
	if index > c.max {
		c.max = index
	}
	c.vals[index] += uint64(p)
}

func (c *constraint) span() uint8 {
	return c.max - c.min
}

func (cb colorBucket) span() (uint8, colorAxis) {
	var R, G, B constraint
	R.min = 255
	G.min = 255
	B.min = 255
	var p uint64
	for _, c := range cb {
		R.update(c.R, c.p)
		G.update(c.G, c.p)
		B.update(c.B, c.p)
		p += uint64(c.p)
	}
	var toCount *constraint
	var span colorAxis
	if R.span() > G.span() && R.span() > B.span() {
		span = red
		toCount = &R
	} else if G.span() > B.span() {
		span = green
		toCount = &G
	} else {
		span = blue
		toCount = &B
	}
	var counted uint64
	var i int
	var c uint64
	for i, c = range toCount.vals {
		if counted > p/2 || counted+c == p {
			break
		}
		counted += c
	}
	return uint8(i), span
}
