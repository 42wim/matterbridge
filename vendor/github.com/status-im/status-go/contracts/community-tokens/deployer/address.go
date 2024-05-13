package communitytokendeployer

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var errorNotAvailableOnChainID = errors.New("deployer contract not available for chainID")

// addresses can be found on https://github.com/status-im/communities-contracts#deployments
var contractAddressByChainID = map[uint64]common.Address{
	1:        common.HexToAddress("0xB3Ef5B0825D5f665bE14394eea41E684CE96A4c5"), // Mainnet
	5:        common.HexToAddress("0x81f4951ff8859d305F47A4574B206cF64C0d2645"), // Goerli
	10:       common.HexToAddress("0x31463D22750324C8721FF7751584EF62F2ff93b3"), // Optimism
	420:      common.HexToAddress("0xfFa8A255D905c909379859eA45B959D090DDC2d4"), // Optimism Goerli
	42161:    common.HexToAddress("0x744Fd6e98dad09Fb8CCF530B5aBd32B56D64943b"), // Arbitrum
	421613:   common.HexToAddress("0x7Ff554af5b6624db2135E4364F416d1D397f43e6"), // Arbitrum Goerli
	11155111: common.HexToAddress("0xCDE984e57cdb88c70b53437cc694345B646371f9"), // Sepolia
	421614:   common.HexToAddress("0x7Ff554af5b6624db2135E4364F416d1D397f43e6"), // Arbitrum Sepolia
	11155420: common.HexToAddress("0xcE2A896eEA2F585BC0C3753DC8116BbE2AbaE541"), // Optimism Sepolia
}

func ContractAddress(chainID uint64) (common.Address, error) {
	addr, exists := contractAddressByChainID[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}
	return addr, nil
}
