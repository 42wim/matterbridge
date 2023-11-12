# RLN Contracts

### Requirements:
- Node.js
- Go 
- jq
- [solcjs 0.8.15](https://github.com/ethereum/solc-js)
- [abigen](https://github.com/ethereum/go-ethereum/tree/master/cmd/abigen) 

### Build
1. Install solcjs with `npm install -g solc@0.8.15`
2. Clone [go-ethereum](https://github.com/ethereum/go-ethereum) and install `abigen`
```
cd $GOPATH/src/github.com/ethereum/go-ethereum
$ go install ./cmd/abigen
```
3. Execute `go generate` to create go bindings for the RLN smart contracts
```
go generate
```

### Notes
Follow https://github.com/vacp2p/rln-contract for updates on solc versions