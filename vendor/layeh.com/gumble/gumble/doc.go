// Package gumble is a client for the Mumble voice chat software.
//
// Getting started
//
// 1. Create a new Config to hold your connection settings:
//
//        config := gumble.NewConfig()
//        config.Username = "gumble-test"
//
// 2. Attach event listeners to the configuration:
//
//        config.Attach(gumbleutil.Listener{
//            TextMessage: func(e *gumble.TextMessageEvent) {
//                fmt.Printf("Received text message: %s\n", e.Message)
//            },
//        })
//
// 3. Connect to the server:
//
//        client, err := gumble.Dial("example.com:64738", config)
//        if err != nil {
//            panic(err)
//        }
//
// Audio codecs
//
// Currently, only the Opus codec (https://www.opus-codec.org/) is supported
// for transmitting and receiving audio. It can be enabled by importing the
// following package for its side effect:
//  import (
//      _ "layeh.com/gumble/opus"
//  )
//
// To ensure that gumble clients can always transmit and receive audio to and
// from your server, add the following line to your murmur configuration file:
//
//  opusthreshold=0
//
// Thread safety
//
// As a general rule, a Client everything that is associated with it
// (Users, Channels, Config, etc.), is thread-unsafe. Accessing or modifying
// those structures should only be done from inside of an event listener or via
// Client.Do.
package gumble
