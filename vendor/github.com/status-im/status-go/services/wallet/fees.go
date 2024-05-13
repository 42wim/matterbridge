package wallet

import (
	"context"
	"math"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/consensus/misc"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	gaspriceoracle "github.com/status-im/status-go/contracts/gas-price-oracle"
	"github.com/status-im/status-go/rpc"
)

type GasFeeMode int

const (
	GasFeeLow GasFeeMode = iota
	GasFeeMedium
	GasFeeHigh
)

type SuggestedFees struct {
	GasPrice             *big.Float `json:"gasPrice"`
	BaseFee              *big.Float `json:"baseFee"`
	MaxPriorityFeePerGas *big.Float `json:"maxPriorityFeePerGas"`
	MaxFeePerGasLow      *big.Float `json:"maxFeePerGasLow"`
	MaxFeePerGasMedium   *big.Float `json:"maxFeePerGasMedium"`
	MaxFeePerGasHigh     *big.Float `json:"maxFeePerGasHigh"`
	L1GasFee             *big.Float `json:"l1GasFee"`
	EIP1559Enabled       bool       `json:"eip1559Enabled"`
}

func (s *SuggestedFees) feeFor(mode GasFeeMode) *big.Float {
	if !s.EIP1559Enabled {
		return s.GasPrice
	}

	if mode == GasFeeLow {
		return s.MaxFeePerGasLow
	}

	if mode == GasFeeHigh {
		return s.MaxFeePerGasHigh
	}

	return s.MaxFeePerGasMedium
}

const inclusionThreshold = 0.95

type TransactionEstimation int

const (
	Unknown TransactionEstimation = iota
	LessThanOneMinute
	LessThanThreeMinutes
	LessThanFiveMinutes
	MoreThanFiveMinutes
)

type FeeHistory struct {
	BaseFeePerGas []string `json:"baseFeePerGas"`
}

type FeeManager struct {
	RPCClient *rpc.Client
}

func weiToGwei(val *big.Int) *big.Float {
	result := new(big.Float)
	result.SetInt(val)

	unit := new(big.Int)
	unit.SetInt64(params.GWei)

	return result.Quo(result, new(big.Float).SetInt(unit))
}

func gweiToEth(val *big.Float) *big.Float {
	return new(big.Float).Quo(val, big.NewFloat(1000000000))
}

func gweiToWei(val *big.Float) *big.Int {
	res, _ := new(big.Float).Mul(val, big.NewFloat(1000000000)).Int(nil)
	return res
}

func (f *FeeManager) suggestedFees(ctx context.Context, chainID uint64) (*SuggestedFees, error) {
	backend, err := f.RPCClient.EthClient(chainID)
	if err != nil {
		return nil, err
	}
	gasPrice, err := backend.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}
	maxPriorityFeePerGas, err := backend.SuggestGasTipCap(ctx)
	if err != nil {
		return &SuggestedFees{
			GasPrice:             weiToGwei(gasPrice),
			BaseFee:              big.NewFloat(0),
			MaxPriorityFeePerGas: big.NewFloat(0),
			MaxFeePerGasLow:      big.NewFloat(0),
			MaxFeePerGasMedium:   big.NewFloat(0),
			MaxFeePerGasHigh:     big.NewFloat(0),
			EIP1559Enabled:       false,
		}, nil
	}

	header, err := backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	config := params.MainnetChainConfig
	baseFee := misc.CalcBaseFee(config, header)

	fees, err := f.getFeeHistorySorted(chainID)
	if err != nil {
		return &SuggestedFees{
			GasPrice:             weiToGwei(gasPrice),
			BaseFee:              weiToGwei(baseFee),
			MaxPriorityFeePerGas: weiToGwei(maxPriorityFeePerGas),
			MaxFeePerGasLow:      weiToGwei(maxPriorityFeePerGas),
			MaxFeePerGasMedium:   weiToGwei(maxPriorityFeePerGas),
			MaxFeePerGasHigh:     weiToGwei(maxPriorityFeePerGas),
			EIP1559Enabled:       false,
		}, nil
	}

	perc10 := fees[int64(0.1*float64(len(fees)))-1]
	perc20 := fees[int64(0.2*float64(len(fees)))-1]

	var maxFeePerGasMedium *big.Int
	if baseFee.Cmp(perc20) >= 0 {
		maxFeePerGasMedium = baseFee
	} else {
		maxFeePerGasMedium = perc20
	}

	if maxPriorityFeePerGas.Cmp(maxFeePerGasMedium) > 0 {
		maxFeePerGasMedium = maxPriorityFeePerGas
	}

	maxFeePerGasHigh := new(big.Int).Mul(maxPriorityFeePerGas, big.NewInt(2))
	twoTimesBaseFee := new(big.Int).Mul(baseFee, big.NewInt(2))
	if twoTimesBaseFee.Cmp(maxFeePerGasHigh) > 0 {
		maxFeePerGasHigh = twoTimesBaseFee
	}

	return &SuggestedFees{
		GasPrice:             weiToGwei(gasPrice),
		BaseFee:              weiToGwei(baseFee),
		MaxPriorityFeePerGas: weiToGwei(maxPriorityFeePerGas),
		MaxFeePerGasLow:      weiToGwei(perc10),
		MaxFeePerGasMedium:   weiToGwei(maxFeePerGasMedium),
		MaxFeePerGasHigh:     weiToGwei(maxFeePerGasHigh),
		EIP1559Enabled:       true,
	}, nil
}

