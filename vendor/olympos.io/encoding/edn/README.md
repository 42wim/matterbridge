# Go implementation of EDN, extensible data notation

[![GoDoc](https://godoc.org/olympos.io/encoding/edn?status.svg)](https://godoc.org/olympos.io/encoding/edn)

go-edn is a Golang library to read and write
[EDN](https://github.com/edn-format/edn) (extensible data notation), a subset of
Clojure used for transferring data between applications, much like JSON or XML.
EDN is also a very good language for configuration files, much like a JSON-like
version of YAML.

This library is heavily influenced by the JSON library that ships with Go, and
people familiar with that package should know the basics of how this library
works. In fact, this should be close to a drop-in replacement for the
`encoding/json` package if you only use basic functionality.

This implementation is complete, stable, and presumably also bug free. This
is why you don't see any changes in the repository.

If you wonder why you should (or should not) use EDN, you can have a look at the
[why](docs/why.md) document.

## Installation and Usage

The import path for the package is `olympos.io/encoding/edn`

To install it, run:

```shell
go get olympos.io/encoding/edn
```

To use it in your project, you import `olympos.io/encoding/edn` and refer to it as `edn`
like this:

```go
import "olympos.io/encoding/edn"

//...

edn.DoStuff()
```

The previous import path of this library was `gopkg.in/edn.v1`, which is still
permanently supported.

## Quickstart

You can follow http://blog.golang.org/json-and-go and replace every occurence of
JSON with EDN (and the JSON data with EDN data), and the text makes almost
perfect sense. The only caveat is that, since EDN is more general than JSON, go-edn
stores arbitrary maps on the form `map[interface{}]interface{}`.

go-edn also ships with keywords, symbols and tags as types.

For a longer introduction on how to use the library, see
[introduction.md](docs/introduction.md). If you're familiar with the JSON
package, then the [API Documentation](https://godoc.org/olympos.io/encoding/edn) might
be the only thing you need.

## Example Usage

Say you want to describe your pet forum's users as EDN. They have the following
types:

```go
type Animal struct {
	Name string
	Type string `edn:"kind"`
}

type Person struct {
	Name      string
	Birthyear int `edn:"born"`
	Pets      []Animal
}
```

With go-edn, we can do as follows to read and write these types:

```go
import "olympos.io/encoding/edn"

//...


func ReturnData() (Person, error) {
	data := `{:name "Hans",
              :born 1970,
              :pets [{:name "Cap'n Jack" :kind "Sparrow"}
                     {:name "Freddy" :kind "Cockatiel"}]}`
	var user Person
	err := edn.Unmarshal([]byte(data), &user)
	// user '==' Person{"Hans", 1970,
	//             []Animal{{"Cap'n Jack", "Sparrow"}, {"Freddy", "Cockatiel"}}}
	return user, err
}
```

If you want to write that user again, just `Marshal` it:

```go
bs, err := edn.Marshal(user)
```

## Dependencies

go-edn has no external dependencies, except the default Go library. However, as
it depends on `math/big.Float`, go-edn requires Go 1.5 or higher.


## License

Copyright © 2015-2019 Jean Niklas L'orange and [contributors](https://github.com/go-edn/edn/graphs/contributors)

Distributed under the BSD 3-clause license, which is available in the file
LICENSE.
