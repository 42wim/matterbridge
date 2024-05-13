package chain

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// The code below is mostly copied from go-ethereum/ethclient (see TransactionInBlock), to keep the exact same behavior as the
// normal Transaction items, but exposing the additional data obtained in the `rpcTransaction` struct`.
// Unfortunately, the functions and classes used are not exposed outside of the package.
type FullTransaction struct {
	Tx *types.Transaction
	TxExtraInfo
}

type TxExtraInfo struct {
	BlockNumber *hexutil.Big    `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash    `json:"blockHash,omitempty"`
	From        *common.Address `json:"from,omitempty"`
}

func callBlockHashByTransaction(ctx context.Context, rpc *rpc.Client, number *big.Int, index uint) (common.Hash, error) {
	var json *FullTransaction
	err := rpc.CallContext(ctx, &json, "eth_getTransactionByBlockNumberAndIndex", toBlockNumArg(number), hexutil.Uint64(index))

	if err != nil {
		return common.HexToHash(""), err
	}
	if json == nil {
		return common.HexToHash(""), ethereum.NotFound
	}
	return *json.BlockHash, nil
}

func (tx *FullTransaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.Tx); err != nil {
		return err
	}
	return json.Unmarshal(msg, &tx.TxExtraInfo)
}

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	pending := big.NewInt(-1)
	if number.Cmp(pending) == 0 {
		return "pending"
	}
	finalized := big.NewInt(int64(rpc.FinalizedBlockNumber))
	if number.Cmp(finalized) == 0 {
		return "finalized"
	}
	safe := big.NewInt(int64(rpc.SafeBlockNumber))
	if number.Cmp(safe) == 0 {
		return "safe"
	}
	return hexutil.EncodeBig(number)
}
