# godown

[![Build Status](https://travis-ci.org/mattn/godown.png?branch=master)](https://travis-ci.org/mattn/godown)
[![Codecov](https://codecov.io/gh/mattn/godown/branch/master/graph/badge.svg)](https://codecov.io/gh/mattn/godown)
[![GoDoc](https://godoc.org/github.com/mattn/godown?status.svg)](http://godoc.org/github.com/mattn/godown)
[![Go Report Card](https://goreportcard.com/badge/github.com/mattn/godown)](https://goreportcard.com/report/github.com/mattn/godown)

Convert HTML into Markdown

This is work in progress.

## Usage

```
err := godown.Convert(w, r)
checkError(err)
```


## Command Line

```
$ godown < index.html > index.md
```

## Installation

```
$ go get github.com/mattn/godown/cmd/godown
```

## TODO

* escape strings in HTML

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
