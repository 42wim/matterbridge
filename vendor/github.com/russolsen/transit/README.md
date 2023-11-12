# transit (go)

[![GoDoc](https://godoc.org/github.com/russolsen/transit?status.svg)](https://godoc.org/github.com/russolsen/transit)

Transit is a data format and a set of libraries for conveying values between applications written in different languages. This library provides support for marshalling Transit data to/from Go.

* [Rationale](http://blog.cognitect.com/blog/2014/7/22/transit)
* [Specification](http://github.com/cognitect/transit-format)

This implementation's major.minor version number corresponds to the version of the Transit specification it supports.

Currently on the JSON formats are implemented.
MessagePack is **not** implemented yet.

_NOTE: Transit is a work in progress and may evolve based on feedback. As a result, while Transit is a great option for transferring data between applications, it should not yet be used for storing data durably over time. This recommendation will change when the specification is complete._

## Usage

Reading data with Transit(go) involves creating a `transit.Decoder` and calling `Decode`:

```go
package main

import (
	"fmt"
	"os"
	"github.com/russolsen/transit"
)

func ReadTransit(path string) interface{} {
	f, err := os.Open(path)

	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}

	decoder := transit.NewDecoder(f)

	value, err := decoder.Decode()

	if err != nil {
		fmt.Printf("Error reading Transit data: %v\n", err)
		return nil
	}

	fmt.Printf("The value read is: %v\n", value)

	return value
}
```

Writing is similar:

```go
func WriteTransit(path string, value interface{}) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)

	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	encoder := transit.NewEncoder(f, false)

	err = encoder.Encode(value)

	if err != nil {
		fmt.Printf("Error writing Transit data: %v\n", err)
		return
	}
}
```


## Default Type Mapping

| Semantic Type | write accepts | read produces |
|:--------------|:--------------|:--------------|
| null| nil | nil |
| string| string | string |
| boolean | bool| bool |
| integer, signed 64 bit| any signed or unsiged int type | int64 |
| floating pt decimal| float32 or float64 | float64 |
| bytes| []byte | []byte |
| keyword | transit.Keyword | transit.Keyword |
| symbol | transit.Symbol | transit.Keyword
| arbitrary precision decimal| big.Float or github.com/shopspring/decimal.Decimal| github.com/shopspring/decimal.Decimal |
| arbitrary precision integer| big.Int | big.Int |
| point in time | time.Time | time.Time |
| point in time RFC 33339 | - | time.Time |
| u| github.com/pborman/uuid UUID| github.com/pborman/uuid UUID|
| uri | net/url URL | net/url URL |
| char | rune | rune |
| special numbers | As defined by math NaN and math.Inf() | TBD
| array | arrays or slices | []interface{} |
| map | map[interface{}]interface{} | map[interface{}]interface{} |
| set |  transit.Set | transit.Set |
| list | container/list List | container/list List |
| map w/ composite keys |  transit.CMap |  transit.CMap |
| link | transit.Link | transit.Link |
| ratio | big.Rat | big.Rat |


## Copyright and License
Copyright © 2016 Russ Olsen

This library is a Go port of the Java version created and maintained by Cognitect, therefore

Copyright © 2014 Cognitect

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

This README file is based on the README from transit-csharp, therefore:

Copyright © 2014 NForza.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.


