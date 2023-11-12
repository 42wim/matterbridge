package communitytokens

import "github.com/status-im/status-go/services/wallet/bigint"

type CollectibleContractData struct {
	TotalSupply    *bigint.BigInt
	Transferable   bool
	RemoteBurnable bool
	InfiniteSupply bool
}
