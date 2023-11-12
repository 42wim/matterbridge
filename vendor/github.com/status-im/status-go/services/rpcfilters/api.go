package rpcfilters

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pborman/uuid"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/log"
	getrpc "github.com/ethereum/go-ethereum/rpc"
)

const (
	defaultFilterLivenessPeriod = 5 * time.Minute
	defaultLogsPeriod           = 3 * time.Second
	defaultLogsQueryTimeout     = 10 * time.Second
)

var (
	errFilterNotFound = errors.New("filter not found")
)

type filter interface {
	add(interface{}) error
	pop() interface{}
	stop()
	deadline() *time.Timer
}

type ChainEvent interface {
	Start() error
	Stop()
	Subscribe() (id int, ch interface{})
	Unsubscribe(id int)
}

// PublicAPI represents filter API that is exported to `eth` namespace
type PublicAPI struct {
	filtersMu sync.Mutex
	filters   map[getrpc.ID]filter

	// filterLivenessLoop defines how often timeout loop is executed
	filterLivenessLoop time.Duration
	// filter liveness increased by this period when changes are requested
	filterLivenessPeriod time.Duration

	client  func() ContextCaller
	chainID func() uint64

	latestBlockChangedEvent        *latestBlockChangedEvent
	transactionSentToUpstreamEvent *transactionSentToUpstreamEvent
}

// NewPublicAPI returns a reference to the PublicAPI object
func NewPublicAPI(s *Service) *PublicAPI {
	api := &PublicAPI{
		filters:                        make(map[getrpc.ID]filter),
		latestBlockChangedEvent:        s.latestBlockChangedEvent,
		transactionSentToUpstreamEvent: s.transactionSentToUpstreamEvent,

		client:               func() ContextCaller { return s.rpc.RPCClient() },
		chainID:              func() uint64 { return s.rpc.RPCClient().UpstreamChainID },
		filterLivenessLoop:   defaultFilterLivenessPeriod,
		filterLivenessPeriod: defaultFilterLivenessPeriod + 10*time.Second,
	}
	go api.timeoutLoop(s.quit)
	return api
}

func (api *PublicAPI) timeoutLoop(quit chan struct{}) {
	for {
		select {
		case <-quit:
			return
		case <-time.After(api.filterLivenessLoop):
			api.filtersMu.Lock()
			for id, f := range api.filters {
				deadline := f.deadline()
				if deadline == nil {
					continue
				}
				select {
				case <-deadline.C:
					delete(api.filters, id)
					f.stop()
				default:
					continue
				}
			}
			api.filtersMu.Unlock()
		}
	}
}

func (api *PublicAPI) NewFilter(crit filters.FilterCriteria) (getrpc.ID, error) {
	id := getrpc.ID(uuid.New())
	ctx, cancel := context.WithCancel(context.Background())
	f := &logsFilter{
		id:           id,
		crit:         ethereum.FilterQuery(crit),
		originalCrit: ethereum.FilterQuery(crit),
		done:         make(chan struct{}),
		timer:        time.NewTimer(api.filterLivenessPeriod),
		ctx:          ctx,
		cancel:       cancel,
		logsCache:    newCache(defaultCacheSize),
	}
	api.filtersMu.Lock()
	api.filters[id] = f
	api.filtersMu.Unlock()
	go pollLogs(api.client(), api.chainID(), f, defaultLogsQueryTimeout, defaultLogsPeriod)
	return id, nil
}

// NewBlockFilter is an implemenation of `eth_newBlockFilter` API
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_newblockfilter
func (api *PublicAPI) NewBlockFilter() getrpc.ID {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	f := newHashFilter()
	id := getrpc.ID(uuid.New())

	api.filters[id] = f

	go func() {
		id, si := api.latestBlockChangedEvent.Subscribe()
		s, ok := si.(chan common.Hash)
		if !ok {
			panic("latestBlockChangedEvent returned wrong type")
		}

		defer api.latestBlockChangedEvent.Unsubscribe(id)

		for {
			select {
			case hash := <-s:
				if err := f.add(hash); err != nil {
					log.Error("error adding value to filter", "hash", hash, "error", err)
				}
			case <-f.done:
				return
			}
		}

	}()

	return id
}

// NewPendingTransactionFilter is an implementation of `eth_newPendingTransactionFilter` API
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_newpendingtransactionfilter
func (api *PublicAPI) NewPendingTransactionFilter() getrpc.ID {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	f := newHashFilter()
	id := getrpc.ID(uuid.New())

	api.filters[id] = f

	go func() {
		id, si := api.transactionSentToUpstreamEvent.Subscribe()
		s, ok := si.(chan *PendingTxInfo)
		if !ok {
			panic("transactionSentToUpstreamEvent returned wrong type")
		}
		defer api.transactionSentToUpstreamEvent.Unsubscribe(id)

		for {
			select {
			case hash := <-s:
				if err := f.add(hash); err != nil {
					log.Error("error adding value to filter", "hash", hash, "error", err)
				}
			case <-f.done:
				return
			}
		}
	}()

	return id

}

// UninstallFilter is an implemenation of `eth_uninstallFilter` API
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_uninstallfilter
func (api *PublicAPI) UninstallFilter(id getrpc.ID) bool {
	api.filtersMu.Lock()
	f, found := api.filters[id]
	if found {
		delete(api.filters, id)
	}
	api.filtersMu.Unlock()

	if found {
		f.stop()
	}

	return found
}

// GetFilterLogs returns the logs for the filter with the given id.
// If the filter could not be found an empty array of logs is returned.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getfilterlogs
func (api *PublicAPI) GetFilterLogs(ctx context.Context, id getrpc.ID) ([]types.Log, error) {
	api.filtersMu.Lock()
	f, exist := api.filters[id]
	api.filtersMu.Unlock()
	if !exist {
		return []types.Log{}, errFilterNotFound
	}
	logs, ok := f.(*logsFilter)
	if !ok {
		return []types.Log{}, fmt.Errorf("filter with ID %v is not of logs type", id)
	}
	ctx, cancel := context.WithTimeout(ctx, defaultLogsQueryTimeout)
	defer cancel()
	rst, err := getLogs(ctx, api.client(), api.chainID(), logs.originalCrit)
	return rst, err
}

// GetFilterChanges returns the hashes for the filter with the given id since
// last time it was called. This can be used for polling.
//
// https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_getfilterchanges
func (api *PublicAPI) GetFilterChanges(id getrpc.ID) (interface{}, error) {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	if f, found := api.filters[id]; found {
		deadline := f.deadline()
		if deadline != nil {
			if !deadline.Stop() {
				// timer expired but filter is not yet removed in timeout loop
				// receive timer value and reset timer
				// see https://golang.org/pkg/time/#Timer.Reset
				<-deadline.C
			}
			deadline.Reset(api.filterLivenessPeriod)
		}
		rst := f.pop()
		if rst == nil {
			return []interface{}{}, nil
		}
		return rst, nil
	}
	return []interface{}{}, errFilterNotFound
}
