# Wallet service API

Wallet service provides RPC API for checking transfers history and other methods related to wallet functionality. To enable service two values need to be changed in the config:

1. Set Enable to true in WalletConfig

```json
{
  "WalletConfig": {
    "Enabled": true,
  }
}
```

2. And expose wallet API with APIModules

```
{
  APIModules: "eth,net,web3,peer,wallet",
}
```

## API

### wallet_getTransfersByAddress

Returns avaiable transfers in a given range.

#### Parameters

- `address`: `HEX` - ethereum address encoded in hex
- `toBlock`: `BIGINT` - end of the range. if nil query will return last transfers.
- `limit`: `BIGINT` - limit of returned transfers.
- `fetchMore`: `BOOLEAN` - if `true`, there are less than `limit` fetched transfers in the database, and zero block is not reached yet, history will be scanned for more transfers. If `false` only transfers which are already fetched to the app's database will be returned.

#### Examples

```json
{
  "jsonrpc":"2.0",
  "id":7,
  "method":"wallet_getTransfersByAddress",
  "params":[
    "0xb81a6845649fa8c042dfaceb3f7a684873406993",
    "0x0",
    "0x5",
    true
  ]
}
```

#### Returns

```json
[
  {
    "id":"0xb1a8adeaa0e6727bf01d6d8431b6238bdefa915e19ae7e8ceb16886c9f5e",
    "type":"eth",
    "address":"0xd65f3cb52605a54a833ae118fb13",
    "blockNumber":"0xb7190",
    "blockhash":"0x8d98aa2297fe322d0093b24372e2ead98414959093b479baf670",
    "timestamp":"0x6048ec6",
    "gasPrice":"0x346308a00",
    "gasLimit":"0x508",
    "gasUsed":"0x520",
    "nonce":"0x13",
    "txStatus":"0x1",
    "input":"0x",
    "txHash":"0x1adeaa0e672d7e67bf01d8431b6238bdef15e19ae7e8ceb16886c",
    "value":"0x1",
    "from":"0x2f865fb5dfdf0dfdf54a833ae118fb1363aaasd",
    "to":"0xaaaaaaf3cb52605a54a833ae118fb1363a123123",
    "contract":"0x0000000000000000000000000000000000000000",
    "NetworkID":1
  },...
]
```

### GetTransfersByAddressAndChainID

Returns avaiable transfers in a given range.

#### Parameters

- `chainID`: `INT` - ethereum chain ID
- `address`: `HEX` - ethereum address encoded in hex
- `toBlock`: `BIGINT` - end of the range. if nil query will return last transfers.
- `limit`: `BIGINT` - limit of returned transfers.
- `fetchMore`: `BOOLEAN` - if `true`, there are less than `limit` fetched transfers in the database, and zero block is not reached yet, history will be scanned for more transfers. If `false` only transfers which are already fetched to the app's database will be returned.

#### Examples

```json
{
  "jsonrpc":"2.0",
  "id":7,
  "method":"wallet_getTransfersByAddressAndChainID",
  "params":[
    1,
    "0xb81a6845649fa8c042dfaceb3f7a684873406993",
    "0x0",
    "0x5",
    true
  ]
}
```

#### Returns

```json
[
  {
    "id":"0xb1a8adeaa0e6727bf01d6d8431b6238bdefa915e19ae7e8ceb16886c9f5e",
    "type":"eth",
    "address":"0xd65f3cb52605a54a833ae118fb13",
    "blockNumber":"0xb7190",
    "blockhash":"0x8d98aa2297fe322d0093b24372e2ead98414959093b479baf670",
    "timestamp":"0x6048ec6",
    "gasPrice":"0x346308a00",
    "gasLimit":"0x508",
    "gasUsed":"0x520",
    "nonce":"0x13",
    "txStatus":"0x1",
    "input":"0x",
    "txHash":"0x1adeaa0e672d7e67bf01d8431b6238bdef15e19ae7e8ceb16886c",
    "value":"0x1",
    "from":"0x2f865fb5dfdf0dfdf54a833ae118fb1363aaasd",
    "to":"0xaaaaaaf3cb52605a54a833ae118fb1363a123123",
    "contract":"0x0000000000000000000000000000000000000000",
    "NetworkID":1
  },...
]
```

