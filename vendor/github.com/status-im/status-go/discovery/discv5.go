package discovery

import (
	"crypto/ecdsa"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/discv5"
)

// NewDiscV5 creates instances of discovery v5 facade.
func NewDiscV5(prv *ecdsa.PrivateKey, laddr string, bootnodes []*discv5.Node) *DiscV5 {
	return &DiscV5{
		prv:       prv,
		laddr:     laddr,
		bootnodes: bootnodes,
	}
}

// DiscV5 is a facade for ethereum discv5 implementation.
type DiscV5 struct {
	mu  sync.Mutex
	net *discv5.Network

	prv       *ecdsa.PrivateKey
	laddr     string
	bootnodes []*discv5.Node
}

// Running returns true if v5 server is started.
func (d *DiscV5) Running() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.net != nil
}

// Start creates v5 server and stores pointer to it.
func (d *DiscV5) Start() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	log.Debug("Starting discovery", "listen address", d.laddr)
	addr, err := net.ResolveUDPAddr("udp", d.laddr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	ntab, err := discv5.ListenUDP(d.prv, conn, "", nil)
	if err != nil {
		return err
	}
	if err := ntab.SetFallbackNodes(d.bootnodes); err != nil {
		return err
	}
	d.net = ntab
	return nil
}

// Stop closes v5 server listener and removes pointer.
func (d *DiscV5) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.net == nil {
		return nil
	}
	d.net.Close()
	d.net = nil
	return nil
}

// Register creates a register request in v5 server.
// It will block until stop is closed.
func (d *DiscV5) Register(topic string, stop chan struct{}) error {
	d.net.RegisterTopic(discv5.Topic(topic), stop)
	return nil
}

// Discover creates search request in v5 server. Results will be published to found channel.
// It will block until period is closed.
func (d *DiscV5) Discover(topic string, period <-chan time.Duration, found chan<- *discv5.Node, lookup chan<- bool) error {
	d.net.SearchTopic(discv5.Topic(topic), period, found, lookup)
	return nil
}
