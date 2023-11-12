package blockchainstate

import (
	"context"
	"sync"
	"time"

	"github.com/status-im/status-go/services/wallet/common"
)

type LatestBlockData struct {
	blockNumber   uint64
	timestamp     time.Time
	blockDuration time.Duration
}

type BlockChainState struct {
	blkMu              sync.RWMutex
	latestBlockNumbers map[uint64]LatestBlockData
	sinceFn            func(time.Time) time.Duration
}

func NewBlockChainState() *BlockChainState {
	return &BlockChainState{
		blkMu:              sync.RWMutex{},
		latestBlockNumbers: make(map[uint64]LatestBlockData),
		sinceFn:            time.Since,
	}
}

func (s *BlockChainState) GetEstimatedLatestBlockNumber(ctx context.Context, chainID uint64) (uint64, error) {
	blockNumber, _ := s.estimateLatestBlockNumber(chainID)
	return blockNumber, nil
}

func (s *BlockChainState) SetLastBlockNumber(chainID uint64, blockNumber uint64) {
	blockDuration, found := common.AverageBlockDurationForChain[common.ChainID(chainID)]
	if !found {
		blockDuration = common.AverageBlockDurationForChain[common.ChainID(common.UnknownChainID)]
	}
	s.setLatestBlockDataForChain(chainID, LatestBlockData{
		blockNumber:   blockNumber,
		timestamp:     time.Now(),
		blockDuration: blockDuration,
	})
}

func (s *BlockChainState) setLatestBlockDataForChain(chainID uint64, latestBlockData LatestBlockData) {
	s.blkMu.Lock()
	defer s.blkMu.Unlock()
	s.latestBlockNumbers[chainID] = latestBlockData
}

func (s *BlockChainState) estimateLatestBlockNumber(chainID uint64) (uint64, bool) {
	s.blkMu.RLock()
	defer s.blkMu.RUnlock()
	blockData, ok := s.latestBlockNumbers[chainID]
	if !ok {
		return 0, false
	}
	timeDiff := s.sinceFn(blockData.timestamp)
	return blockData.blockNumber + uint64((timeDiff / blockData.blockDuration)), true
}
