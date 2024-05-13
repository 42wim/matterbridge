package snt

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var errorNotAvailableOnChainID = errors.New("not available for chainID")

var contractAddressByChainID = map[uint64]common.Address{
	1:        common.HexToAddress("0x744d70fdbe2ba4cf95131626614a1763df805b9e"), // mainnet
	5:        common.HexToAddress("0x3d6afaa395c31fcd391fe3d562e75fe9e8ec7e6a"), // goerli
	11155111: common.HexToAddress("0xE452027cdEF746c7Cd3DB31CB700428b16cD8E51"), // sepolia
}

func ContractAddress(chainID uint64) (common.Address, error) {
	addr, exists := contractAddressByChainID[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}
	return addr, nil
}
