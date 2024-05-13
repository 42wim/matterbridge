package gaspriceoracle

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	wallet_common "github.com/status-im/status-go/services/wallet/common"
)

var ErrorNotAvailableOnChainID = errors.New("not available for chainID")

var contractAddressByChainID = map[uint64]common.Address{
	wallet_common.OptimismMainnet: common.HexToAddress("0x8527c030424728cF93E72bDbf7663281A44Eeb22"),
	wallet_common.OptimismSepolia: common.HexToAddress("0x5230210c2b4995FD5084b0F5FD0D7457aebb5010"),
}

func ContractAddress(chainID uint64) (common.Address, error) {
	addr, exists := contractAddressByChainID[chainID]
	if !exists {
		return *new(common.Address), ErrorNotAvailableOnChainID
	}
	return addr, nil
}
