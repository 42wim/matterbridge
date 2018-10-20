HTML Strip tags
=====================

[![Build Status][build-status-svg]][build-status-link]
[![Docs][docs-godoc-svg]][docs-godoc-link]
[![Go Report Card][goreport-svg]][goreport-link]
[![License][license-svg]][license-link]

This is a Go package which strip HTML tags from a string. Also, you can provide an array of `allowableTags` that can be
skipped.
Strip HTML tags library is very useful if you work with web crawlers, or just want to strip all or specific tags from
a string.

```go
nodes, err := Strip(content string, allowableTags []string, stripInlineAttributes bool) (Nodes, error)
nodes.Elements //HTML nodes structure of type *html.Node
nodes.ToString() //returns stripped HTML string
```

## Installation

```bash
$ go get github.com/darkoatanasovski/htmltags
``` 

## Parameters

```go
input                   - string
allowableTags           - []string{} //array of strings e.g. []string{"p", "span"}
removeInlineAttributes  - bool // true/false
```

## Return values

Returns `node` structure. You can get the stripped string with `nodes.ToString()`. If there are errors, it will return
the first error message

## Usage

If you want to keep the inline attributes of the tags, set the third parameter to `false`
```go
stripped, err := htmltags.Strip("<h1>Header text with <span style=\"color:red\">color</span></h1>", []string{"span"}, false)
```

Or if you want to strip all tags from the string, and get a pure text, the second parameter has to be
empty array

```go
stripped, err := htmltags.Strip("<h1>Header text with <span style=\"color:red\">color</span></h1>", []string{}, false)
```

A working example
```go
package main

import(
    "fmt"
    "github.com/darkoatanasovski/htmltags"
)

func main() {
    original := "<div>This is <strong style=\"font-size:50px\">complex</strong> text with <span>children <i>nodes</i></span></div>"
    allowableTags := []string{"strong", "i"}
    removeInlineAttributes := false
    stripped, _ := htmltags.Strip(original, allowableTags, removeInlineAttributes)
    
    fmt.Println(stripped) //output: Node structure
    fmt.Println(stripped.ToString()) //output string: This is <strong>complex</strong> text with children <i>nodes</i>
}
```

## Development
If you have cloned this repo you will probably need the dependency:

`go get golang.org/x/net/html`

## Notes
> The broken or partial html will be fixed. If your input HTML string is `<p>Content <i>italic`, 
> the fixed string will be `<p>Content <i>italic</i></i>` 


[build-status-svg]: https://api.travis-ci.org/darkoatanasovski/htmltags.svg?branch=master
[build-status-link]: https://travis-ci.org/darkoatanasovski/htmltags
[docs-godoc-svg]: https://img.shields.io/badge/docs-godoc-blue.svg
[docs-godoc-link]: https://godoc.org/github.com/darkoatanasovski/htmltags
[goreport-svg]: https://goreportcard.com/badge/github.com/darkoatanasovski/htmltags
[goreport-link]: https://goreportcard.com/report/github.com/darkoatanasovski/htmltags
[license-svg]: https://img.shields.io/badge/license-BSD--style+patent--grant-blue.svg
[license-link]: https://github.com/darkoatanasovski/htmltags/blob/master/LICENSE