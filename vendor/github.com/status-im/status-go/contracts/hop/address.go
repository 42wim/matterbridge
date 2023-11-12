package hop

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var errorNotAvailableOnChainID = errors.New("not available for chainID")

var l2SaddleSwapContractAddresses = map[uint64]map[string]common.Address{
	10: {
		"USDC": common.HexToAddress("0x3c0FFAca566fCcfD9Cc95139FEF6CBA143795963"),
		"USDT": common.HexToAddress("0xeC4B41Af04cF917b54AEb6Df58c0f8D78895b5Ef"),
		"DAI":  common.HexToAddress("0xF181eD90D6CfaC84B8073FdEA6D34Aa744B41810"),
		"ETH":  common.HexToAddress("0xaa30D6bba6285d0585722e2440Ff89E23EF68864"),
		"WBTC": common.HexToAddress("0x46fc3Af3A47792cA3ED06fdF3D657145A675a8D8"),
	},
	42161: {
		"USDC": common.HexToAddress("0x10541b07d8Ad2647Dc6cD67abd4c03575dade261"),
		"USDT": common.HexToAddress("0x18f7402B673Ba6Fb5EA4B95768aABb8aaD7ef18a"),
		"DAI":  common.HexToAddress("0xa5A33aB9063395A90CCbEa2D86a62EcCf27B5742"),
		"ETH":  common.HexToAddress("0x652d27c0F72771Ce5C76fd400edD61B406Ac6D97"),
		"WBTC": common.HexToAddress("0x7191061D5d4C60f598214cC6913502184BAddf18"),
	},
	420: {
		"USDC": common.HexToAddress("0xE4757dD81AFbecF61E51824AB9238df6691c3D0e"),
		"ETH":  common.HexToAddress("0xa50395bdEaca7062255109fedE012eFE63d6D402"),
	},
	421613: {
		"USDC": common.HexToAddress("0x83f6244Bd87662118d96D9a6D44f09dffF14b30E"),
		"ETH":  common.HexToAddress("0x69a71b7F6Ff088a0310b4f911b4f9eA11e2E9740"),
	},
}

var l2AmmWrapperContractAddress = map[uint64]map[string]common.Address{
	10: {
		"USDC": common.HexToAddress("0x2ad09850b0CA4c7c1B33f5AcD6cBAbCaB5d6e796"),
		"USDT": common.HexToAddress("0x7D269D3E0d61A05a0bA976b7DBF8805bF844AF3F"),
		"DAI":  common.HexToAddress("0xb3C68a491608952Cb1257FC9909a537a0173b63B"),
		"ETH":  common.HexToAddress("0x86cA30bEF97fB651b8d866D45503684b90cb3312"),
		"WBTC": common.HexToAddress("0x2A11a98e2fCF4674F30934B5166645fE6CA35F56"),
	},
	42161: {
		"USDC": common.HexToAddress("0xe22D2beDb3Eca35E6397e0C6D62857094aA26F52"),
		"USDT": common.HexToAddress("0xCB0a4177E0A60247C0ad18Be87f8eDfF6DD30283"),
		"DAI":  common.HexToAddress("0xe7F40BF16AB09f4a6906Ac2CAA4094aD2dA48Cc2"),
		"ETH":  common.HexToAddress("0x33ceb27b39d2Bb7D2e61F7564d3Df29344020417"),
		"WBTC": common.HexToAddress("0xC08055b634D43F2176d721E26A3428D3b7E7DdB5"),
	},
	420: {
		"USDC": common.HexToAddress("0xfF21e82a4Bc305BCE591530A68628192b5b6B6FD"),
		"ETH":  common.HexToAddress("0xC1985d7a3429cDC85E59E2E4Fcc805b857e6Ee2E"),
	},
	421613: {
		"USDC": common.HexToAddress("0x32219766597DFbb10297127238D921E7CCF5D920"),
		"ETH":  common.HexToAddress("0xa832293f2DCe2f092182F17dd873ae06AD5fDbaF"),
	},
}

var l1BridgeContractAddress = map[uint64]map[string]common.Address{
	1: {
		"USDC": common.HexToAddress("0x3666f603Cc164936C1b87e207F36BEBa4AC5f18a"),
		"USDT": common.HexToAddress("0x3666f603Cc164936C1b87e207F36BEBa4AC5f18a"),
		"DAI":  common.HexToAddress("0x3d4Cc8A61c7528Fd86C55cfe061a78dCBA48EDd1"),
		"ETH":  common.HexToAddress("0xb8901acB165ed027E32754E0FFe830802919727f"),
		"WBTC": common.HexToAddress("0xb98454270065A31D71Bf635F6F7Ee6A518dFb849"),
	},
	5: {
		"USDC": common.HexToAddress("0x7D269D3E0d61A05a0bA976b7DBF8805bF844AF3F"),
		"ETH":  common.HexToAddress("0xC8A4FB931e8D77df8497790381CA7d228E68a41b"),
	},
}

func L2SaddleSwapContractAddress(chainID uint64, symbol string) (common.Address, error) {
	tokens, exists := l2SaddleSwapContractAddresses[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}

	addr, exists := tokens[symbol]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}

	return addr, nil
}

func L2AmmWrapperContractAddress(chainID uint64, symbol string) (common.Address, error) {
	tokens, exists := l2AmmWrapperContractAddress[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}

	addr, exists := tokens[symbol]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}

	return addr, nil
}

func L1BridgeContractAddress(chainID uint64, symbol string) (common.Address, error) {
	tokens, exists := l1BridgeContractAddress[chainID]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}

	addr, exists := tokens[symbol]
	if !exists {
		return *new(common.Address), errorNotAvailableOnChainID
	}

	return addr, nil
}
