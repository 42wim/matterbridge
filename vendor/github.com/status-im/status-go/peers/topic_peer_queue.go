package peers

import (
	"container/heap"
)

type peerInfoItem struct {
	*peerInfo
	index int
}

type peerPriorityQueue []*peerInfoItem

var _ heap.Interface = (*peerPriorityQueue)(nil)

func (q peerPriorityQueue) Len() int { return len(q) }

func (q peerPriorityQueue) Less(i, j int) bool {
	return q[i].discoveredTime.After(q[j].discoveredTime)
}

func (q peerPriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *peerPriorityQueue) Push(x interface{}) {
	item := x.(*peerInfoItem)
	item.index = len(*q)
	*q = append(*q, item)
}

func (q *peerPriorityQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	item.index = -1
	*q = old[0 : n-1]
	return item
}
