# `waku`

## Table of contents

- [What is Waku?](#what-is-waku)
- [Waku versioning](#waku-versioning)
- [What does this package do?](#what-does-this-package-do)
- [Waku package files](#waku-package-files)

## What is Waku?

Waku is a communication protocol for sending messages between Dapps. Waku is a fork of the [Ethereum Whisper subprotocol](https://github.com/ethereum/wiki/wiki/Whisper), although not directly compatible with Whisper, both Waku and Whisper subprotocols can communicate [via bridging](https://github.com/vacp2p/specs/blob/master/specs/waku/waku-1.md#backwards-compatibility).

Waku was [created to solve scaling issues with Whisper](https://discuss.status.im/t/fixing-whisper-for-great-profit/1419) and [currently diverges](https://github.com/vacp2p/specs/blob/master/specs/waku/waku-1.md#differences-between-shh6-and-waku1) from Whisper in the following ways:

- RLPx subprotocol is changed from `shh/6` to `waku/1`.
- Light node capability is added.*
- Optional rate limiting is added.
- Status packet has following additional parameters: light-node, confirmations-enabled and rate-limits
- Mail Server and Mail Client functionality is now part of the specification.
- P2P Message packet contains a list of envelopes instead of a single envelope.

*As per [vacp2p/specs#117](https://github.com/vacp2p/specs/pull/117) Waku de jure introduced light nodes as far as updates to the written Whisper specifications. Though the de facto case is that the `go-ethereum` Whisper implementation had [already implemented light nodes](https://github.com/ethereum/go-ethereum/blob/510b6f90db406b697610fe0ff2eee66d173673b2/whisper/whisperv6/whisper.go#L291) and weren't a new feature in code.

Although `status-go`'s Waku light node functionality is a direct fork of `go-ethereum`'s Whisper light node functionality, technically, as far as specifications are concerned, light nodes are considered a new feature introduced in Waku.  

## Waku versioning

This package follows a versioning pattern that makes clean separation between breaking versions. As [detailed in the PR](https://github.com/status-im/status-go/pull/1947#issue-407073908) that introduced this strategy to the package.

>... the way we will move across versions is to maintain completely separate codebases and eventually remove those that are not supported anymore.
>
>This has the drawback of some code duplication, but the advantage is that is more explicit what each version requires, and changes in one version will not impact the other, so we won't pile up backward compatible code. This is the same strategy used by whisper in go ethereum and is influenced by https://www.youtube.com/watch?v=oyLBGkS5ICk.

Familiarise yourself with the [Spec-ulation Keynote by Rich Hickey](https://www.youtube.com/watch?v=oyLBGkS5ICk), if you wish to more deeply understand the rationale for this versioning implementation. 

This means that breaking changes will necessitate a new protocol version and a new version sub-package. The packages follow the naming convention of `v*` where `*` represents the major / breaking version number of the protocol.

Currently the package has the following version sub-packages:

- [version 0 - `v0`](./v0)
- [version 1 - `v1`](./v1)

## What does this package do? 

The basic function of this package is to implement the [waku specifications](https://github.com/vacp2p/specs/blob/master/specs/waku/waku-1.md), and provide the `status-go` binary with the ability to send and receive messages via Waku.

## Waku package files

  - [waku.go](#wakugo)
  - [api.go](#apigo)
  - [config.go](#configgo)
  - [mailserver.go](#mailservergo)
  - [common](#common)
    - [bloomfilter.go](#bloomfiltergo)
    - [const.go](#constgo)
    - [envelope.go](#envelopego)
    - [errors.go](#errorsgo)
    - [events.go](#eventsgo)
    - [filter.go](#filtergo)
    - [helpers.go](#helpersgo)
    - [message.go](#messagego)
    - [metrics.go](#metricsgo)
    - [protocol.go](#protocolgo)
    - [rate_limiter.go](#rate_limitergo)
    - [topic.go](#topicgo)
  - [Versioned](#versioned)
     - [const.go](#version-constgo)
     - [init.go](#version-initgo)
     - [message.go](#version-messagego)
     - [peer.go](#version-peergo)
     - [status_options.go](#version-status_optionsgo)

## Root

### `waku.go`

[`waku.go`](./waku.go) serves as the main entry point for the package and where the main `Waku{}` struct lives.

---

### `api.go`

[`api.go`](./api.go) is home to the `PublicWakuAPI{}` struct which provides the waku RPC service that can be used publicly without security implications.

`PublicWakuAPI{}` wraps the main `Waku{}`, making the `Waku{}` functionality suitable for external consumption.

#### Consumption

`PublicWakuAPI{}` is wrapped by `eth-node\bridge\geth.gethPublicWakuAPIWrapper{}`, which is initialised via `eth-node\bridge\geth.NewGethPublicWakuAPIWrapper()` and exposed via `gethWakuWrapper.PublicWakuAPI()` and is finally consumed by wider parts of the application.

#### Notes

It is worth noting that each function of `PublicWakuAPI{}` received an unused `context.Context` parameter. This is originally passed in way higher up the food-chain and without significant refactoring is not a simple thing to remove / change. Mobile bindings depend on the ability to pass in a context.

---

### `config.go`

[`config.go`](./config.go) is home to the `Config{}` struct and the declaration of `DefaultConfig`.

`Config{}` is used to initialise the settings of an instantiated `Waku{}`. `waku.New()` creates a new instance of a `Waku{}` and takes a `Config{}` as a parameter, if nil is passed instead of an instance of `Config{}`, `DefaultConfig` is used. 

#### Configuration values

|Name                      |Type     |Description|
|--------------------------|---------|---|
|`MaxMessageSize`          |`uint32` |Sets the maximum size of a waku message in bytes|
|`MinimumAcceptedPoW`      |`float64`|Sets the minimum amount of work a message needs to have to be accepted by the waku node|
|`BloomFilterMode`         |`bool`   |When true, the waku node only matches against bloom filter|
|`LightClient`             |`bool`   |When true, the waku node does not forward messages|
|`FullNode`                |`bool`   |When true, the waku node forwards all messages|
|`RestrictLightClientsConn`|`bool`   |When true, the waku node does not accept light clients as peers if it is a light client itself|
|`EnableConfirmations`     |`bool`   |When true, sends message confirmations|

#### Default

The default configuration for a `status-go` Waku node is:

	MaxMessageSize           : 1Mb
	MinimumAcceptedPoW       : 0.2
	RestrictLightClientsConn : true

---

### `mailserver.go`

[`mailserver.go`](./mailserver.go) is home to `MailServer` interface, which is implemented by `mailserver.WakuMailServer{}` found in the package file [`mailserver/mailserver.go`](../mailserver/mailserver.go). `MailServer` represents a mail server, capable of receiving and archiving messages for subsequent delivery to the peers.

Additionally this package is home to `MailServerResponse{}` which represents the response payload sent by the mail-server. `MailServerResponse{}` is ultimately initialised by `CreateMailServerEvent()`, which is tied to the main `Waku{}` via the `Waku.OnP2PRequestCompleted()` function. This is ultimately accessed via the `Peer.Run()` function and is made available outside of the package with the `waku.HandlePeer()` function via `Waku.protocol.Run := waku.HandlePeer`. 

---

## Common

### `bloomfilter.go`

[`bloomfilter.go`](./common/bloomfilter.go) holds a few bloomfilter specific functions.

---

### `const.go`

[`const.go`](./common/const.go), originally a hangover from the [`go-ethereum` `whisperv6/doc.go` package file](https://github.com/ethereum/go-ethereum/blob/master/whisper/whisperv6/doc.go) later [refactored](https://github.com/status-im/status-go/pull/1950), is home to the common Waku constants.

#### Notes

Versions also have version specific `const.go` files.

---

### `envelope.go`

[`envelope.go`](./common/envelope.go) is home to the `Evelope{}` and `EnvelopeError{}` structs. `Envelope{}` is used as the data packet in which message data is sent through the Waku network.

`Envelope{}` is accessed via the initialisation function `NewEnvelope()`, which is exclusively consumed by `Message.Wrap()` that prepares a message to be sent via Waku. 

---

### `errors.go`

[`errors.go`](./common/errors.go) holds generic package errors.

---

### `events.go`

[`events.go`](./common/events.go) handles data related to Waku events. This file contains string type `const`s that identify known Waku events.

Additionally, the file contains `EnvelopeEvent{}`, which serves as a representation of events created by envelopes. `EnvelopeEvent{}`s are initialised exclusively within the `waku` package.  

--- 

### `filter.go`

[`filter.go`](./common/filter.go) is home to `Filter{}` which represents a waku filter.

#### Usage

A `status-go` node will install / register filters through RPC calls from a client (eg `status-mobile`). The basic implementation of a filter requires at least 2 things:

1) An encryption key, example "`superSafeEncryptionKey`"
2) A 4 byte topic (`TopicType`), example "`0x1234`"

The node will install the filter `[0x1234][{"superSafeEncryptionKey"}]` on an instance of `Filters{}` and will notify its peers of this event

When a node receives an envelope it will attempt to match the topics against the installed filters, and then try to decrypt the envelope if the topic matches.

For example, if a node receives an envelope with topic `0x1234`, the node will try to use the installed filter key `superSafeEncryptionKey` to decrypt the message. On success the node passes the decrypted message to the client.

In addition to the basic example above `Filter{}` allows for richer filtering:

|Field Name  |Type               |Description                                |
|------------|-------------------|-------------------------------------------|
|`Src`       |`*ecdsa.PublicKey` |Sender of the message. *Currently not used*|
|`KeyAsym`   |`*ecdsa.PrivateKey`|Private Key of recipient                   |
|`KeySym`    |`[]byte`           |Key associated with the Topic              |
|`Topics`    |`[][]byte`         |Topics to filter messages with             |
|`PoW`       |`float64`          |Proof of work as [described in the Waku specs](https://github.com/vacp2p/specs/blob/master/specs/waku/waku-1.md#pow-requirement-update) .<br/><br/>**Note:** *In `status-mobile` each client listens to the topic hash(pk), if a client wants to send a message to hash(pk1) they will also need to listen the hash(pk1) topic. However if the client doesn't want to receive envelopes for topic hash(pk1), the client may set the PoW to 1 so that all envelopes for topic hash(pk1) are discarded.*|
|`AllowP2P`  |`bool`             |Indicates whether this filter is interested in direct peer-to-peer messages.<br/><br/>**Note:** *Typically set to true, we always want to receive P2P envelopes on a filter from trusted peers*|
|`SymKeyHash`|`common.Hash`      |The Keccak256Hash of the symmetric key, needed for optimization|   

**Waku / Whisper divergence**

Whisper, will process all the installed filters that the node has, and build a `BloomFilter` from all the topics of each installed filter (i.e. `func ToBloomFilter(topics []TopicType) []byte { ... }`). When a peer receives this BloomFilter, it will match the topic on each envelope that they receive against the BloomFilter, if it matches, it will forward this to the peer.

Waku, by default, does not send a BloomFilter, instead sends the topic in a clear array of `[]TopicType`. This is an improvement on Whisper's usage as a BloomFilter may include false positives, which increase bandwidth usage. In contrast, clear topics are matched exactly and therefore don't create redundant bandwidth usage.

---

### `helpers.go`

[`helpers.go`](./common/helpers.go) holds the package's generic functions.

---

### `message.go`

[`message.go`](./common/message.go) is home to all message related functionality and contains a number of structs:

|Name|Description|
|---|---|
|`MessageParams{}`|Specifies the exact way a message should be wrapped into an Envelope|
|`sentMessage{}`|Represents an end-user data packet to transmit through the Waku protocol. These are wrapped into Envelopes that need not be understood by intermediate nodes, just forwarded.|
|`ReceivedMessage{}`|Represents a data packet to be received through the Waku protocol and successfully decrypted.|
|`MessagesRequest{}`|Contains details of a request for historic messages.|
|`MessagesResponse{}`|Represents a request response sent after processing batch of envelopes.|
|`MessageStore` `interface`|Implemented by `MemoryMessageStore{}`|
|`MemoryMessageStore{}`|Represents messages stored in a memory hash table.|

---

### `metrics.go`

[`metrics.go`](./common/metrics.go) is home to [Prometheus](https://prometheus.io/) metric hooks, for counting a range of Waku related metrics.

---

### `protocol.go`

[`protocol.go`](./common/protocol.go) houses the `Peer` and `WakuHost` interfaces.

`Peer` represents a remote Waku client with which the local host waku instance exchanges data / messages. 

`WakuHost` is the local instance of waku, which both interacts with remote clients (peers) and local clients (like `status-mobile`, via a RPC API).

---

### `rate_limiter.go`

[`rate_limiter.go`](./common/rate_limiter.go) was introduced as an improvement to Whisper allowing Waku nodes to limit the rate at which data is transferred. These limits are defined by the `RateLimits{}` which allows nodes to limit the following:

|Limit Type|Description|
|---|---|
|`IPLimits`|Messages per second from a single IP (default 0, no limits)|
|`PeerIDLimits`|Messages per second from a single peer ID (default 0, no limits)|
|`TopicLimits`|Messages per second from a single topic (default 0, no limits)|

In addition to the `RateLimits{}` this file also contains the following interfaces and structs.

|Name|Description|
|---|---|
|`RateLimiterPeer` `interface`|Represents a `Peer{}` that is capable of being rate limited|
|`RateLimiterHandler` `interface`|Represents handler functionality for a Rate Limiter in the cases of exceeding a peer limit and exceeding an IP limit|
|`MetricsRateLimiterHandler{}`|Implements `RateLimiterHandler`, represents a handler for reporting rate limit Exceed data to the metrics collection service (currently prometheus)|
|`DropPeerRateLimiterHandler{}`|Implements `RateLimiterHandler`, represents a handler that introduces Tolerance to the number of Peer connections before Limit Exceeded errors are returned.|
|`RateLimits{}`|Represents rate limit settings exchanged using rateLimitingCode packet or in the handshake.|
|`PeerRateLimiterConfig{}`|Represents configurations for initialising a PeerRateLimiter|
|`PeerRateLimiter{}`|Represents a rate limiter that limits communication between Peers|

The default PeerRateLimiterConfig is:

```text
LimitPerSecIP:      10,
LimitPerSecPeerID:  5,
WhitelistedIPs:     nil,
WhitelistedPeerIDs: nil,
```

---

### `topic.go`

[`topic.go`](./common/topic.go) houses the `TopicType` type.

`TopicType` represents a cryptographically secure, probabilistic partial classification of a message, determined as the first (leftmost) 4 bytes of the SHA3 hash of some arbitrary data given by the original author of a message.

Topics are used to filter incoming messages that the host's user has registered interest in. For further details on filtering see [filter.go](#filtergo).

---

## Versioned

For details about the divergence between versions please consult the `README`s of each version package.

- [version 0](./v0)
- [version 1](./v1)

---

### Version `const.go`

`const.go` is home to the version sub-package's `const`s. These constants are version dependant and, as expected, may change from version to version.

---

### Version `init.go`

`init.go` is home to the version sub-package's initialisation, and is used to initialise struct based variables at runtime. 

---

### Version `message.go`

`message.go` is home to the `MultiVersionResponse{}` and `Version1MessageResponse{}` structs, both of which are exclusively consumed by the version subpackage's `Peer{}`.

Both of these structs are used for handling Waku message responses, also known as message confirmations.

`Version1MessageResponse{}` is used for sending message responses. `MultiVersionResponse{}` is used for handling incoming message responses. 

#### Usage

Message confirmations are used to inform a user, via the UI, that a message has been sent.
Initially the message is marked as "Pending" and eventually as "Sent".

In order to trigger the message state transition from "pending" to "sent",
Waku uses `MessageResponse{}`. Each peer on receiving a message will send back a `MessageResponse`,
see the [`NewMessagesResponse()` function](https://github.com/status-im/status-go/blob/4d00656c41909ccdd80a8a77a0982bd66f74d29e/waku/v1/message.go#L31).

The Waku host checks that the peer is a mailserver and if this mailserver was selected by the user,
if so the message will be marked as "Sent" in the UI.

For further details [read the Waku specification section](https://github.com/status-im/specs/blob/master/docs/stable/3-whisper-usage.md#message-confirmations) on the subject.

#### Notes

Versioning at the `MessageResponse` level (see the struct Version field) should be phased out and defer to the subpackage's version number. Consider removal once we decide to move to a new major Waku version (i.e. `waku/2`).

---

### Version `peer.go`

`peer.go` holds the version's sub-package `Peer{}` implementation of the `common.Peer` interface. 

---

### Version `status_options.go`

`status_options.go` holds the version's sub-package `StatusOptions{}` which implements the `ethereum/go-ethereum/rlp` `Decoder` and `Encoder` interfaces.

`StatusOptions` defines additional information shared between peers during the handshake. There might be more options provided than fields in `StatusOptions`, and as per the specs, should be ignored during deserialisation to stay forward compatible. In the case of RLP, options should be serialised to an array of tuples where the first item is a field name and the second is a RLP-serialised value.

For further details on RLP see:
- https://github.com/ethereum/wiki/wiki/RLP
- https://specs.vac.dev/specs/waku/waku.html#use-of-rlpx-transport-protocol
