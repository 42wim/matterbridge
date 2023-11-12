package torrent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent/metainfo"
	pp "github.com/anacrolix/torrent/peer_protocol"
	"github.com/anacrolix/torrent/webseed"
)

type webseedPeer struct {
	// First field for stats alignment.
	peer           Peer
	client         webseed.Client
	activeRequests map[Request]webseed.Request
	requesterCond  sync.Cond
	// Number of requester routines.
	maxRequests int
}

var _ peerImpl = (*webseedPeer)(nil)

func (me *webseedPeer) connStatusString() string {
	return me.client.Url
}

func (ws *webseedPeer) String() string {
	return fmt.Sprintf("webseed peer for %q", ws.client.Url)
}

func (ws *webseedPeer) onGotInfo(info *metainfo.Info) {
	ws.client.SetInfo(info)
	// There should be probably be a callback in Client instead, so it can remove pieces at its whim
	// too.
	ws.client.Pieces.Iterate(func(x uint32) bool {
		ws.peer.t.incPieceAvailability(pieceIndex(x))
		return true
	})
}

func (ws *webseedPeer) writeInterested(interested bool) bool {
	return true
}

func (ws *webseedPeer) _cancel(r RequestIndex) bool {
	if active, ok := ws.activeRequests[ws.peer.t.requestIndexToRequest(r)]; ok {
		active.Cancel()
		// The requester is running and will handle the result.
		return true
	}
	// There should be no requester handling this, so no further events will occur.
	return false
}

func (ws *webseedPeer) intoSpec(r Request) webseed.RequestSpec {
	return webseed.RequestSpec{ws.peer.t.requestOffset(r), int64(r.Length)}
}

func (ws *webseedPeer) _request(r Request) bool {
	ws.requesterCond.Signal()
	return true
}

func (ws *webseedPeer) doRequest(r Request) error {
	webseedRequest := ws.client.NewRequest(ws.intoSpec(r))
	ws.activeRequests[r] = webseedRequest
	err := func() error {
		ws.requesterCond.L.Unlock()
		defer ws.requesterCond.L.Lock()
		return ws.requestResultHandler(r, webseedRequest)
	}()
	delete(ws.activeRequests, r)
	return err
}

func (ws *webseedPeer) requester(i int) {
	ws.requesterCond.L.Lock()
	defer ws.requesterCond.L.Unlock()
start:
	for !ws.peer.closed.IsSet() {
		restart := false
		ws.peer.requestState.Requests.Iterate(func(x uint32) bool {
			r := ws.peer.t.requestIndexToRequest(x)
			if _, ok := ws.activeRequests[r]; ok {
				return true
			}
			err := ws.doRequest(r)
			ws.requesterCond.L.Unlock()
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("requester %v: error doing webseed request %v: %v", i, r, err)
			}
			restart = true
			if errors.Is(err, webseed.ErrTooFast) {
				time.Sleep(time.Duration(rand.Int63n(int64(10 * time.Second))))
			}
			ws.requesterCond.L.Lock()
			return false
		})
		if restart {
			goto start
		}
		ws.requesterCond.Wait()
	}
}

func (ws *webseedPeer) connectionFlags() string {
	return "WS"
}

// TODO: This is called when banning peers. Perhaps we want to be able to ban webseeds too. We could
// return bool if this is even possible, and if it isn't, skip to the next drop candidate.
func (ws *webseedPeer) drop() {}

func (ws *webseedPeer) handleUpdateRequests() {
	// Because this is synchronous, webseed peers seem to get first dibs on newly prioritized
	// pieces.
	go func() {
		ws.peer.t.cl.lock()
		defer ws.peer.t.cl.unlock()
		ws.peer.maybeUpdateActualRequestState()
	}()
}

func (ws *webseedPeer) onClose() {
	ws.peer.logger.Levelf(log.Debug, "closing")
	// Just deleting them means we would have to manually cancel active requests.
	ws.peer.cancelAllRequests()
	ws.peer.t.iterPeers(func(p *Peer) {
		if p.isLowOnRequests() {
			p.updateRequests("webseedPeer.onClose")
		}
	})
	ws.requesterCond.Broadcast()
}

func (ws *webseedPeer) requestResultHandler(r Request, webseedRequest webseed.Request) error {
	result := <-webseedRequest.Result
	close(webseedRequest.Result) // one-shot
	// We do this here rather than inside receiveChunk, since we want to count errors too. I'm not
	// sure if we can divine which errors indicate cancellation on our end without hitting the
	// network though.
	if len(result.Bytes) != 0 || result.Err == nil {
		// Increment ChunksRead and friends
		ws.peer.doChunkReadStats(int64(len(result.Bytes)))
	}
	ws.peer.readBytes(int64(len(result.Bytes)))
	ws.peer.t.cl.lock()
	defer ws.peer.t.cl.unlock()
	if ws.peer.t.closed.IsSet() {
		return nil
	}
	err := result.Err
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
		case errors.Is(err, webseed.ErrTooFast):
		case ws.peer.closed.IsSet():
		default:
			ws.peer.logger.Printf("Request %v rejected: %v", r, result.Err)
			// // Here lies my attempt to extract something concrete from Go's error system. RIP.
			// cfg := spew.NewDefaultConfig()
			// cfg.DisableMethods = true
			// cfg.Dump(result.Err)
			log.Printf("closing %v", ws)
			ws.peer.close()
		}
		if !ws.peer.remoteRejectedRequest(ws.peer.t.requestIndexFromRequest(r)) {
			panic("invalid reject")
		}
		return err
	}
	err = ws.peer.receiveChunk(&pp.Message{
		Type:  pp.Piece,
		Index: r.Index,
		Begin: r.Begin,
		Piece: result.Bytes,
	})
	if err != nil {
		panic(err)
	}
	return err
}

func (me *webseedPeer) peerPieces() *roaring.Bitmap {
	return &me.client.Pieces
}

func (cn *webseedPeer) peerHasAllPieces() (all, known bool) {
	if !cn.peer.t.haveInfo() {
		return true, false
	}
	return cn.client.Pieces.GetCardinality() == uint64(cn.peer.t.numPieces()), true
}
