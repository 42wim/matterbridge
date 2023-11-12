// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type RollupGasData struct {
	Zeroes, Ones uint64
}

/*
  // Unused, depends on other unneeded changes
	func (r RollupGasData) DataGas(time uint64, cfg *params.ChainConfig) (gas uint64) {
		gas = r.Zeroes * params.TxDataZeroGas
		if cfg.IsRegolith(time) {
			gas += r.Ones * params.TxDataNonZeroGasEIP2028
		} else {
			gas += (r.Ones + 68) * params.TxDataNonZeroGasEIP2028
		}
		return gas
	}
*/

type StateGetter interface {
	GetState(common.Address, common.Hash) common.Hash
}

// L1CostFunc is used in the state transition to determine the cost of a rollup message.
// Returns nil if there is no cost.
type L1CostFunc func(blockNum uint64, blockTime uint64, dataGas RollupGasData, isDepositTx bool) *big.Int

var (
	L1BaseFeeSlot = common.BigToHash(big.NewInt(1))
	OverheadSlot  = common.BigToHash(big.NewInt(5))
	ScalarSlot    = common.BigToHash(big.NewInt(6))
)

var L1BlockAddr = common.HexToAddress("0x4200000000000000000000000000000000000015")

// NewL1CostFunc returns a function used for calculating L1 fee cost.
// This depends on the oracles because gas costs can change over time.
// It returns nil if there is no applicable cost function.
/*
// Unused, depends on other unneeded changes
func NewL1CostFunc(config *params.ChainConfig, statedb StateGetter) L1CostFunc {
	cacheBlockNum := ^uint64(0)
	var l1BaseFee, overhead, scalar *big.Int
	return func(blockNum uint64, blockTime uint64, dataGas RollupGasData, isDepositTx bool) *big.Int {
		rollupDataGas := dataGas.DataGas(blockTime, config) // Only fake txs for RPC view-calls are 0.
		if config.Optimism == nil || isDepositTx || rollupDataGas == 0 {
			return nil
		}
		if blockNum != cacheBlockNum {
			l1BaseFee = statedb.GetState(L1BlockAddr, L1BaseFeeSlot).Big()
			overhead = statedb.GetState(L1BlockAddr, OverheadSlot).Big()
			scalar = statedb.GetState(L1BlockAddr, ScalarSlot).Big()
			cacheBlockNum = blockNum
		}
		return L1Cost(rollupDataGas, l1BaseFee, overhead, scalar)
	}
}
*/

func L1Cost(rollupDataGas uint64, l1BaseFee, overhead, scalar *big.Int) *big.Int {
	l1GasUsed := new(big.Int).SetUint64(rollupDataGas)
	l1GasUsed = l1GasUsed.Add(l1GasUsed, overhead)
	l1Cost := l1GasUsed.Mul(l1GasUsed, l1BaseFee)
	l1Cost = l1Cost.Mul(l1Cost, scalar)
	return l1Cost.Div(l1Cost, big.NewInt(1_000_000))
}
