# Wray, faye client for Go

Wray is a client for the [Faye](http://faye.jcoglan.com) publish-subscribe messaging service by [James Coglan](https://twitter.com/jcoglan).

## Current status

In progress.

## Getting Started

Wray is only a client for Faye. You will need to setup a server using Ruby or Node.js first. Instructions that can be found on the [Faye project pages](http://faye.jcoglan.com/).

###Subscribing to channels

```
package main

import "github.com/pythonandchips/wray"
import "fmt"

func main() {
  //register the types of transport you want available. Only long-polling is currently supported
  wray.RegisterTransports([]gofaye.Transport{ &gofaye.HttpTransport{} })

  //create a new client
  client := wray.NewFayeClient("http://localhost:5000/faye")

  //subscribe to the channels you want to listen to
  client.Subscribe("/foo", false, func(message wray.Message) {
    fmt.Println("-------------------------------------------")
    fmt.Println(message.Data)
  })

  //wildcards can be used to subscribe to multipule channels
  client.Subscribe("/foo/*", false, func(message wray.Message) {
    fmt.Println("-------------------------------------------")
    fmt.Println(message.Data)
  })

  //start listening on all subscribed channels and hold the process open
  client.Listen()
}
```

###Publishing to channels
```
package main

import "github.com/pythonandchips/wray"
import "fmt"

func main() {
  //register the types of transport you want available. Only long-polling is currently supported
  wray.RegisterTransports([]wray.Transport{ &gofaye.HttpTransport{} })

  //create a new client
  client := wray.NewFayeClient("http://localhost:5000/faye")

  params := map[string]interface{}{"hello": "from golang"}

  //send message to server
  client.Publish("/foo", params)
}
```

Simple examples are availabe in the examples folder.

##Future Work
There is still a lot to do to bring Wray in line with Faye functionality. This is a less than exhaustive list of work to be completed:-

- web socket support
- eventsource support
- logging
- middleware additions
- correctly handle disconnect and server down
- promises for subscription and publishing
- automated integrations test to ensure Wray continues to work with Faye

## Bugs/Features/Prase

It you find any bugs or have some feature requests please add an issue on the repository. Or if you just want to get in touch and tell me how awesome wray is you can get me on twitter @colin_gemmell or drop me an email at pythonandchips{at}gmail.com.

## Contributing to Wray

* Check out the latest master to make sure the feature hasn't been implemented or the bug hasn't been fixed yet
* Check out the issue tracker to make sure someone already hasn't requested it and/or contributed it
* Fork the project
* Start a feature/bugfix branch
* Commit and push until you are happy with your contribution
* Make sure to add tests for it. This is important so I don't break it in a future version unintentionally.

## Copyright

Copyright (c) 2014 Colin Gemmell

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
