# go-multicodec

[![Tag](https://img.shields.io/github/tag/wealdtech/go-multicodec.svg)](https://github.com/wealdtech/go-multicodec/releases/)
[![License](https://img.shields.io/github/license/wealdtech/go-multicodec.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/go-multicodec?status.svg)](https://godoc.org/github.com/wealdtech/go-multicodec)
[![Travis CI](https://img.shields.io/travis/wealdtech/go-multicodec.svg)](https://travis-ci.org/wealdtech/go-multicodec)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/go-multicodec.svg)](https://codecov.io/github/wealdtech/go-multicodec)

Go utility library to provide encoding and decoding of [multicodec](https://github.com/multiformats/multicodec) values.


## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`go-multicodec` is a standard Go module which can be installed with:

```sh
go get github.com/wealdtech/go-multicodec
```

## Usage

### Example

```go
import (
    "bytes"
    "encoding/hex"
    "errors"

    multicodec "github.com/wealdtech/go-multicodec"
)

func main() {
    // Data in this case is an IPFS hash in multihash format with a dag-pb content type
    data, err := hex.DecodeString("70122029f2d17be6139079dc48696d1f582a8530eb9805b561eda517e22a892c7e3f1f")
    if err != nil {
        panic(err)
    }

    // Add the "ipfs-ns" namespace codec
    dataWithCodec, err := multicodec.AddCodec("ipfs-ns", data)
    if err != nil {
        panic(err)
    }

    if !multicodec.IsCodec("ipfs-ns", dataWithCodec) {
        panic(errors.New("data does not have correct codec prefix"))
    }

    // Remove the codec
    dataWithoutCodec, _, err := multicodec.RemoveCodec(dataWithCodec)
    if err != nil {
        panic(err)
    }

    if !bytes.Equal(data, dataWithoutCodec) {
        panic(errors.New("data mismatch"))
    }
}
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/go-multicodec/issues).

## License

[Apache-2.0](LICENSE) © 2019 Weald Technology Trading Ltd
