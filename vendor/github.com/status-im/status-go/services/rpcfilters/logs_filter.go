package rpcfilters

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

type logsFilter struct {
	mu   sync.RWMutex
	logs []types.Log
	crit ethereum.FilterQuery // will be modified and different from original

	originalCrit ethereum.FilterQuery // not modified version of the criteria

	logsCache *cache

	id    rpc.ID
	timer *time.Timer

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func (f *logsFilter) criteria() ethereum.FilterQuery {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.crit
}

func (f *logsFilter) add(data interface{}) error {
	logs, ok := data.([]types.Log)
	if !ok {
		return fmt.Errorf("can't cast %v to types.Log", data)
	}
	filtered := filterLogs(logs, f.crit)
	if len(filtered) > 0 {
		f.mu.Lock()
		defer f.mu.Unlock()
		added, replaced, err := f.logsCache.add(filtered)
		if err != nil {
			return err
		}
		for _, log := range replaced {
			log.Removed = true
			f.logs = append(f.logs, log)
		}
		if len(added) > 0 {
			f.logs = append(f.logs, added...)
		}
		// if there was no replaced logs - keep polling only latest logs
		if len(replaced) == 0 {
			adjustFromBlock(&f.crit)
		} else {
			// otherwise poll earliest known block in cache
			earliest := f.logsCache.earliestBlockNum()
			if earliest != 0 {
				f.crit.FromBlock = new(big.Int).SetUint64(earliest)
			}
		}
	}
	return nil
}

func (f *logsFilter) pop() interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	rst := f.logs
	f.logs = nil
	return rst
}

func (f *logsFilter) stop() {
	select {
	case <-f.done:
		return
	default:
		close(f.done)
		if f.cancel != nil {
			f.cancel()
		}
	}
}

func (f *logsFilter) deadline() *time.Timer {
	return f.timer
}

// adjustFromBlock adjusts crit.FromBlock to latest to avoid querying same logs.
func adjustFromBlock(crit *ethereum.FilterQuery) {
	latest := big.NewInt(rpc.LatestBlockNumber.Int64())
	// don't adjust if filter is not interested in newer blocks
	if crit.ToBlock != nil && crit.ToBlock.Cmp(latest) == 1 {
		return
	}
	// don't adjust if from block is already pending
	if crit.FromBlock != nil && crit.FromBlock.Cmp(latest) == -1 {
		return
	}
	crit.FromBlock = latest
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}
	return false
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []types.Log, crit ethereum.FilterQuery) (
	ret []types.Log) {
	for _, log := range logs {
		if matchLog(log, crit) {
			ret = append(ret, log)
		}
	}
	return
}

func matchLog(log types.Log, crit ethereum.FilterQuery) bool {
	if crit.FromBlock != nil && crit.FromBlock.Int64() >= 0 && crit.FromBlock.Uint64() > log.BlockNumber {
		return false
	}
	if crit.ToBlock != nil && crit.ToBlock.Int64() >= 0 && crit.ToBlock.Uint64() < log.BlockNumber {
		return false
	}
	if len(crit.Addresses) > 0 && !includes(crit.Addresses, log.Address) {
		return false
	}
	if len(crit.Topics) > len(log.Topics) {
		return false
	}
	return matchTopics(log, crit.Topics)
}

func matchTopics(log types.Log, topics [][]common.Hash) bool {
	for i, sub := range topics {
		match := len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if log.Topics[i] == topic {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}
	return true
}
