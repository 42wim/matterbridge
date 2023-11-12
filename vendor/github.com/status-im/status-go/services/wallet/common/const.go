package common

import (
	"strconv"
	"time"
)

type ChainID uint64

const (
	UnknownChainID  uint64 = 0
	EthereumMainnet uint64 = 1
	EthereumGoerli  uint64 = 5
	EthereumSepolia uint64 = 11155111
	OptimismMainnet uint64 = 10
	OptimismGoerli  uint64 = 420
	OptimismSepolia uint64 = 11155420
	ArbitrumMainnet uint64 = 42161
	ArbitrumGoerli  uint64 = 421613
	ArbitrumSepolia uint64 = 421614
)

type ContractType byte

const (
	ContractTypeUnknown ContractType = iota
	ContractTypeERC20
	ContractTypeERC721
	ContractTypeERC1155
)

func (c ChainID) String() string {
	return strconv.Itoa(int(c))
}

func (c ChainID) IsMainnet() bool {
	switch uint64(c) {
	case EthereumMainnet, OptimismMainnet, ArbitrumMainnet:
		return true
	case EthereumGoerli, EthereumSepolia, OptimismGoerli, OptimismSepolia, ArbitrumGoerli, ArbitrumSepolia:
		return false
	case UnknownChainID:
		return false
	}
	return false
}

func AllChainIDs() []ChainID {
	return []ChainID{
		ChainID(EthereumMainnet),
		ChainID(EthereumGoerli),
		ChainID(EthereumSepolia),
		ChainID(OptimismMainnet),
		ChainID(OptimismGoerli),
		ChainID(OptimismSepolia),
		ChainID(ArbitrumMainnet),
		ChainID(ArbitrumGoerli),
		ChainID(ArbitrumSepolia),
	}
}

var AverageBlockDurationForChain = map[ChainID]time.Duration{
	ChainID(UnknownChainID):  time.Duration(12000) * time.Millisecond,
	ChainID(EthereumMainnet): time.Duration(12000) * time.Millisecond,
	ChainID(EthereumGoerli):  time.Duration(12000) * time.Millisecond,
	ChainID(OptimismMainnet): time.Duration(400) * time.Millisecond,
	ChainID(OptimismGoerli):  time.Duration(2000) * time.Millisecond,
	ChainID(ArbitrumMainnet): time.Duration(300) * time.Millisecond,
	ChainID(ArbitrumGoerli):  time.Duration(1500) * time.Millisecond,
}
