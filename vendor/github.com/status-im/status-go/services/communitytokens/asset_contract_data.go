package communitytokens

import "github.com/status-im/status-go/services/wallet/bigint"

type AssetContractData struct {
	TotalSupply    *bigint.BigInt
	InfiniteSupply bool
}
