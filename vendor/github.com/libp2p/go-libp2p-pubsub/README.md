# go-libp2p-pubsub

<p align="left">
  <a href="http://protocol.ai"><img src="https://img.shields.io/badge/made%20by-Protocol%20Labs-blue.svg?style=flat-square" /></a>
  <a href="http://libp2p.io/"><img src="https://img.shields.io/badge/project-libp2p-yellow.svg?style=flat-square" /></a>
  <a href="http://webchat.freenode.net/?channels=%23libp2p"><img src="https://img.shields.io/badge/freenode-%23libp2p-yellow.svg?style=flat-square" /></a>
  <a href="https://discuss.libp2p.io"><img src="https://img.shields.io/discourse/https/discuss.libp2p.io/posts.svg?style=flat-square"/></a>
</p>

<p align="left">
  <a href="https://codecov.io/gh/libp2p/go-libp2p-pubsub"><img src="https://codecov.io/gh/libp2p/go-libp2p-pubsub/branch/master/graph/badge.svg"></a>
  <a href="https://goreportcard.com/report/github.com/libp2p/go-libp2p-pubsub"><img src="https://goreportcard.com/badge/github.com/libp2p/go-libp2p-pubsub" /></a>
  <a href="https://github.com/RichardLitt/standard-readme"><img src="https://img.shields.io/badge/readme%20style-standard-brightgreen.svg?style=flat-square" /></a>
  <a href="https://godoc.org/github.com/libp2p/go-libp2p-pubsub"><img src="http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square" /></a>
  <a href=""><img src="https://img.shields.io/badge/golang-%3E%3D1.14.0-orange.svg?style=flat-square" /></a>
  <br>
</p>

This repo contains the canonical pubsub implementation for libp2p. We currently provide three message router options:
- Floodsub, which is the baseline flooding protocol.
- Randomsub, which is a simple probabilistic router that propagates to random subsets of peers.
- Gossipsub, which is a more advanced router with mesh formation and gossip propagation. See [spec](https://github.com/libp2p/specs/tree/master/pubsub/gossipsub) and  [implementation](https://github.com/libp2p/go-libp2p-pubsub/blob/master/gossipsub.go) for more details.


## Repo Lead Maintainer

[@vyzo](https://github.com/vyzo/)

> This repo follows the [Repo Lead Maintainer Protocol](https://github.com/ipfs/team-mgmt/blob/master/LEAD_MAINTAINER_PROTOCOL.md)

## Table of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Install](#install)
- [Usage](#usage)
- [Example](#example)
- [Documentation](#documentation)
- [Tracing](#tracing)
- [Contribute](#contribute)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Install

```
go get github.com/libp2p/go-libp2p-pubsub
```

## Usage

To be used for messaging in p2p instrastructure (as part of libp2p) such as IPFS, Ethereum, other blockchains, etc.

### Example

https://github.com/libp2p/go-libp2p/tree/master/examples/pubsub

## Documentation

See the [libp2p specs](https://github.com/libp2p/specs/tree/master/pubsub) for high level documentation and [godoc](https://godoc.org/github.com/libp2p/go-libp2p-pubsub) for API documentation.

### In this repo, you will find

```
.
├── LICENSE
├── README.md
# Regular Golang repo set up
├── codecov.yml
├── pb
├── go.mod
├── go.sum
├── doc.go
# PubSub base
├── pubsub.go
├── blacklist.go
├── notify.go
├── comm.go
├── discovery.go
├── sign.go
├── subscription.go
├── topic.go
├── trace.go
├── tracer.go
├── validation.go
# Floodsub router
├── floodsub.go
# Randomsub router
├── randomsub.go
# Gossipsub router
├── gossipsub.go
├── score.go
├── score_params.go
└── mcache.go
```

### Tracing

The pubsub system supports _tracing_, which collects all events pertaining to the internals of the system. This allows you to recreate the complete message flow and state of the system for analysis purposes.

To enable tracing, instantiate the pubsub system using the `WithEventTracer` option; the option accepts a tracer with three available implementations in-package (trace to json, pb, or a remote peer).
If you want to trace using a remote peer, you can do so using the `traced` daemon from [go-libp2p-pubsub-tracer](https://github.com/libp2p/go-libp2p-pubsub-tracer). The package also includes a utility program, `tracestat`, for analyzing the traces collected by the daemon.

For instance, to capture the trace as a json file, you can use the following option:
```go
tracer, err := pubsub.NewJSONTracer("/path/to/trace.json")
if err != nil {
  panic(err)
}

pubsub.NewGossipSub(..., pubsub.WithEventTracer(tracer))
```

To capture the trace as a protobuf, you can use the following option:
```go
tracer, err := pubsub.NewPBTracer("/path/to/trace.pb")
if err != nil {
  panic(err)
}

pubsub.NewGossipSub(..., pubsub.WithEventTracer(tracer))
```

Finally, to use the remote tracer, you can use the following incantations:
```go
// assuming that your tracer runs in x.x.x.x and has a peer ID of QmTracer
pi, err := peer.AddrInfoFromP2pAddr(ma.StringCast("/ip4/x.x.x.x/tcp/4001/p2p/QmTracer"))
if err != nil {
  panic(err)
}

tracer, err := pubsub.NewRemoteTracer(ctx, host, pi)
if err != nil {
  panic(err)
}

ps, err := pubsub.NewGossipSub(..., pubsub.WithEventTracer(tracer))
```

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/libp2p/go-libp2p-pubsub/issues).

Check out our [contributing document](https://github.com/libp2p/community/blob/master/contributing.md) for more information on how we work, and about contributing in general. Please be aware that all interactions related to multiformats are subject to the IPFS [Code of Conduct](https://github.com/ipfs/community/blob/master/code-of-conduct.md).

Small note: If editing the README, please conform to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

The go-libp2p-pubsub project is dual-licensed under Apache 2.0 and MIT terms:

- Apache License, Version 2.0, ([LICENSE-APACHE](./LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
- MIT license ([LICENSE-MIT](./LICENSE-MIT) or http://opensource.org/licenses/MIT)
