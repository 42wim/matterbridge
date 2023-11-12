Peer pool signals
=================

Peer pool sends 3 types of signals.

Discovery started signal will be sent once discovery server is started.
And every time node will have to re-start discovery server because peer number dropped too low.

```json
{
  "type": "discovery.started",
  "event": null
}
```


Discovery stopped signal will be sent once discovery found max limit of peers
for every registered topic.

```json
{
  "type": "discovery.stopped",
  "event": null
}
```


Discovery summary signal will be sent every time new peer is added or removed
from a cluster. It will contain a map with capability as a key and total numbers
of peers with that capability as a value.

```json
{
  "type": "discovery.summary",
  "event": [
    {
      "id": "339c84c816b5f17a622c8d7ab9498f9998e942a274f70794af934bf5d3d02e14db8ddca2170e4edccede29ea6d409b154c141c34c01006e76c95e17672a27454",
      "name": "peer-0/v1.0/darwin/go1.10.1",
      "caps": [
        "shh/6"
      ],
      "network": {
        "localAddress": "127.0.0.1:61049",
        "remoteAddress": "127.0.0.1:33732",
        "inbound": false,
        "trusted": false,
        "static": true
      },
      "protocols": {
        "shh": "unknown"
      }
    }
  ]
}
```

Or if we don't have any peers:

```json
{
  "type": "discovery.summary",
  "event": []
}
```
