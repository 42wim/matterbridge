package traversal

import (
	"context"
	"sync/atomic"

	"github.com/anacrolix/chansync/events"
	"github.com/anacrolix/dht/v2/containers"
	"github.com/anacrolix/sync"

	"github.com/anacrolix/chansync"

	"github.com/anacrolix/dht/v2/int160"
	k_nearest_nodes "github.com/anacrolix/dht/v2/k-nearest-nodes"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/dht/v2/types"
)

type QueryResult struct {
	// A node that should be considered for a closest entry.
	ResponseFrom *krpc.NodeInfo
	// Data associated with a closest node.
	ClosestData interface{}
	Nodes       []krpc.NodeInfo
	Nodes6      []krpc.NodeInfo
}

type OperationInput struct {
	Target     krpc.ID
	Alpha      int
	K          int
	DoQuery    func(context.Context, krpc.NodeAddr) QueryResult
	NodeFilter func(types.AddrMaybeId) bool
}

type defaultsAppliedOperationInput OperationInput

func Start(input OperationInput) *Operation {
	herp := defaultsAppliedOperationInput(input)
	if herp.Alpha == 0 {
		herp.Alpha = 3
	}
	if herp.K == 0 {
		herp.K = 8
	}
	if herp.NodeFilter == nil {
		herp.NodeFilter = func(types.AddrMaybeId) bool {
			return true
		}
	}
	targetInt160 := herp.Target.Int160()
	op := &Operation{
		targetInt160: targetInt160,
		input:        herp,
		queried:      make(map[addrString]struct{}),
		closest:      k_nearest_nodes.New(targetInt160, herp.K),
		unqueried:    containers.NewImmutableAddrMaybeIdsByDistance(targetInt160),
	}
	go op.run()
	return op
}

type addrString string

type Operation struct {
	stats        Stats
	mu           sync.Mutex
	unqueried    containers.AddrMaybeIdsByDistance
	queried      map[addrString]struct{}
	closest      k_nearest_nodes.Type
	targetInt160 int160.T
	input        defaultsAppliedOperationInput
	outstanding  int
	cond         chansync.BroadcastCond
	stalled      chansync.LevelTrigger
	stopping     chansync.SetOnce
	stopped      chansync.SetOnce
}

func (op *Operation) Stats() *Stats {
	return &op.stats
}

func (op *Operation) Stop() {
	if op.stopping.Set() {
		go func() {
			defer op.stopped.Set()
			op.mu.Lock()
			defer op.mu.Unlock()
			for {
				if op.outstanding == 0 {
					break
				}
				cond := op.cond.Signaled()
				op.mu.Unlock()
				<-cond
				op.mu.Lock()
			}
		}()
	}
}

func (op *Operation) Stopped() events.Done {
	return op.stopped.Done()
}

func (op *Operation) Stalled() events.Active {
	return op.stalled.Active()
}

func (op *Operation) AddNodes(nodes []types.AddrMaybeId) (added int) {
	op.mu.Lock()
	defer op.mu.Unlock()
	before := op.unqueried.Len()
	for _, n := range nodes {
		if _, ok := op.queried[addrString(n.Addr.String())]; ok {
			continue
		}
		if !op.input.NodeFilter(n) {
			continue
		}
		op.unqueried = op.unqueried.Add(n)
	}
	op.cond.Broadcast()
	return op.unqueried.Len() - before
}

func (op *Operation) markQueried(addr krpc.NodeAddr) {
	op.queried[addrString(addr.String())] = struct{}{}
}

func (op *Operation) closestUnqueried() (ret types.AddrMaybeId) {
	return op.unqueried.Next()
}

func (op *Operation) popClosestUnqueried() types.AddrMaybeId {
	ret := op.closestUnqueried()
	op.unqueried = op.unqueried.Delete(ret)
	return ret
}

func (op *Operation) haveQuery() bool {
	if op.unqueried.Len() == 0 {
		return false
	}
	if !op.closest.Full() {
		return true
	}
	cu := op.closestUnqueried()
	if cu.Id == nil {
		return false
	}
	return cu.Id.Distance(op.targetInt160).Cmp(op.closest.Farthest().ID.Int160().Distance(op.targetInt160)) <= 0
}

func (op *Operation) run() {
	defer close(op.stalled.Signal())
	op.mu.Lock()
	defer op.mu.Unlock()
	for {
		if op.stopping.IsSet() {
			return
		}
		for op.outstanding < op.input.Alpha && op.haveQuery() {
			op.startQuery()
		}
		var stalled events.Signal
		if (!op.haveQuery() || op.input.Alpha == 0) && op.outstanding == 0 {
			stalled = op.stalled.Signal()
		}
		queryCondSignaled := op.cond.Signaled()
		op.mu.Unlock()
		select {
		case stalled <- struct{}{}:
		case <-op.stopping.Done():
		case <-queryCondSignaled:
		}
		op.mu.Lock()
	}
}

func (op *Operation) addClosest(node krpc.NodeInfo, data interface{}) {
	var ami types.AddrMaybeId
	ami.FromNodeInfo(node)
	if !op.input.NodeFilter(ami) {
		return
	}
	op.closest = op.closest.Push(k_nearest_nodes.Elem{
		Key:  node,
		Data: data,
	})
}

func (op *Operation) Closest() *k_nearest_nodes.Type {
	return &op.closest
}

func (op *Operation) startQuery() {
	a := op.popClosestUnqueried()
	op.markQueried(a.Addr)
	op.outstanding++
	go func() {
		defer func() {
			op.mu.Lock()
			defer op.mu.Unlock()
			op.outstanding--
			op.cond.Broadcast()
		}()
		// log.Printf("traversal querying %v", a)
		atomic.AddUint32(&op.stats.NumAddrsTried, 1)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			select {
			case <-ctx.Done():
			case <-op.stopping.Done():
				cancel()
			}
		}()
		res := op.input.DoQuery(ctx, a.Addr)
		cancel()
		if res.ResponseFrom != nil {
			func() {
				op.mu.Lock()
				defer op.mu.Unlock()
				atomic.AddUint32(&op.stats.NumResponses, 1)
				op.addClosest(*res.ResponseFrom, res.ClosestData)
			}()
		}
		op.AddNodes(types.AddrMaybeIdSliceFromNodeInfoSlice(res.Nodes))
		op.AddNodes(types.AddrMaybeIdSliceFromNodeInfoSlice(res.Nodes6))
	}()
}
