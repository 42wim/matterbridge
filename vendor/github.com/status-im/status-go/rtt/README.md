# Description

This module is used by the Status app to select mailservers based on their RTT(Round Trip Time).

It is exposed via the JSON RPC endpoint in the [`services/mailservers/tcp_ping.go`](../services/mailservers/tcp_ping.go) file.

# Usage

The simplest way to use the `mailserver_Ping` RPC command is using `curl`.

The call takes one struct argument which contains two attributes:

* `addresses` - A list of `enode` addresses to ping.
* `timeoutMs` - Call timeout given in milliseconds.

The return value consists of a list of objects representing a result for each mailserver, each containing following attributes:

* `address` - The `enode` address of given mailserver.
* `rttMs` - Round Trip Time given in milliseconds. Set to `null` in case of an error.`
* `error` - A text of error that caused the ping failure.

# Example

```bash
 $ cat >payload.json <<EOL
{
  "jsonrpc": "2.0",
  "method": "mailservers_ping",
  "params": [
    {
      "addresses": [
        "enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:443",
        "enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:999"
      ],
      "timeoutMs": 500
    }
  ],
  "id": 1
}
EOL

 $ curl -s localhost:8545 -H 'content-type: application/json' -d @payload.json
```
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": [
    {
      "address": "enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:443",
      "rttMs": 31,
      "error": null
    },
    {
      "address": "enode://c42f368a23fa98ee546fd247220759062323249ef657d26d357a777443aec04db1b29a3a22ef3e7c548e18493ddaf51a31b0aed6079bd6ebe5ae838fcfaf3a49@206.189.243.162:999",
      "rttMs": null,
      "error": "tcp check timeout: I/O timeout"
    }
  ]
}
```

# Links

* https://github.com/status-im/status-mobile/issues/9394
* https://github.com/status-im/status-go/pull/1672
* https://github.com/status-im/tcp-shaker
