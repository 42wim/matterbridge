package rpcfilters

import (
	"fmt"
	"sort"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	defaultCacheSize = 20
)

type cacheRecord struct {
	block uint64
	hash  common.Hash
	logs  []types.Log
}

func newCache(size int) *cache {
	return &cache{
		records: make([]cacheRecord, 0, size),
		size:    size,
	}
}

type cache struct {
	mu      sync.RWMutex
	size    int // length of the records
	records []cacheRecord
}

// add inserts logs into cache and returns added and replaced logs.
// replaced logs with will be returned with Removed=true.
func (c *cache) add(logs []types.Log) (added, replaced []types.Log, err error) {
	if len(logs) == 0 {
		return nil, nil, nil
	}
	aggregated := aggregateLogs(logs, c.size) // size doesn't change
	if len(aggregated) == 0 {
		return nil, nil, nil
	}
	if err := checkLogsAreInOrder(aggregated); err != nil {
		return nil, nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	// find common block. e.g. [3,4] and [1,2,3,4] = 3
	last := 0
	if len(c.records) > 0 {
		last = len(c.records) - 1
		for aggregated[0].block < c.records[last].block && last > 0 {
			last--
		}
	}
	c.records, added, replaced = merge(last, c.records, aggregated)
	if lth := len(c.records); lth > c.size {
		copy(c.records, c.records[lth-c.size:])
	}
	return added, replaced, nil
}

func (c *cache) earliestBlockNum() uint64 {
	if len(c.records) == 0 {
		return 0
	}
	return c.records[0].block
}

func checkLogsAreInOrder(records []cacheRecord) error {
	for prev, i := 0, 1; i < len(records); i++ {
		if records[prev].block == records[i].block-1 {
			prev = i
		} else {
			return fmt.Errorf(
				"logs must be delivered straight in order. gaps between blocks '%d' and '%d'",
				records[prev].block, records[i].block,
			)
		}
	}
	return nil
}

// merge merges received records into old slice starting at provided position, example:
// [1, 2, 3]
//
//	[2, 3, 4]
//
// [1, 2, 3, 4]
// if hash doesn't match previously received hash - such block was removed due to reorg
// logs that were a part of that block will be returned with Removed set to true
func merge(last int, old, received []cacheRecord) ([]cacheRecord, []types.Log, []types.Log) {
	var (
		added, replaced []types.Log
		block           uint64
		hash            common.Hash
	)
	for i := range received {
		record := received[i]
		if last < len(old) {
			block = old[last].block
			hash = old[last].hash
		}
		if record.block > block {
			// simply add new records
			added = append(added, record.logs...)
			old = append(old, record)
		} else if record.hash != hash && record.block == block {
			// record hash is not equal to previous record hash at the same height
			// replace record in hash and add logs as replaced
			replaced = append(replaced, old[last].logs...)
			added = append(added, record.logs...)
			old[last] = record
		}
		last++
	}
	return old, added, replaced
}

// aggregateLogs creates at most requested amount of cacheRecords from provided logs.
// cacheRecords will be sorted in ascending order, starting from lowest block to highest.
func aggregateLogs(logs []types.Log, limit int) []cacheRecord {
	// sort in reverse order, so that iteration will start from latest blocks
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].BlockNumber > logs[j].BlockNumber
	})
	rst := make([]cacheRecord, limit)
	pos, start := len(rst)-1, 0
	var hash common.Hash
	for i := range logs {
		log := logs[i]
		if (hash != common.Hash{}) && hash != log.BlockHash {
			rst[pos].logs = logs[start:i]
			start = i
			if pos-1 < 0 {
				break
			}
			pos--
		}
		rst[pos].logs = logs[start:]
		rst[pos].block = log.BlockNumber
		rst[pos].hash = log.BlockHash
		hash = log.BlockHash
	}
	return rst[pos:]
}
