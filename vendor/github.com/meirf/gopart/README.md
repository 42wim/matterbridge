# Go Type-Agnostic Collection Partitioning

[![GoDoc](https://godoc.org/github.com/meirf/gopart?status.png)](https://godoc.org/github.com/meirf/gopart) [![Travis](https://travis-ci.org/meirf/gopart.svg?branch=master)](https://travis-ci.org/meirf/gopart)

Type-agnostic partitioning for anything that can be indexed in Go - slices, arrays,`string`s. Inspired by Guava's `Lists.partition`. This tiny library alleviates the issue of partitioning collections with wide ranging types - Go lacks generics - by returning consecutive index ranges that can be used on any indexable object.

## Usage

```go
    ...
    // bigList can be any type
    for idxRange := range gopart.Partition(len(bigList), partitionSize) {
        bulkOperation(bigList[idxRange.Low:idxRange.High])
    }
    ...
```
[Full Executable Example](http://play.golang.org/p/WlVPpejxFV)

## Installation

    # install the library:
    go get github.com/meirf/gopart
    
    // use in your .go code:
    import (
        "github.com/meirf/gopart"
    )

## Implementation

The partitioning is done with a separate goroutine that passes the index ranges to a channel. This requires the use of a for...range loop, but adds concurrency and lowers memory usage (no slice of index ranges is stored anywhere).