### wallet_watchTransaction

Starts watching for transaction confirmation/rejection. If transaction was not confirmed/rejected in 10 minutes the call is timed out with error.

#### Parameters

- `tx-id`: `HEX` - transaction hash

#### Example

```json
{
  "jsonrpc":"2.0",
  "id":7,
  "method":"wallet_watchTransaction",
  "params":[
    "0xaaaaaaaa11111112222233333333"
  ]
}
```

### wallet_watchTransactionByChainID

Starts watching for transaction confirmation/rejection. If transaction was not confirmed/rejected in 10 minutes the call is timed out with error.

#### Parameters

- `chainID`: `HEX` - ethereum chain id
- `tx-id`: `HEX` - transaction hash

#### Example

```json
{
  "jsonrpc":"2.0",
  "id":7,
  "method":"wallet_watchTransactionByChainID",
  "params":[
    1,
    "0xaaaaaaaa11111112222233333333"
  ]
}
```

### `wallet_checkRecentHistory`

#### Parameters

- `addresses`: `[]HEX` - array of addresses to be checked

#### Example

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_checkRecentHistory",
  "params":[
    [
      "0x23458d65f3cB52605a54AaA833ae118fb1111aaa",
      "0x24568B4166D11aaa1194097C60Cdc714F7e11111"
    ]
  ]
}
```

### `wallet_checkRecentHistoryForChainIDs`

#### Parameters

- `chainIDs`: `[]INT` - array of ethereum chain ID to be checked
- `addresses`: `[]HEX` - array of addresses to be checked

#### Example

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_checkRecentHistoryForChainIDs",
  "params":[
    [1, 2],
    [
      "0x23458d65f3cB52605a54AaA833ae118fb1111aaa",
      "0x24568B4166D11aaa1194097C60Cdc714F7e11111"
    ]
  ]
}
```

### `wallet_getTokensBalancesForChainIDs`

Returns tokens balances mapping for every account. See section below for the response example.

#### Parameters

- `chainIDs`: `[]INT` - array of ethereum chain ID
- `accounts` `HEX` - list of ethereum addresses encoded in hex
- `tokens` `HEX` - list of ethereum addresses encoded in hex

#### Request

```json
{"jsonrpc":"2.0","id":11,"method":"wallet_getTokensBalancesForChainIDs","params":[
  [1, 2]
  ["0x066ed5c2ed45d70ad72f40de0b4dd97bd67d84de", "0x0ed535be4c0aa276942a1a782669790547ad8768"], 
  ["0x5e4bbdc178684478a615354d83c748a4393b20f0", "0x5e4bbdc178684478a615354d83c748a4393b20f0"]]
}
```

#### Returns

First level keys accounts, second level keys are tokens.

```json
{
  "0x066ed5c2ed45d70ad72f40de0b4dd97bd67d84de": {
    "0x1dfb2099f936b3e98bfc9b7059a8fb04edcce5b3": 12,
    "0x5e4bbdc178684478a615354d83c748a4393b20f0": 12
  },
  "0x0ed535be4c0aa276942a1a782669790547ad8768": {
    "0x1dfb2099f936b3e98bfc9b7059a8fb04edcce5b3": 14,
    "0x5e4bbdc178684478a615354d83c748a4393b20f0": 14
  }
}
```

### `wallet_getTokensBalances`

Returns tokens balances mapping for every account. See section below for the response example.

#### Parameters

- `accounts` `HEX` - list of ethereum addresses encoded in hex
- `tokens` `HEX` - list of ethereum addresses encoded in hex

#### Request

```json
{"jsonrpc":"2.0","id":11,"method":"wallet_getTokensBalances","params":[["0x066ed5c2ed45d70ad72f40de0b4dd97bd67d84de", "0x0ed535be4c0aa276942a1a782669790547ad8768"], ["0x5e4bbdc178684478a615354d83c748a4393b20f0", "0x5e4bbdc178684478a615354d83c748a4393b20f0"]]}
```

