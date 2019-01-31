# :book: emojilib

[![Build Status](https://travis-ci.org/peterhellberg/emojilib.svg?branch=master)](https://travis-ci.org/peterhellberg/emojilib)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/peterhellberg/emojilib)
[![License MIT](https://img.shields.io/badge/license-MIT-lightgrey.svg?style=flat)](https://github.com/peterhellberg/emojilib#license-mit)

The [Emoji keyword library](https://github.com/muan/emojilib) by [@muan](https://github.com/muan/) ported to Go. (using `go generate`)

## Installation

    go get -u github.com/peterhellberg/emojilib

## Usage

```go
package main

import (
	"fmt"

	"github.com/peterhellberg/emojilib"
)

func main() {
	fmt.Println(emojilib.ReplaceWithPadding("I :green_heart: You!"))
}
```

## Generating a new version

```bash
$ go generate
```

This will download the latest version of [emojis.json](https://raw.githubusercontent.com/muan/emojilib/master/emojis.json)
and generate a new version of `generated.go`

_You’ll need to have the [golang.org/x/tools/imports](https://golang.org/x/tools/imports) package installed in order to run the generator._

## License (MIT)

Copyright (c) 2015-2019 [Peter Hellberg](https://c7.se)

> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:

> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.

> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
