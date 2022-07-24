# String globbing in Go

[![GoDoc](https://godoc.org/github.com/zyedidia/glob?status.svg)](http://godoc.org/github.com/zyedidia/glob)

This package adds support for globs in Go.

It simply converts glob expressions to regexps. I try to follow the standard defined [here](http://pubs.opengroup.org/onlinepubs/009695399/utilities/xcu_chap02.html#tag_02_13).

# Example

```go
package main

import "github.com/zyedidia/glob"

func main() {
    glob, err := glob.Compile("{*.go,*.c}")
    if err != nil {
        // Error
    }

    glob.Match([]byte("test.c"))   // true
    glob.Match([]byte("hello.go")) // true
    glob.Match([]byte("test.d"))   // false
}
```

You can call all the same functions on a glob that you can call on a regexp.