#### Returns

First level keys accounts, second level keys are tokens.

```json
{
  "0x066ed5c2ed45d70ad72f40de0b4dd97bd67d84de": {
    "0x1dfb2099f936b3e98bfc9b7059a8fb04edcce5b3": 12,
    "0x5e4bbdc178684478a615354d83c748a4393b20f0": 12
  },
  "0x0ed535be4c0aa276942a1a782669790547ad8768": {
    "0x1dfb2099f936b3e98bfc9b7059a8fb04edcce5b3": 14,
    "0x5e4bbdc178684478a615354d83c748a4393b20f0": 14
  }
}
```

### `wallet_storePendingTransaction`

Stores pending transation in the database.

#### Parameters

- `transaction` `OBJECT` - list of ethereum addresses encoded in hex

##### Transaction

- `hash` `HEX`
- `timestamp` ``INT`
- `from` `HEX`
- `to` `HEX`
- `symbol` `VARCHAR` - `"ETH"` for ethereum, otherwise ERC20 tokaen name, `null` for contract call 
- `gasPrice` `BIGINT`
- `gasLimit` `BIGINT`
- `value` `BIGINT`
- `data` `TEXT` - transaction's `data` field
- `type` `VARCHAR`
- `additionalData` `TEXT` - arbitrary additional data
- `network_id` `INT` - an optional network id

#### Request example

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_storePendingTransaction",
  "params":[
    {
      "hash":"0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e",
      "symbol":"ETH",
      "gasPrice":"2000000000",
      "value":"1000000000000000",
      "from":"0xaaaad65f3cB52605433ae118fb1363aaaaad2",
      "timestamp":1618584138787,
      "gasLimit":"21000",
      "to":"0x237f8B4166D64a2b94097C60Cdc714F7eC3aa079",
      "data":null
    }
  ]
}
```

### `wallet_getPendingTransactions`

Returns all stored pending transactions.

#### Request

```json
{"jsonrpc":"2.0","id":1,"method":"wallet_getPendingTransactions","params":[]}
```

#### Returns

First level keys accounts, second level keys are tokens.

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "result":[
    {
      "hash":"0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e",
      "timestamp":1618584138787,
      "value":"1000000000000000",
      "from":"0xaaaaaaaa605a54a833ae118fb1aaaaaaaaaaa",
      "to":"0x237f8b4166d64a2b94097c60cdc714f7ec3aa079",
      "data":"",
      "symbol":"ETH",
      "gasPrice":"2000000000",
      "gasLimit":"21000",
      "type":"",
      "additionalData":""
    },
    ...
  ]
}
```

### `wallet_getPendingTransactionsByChainID`

Returns all stored pending transactions.

#### Parameters

- `chainID` `INT` - ethereum chain ID

#### Request

```json
{"jsonrpc":"2.0","id":1,"method":"wallet_getPendingTransactions","params":[1]}
```

#### Returns

First level keys accounts, second level keys are tokens.

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "result":[
    {
      "hash":"0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e",
      "timestamp":1618584138787,
      "value":"1000000000000000",
      "from":"0xaaaaaaaa605a54a833ae118fb1aaaaaaaaaaa",
      "to":"0x237f8b4166d64a2b94097c60cdc714f7ec3aa079",
      "data":"",
      "symbol":"ETH",
      "gasPrice":"2000000000",
      "gasLimit":"21000",
      "type":"",
      "additionalData":"",
      "network_id": 1
    },
    ...
  ]
}
```

### `wallet_getPendingOutboundTransactionsByAddress`

Returns all stored pending transaction sent from `address`.

#### Parameters

- `address` `HEX` 

