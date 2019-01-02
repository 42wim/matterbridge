markdown [![GoDoc](http://godoc.org/gitlab.com/golang-commonmark/markdown?status.svg)](http://godoc.org/gitlab.com/golang-commonmark/markdown) [![License](https://img.shields.io/badge/licence-BSD--2--Clause-blue.svg)](https://opensource.org/licenses/BSD-2-Clause) [![Pipeline status](https://gitlab.com/golang-commonmark/markdown/badges/master/pipeline.svg)](https://gitlab.com/golang-commonmark/markdown/commits/master) [![Coverage report](https://gitlab.com/golang-commonmark/markdown/badges/master/coverage.svg)](https://gitlab.com/golang-commonmark/markdown/commits/master)
========

Package golang-commonmark/markdown provides a CommonMark-compliant markdown parser and renderer, written in Go.

## Installation

    go get -u gitlab.com/golang-commonmark/markdown

You can also go get [mdtool](https://gitlab.com/golang-commonmark/mdtool), an example command-line tool:

    go get -u gitlab.com/golang-commonmark/mdtool

## Standards support

Currently supported CommonMark spec: [v0.28](http://spec.commonmark.org/0.28/).

## Extensions

Besides the features required by CommonMark, golang-commonmark/markdown supports:

  * Tables (GFM)
  * Strikethrough (GFM)
  * Autoconverting plain-text URLs to links
  * Typographic replacements (smart quotes and other)

## Usage

``` go
md := markdown.New(markdown.XHTMLOutput(true))
fmt.Println(md.RenderToString([]byte("Header\n===\nText")))
```

Check out [the source of mdtool](https://gitlab.com/golang-commonmark/mdtool/blob/master/main.go) for a more complete example.

The following options are currently supported:

  Name            |  Type     |                        Description                          | Default
  --------------- | --------- | ----------------------------------------------------------- | ---------
  HTML            | bool      | whether to enable raw HTML                                  | false
  Tables          | bool      | whether to enable GFM tables                                | true
  Linkify         | bool      | whether to autoconvert plain-text URLs to links             | true
  Typographer     | bool      | whether to enable typographic replacements                  | true
  Quotes          | string / []string | double + single quote replacement pairs for the typographer | “”‘’
  MaxNesting      | int       | maximum nesting level                                       | 20
  LangPrefix      | string    | CSS language prefix for fenced blocks                       | language-
  Breaks          | bool      | whether to convert newlines inside paragraphs into `<br>`   | false
  XHTMLOutput     | bool      | whether to output XHTML instead of HTML                     | false

## Benchmarks

Rendering spec/spec-0.28.txt on a Intel(R) Core(TM) i5-2400 CPU @ 3.10GHz

    BenchmarkRenderSpecNoHTML         100    10254720 ns/op    2998037 B/op    18225 allocs/op
    BenchmarkRenderSpec               100    10180241 ns/op    2997307 B/op    18214 allocs/op
    BenchmarkRenderSpecBlackFriday    200     7241749 ns/op    2834340 B/op    17101 allocs/op
    BenchmarkRenderSpecBlackFriday2   200     7448256 ns/op    2991202 B/op    16705 allocs/op

## See also

https://github.com/jgm/CommonMark — the reference CommonMark implementations in C and JavaScript,
  also contains the latest spec and an online demo.

http://talk.commonmark.org — the CommonMark forum, a good place to join together the efforts of the developers.
