// Copyright 2018 kts of kettek / Ketchetwahmeegwun Tecumseh Southall. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package apng

type APNG struct {
	Frames []Frame
	// LoopCount defines the number of times an animation will be
	// restarted during display.
	// A LoopCount of 0 means to loop forever
	LoopCount uint
}
