// Copyright 2019 The Waku Library Authors.
//
// The Waku library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Waku library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty off
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Waku library. If not, see <http://www.gnu.org/licenses/>.
//
// This software uses the go-ethereum library, which is licensed
// under the GNU Lesser General Public Library, version 3 or any later.

package common

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/tsenart/tb"

	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var errRateLimitExceeded = errors.New("rate limit has been exceeded")

type runLoop func(rw p2p.MsgReadWriter) error

// RateLimiterPeer interface represents a Peer that is capable of being rate limited
type RateLimiterPeer interface {
	ID() []byte
	IP() net.IP
}

// RateLimiterHandler interface represents handler functionality for a Rate Limiter in the cases of
// exceeding a peer limit and exceeding an IP limit
type RateLimiterHandler interface {
	ExceedPeerLimit() error
	ExceedIPLimit() error
}

// MetricsRateLimiterHandler implements RateLimiterHandler, represents a handler for reporting rate limit Exceed data
// to the metrics collection service (currently prometheus)
type MetricsRateLimiterHandler struct{}

func (MetricsRateLimiterHandler) ExceedPeerLimit() error {
	RateLimitsExceeded.WithLabelValues("peer_id").Inc()
	return nil
}
func (MetricsRateLimiterHandler) ExceedIPLimit() error {
	RateLimitsExceeded.WithLabelValues("ip").Inc()
	return nil
}

// RateLimits contains information about rate limit settings.
// It's agnostic on what it's being rate limited on (bytes or number of packets currently)
// It's exchanged with the status-update packet code
type RateLimits struct {
	IPLimits     uint64 // amount per second from a single IP (default 0, no limits)
	PeerIDLimits uint64 // amount per second from a single peer ID (default 0, no limits)
	TopicLimits  uint64 // amount per second from a single topic (default 0, no limits)
}

func (r RateLimits) IsZero() bool {
	return r == (RateLimits{})
}

// DropPeerRateLimiterHandler implements RateLimiterHandler, represents a handler that introduces Tolerance to the
// number of Peer connections before Limit Exceeded errors are returned.
type DropPeerRateLimiterHandler struct {
	// Tolerance is a number by which a limit must be exceeded before a peer is dropped.
	Tolerance int64

	peerLimitExceeds int64
	ipLimitExceeds   int64
}

func (h *DropPeerRateLimiterHandler) ExceedPeerLimit() error {
	h.peerLimitExceeds++
	if h.Tolerance > 0 && h.peerLimitExceeds >= h.Tolerance {
		return errRateLimitExceeded
	}
	return nil
}

func (h *DropPeerRateLimiterHandler) ExceedIPLimit() error {
	h.ipLimitExceeds++
	if h.Tolerance > 0 && h.ipLimitExceeds >= h.Tolerance {
		return errRateLimitExceeded
	}
	return nil
}

// PeerRateLimiterConfig represents configurations for initialising a PeerRateLimiter
type PeerRateLimiterConfig struct {
	PacketLimitPerSecIP     int64
	PacketLimitPerSecPeerID int64
	BytesLimitPerSecIP      int64
	BytesLimitPerSecPeerID  int64
	WhitelistedIPs          []string
	WhitelistedPeerIDs      []enode.ID
}

var defaultPeerRateLimiterConfig = PeerRateLimiterConfig{
	PacketLimitPerSecIP:     10,
	PacketLimitPerSecPeerID: 5,
	BytesLimitPerSecIP:      1048576, // 1MB
	BytesLimitPerSecPeerID:  1048576, // 1MB
	WhitelistedIPs:          nil,
	WhitelistedPeerIDs:      nil,
}

// PeerRateLimiter represents a rate limiter that limits communication between Peers
type PeerRateLimiter struct {
	packetThrottler *tb.Throttler
	bytesThrottler  *tb.Throttler

	PacketLimitPerSecIP     int64
	PacketLimitPerSecPeerID int64

	BytesLimitPerSecIP     int64
	BytesLimitPerSecPeerID int64

	whitelistedPeerIDs []enode.ID
	whitelistedIPs     []string

	handlers []RateLimiterHandler
}

func NewPeerRateLimiter(cfg *PeerRateLimiterConfig, handlers ...RateLimiterHandler) *PeerRateLimiter {
	if cfg == nil {
		cfgCopy := defaultPeerRateLimiterConfig
		cfg = &cfgCopy
	}

	return &PeerRateLimiter{
		packetThrottler:         tb.NewThrottler(time.Millisecond * 100),
		bytesThrottler:          tb.NewThrottler(time.Millisecond * 100),
		PacketLimitPerSecIP:     cfg.PacketLimitPerSecIP,
		PacketLimitPerSecPeerID: cfg.PacketLimitPerSecPeerID,
		BytesLimitPerSecIP:      cfg.BytesLimitPerSecIP,
		BytesLimitPerSecPeerID:  cfg.BytesLimitPerSecPeerID,
		whitelistedPeerIDs:      cfg.WhitelistedPeerIDs,
		whitelistedIPs:          cfg.WhitelistedIPs,
		handlers:                handlers,
	}
}

