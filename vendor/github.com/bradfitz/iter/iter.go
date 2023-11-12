// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package iter provides a syntactically different way to iterate over
// integers. That's it.
//
// This package was intended to be an educational joke when it was
// released in 2014. People didn't get the joke part and started
// depending on it. That's fine, I guess. (This is the Internet.) But
// it's kinda weird. It's one line, and not even idiomatic Go style. I
// encourage you not to depend on this or write code like this, but I
// do encourage you to read the code and think about the
// representation of Go slices and why it doesn't allocate.
package iter

// N returns a slice of n 0-sized elements, suitable for ranging over.
//
// For example:
//
//    for i := range iter.N(10) {
//        fmt.Println(i)
//    }
//
// ... will print 0 to 9, inclusive.
//
// It does not cause any allocations.
func N(n int) []struct{} {
	return make([]struct{}, n)
}
