package rendezvous

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	ethv4 "github.com/status-im/go-multiaddr-ethv4"
	"github.com/status-im/rendezvous/protocol"
)

var logger = log.New("package", "rendezvous/client")

func NewEphemeral() (c Client, err error) {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Secp256k1, 0, rand.Reader) // bits are ignored with edwards or secp251k1
	if err != nil {
		return Client{}, err
	}
	return New(priv)
}

func New(identity crypto.PrivKey) (c Client, err error) {
	opts := []libp2p.Option{
		libp2p.Identity(identity),
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		return c, err
	}
	return Client{
		h: h,
	}, nil
}

func NewWithHost(h host.Host) (c Client, err error) {
	return Client{
		h: h,
	}, nil
}

type Client struct {
	h host.Host
}

func (c Client) Register(ctx context.Context, srv ma.Multiaddr, topic string, record enr.Record, ttl time.Duration) error {
	s, err := c.newStream(ctx, srv)
	if err != nil {
		return err
	}
	defer s.Close()
	if err = rlp.Encode(s, protocol.REGISTER); err != nil {
		return err
	}
	if err = rlp.Encode(s, protocol.Register{Topic: topic, Record: record, TTL: uint64(ttl)}); err != nil {
		return err
	}
	rs := rlp.NewStream(s, 0)
	typ, err := rs.Uint()
	if err != nil {
		return err
	}
	if protocol.MessageType(typ) != protocol.REGISTER_RESPONSE {
		return fmt.Errorf("expected %v as response, but got %v", protocol.REGISTER_RESPONSE, typ)
	}
	var val protocol.RegisterResponse
	if err = rs.Decode(&val); err != nil {
		return err
	}
	logger.Debug("received response to register", "status", val.Status, "message", val.Message)
	if val.Status != protocol.OK {
		return fmt.Errorf("register failed. status code %v", val.Status)
	}
	return nil
}

func (c Client) Discover(ctx context.Context, srv ma.Multiaddr, topic string, limit int) (rst []enr.Record, err error) {
	s, err := c.newStream(ctx, srv)
	if err != nil {
		return
	}
	defer s.Close()

	if err = rlp.Encode(s, protocol.DISCOVER); err != nil {
		return
	}
	if err = rlp.Encode(s, protocol.Discover{Topic: topic, Limit: uint(limit)}); err != nil {
		return
	}
	rs := rlp.NewStream(s, 0)
	typ, err := rs.Uint()
	if err != nil {
		return nil, err
	}
	if protocol.MessageType(typ) != protocol.DISCOVER_RESPONSE {
		return nil, fmt.Errorf("expected %v as response, but got %v", protocol.REGISTER_RESPONSE, typ)
	}
	var val protocol.DiscoverResponse
	if err = rs.Decode(&val); err != nil {
		return
	}
	if val.Status != protocol.OK {
		return nil, fmt.Errorf("discover request failed. status code %v", val.Status)
	}
	logger.Debug("received response to discover request", "status", val.Status, "records lth", len(val.Records))
	return val.Records, nil
}

func (c Client) RemoteIp(ctx context.Context, srv ma.Multiaddr) (value string, err error) {
	s, err := c.newStream(ctx, srv)
	if err != nil {
		return
	}
	defer s.Close()

	if err = rlp.Encode(s, protocol.REMOTEIP); err != nil {
		return
	}

	rs := rlp.NewStream(s, 0)
	typ, err := rs.Uint()
	if err != nil {
		return
	}
	if protocol.MessageType(typ) != protocol.REMOTEIP_RESPONSE {
		err = fmt.Errorf("expected %v as response, but got %v", protocol.REMOTEIP_RESPONSE, typ)
		return
	}
	var val protocol.RemoteIpResponse
	if err = rs.Decode(&val); err != nil {
		return
	}
	if val.Status != protocol.OK {
		err = fmt.Errorf("remoteip request failed. status code %v", val.Status)
		return
	}
	logger.Debug("received response to remoteip request", "status", val.Status, "ip", val.IP)
	value = val.IP

	return
}

func (c Client) newStream(ctx context.Context, srv ma.Multiaddr) (rw network.Stream, err error) {
	pid, err := srv.ValueForProtocol(ethv4.P_ETHv4)
	if err != nil {
		return
	}
	peerid, err := peer.Decode(pid)
	if err != nil {
		return
	}
	// TODO there must be a better interface
	targetPeerAddr, err := ma.NewMultiaddr(fmt.Sprintf("/ethv4/%s", pid))
	if err != nil {
		return
	}
	targetAddr := srv.Decapsulate(targetPeerAddr)
	c.h.Peerstore().AddAddr(peerid, targetAddr, 5*time.Second)
	s, err := c.h.NewStream(ctx, peerid, "/rend/0.1.0")
	if err != nil {
		return nil, err
	}
	return &InstrumentedStream{s}, nil
}

// Close shutdowns the host and all open connections.
func (c Client) Close() error {
	return c.h.Close()
}