func (f *FeeManager) transactionEstimatedTime(ctx context.Context, chainID uint64, maxFeePerGas *big.Float) TransactionEstimation {
	fees, err := f.getFeeHistorySorted(chainID)
	if err != nil {
		return Unknown
	}

	maxFeePerGasWei := gweiToWei(maxFeePerGas)
	// pEvent represents the probability of the transaction being included in a block,
	// we assume this one is static over time, in reality it is not.
	pEvent := 0.0
	for idx, fee := range fees {
		if fee.Cmp(maxFeePerGasWei) == 1 || idx == len(fees)-1 {
			pEvent = float64(idx) / float64(len(fees))
			break
		}
	}

	// Probability of next 4 blocks including the transaction (less than 1 minute)
	// Generalising the formula: P(AUB) = P(A) + P(B) - P(A∩B) for 4 events and in our context P(A) == P(B) == pEvent
	// The factors are calculated using the combinations formula
	probability := pEvent*4 - 6*(math.Pow(pEvent, 2)) + 4*(math.Pow(pEvent, 3)) - (math.Pow(pEvent, 4))
	if probability >= inclusionThreshold {
		return LessThanOneMinute
	}

	// Probability of next 12 blocks including the transaction (less than 5 minutes)
	// Generalising the formula: P(AUB) = P(A) + P(B) - P(A∩B) for 20 events and in our context P(A) == P(B) == pEvent
	// The factors are calculated using the combinations formula
	probability = pEvent*12 -
		66*(math.Pow(pEvent, 2)) +
		220*(math.Pow(pEvent, 3)) -
		495*(math.Pow(pEvent, 4)) +
		792*(math.Pow(pEvent, 5)) -
		924*(math.Pow(pEvent, 6)) +
		792*(math.Pow(pEvent, 7)) -
		495*(math.Pow(pEvent, 8)) +
		220*(math.Pow(pEvent, 9)) -
		66*(math.Pow(pEvent, 10)) +
		12*(math.Pow(pEvent, 11)) -
		math.Pow(pEvent, 12)
	if probability >= inclusionThreshold {
		return LessThanThreeMinutes
	}

	// Probability of next 20 blocks including the transaction (less than 5 minutes)
	// Generalising the formula: P(AUB) = P(A) + P(B) - P(A∩B) for 20 events and in our context P(A) == P(B) == pEvent
	// The factors are calculated using the combinations formula
	probability = pEvent*20 -
		190*(math.Pow(pEvent, 2)) +
		1140*(math.Pow(pEvent, 3)) -
		4845*(math.Pow(pEvent, 4)) +
		15504*(math.Pow(pEvent, 5)) -
		38760*(math.Pow(pEvent, 6)) +
		77520*(math.Pow(pEvent, 7)) -
		125970*(math.Pow(pEvent, 8)) +
		167960*(math.Pow(pEvent, 9)) -
		184756*(math.Pow(pEvent, 10)) +
		167960*(math.Pow(pEvent, 11)) -
		125970*(math.Pow(pEvent, 12)) +
		77520*(math.Pow(pEvent, 13)) -
		38760*(math.Pow(pEvent, 14)) +
		15504*(math.Pow(pEvent, 15)) -
		4845*(math.Pow(pEvent, 16)) +
		1140*(math.Pow(pEvent, 17)) -
		190*(math.Pow(pEvent, 18)) +
		20*(math.Pow(pEvent, 19)) -
		math.Pow(pEvent, 20)
	if probability >= inclusionThreshold {
		return LessThanFiveMinutes
	}

	return MoreThanFiveMinutes
}

func (f *FeeManager) getFeeHistorySorted(chainID uint64) ([]*big.Int, error) {
	var feeHistory FeeHistory

	err := f.RPCClient.Call(&feeHistory, chainID, "eth_feeHistory", 101, "latest", nil)
	if err != nil {
		return nil, err
	}

	fees := []*big.Int{}
	for _, fee := range feeHistory.BaseFeePerGas {
		i := new(big.Int)
		i.SetString(strings.Replace(fee, "0x", "", 1), 16)
		fees = append(fees, i)
	}

	sort.Slice(fees, func(i, j int) bool { return fees[i].Cmp(fees[j]) < 0 })
	return fees, nil
}

func (f *FeeManager) getL1Fee(ctx context.Context, chainID uint64, tx *ethTypes.Transaction) (uint64, error) {

	ethClient, err := f.RPCClient.EthClient(chainID)
	if err != nil {
		return 0, err
	}

	contractAddress, err := gaspriceoracle.ContractAddress(chainID)
	if err != nil {
		return 0, err
	}

	contract, err := gaspriceoracle.NewGaspriceoracleCaller(contractAddress, ethClient)
	if err != nil {
		return 0, err
	}

	callOpt := &bind.CallOpts{}

	data, err := tx.MarshalBinary()
	if err != nil {
		return 0, err
	}

	result, err := contract.GetL1Fee(callOpt, data)
	if err != nil {
		return 0, err
	}

	return result.Uint64(), nil
}
