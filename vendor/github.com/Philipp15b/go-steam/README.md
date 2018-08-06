# Steam for Go

This library implements Steam's protocol to allow automation of different actions on Steam without running an actual Steam client. It is based on [SteamKit2](https://github.com/SteamRE/SteamKit), a .NET library.

In addition, it contains APIs to Steam Community features, like trade offers and inventories.

Some of the currently implemented features:

  * Trading and trade offers, including inventories and notifications
  * Friend and group management
  * Chatting with friends
  * Persona states (online, offline, looking to trade, etc.)
  * SteamGuard with two-factor authentication
  * Team Fortress 2: Crafting, moving, naming and deleting items

If this is useful to you, there's also the [go-steamapi](https://github.com/Philipp15b/go-steamapi) package that wraps some of the official Steam Web API's types.

## Installation

    go get github.com/Philipp15b/go-steam

## Usage

You can view the documentation with the [`godoc`](http://golang.org/cmd/godoc) tool or
[online on godoc.org](http://godoc.org/github.com/Philipp15b/go-steam).

You should also take a look at the following sub-packages:

  * [`gsbot`](http://godoc.org/github.com/Philipp15b/go-steam/gsbot) utilites that make writing bots easier
  * [example bot](http://godoc.org/github.com/Philipp15b/go-steam/gsbot/gsbot) and [its source code](https://github.com/Philipp15b/go-steam/blob/master/gsbot/gsbot/gsbot.go)
  * [`trade`](http://godoc.org/github.com/Philipp15b/go-steam/trade) for trading
  * [`tradeoffer`](http://godoc.org/github.com/Philipp15b/go-steam/tradeoffer) for trade offers
  * [`economy/inventory`](http://godoc.org/github.com/Philipp15b/go-steam/economy/inventory) for inventories
  * [`tf2`](http://godoc.org/github.com/Philipp15b/go-steam/tf2) for Team Fortress 2 related things

## Working with go-steam

Whether you want to develop your own Steam bot or directly work on go-steam itself, there are are few things to know.

 * If something is not working, check first if the same operation works (under the same conditions!) in the Steam client on that account. Maybe there's something go-steam doesn't handle correctly or you're missing a warning that's not obviously shown in go-steam. This is particularly important when working with trading since there are [restrictions](https://support.steampowered.com/kb_article.php?ref=1047-edfm-2932), for example newly authorized devices will not be able to trade for seven days.
 * Since Steam does not maintain a public API for most of the things go-steam implements, you can expect that sometimes things break randomly. Especially the `trade` and `tradeoffer` packages have been affected in the past.
 * Always gather as much information as possible. When you file an issue, be as precise and complete as you can. This makes debugging way easier.
 * If you haven't noticed yet, expect to find lots of things out yourself. Debugging can be complicated and Steam's internals are too.
 * Sometimes things break and other [SteamKit ports](https://github.com/SteamRE/SteamKit/wiki/Ports) are fixed already. Maybe take a look what people are saying over there? There's also the [SteamKit IRC channel](https://github.com/SteamRE/SteamKit/wiki#contact).

## Updating go-steam to a new SteamKit version

To update go-steam to a new version of SteamKit, do the following:

	go get github.com/golang/protobuf/protoc-gen-go/
    git submodule init && git submodule update
    cd generator
    go run generator.go clean proto steamlang

Make sure that `$GOPATH/bin` / `protoc-gen-go` is in your `$PATH`. You'll also need [`protoc`](https://developers.google.com/protocol-buffers/docs/downloads), the protocol buffer compiler. At the moment, we use Protocol Buffers 2.6.1 with `proco-gen-go`-[2402d76](https://github.com/golang/protobuf/tree/2402d76f3d41f928c7902a765dfc872356dd3aad).

To compile the Steam Language files, you also need the [.NET Framework](https://www.microsoft.com/net/downloads)
on Windows or [mono](http://www.go-mono.com/mono-downloads/download.html) on other operating systems.

Apply the protocol changes where necessary.

## License

Steam for Go is licensed under the New BSD License. More information can be found in LICENSE.txt.