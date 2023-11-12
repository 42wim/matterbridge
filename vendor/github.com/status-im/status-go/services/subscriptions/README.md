# Signal Subscriptions

This package implements subscriptions mechanics using [`signal`](../../signal) package.

It defines 3 new RPC methods in the `eth` namespace and 2 signals.

## Methods

###`eth_subscribeSignal`
Creates a new filter and subscribes to its changes via signals.

Parameters: receives the method name and parameters for the filter that is created.

Example 1:
```json
{
  "jsonrpc": "2.0", 
  "id": 1,
  "method": "eth_subscribeSignal", 
  "params": ["eth_newPendingTransactionFilter", []]
}
```

Example 2:
```json
{
  "jsonrpc": "2.0", 
  "id": 2,
  "method": "eth_subscribeSignal", 
  "params": [
    "shh_newMessageFilter",
    [{ "symKeyID":"abcabcabcabc", "topics": ["0x12341234"] }]
  ]
}
```

Supported filters: `shh_newMessageFilter`, `eth_newFilter`, `eth_newBlockFilter`, `eth_newPendingTransactionFilter`
(see [Ethereum documentation](https://github.com/ethereum/wiki/wiki/JSON-RPC) for respective parameters).

Returns: error or `subscriptionID`.


###`eth_unsubscribeSignal`
Unsubscribes and removes one filter by its ID.
NOTE: Unsubscribing from a filter removes it.

Parameters: `subscriptionID` obtained from `eth_subscribeSignal`
Returns: error if something went wrong while unsubscribing.


## Signals

1. Subscription data received

```json
{
  "type": "subscriptions.data",
  "event": {
    "subscription_id": "shh_0x01",
    "data": {
        <whisper envelope 01>,
        <whisper envelope 02>,
        ...
    }
}
```

2. Subscription error received

```json
{
  "type": "subscriptions.error",
  "event": {
    "subscription_id": "shh_0x01",
    "error_message": "can not find filter with id: 0x01"
  }
}
```

