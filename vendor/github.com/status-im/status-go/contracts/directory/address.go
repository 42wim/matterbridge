package directory

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var errorNotAvailableOnChainID = errors.New("not available for chainID")

var contractAddressByChainID = map[uint64]common.Address{
	10:  common.HexToAddress("0xA8d270048a086F5807A8dc0a9ae0e96280C41e3A"), // optimism mainnet
	420: common.HexToAddress("0xB3Ef5B0825D5f665bE14394eea41E684CE96A4c5"), // optimism goerli testnet
}

func ContractAddress(chainID uint64) (common.Address, error) {
	addr, exists := contractAddressByChainID[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}
	return addr, nil
}
