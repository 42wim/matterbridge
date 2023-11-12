package ethscan

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var errorNotAvailableOnChainID = errors.New("not available for chainID")

type ContractData struct {
	Address        common.Address
	CreatedAtBlock uint
}

var contractDataByChainID = map[uint64]ContractData{
	1:        {common.HexToAddress("0x08A8fDBddc160A7d5b957256b903dCAb1aE512C5"), 12_194_222}, // mainnet
	5:        {common.HexToAddress("0x08A8fDBddc160A7d5b957256b903dCAb1aE512C5"), 4_578_854},  // goerli
	10:       {common.HexToAddress("0x9e5076df494fc949abc4461f4e57592b81517d81"), 34_421_097}, // optimism
	420:      {common.HexToAddress("0xf532c75239fa61b66d31e73f44300c46da41aadd"), 2_236_534},  // goerli optimism
	42161:    {common.HexToAddress("0xbb85398092b83a016935a17fc857507b7851a071"), 70_031_945}, // arbitrum
	421613:   {common.HexToAddress("0xec21ebe1918e8975fc0cd0c7747d318c00c0acd5"), 818_155},    // goerli arbitrum
	777333:   {common.HexToAddress("0x0000000000000000000000000000000000777333"), 50},         // unit tests
	11155111: {common.HexToAddress("0xec21ebe1918e8975fc0cd0c7747d318c00c0acd5"), 4_366_506},  // sepolia
	421614:   {common.HexToAddress("0xec21Ebe1918E8975FC0CD0c7747D318C00C0aCd5"), 553_947},    // sepolia arbitrum
	11155420: {common.HexToAddress("0xec21ebe1918e8975fc0cd0c7747d318c00c0acd5"), 7_362_011},  // sepolia optimism
}

func ContractAddress(chainID uint64) (common.Address, error) {
	contract, exists := contractDataByChainID[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}
	return contract.Address, nil
}

func ContractCreatedAt(chainID uint64) (uint, error) {
	contract, exists := contractDataByChainID[chainID]
	if !exists {
		return 0, errorNotAvailableOnChainID
	}
	return contract.CreatedAtBlock, nil
}
