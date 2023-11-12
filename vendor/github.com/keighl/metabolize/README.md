# Metabolize

[![Build Status](https://travis-ci.org/keighl/metabolize.png?branch=master)](https://travis-ci.org/keighl/metabolize) [![Coverage Status](https://coveralls.io/repos/keighl/metabolize/badge.svg)](https://coveralls.io/r/keighl/metabolize)

Decodes HTML <meta> values into a Golang struct. Great for quickly grabbing [open graph](http://ogp.me/) data.

### Installation

    go get -u github.com/keighl/metabolize

### Usage

Use `meta:"xxx"` tags on your struct to tell metabolize how to decode metadata from an HTML document.

```go
type MetaData struct {
    Title string `meta:"og:title"`
    // If no `og:description`, will fall back to `description`
    Description string `meta:"og:description,description"`
}
```

Example

```go
package main

import (
    "fmt"
    m "github.com/keighl/metabolize"
    "net/http"
    "net/url"
)

type MetaData struct {
    Title       string  `meta:"og:title"`
    Description string  `meta:"og:description,description"`
    Type        string  `meta:"og:type"`
    URL         url.URL `meta:"og:url"`
    VideoWidth  int64   `meta:"og:video:width"`
    VideoHeight int64   `meta:"og:video:height"`
}

func main() {
    res, _ := http.Get("https://www.youtube.com/watch?v=FzRH3iTQPrk")

    data := new(MetaData)

    err := m.Metabolize(res.Body, data)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Title: %s\n", data.Title)
    fmt.Printf("Description: %s\n", data.Description)
    fmt.Printf("Type: %s\n", data.Type)
    fmt.Printf("URL: %s\n", data.URL.String())
    fmt.Printf("VideoWidth: %d\n", data.VideoWidth)
    fmt.Printf("VideoHeight: %d\n", data.VideoHeight)
}
```

Outputs:

```
Title: The Sneezing Baby Panda
Description: A Baby Panda Sneezing Original footage taken and being used with kind permission of LJM Productions Pty. Ltd.,/Wild Candy Pty. Ltd. Authentic t-shirts http:/...
Type: video
URL: http://www.youtube.com/watch?v=FzRH3iTQPrk
VideoWidth: 480
VideoHeight: 360
```

### Supported types

* `string`
* `bool`
* `float64`
* `int64`
* `time.Time`
* `url.URL`

