package rpcstats

import (
	"sync"
)

type RPCUsageStats struct {
	total            uint
	counterPerMethod map[string]uint
	rw               sync.RWMutex
}

var stats *RPCUsageStats

func getInstance() *RPCUsageStats {
	if stats == nil {
		stats = &RPCUsageStats{
			total:            0,
			counterPerMethod: map[string]uint{},
		}
	}
	return stats
}

func getStats() (uint, map[string]uint) {
	stats := getInstance()
	stats.rw.RLock()
	defer stats.rw.RUnlock()
	return stats.total, stats.counterPerMethod
}

func resetStats() {
	stats := getInstance()
	stats.rw.Lock()
	defer stats.rw.Unlock()

	stats.total = 0
	stats.counterPerMethod = map[string]uint{}
}

func CountCall(method string) {
	stats := getInstance()
	stats.rw.Lock()
	defer stats.rw.Unlock()

	stats.total++
	stats.counterPerMethod[method]++
}
