package peerstore

import (
	"context"
	"time"

	ic "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/record"
	ma "github.com/multiformats/go-multiaddr"
)

// Contains all interface methods from a libp2p peerstore

func (ps *WakuPeerstoreImpl) AddAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration) {
	ps.peerStore.AddAddr(p, addr, ttl)
}

func (ps *WakuPeerstoreImpl) AddAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration) {
	ps.peerStore.AddAddrs(p, addrs, ttl)
}

func (ps *WakuPeerstoreImpl) SetAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration) {
	ps.peerStore.SetAddr(p, addr, ttl)
}

func (ps *WakuPeerstoreImpl) SetAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration) {
	ps.peerStore.SetAddrs(p, addrs, ttl)
}

func (ps *WakuPeerstoreImpl) UpdateAddrs(p peer.ID, oldTTL time.Duration, newTTL time.Duration) {
	ps.peerStore.UpdateAddrs(p, oldTTL, newTTL)
}

func (ps *WakuPeerstoreImpl) Addrs(p peer.ID) []ma.Multiaddr {
	return ps.peerStore.Addrs(p)
}

func (ps *WakuPeerstoreImpl) AddrStream(ctx context.Context, p peer.ID) <-chan ma.Multiaddr {
	return ps.peerStore.AddrStream(ctx, p)
}

func (ps *WakuPeerstoreImpl) ClearAddrs(p peer.ID) {
	ps.peerStore.ClearAddrs(p)
}

func (ps *WakuPeerstoreImpl) PeersWithAddrs() peer.IDSlice {
	return ps.peerStore.PeersWithAddrs()
}

func (ps *WakuPeerstoreImpl) PeerInfo(peerID peer.ID) peer.AddrInfo {
	return ps.peerStore.PeerInfo(peerID)
}

func (ps *WakuPeerstoreImpl) Peers() peer.IDSlice {
	return ps.peerStore.Peers()
}

func (ps *WakuPeerstoreImpl) Close() error {
	return ps.peerStore.Close()
}

func (ps *WakuPeerstoreImpl) PubKey(p peer.ID) ic.PubKey {
	return ps.peerStore.PubKey(p)
}

func (ps *WakuPeerstoreImpl) AddPubKey(p peer.ID, pubk ic.PubKey) error {
	return ps.peerStore.AddPubKey(p, pubk)
}

func (ps *WakuPeerstoreImpl) PrivKey(p peer.ID) ic.PrivKey {
	return ps.peerStore.PrivKey(p)
}

func (ps *WakuPeerstoreImpl) AddPrivKey(p peer.ID, privk ic.PrivKey) error {
	return ps.peerStore.AddPrivKey(p, privk)
}

func (ps *WakuPeerstoreImpl) PeersWithKeys() peer.IDSlice {
	return ps.peerStore.PeersWithKeys()
}

func (ps *WakuPeerstoreImpl) RemovePeer(p peer.ID) {
	ps.peerStore.RemovePeer(p)
}

func (ps *WakuPeerstoreImpl) Get(p peer.ID, key string) (interface{}, error) {
	return ps.peerStore.Get(p, key)
}

func (ps *WakuPeerstoreImpl) Put(p peer.ID, key string, val interface{}) error {
	return ps.peerStore.Put(p, key, val)

}

func (ps *WakuPeerstoreImpl) RecordLatency(p peer.ID, t time.Duration) {
	ps.peerStore.RecordLatency(p, t)
}

func (ps *WakuPeerstoreImpl) LatencyEWMA(p peer.ID) time.Duration {
	return ps.peerStore.LatencyEWMA(p)
}

func (ps *WakuPeerstoreImpl) GetProtocols(p peer.ID) ([]protocol.ID, error) {
	return ps.peerStore.GetProtocols(p)
}

func (ps *WakuPeerstoreImpl) AddProtocols(p peer.ID, proto ...protocol.ID) error {
	return ps.peerStore.AddProtocols(p, proto...)
}

func (ps *WakuPeerstoreImpl) SetProtocols(p peer.ID, proto ...protocol.ID) error {
	return ps.peerStore.SetProtocols(p, proto...)
}

func (ps *WakuPeerstoreImpl) RemoveProtocols(p peer.ID, proto ...protocol.ID) error {
	return ps.peerStore.RemoveProtocols(p, proto...)
}

func (ps *WakuPeerstoreImpl) SupportsProtocols(p peer.ID, proto ...protocol.ID) ([]protocol.ID, error) {
	return ps.peerStore.SupportsProtocols(p, proto...)
}

func (ps *WakuPeerstoreImpl) FirstSupportedProtocol(p peer.ID, proto ...protocol.ID) (protocol.ID, error) {
	return ps.peerStore.FirstSupportedProtocol(p, proto...)
}

func (ps *WakuPeerstoreImpl) ConsumePeerRecord(s *record.Envelope, ttl time.Duration) (accepted bool, err error) {
	return ps.peerStore.(peerstore.CertifiedAddrBook).ConsumePeerRecord(s, ttl)
}

// GetPeerRecord returns a Envelope containing a PeerRecord for the
// given peer id, if one exists.
// Returns nil if no signed PeerRecord exists for the peer.
func (ps *WakuPeerstoreImpl) GetPeerRecord(p peer.ID) *record.Envelope {
	return ps.peerStore.(peerstore.CertifiedAddrBook).GetPeerRecord(p)
}
