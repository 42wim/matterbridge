Rendezvous server
=================

In order to build a docker image, run:

```bash
make image
```

Server usage:

```
  -a, --address string     listener ip address (default "0.0.0.0")
  -d, --data string        path where ENR infos will be stored. (default "/tmp/rendevouz")
  -g, --generate           dump private key and exit.
  -h, --keyhex string      private key hex
  -k, --keypath string     path to load private key
  -p, --port int           listener port (default 9090)
  -v, --verbosity string   verbosity level, options: crit, error, warning, info, debug (default "info")
```

Option `-g` can be used to generate hex of the private key for convenience.
Option `-h` should be used only in tests.

The only mandatory parameter is keypath `-k`, and not mandatory but i suggest to change data path `-d` not to a temporary
directory.


# Differences with original rendezvous

Original rendezvous description by members of libp2p team - [rendezvous](https://github.com/libp2p/specs/pull/56).
We are using current implementation for a similar purposes, but mainly as a light-peer discovery protocol for mobile
devices. Discovery v5 that depends on the kademlia implementation was too slow for mobile and consumed noticeable amount
of traffic to find peers.

Some differences with original implementation:
1. We are using ENR ([Ethereum Node Records](https://eips.ethereum.org/EIPS/eip-778)) for encoding information
about peers. ENR must be signed.
2. We are using RLP instead of protobuf. Mainly for convenience, because ENR already had util for rlp serialization.
3. Smaller liveness TTL for records. At the time of writing liveness TTL is set to be 20s.
This way we want to provide minimal guarantees that peer is online and dialable.
4. ENRs are fetched from storage randomly. And we don't provide a way to fetch "new" records.
It was done as a naive measure against spamming rendezvous servers with invalid records.
And at the same time spread load of new peers between multiple servers.
5. We don't use UNREGISTER request, since we assume that TTL is very low.

Those are mostly implementation details while idea is pretty much the same, but it is important to note that this implementation
is not compatible with one from libp2p team.