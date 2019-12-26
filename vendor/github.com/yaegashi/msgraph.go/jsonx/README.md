# jsonx

It's modified version of [encoding/json](https://golang.org/pkg/encoding/json/)
which enables extra map field (with `jsonx:"true"` tag) to catch all other fields not declared in the struct.

`jsonx` is a code name for internal use
and not related to [JSONx](https://tools.ietf.org/html/draft-rsalz-jsonx-00).

Example ([Run on playgroud](https://play.golang.org/p/TZi0JeHYG69))
```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/yaegashi/msgraph.go/jsonx"
)

type Extra struct {
	X     string
	Y     int
	Extra map[string]interface{} `json:"-" jsonx:"true"`
}

func main() {
	var x1, x2 Extra
	b := []byte(`{"X":"123","Y":123,"A":"123","B":123}`)
	fmt.Printf("\nUnmarshal input: %s\n", string(b))
	json.Unmarshal(b, &x1)
	fmt.Printf(" json.Unmarshal: %#v\n", x1)
	jsonx.Unmarshal(b, &x2)
	fmt.Printf("jsonx.Unmarshal: %#v\n", x2)

	x := Extra{X: "456", Y: 456, Extra: map[string]interface{}{"A": "456", "B": 456}}
	fmt.Printf("\nMarshal input: %#v\n", x)
	b1, _ := json.Marshal(x)
	fmt.Printf(" json.Marshal: %s\n", string(b1))
	b2, _ := jsonx.Marshal(x)
	fmt.Printf("jsonx.Marshal: %s\n", string(b2))
}
```

Result

```text
Unmarshal input: {"X":"123","Y":123,"A":"123","B":123}
 json.Unmarshal: main.Extra{X:"123", Y:123, Extra:map[string]interface {}(nil)}
jsonx.Unmarshal: main.Extra{X:"123", Y:123, Extra:map[string]interface {}{"A":"123", "B":123}}

Marshal input: main.Extra{X:"456", Y:456, Extra:map[string]interface {}{"A":"456", "B":456}}
 json.Marshal: {"X":"456","Y":456}
jsonx.Marshal: {"X":"456","Y":456,"A":"456","B":456}
```

