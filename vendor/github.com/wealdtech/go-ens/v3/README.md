# go-ens

[![Tag](https://img.shields.io/github/tag/wealdtech/go-ens.svg)](https://github.com/wealdtech/go-ens/releases/)
[![License](https://img.shields.io/github/license/wealdtech/go-ens.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/go-ens?status.svg)](https://godoc.org/github.com/wealdtech/go-ens)
[![Travis CI](https://img.shields.io/travis/wealdtech/go-ens.svg)](https://travis-ci.org/wealdtech/go-ens)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/go-ens.svg)](https://codecov.io/github/wealdtech/go-ens)
[![Go Report Card](https://goreportcard.com/badge/github.com/wealdtech/go-ens)](https://goreportcard.com/report/github.com/wealdtech/go-ens)

Go module to simplify interacting with the [Ethereum Name Service](https://ens.domains/) contracts.


## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`go-ens` is a standard Go module which can be installed with:

```sh
go get github.com/wealdtech/go-ens
```

## Usage

`go-ens` provides simple access to the [Ethereum Name Service](https://ens.domains/) (ENS) contracts.

### Resolution

The most commonly-used feature of ENS is resolution: converting an ENS name to an Ethereum address.  `go-ens` provides a simple call to allow this:

```go
address, err := ens.Resolve(client, domain)
```

where `client` is a connection to an Ethereum client and `domain` is the fully-qualified name you wish to resolve (e.g. `foo.mydomain.eth`) (full examples for using this are given in the [Example](#Example) section below).

The reverse process, converting an address to an ENS name, is just as simple:

```go
domain, err := ens.ReverseResolve(client, address)
```

Note that if the address does not have a reverse resolution this will return "".  If you just want a string version of an address for on-screen display then you can use `ens.Format()`, for example:

```go
fmt.Printf("The address is %s\n", ens.Format(client, address))
```

This will carry out reverse resolution of the address and print the name if present; if not it will print a formatted version of the address.


### Management of names

A top-level name is one that sits directly underneath `.eth`, for example `mydomain.eth`.  Lower-level names, such as `foo.mydomain.eth` are covered in the following section.  `go-ens` provides a simplified `Name` interface to manage top-level, removing the requirement to understand registrars, controllers, _etc._

Starting out with names in `go-ens` is easy:

```go
client, err := ethclient.Dial("https://infura.io/v3/SECRET")
name, err := ens.NewName(client, "mydomain.eth")
```

Addresses can be set and obtained using the address functions, for example to get an address:

```go
COIN_TYPE_ETHEREUM := uint64(60)
address, err := name.Address(COIN_TYPE_ETHEREUM)
```

ENS supports addresses for multiple coin types; values of coin types can be found at https://github.com/satoshilabs/slips/blob/master/slip-0044.md

### Registering and extending names

Most operations on a domain will involve setting resolvers and resolver information.


### Management of subdomains

Because subdomains have their own registrars they do not work with the `Name` interface.

### Example

```go
package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	ens "github.com/wealdtech/go-ens/v3"
)

func main() {
	// Replace SECRET with your own access token for this example to work.
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/SECRET")
	if err != nil {
		panic(err)
	}

	// Resolve a name to an address
	domain := "avsa.eth"
	address, err := ens.Resolve(client, domain)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Address of %s is %s\n", domain, address.Hex())

	// Reverse resolve an address to a name
	reverse, err := ens.ReverseResolve(client, address)
	if err != nil {
		panic(err)
	}
	if reverse == "" {
		fmt.Printf("%s has no reverse lookup\n", address.Hex())
	} else {
		fmt.Printf("Name of %s is %s\n", address.Hex(), reverse)
	}
}

package main

import (
    "github.com/ethereum/go-ethereum/ethclient"
	ens "github.com/wealdtech/go-ens/v3"
)

func main() {
    client, err := ethclient.Dial("https://mainnet.infura.io/v3/SECRET")
    if err != nil {
        panic(err)
    }

    // Resolve a name to an address
    domain := "wealdtech.eth"
    address, err := ens.Resolve(client, domain)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Address of %s is %s\n", domain, address.Hex())

    // Reverse resolve an address to a name
    reverse, err := ens.ReverseResolve(client, address)
    if err != nil {
        panic(err)
    }
    if reverse == "" {
      fmt.Printf("%s has no reverse lookup\n", address.Hex())
    } else {
      fmt.Printf("Name of %s is %s\n", address.Hex(), reverse)
    }
}
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/go-ens/issues).

## License

[Apache-2.0](LICENSE) © 2019 Weald Technology Trading Ltd