func (r *PeerRateLimiter) Decorate(p RateLimiterPeer, rw p2p.MsgReadWriter, runLoop runLoop) error {
	errC := make(chan error, 1)

	in, out := p2p.MsgPipe()
	defer func() {
		if err := in.Close(); err != nil {
			// Don't block as otherwise we might leak go routines
			select {
			case errC <- err:
			default:
			}
		}
	}()
	defer func() {
		if err := out.Close(); err != nil {
			errC <- err
		}
	}()

	// Read from the original reader and write to the message pipe.
	go func() {
		for {
			packet, err := rw.ReadMsg()
			if err != nil {
				// Don't block as otherwise we might leak go routines
				select {
				case errC <- fmt.Errorf("failed to read packet: %v", err):
					return
				default:
					return
				}
			}

			RateLimitsProcessed.Inc()

			var ip string
			if p != nil {
				// this relies on <nil> being the string representation of nil
				// as IP() might return a nil value
				ip = p.IP().String()
			}
			if halted := r.throttleIP(ip, packet.Size); halted {
				for _, h := range r.handlers {
					if err := h.ExceedIPLimit(); err != nil {
						// Don't block as otherwise we might leak go routines
						select {

						case errC <- fmt.Errorf("exceed rate limit by IP: %v", err):
							return
						default:
							return
						}
					}
				}
			}

			var peerID []byte
			if p != nil {
				peerID = p.ID()
			}
			if halted := r.throttlePeer(peerID, packet.Size); halted {
				for _, h := range r.handlers {
					if err := h.ExceedPeerLimit(); err != nil {
						// Don't block as otherwise we might leak go routines
						select {
						case errC <- fmt.Errorf("exceeded rate limit by peer: %v", err):
							return
						default:
							return
						}
					}
				}
			}

			if err := in.WriteMsg(packet); err != nil {
				// Don't block as otherwise we might leak go routines
				select {
				case errC <- fmt.Errorf("failed to write packet to pipe: %v", err):
					return
				default:
					return
				}
			}
		}
	}()

	// Read from the message pipe and write to the original writer.
	go func() {
		for {
			packet, err := in.ReadMsg()
			if err != nil {
				// Don't block as otherwise we might leak go routines
				select {
				case errC <- fmt.Errorf("failed to read packet from pipe: %v", err):
					return
				default:
					return
				}
			}
			if err := rw.WriteMsg(packet); err != nil {
				// Don't block as otherwise we might leak go routines
				select {
				case errC <- fmt.Errorf("failed to write packet: %v", err):
					return
				default:
					return
				}
			}
		}
	}()

	go func() {
		// Don't block as otherwise we might leak go routines
		select {
		case errC <- runLoop(out):
			return
		default:
			return
		}
	}()

	return <-errC
}

// throttleIP throttles packets incoming from a given IP.
func (r *PeerRateLimiter) throttleIP(ip string, size uint32) bool {
	if stringSliceContains(r.whitelistedIPs, ip) {
		return false
	}

	var packetLimiterResponse bool
	var bytesLimiterResponse bool

	if r.PacketLimitPerSecIP != 0 {
		packetLimiterResponse = r.packetThrottler.Halt(ip, 1, r.PacketLimitPerSecIP)
	}
	if r.BytesLimitPerSecIP != 0 {
		bytesLimiterResponse = r.bytesThrottler.Halt(ip, int64(size), r.BytesLimitPerSecIP)
	}

	return packetLimiterResponse || bytesLimiterResponse
}

// throttlePeer throttles packets incoming from a peer.
func (r *PeerRateLimiter) throttlePeer(peerID []byte, size uint32) bool {
	var id enode.ID
	copy(id[:], peerID)
	if enodeIDSliceContains(r.whitelistedPeerIDs, id) {
		return false
	}

	var packetLimiterResponse bool
	var bytesLimiterResponse bool

	if r.PacketLimitPerSecPeerID != 0 {
		packetLimiterResponse = r.packetThrottler.Halt(id.String(), 1, r.PacketLimitPerSecPeerID)
	}

	if r.BytesLimitPerSecPeerID != 0 {
		bytesLimiterResponse = r.bytesThrottler.Halt(id.String(), int64(size), r.BytesLimitPerSecPeerID)
	}

	return packetLimiterResponse || bytesLimiterResponse
}

func stringSliceContains(s []string, searched string) bool {
	for _, item := range s {
		if item == searched {
			return true
		}
	}
	return false
}

func enodeIDSliceContains(s []enode.ID, searched enode.ID) bool {
	for _, item := range s {
		if bytes.Equal(item.Bytes(), searched.Bytes()) {
			return true
		}
	}
	return false
}
