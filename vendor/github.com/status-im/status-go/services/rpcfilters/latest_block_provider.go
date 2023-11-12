package rpcfilters

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/status-im/status-go/rpc"
)

type rpcProvider interface {
	RPCClient() *rpc.Client
}

// blockInfo contains the hash and the number of the latest block
type blockInfo struct {
	Hash        common.Hash   `json:"hash"`
	NumberBytes hexutil.Bytes `json:"number"`
}

// Number returns a big.Int representation of the encoded block number.
func (i blockInfo) Number() *big.Int {
	number := big.NewInt(0)
	number.SetBytes(i.NumberBytes)
	return number
}

// latestBlockProvider provides the latest block info from the blockchain
type latestBlockProvider interface {
	GetLatestBlock() (blockInfo, error)
}

// latestBlockProviderRPC is an implementation of latestBlockProvider interface
// that requests a block using an RPC client provided
type latestBlockProviderRPC struct {
	rpc rpcProvider
}

// GetLatestBlock returns the block info
func (p *latestBlockProviderRPC) GetLatestBlock() (blockInfo, error) {
	rpcClient := p.rpc.RPCClient()

	if rpcClient == nil {
		return blockInfo{}, errors.New("no active RPC client: is the node running?")
	}

	var result blockInfo

	err := rpcClient.Call(&result, rpcClient.UpstreamChainID, "eth_getBlockByNumber", "latest", false)

	if err != nil {
		return blockInfo{}, err
	}

	return result, nil
}