#### Request

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_getPendingOutboundTransactionsByAddress",
  "params":[
    "0xaaaaaaaa605a54a833ae118fb1aaaaaaaaaaa"
  ]
}
```

#### Returns

First level keys accounts, second level keys are tokens.

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "result":[
    {
      "hash":"0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e",
      "timestamp":1618584138787,
      "value":"1000000000000000",
      "from":"0xaaaaaaaa605a54a833ae118fb1aaaaaaaaaaa",
      "to":"0x237f8b4166d64a2b94097c60cdc714f7ec3aa079",
      "data":"",
      "symbol":"ETH",
      "gasPrice":"2000000000",
      "gasLimit":"21000",
      "type":"",
      "additionalData":""
    },
    ...
  ]
}
```

### `wallet_getPendingOutboundTransactionsByAddressAndChainID`

Returns all stored pending transaction sent from `address`.

#### Parameters

- `chainID` `INT` 
- `address` `HEX` 

#### Request

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_getPendingOutboundTransactionsByAddress",
  "params":[
    1,
    "0xaaaaaaaa605a54a833ae118fb1aaaaaaaaaaa"
  ]
}
```

#### Returns

First level keys accounts, second level keys are tokens.

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "result":[
    {
      "hash":"0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e",
      "timestamp":1618584138787,
      "value":"1000000000000000",
      "from":"0xaaaaaaaa605a54a833ae118fb1aaaaaaaaaaa",
      "to":"0x237f8b4166d64a2b94097c60cdc714f7ec3aa079",
      "data":"",
      "symbol":"ETH",
      "gasPrice":"2000000000",
      "gasLimit":"21000",
      "type":"",
      "additionalData":"",
      "network_id": 1
    },
    ...
  ]
}
```

### `wallet_deletePendingTransaction`

Deletes pending transaction from the database by `hash`.

#### Parameters

- `hash` `HEX` 

#### Request

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_deletePendingTransaction",
  "params":[
    "0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e"
  ]
}
```

### `wallet_deletePendingTransactionByChainID`

Deletes pending transaction from the database by `hash`.

#### Parameters

- `chainID` `INT` 
- `hash` `HEX` 

#### Request

```json
{
  "jsonrpc":"2.0",
  "id":1,
  "method":"wallet_deletePendingTransaction",
  "params":[
    1,
    "0x3bce2c2d0fffbd2862ef3ec61a62872e54954551585fa0072d8e5c2f6be3523e"
  ]
}
```

## Signals
-------

All events are of the same format:

```json
{
  "type": "wallet",
  "event": {
    "type": "event-type",
    "blockNumber": 0,
    "accounts": [
      "0x42c8f505b4006d417dd4e0ba0e880692986adbd8",
      "0x3129mdasmeo132128391fml1130410k312312mll"
    ],
    "message": "something might be here"
  }
}
```

1. `new-transfers`

Emitted when transfers are detected. In this case block number is a block number of the latest found transfer.
Client expected to request transfers starting from received block.

2. `recent-history-fetching`

Emitted when history scanning is started.

3. `recent-history-ready`

Emitted when history scanning is ended.

4. `fetching-history-error`

Emitted when when history can't be fetched because some error. Error's decritption can be found in `message` field.

5. `non-archival-node-detected`

Emitted when the application is connected to a non-archival node.

## Flows

### Account creation

When a new multiaccount is created corresponding address will not contain any transaction. Thus no point in checking history, it will be empty.

1. Call `wallet_checkRecentHistory`
2. On `recent-history-ready` request transactions via `wallet_getTransfersByAddress`
3. Repeat `wallet_checkRecentHistory` in N minutes (currently 20 minutes in `status-mobile` for upstream RPC node. If a custom node is used interval can be arbitrary)

### Logging into application
1. Call `wallet_checkRecentHistory`
2. On `recent-history-ready` request transactions via `wallet_getTransfersByAddress`
3. Repeat `wallet_checkRecentHistory` in N minutes (currently 20 minutes in `status-mobile` for upstream RPC node. If a custom node is used interval can be arbitrary)

### Watching transaction
1. Call `wallet_watchTransaction`
2. On success call `wallet_checkRecentHistory`
3. On `recent-history-ready` request transactions via `wallet_getTransfersByAddress`
