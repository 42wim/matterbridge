package rendezvous

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"github.com/multiformats/go-multiaddr"
	"github.com/waku-org/go-waku/waku/v2/utils"
)

type RendezvousPointIterator struct {
	rendezvousPoints []*RendezvousPoint
}

// NewRendezvousPointIterator creates an iterator with a backoff mechanism to use random rendezvous points taking into account successful/unsuccesful connection attempts
func NewRendezvousPointIterator(rendezvousPoints []multiaddr.Multiaddr) *RendezvousPointIterator {
	var rendevousPoints []*RendezvousPoint
	for _, rp := range rendezvousPoints {
		peerID, err := utils.GetPeerID(rp)
		if err == nil {
			rendevousPoints = append(rendevousPoints, NewRendezvousPoint(peerID))
		}
	}

	return &RendezvousPointIterator{
		rendezvousPoints: rendevousPoints,
	}
}

// RendezvousPoints returns the list of rendezvous points registered in this iterator
func (r *RendezvousPointIterator) RendezvousPoints() []*RendezvousPoint {
	return r.rendezvousPoints
}

// Next will return a channel that will be triggered as soon as the next rendevous point is available to be used (depending on backoff time)
func (r *RendezvousPointIterator) Next(ctx context.Context) <-chan *RendezvousPoint {
	var dialableRP []*RendezvousPoint
	now := time.Now()
	for _, rp := range r.rendezvousPoints {
		if now.After(rp.NextTry()) {
			dialableRP = append(dialableRP, rp)
		}
	}

	result := make(chan *RendezvousPoint, 1)

	if len(dialableRP) > 0 {
		result <- r.rendezvousPoints[rand.Intn(len(r.rendezvousPoints))] // nolint: gosec
	} else {
		if len(r.rendezvousPoints) > 0 {
			sort.Slice(r.rendezvousPoints, func(i, j int) bool {
				return r.rendezvousPoints[i].nextTry.Before(r.rendezvousPoints[j].nextTry)
			})

			tryIn := r.rendezvousPoints[0].NextTry().Sub(now)
			timer := time.NewTimer(tryIn)
			defer timer.Stop()

			select {
			case <-ctx.Done():
				break
			case <-timer.C:
				result <- r.rendezvousPoints[0]
			}
		}
	}

	close(result)
	return result
}
