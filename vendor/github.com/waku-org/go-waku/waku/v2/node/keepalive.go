package node

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	"github.com/waku-org/go-waku/logging"
	"go.uber.org/zap"
)

const maxAllowedPingFailures = 2

// If the difference between the last time the keep alive code was executed and now is greater
// than sleepDectectionIntervalFactor * keepAlivePeriod, force the ping verification to disconnect
// the peers if they don't reply back
const sleepDetectionIntervalFactor = 3

// startKeepAlive creates a go routine that periodically pings connected peers.
// This is necessary because TCP connections are automatically closed due to inactivity,
// and doing a ping will avoid this (with a small bandwidth cost)
func (w *WakuNode) startKeepAlive(ctx context.Context, t time.Duration) {
	defer w.wg.Done()
	w.log.Info("setting up ping protocol", zap.Duration("duration", t))
	ticker := time.NewTicker(t)
	defer ticker.Stop()

	lastTimeExecuted := w.timesource.Now()

	sleepDetectionInterval := int64(t) * sleepDetectionIntervalFactor

	for {
		select {
		case <-ticker.C:
			difference := w.timesource.Now().UnixNano() - lastTimeExecuted.UnixNano()
			forceDisconnectOnPingFailure := false
			if difference > sleepDetectionInterval {
				forceDisconnectOnPingFailure = true
				lastTimeExecuted = w.timesource.Now()
				w.log.Warn("keep alive hasnt been executed recently. Killing connections to peers if ping fails")
				continue
			}

			// Network's peers collection,
			// contains only currently active peers
			pingWg := sync.WaitGroup{}
			peersToPing := w.host.Network().Peers()
			pingWg.Add(len(peersToPing))
			for _, p := range peersToPing {
				if p != w.host.ID() {
					go w.pingPeer(ctx, &pingWg, p, forceDisconnectOnPingFailure)
				}
			}
			pingWg.Wait()

			lastTimeExecuted = w.timesource.Now()
		case <-ctx.Done():
			w.log.Info("stopping ping protocol")
			return
		}
	}
}

func (w *WakuNode) pingPeer(ctx context.Context, wg *sync.WaitGroup, peerID peer.ID, forceDisconnectOnFail bool) {
	defer wg.Done()

	ctx, cancel := context.WithTimeout(ctx, 7*time.Second)
	defer cancel()

	logger := w.log.With(logging.HostID("peer", peerID))
	logger.Debug("pinging")
	pr := ping.Ping(ctx, w.host, peerID)
	select {
	case res := <-pr:
		if res.Error != nil {
			w.keepAliveMutex.Lock()
			w.keepAliveFails[peerID]++
			w.keepAliveMutex.Unlock()
			logger.Debug("could not ping", zap.Error(res.Error))
		} else {
			w.keepAliveMutex.Lock()
			delete(w.keepAliveFails, peerID)
			w.keepAliveMutex.Unlock()
		}
	case <-ctx.Done():
		w.keepAliveMutex.Lock()
		w.keepAliveFails[peerID]++
		w.keepAliveMutex.Unlock()
		logger.Debug("could not ping (context done)", zap.Error(ctx.Err()))
	}

	w.keepAliveMutex.Lock()
	if (forceDisconnectOnFail || w.keepAliveFails[peerID] > maxAllowedPingFailures) && w.host.Network().Connectedness(peerID) == network.Connected {
		logger.Info("disconnecting peer")
		if err := w.host.Network().ClosePeer(peerID); err != nil {
			logger.Debug("closing conn to peer", zap.Error(err))
		}
		w.keepAliveFails[peerID] = 0
	}
	w.keepAliveMutex.Unlock()
}
