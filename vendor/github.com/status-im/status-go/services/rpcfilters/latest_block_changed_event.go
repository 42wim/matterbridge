package rpcfilters

import (
	"errors"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

const (
	defaultTickerPeriod      = 3 * time.Second
	defaultReportHistorySize = 20
)

// ringArray represents a thread-safe capped collection of hashes.
type ringArray struct {
	mu           sync.Mutex
	maxCount     int
	currentIndex int
	blocks       []common.Hash
}

func newRingArray(maxCount int) *ringArray {
	return &ringArray{
		maxCount: maxCount,
		blocks:   make([]common.Hash, maxCount),
	}
}

// TryAddUnique adds a hash to the array if the array doesn't have it.
// Returns true if the element was added.
func (r *ringArray) TryAddUnique(hash common.Hash) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.has(hash) {
		return false
	}

	r.blocks[r.currentIndex] = hash
	r.currentIndex++
	if r.currentIndex >= len(r.blocks) {
		r.currentIndex = 0
	}
	return true
}

// has returns `true` if the hash is in the array.
// It has linear complexity but on short arrays it isn't worth optimizing.
func (r *ringArray) has(hash common.Hash) bool {
	for _, h := range r.blocks {
		if h == hash {
			return true
		}
	}
	return false
}

// latestBlockChangedEvent represents an event that one can subscribe to
type latestBlockChangedEvent struct {
	sxMu sync.Mutex
	sx   map[int]chan common.Hash

	reportedBlocks *ringArray

	provider     latestBlockProvider
	quit         chan struct{}
	tickerPeriod time.Duration
}

func (e *latestBlockChangedEvent) Start() error {
	if e.quit != nil {
		return errors.New("latest block changed event is already started")
	}

	e.quit = make(chan struct{})

	go func() {
		ticker := time.NewTicker(e.tickerPeriod)
		for {
			select {
			case <-ticker.C:
				if e.numberOfSubscriptions() == 0 {
					continue
				}
				latestBlock, err := e.provider.GetLatestBlock()
				if err != nil {
					log.Error("error while receiving latest block", "error", err)
					continue
				}

				e.processLatestBlock(latestBlock)
			case <-e.quit:
				return
			}
		}
	}()

	return nil
}

func (e *latestBlockChangedEvent) numberOfSubscriptions() int {
	e.sxMu.Lock()
	defer e.sxMu.Unlock()
	return len(e.sx)
}

func (e *latestBlockChangedEvent) processLatestBlock(latestBlock blockInfo) {
	// if we received the hash we already received before, don't add it
	if !e.reportedBlocks.TryAddUnique(latestBlock.Hash) {
		return
	}

	e.sxMu.Lock()
	defer e.sxMu.Unlock()

	for _, channel := range e.sx {
		channel <- latestBlock.Hash
	}
}

func (e *latestBlockChangedEvent) Stop() {
	if e.quit == nil {
		return
	}

	select {
	case <-e.quit:
		e.quit = nil
		return
	default:
		close(e.quit)
	}

	e.quit = nil
}

func (e *latestBlockChangedEvent) Subscribe() (int, interface{}) {
	e.sxMu.Lock()
	defer e.sxMu.Unlock()

	channel := make(chan common.Hash)
	id := len(e.sx)
	e.sx[id] = channel
	return id, channel
}

func (e *latestBlockChangedEvent) Unsubscribe(id int) {
	e.sxMu.Lock()
	defer e.sxMu.Unlock()

	delete(e.sx, id)
}

func newLatestBlockChangedEvent(provider latestBlockProvider) *latestBlockChangedEvent {
	return &latestBlockChangedEvent{
		sx:             make(map[int]chan common.Hash),
		provider:       provider,
		reportedBlocks: newRingArray(defaultReportHistorySize),
		tickerPeriod:   defaultTickerPeriod,
	}
}
