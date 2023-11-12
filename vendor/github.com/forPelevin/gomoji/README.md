# GoMoji
<p align="center">work with emoji in the most convenient way</p>

GoMoji is a Go package that provides a [fast](#performance) and [simple](#check-string-contains-emoji) way to work with emojis in strings.
It has features such as:
 * [check whether string contains emoji](#check-string-contains-emoji)
 * [find all emojis in string](#find-all)
 * [get all emojis](#get-all) 
 * [remove all emojis from string](#remove-all-emojis)
 * [get emoji description](#get-emoji-info) 

Getting Started
===============

## Installing

To start using GoMoji, install Go and run `go get`:

```sh
$ go get -u github.com/forPelevin/gomoji
```

This will retrieve the package.

## Check string contains emoji
```go
package main

import (
    "github.com/forPelevin/gomoji"
)

func main() {
    res := gomoji.ContainsEmoji("hello world")
    println(res) // false
    
    res = gomoji.ContainsEmoji("hello world 🤗")
    println(res) // true
}
```

## Find all
The function searches for all emoji occurrences in a string. It returns a nil slice if there are no emojis.
```go
package main

import (
    "github.com/forPelevin/gomoji"
)

func main() {
    res := gomoji.FindAll("🧖 hello 🦋 world")
    println(res)
}
```

Result:

```go
[]gomoji.Emoji{
    {
        Slug:        "person-in-steamy-room",
        Character:   "🧖",
        UnicodeName: "E5.0 person in steamy room",
        CodePoint:   "1F9D6",
        Group:       "People & Body",
        SubGroup:    "person-activity",
    },
    {
        Slug:        "butterfly",
        Character:   "🦋",
        UnicodeName: "E3.0 butterfly",
        CodePoint:   "1F98B",
        Group:       "Animals & Nature",
        SubGroup:    "animal-bug",
    },
}
```

## Get all
The function returns all existing emojis. You can do whatever you need with the list.
 ```go
 package main
 
 import (
     "github.com/forPelevin/gomoji"
 )
 
 func main() {
     emojis := gomoji.AllEmojis()
     println(emojis)
 }
 ```

## Remove all emojis

The function removes all emojis from given string:

```go
res := gomoji.RemoveEmojis("🧖 hello 🦋world")
println(res) // "hello world"
```

## Get emoji info

The function returns info about provided emoji:

```go
info, err := gomoji.GetInfo("1") // error: the string is not emoji
info, err := gomoji.GetInfo("1️⃣")
println(info)
```

Result:

```go
gomoji.Entity{
    Slug:        "keycap-1",
    Character:   "1️⃣",
    UnicodeName: "E0.6 keycap: 1",
    CodePoint:   "0031 FE0F 20E3",
    Group:       "Symbols",
    SubGroup:    "keycap",
}
```

## Emoji entity
All searching methods return the Emoji entity which contains comprehensive info about emoji.
```go
type Emoji struct {
    Slug        string `json:"slug"`
    Character   string `json:"character"`
    UnicodeName string `json:"unicode_name"`
    CodePoint   string `json:"code_point"`
    Group       string `json:"group"`
    SubGroup    string `json:"sub_group"`
}
 ```
Example:
```go
[]gomoji.Emoji{
    {
        Slug:        "butterfly",
        Character:   "🦋",
        UnicodeName: "E3.0 butterfly",
        CodePoint:   "1F98B",
        Group:       "Animals & Nature",
        SubGroup:    "animal-bug",
    },
    {
        Slug:        "roll-of-paper",
        Character:   "🧻",
        UnicodeName: "E11.0 roll of paper",
        CodePoint:   "1F9FB",
        Group:       "Objects",
        SubGroup:    "household",
    },
}
 ```

## Performance

GoMoji Benchmarks

```
goos: darwin
goarch: amd64
pkg: github.com/forPelevin/gomoji
cpu: Intel(R) Core(TM) i5-8257U CPU @ 1.40GHz
BenchmarkContainsEmojiParallel
BenchmarkContainsEmojiParallel-8   	 7439398	       159.2 ns/op	     144 B/op	       3 allocs/op
BenchmarkContainsEmoji
BenchmarkContainsEmoji-8           	 2457042	       482.2 ns/op	     144 B/op	       3 allocs/op
BenchmarkRemoveEmojisParallel
BenchmarkRemoveEmojisParallel-8    	 4589841	       265.8 ns/op	     236 B/op	       5 allocs/op
BenchmarkRemoveEmojis
BenchmarkRemoveEmojis-8            	 1456464	       831.9 ns/op	     236 B/op	       5 allocs/op
BenchmarkGetInfoParallel
BenchmarkGetInfoParallel-8         	272416886	         4.433 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetInfo
BenchmarkGetInfo-8                 	64521932	        19.86 ns/op	       0 B/op	       0 allocs/op
BenchmarkFindAllParallel
BenchmarkFindAllParallel-8         	 3989124	       295.9 ns/op	     456 B/op	       5 allocs/op
BenchmarkFindAll
BenchmarkFindAll-8                 	 1304463	       913.7 ns/op	     456 B/op	       5 allocs/op
```

## Contact
Vlad Gukasov [@vgukasov](https://www.facebook.com/vgukasov)

## License

GoMoji source code is available under the MIT [License](/LICENSE).