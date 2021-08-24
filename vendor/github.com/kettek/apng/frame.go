// Copyright 2018 kts of kettek / Ketchetwahmeegwun Tecumseh Southall. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apng

import (
	"image"
)

// dispose_op values, as per the APNG spec.
const (
	DISPOSE_OP_NONE       = 0
	DISPOSE_OP_BACKGROUND = 1
	DISPOSE_OP_PREVIOUS   = 2
)

// blend_op values, as per the APNG spec.
const (
	BLEND_OP_SOURCE = 0
	BLEND_OP_OVER   = 1
)

type Frame struct {
	Image              image.Image
	width, height    int
	XOffset, YOffset int
	DelayNumerator   uint16
	DelayDenominator uint16
	DisposeOp        byte
	BlendOp          byte
	// IsDefault indicates if the Frame is a default image that
	// should not be used in the animation. IsDefault can only
	// be true on the first frame.
	IsDefault bool
}

// GetDelay returns the number of seconds in the frame.
func (f *Frame) GetDelay() float64 {
	d := uint16(0)
	if f.DelayDenominator == 0 {
		d = 100
	} else {
		d = f.DelayDenominator
	}
	return float64(f.DelayNumerator) / float64(d)
}
