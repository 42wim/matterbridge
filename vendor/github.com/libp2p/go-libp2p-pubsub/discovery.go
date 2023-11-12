package pubsub

import (
	"context"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	discimpl "github.com/libp2p/go-libp2p/p2p/discovery/backoff"
)

var (
	// poll interval

	// DiscoveryPollInitialDelay is how long the discovery system waits after it first starts before polling
	DiscoveryPollInitialDelay = 0 * time.Millisecond
	// DiscoveryPollInterval is approximately how long the discovery system waits in between checks for whether the
	// more peers are needed for any topic
	DiscoveryPollInterval = 1 * time.Second
)

// interval at which to retry advertisements when they fail.
const discoveryAdvertiseRetryInterval = 2 * time.Minute

type DiscoverOpt func(*discoverOptions) error

type discoverOptions struct {
	connFactory BackoffConnectorFactory
	opts        []discovery.Option
}

func defaultDiscoverOptions() *discoverOptions {
	rngSrc := rand.NewSource(rand.Int63())
	minBackoff, maxBackoff := time.Second*10, time.Hour
	cacheSize := 100
	dialTimeout := time.Minute * 2
	discoverOpts := &discoverOptions{
		connFactory: func(host host.Host) (*discimpl.BackoffConnector, error) {
			backoff := discimpl.NewExponentialBackoff(minBackoff, maxBackoff, discimpl.FullJitter, time.Second, 5.0, 0, rand.New(rngSrc))
			return discimpl.NewBackoffConnector(host, cacheSize, dialTimeout, backoff)
		},
	}

	return discoverOpts
}

// discover represents the discovery pipeline.
// The discovery pipeline handles advertising and discovery of peers
type discover struct {
	p *PubSub

	// discovery assists in discovering and advertising peers for a topic
	discovery discovery.Discovery

	// advertising tracks which topics are being advertised
	advertising map[string]context.CancelFunc

	// discoverQ handles continuing peer discovery
	discoverQ chan *discoverReq

	// ongoing tracks ongoing discovery requests
	ongoing map[string]struct{}

	// done handles completion of a discovery request
	done chan string

	// connector handles connecting to new peers found via discovery
	connector *discimpl.BackoffConnector

	// options are the set of options to be used to complete struct construction in Start
	options *discoverOptions
}

// MinTopicSize returns a function that checks if a router is ready for publishing based on the topic size.
// The router ultimately decides the whether it is ready or not, the given size is just a suggestion. Note
// that the topic size does not include the router in the count.
func MinTopicSize(size int) RouterReady {
	return func(rt PubSubRouter, topic string) (bool, error) {
		return rt.EnoughPeers(topic, size), nil
	}
}

// Start attaches the discovery pipeline to a pubsub instance, initializes discovery and starts event loop
func (d *discover) Start(p *PubSub, opts ...DiscoverOpt) error {
	if d.discovery == nil || p == nil {
		return nil
	}

	d.p = p
	d.advertising = make(map[string]context.CancelFunc)
	d.discoverQ = make(chan *discoverReq, 32)
	d.ongoing = make(map[string]struct{})
	d.done = make(chan string)

	conn, err := d.options.connFactory(p.host)
	if err != nil {
		return err
	}
	d.connector = conn

	go d.discoverLoop()
	go d.pollTimer()

	return nil
}

func (d *discover) pollTimer() {
	select {
	case <-time.After(DiscoveryPollInitialDelay):
	case <-d.p.ctx.Done():
		return
	}

	select {
	case d.p.eval <- d.requestDiscovery:
	case <-d.p.ctx.Done():
		return
	}

	ticker := time.NewTicker(DiscoveryPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case d.p.eval <- d.requestDiscovery:
			case <-d.p.ctx.Done():
				return
			}
		case <-d.p.ctx.Done():
			return
		}
	}
}

func (d *discover) requestDiscovery() {
	for t := range d.p.myTopics {
		if !d.p.rt.EnoughPeers(t, 0) {
			d.discoverQ <- &discoverReq{topic: t, done: make(chan struct{}, 1)}
		}
	}
}

func (d *discover) discoverLoop() {
	for {
		select {
		case discover := <-d.discoverQ:
			topic := discover.topic

			if _, ok := d.ongoing[topic]; ok {
				discover.done <- struct{}{}
				continue
			}

			d.ongoing[topic] = struct{}{}

			go func() {
				d.handleDiscovery(d.p.ctx, topic, discover.opts)
				select {
				case d.done <- topic:
				case <-d.p.ctx.Done():
				}
				discover.done <- struct{}{}
			}()
		case topic := <-d.done:
			delete(d.ongoing, topic)
		case <-d.p.ctx.Done():
			return
		}
	}
}

// Advertise advertises this node's interest in a topic to a discovery service. Advertise is not thread-safe.
func (d *discover) Advertise(topic string) {
	if d.discovery == nil {
		return
	}

	advertisingCtx, cancel := context.WithCancel(d.p.ctx)

	if _, ok := d.advertising[topic]; ok {
		cancel()
		return
	}
	d.advertising[topic] = cancel

	go func() {
		next, err := d.discovery.Advertise(advertisingCtx, topic)
		if err != nil {
			log.Warnf("bootstrap: error providing rendezvous for %s: %s", topic, err.Error())
			if next == 0 {
				next = discoveryAdvertiseRetryInterval
			}
		}

		t := time.NewTimer(next)
		defer t.Stop()

		for advertisingCtx.Err() == nil {
			select {
			case <-t.C:
				next, err = d.discovery.Advertise(advertisingCtx, topic)
				if err != nil {
					log.Warnf("bootstrap: error providing rendezvous for %s: %s", topic, err.Error())
					if next == 0 {
						next = discoveryAdvertiseRetryInterval
					}
				}
				t.Reset(next)
			case <-advertisingCtx.Done():
				return
			}
		}
	}()
}

// StopAdvertise stops advertising this node's interest in a topic. StopAdvertise is not thread-safe.
func (d *discover) StopAdvertise(topic string) {
	if d.discovery == nil {
		return
	}

	if advertiseCancel, ok := d.advertising[topic]; ok {
		advertiseCancel()
		delete(d.advertising, topic)
	}
}

// Discover searches for additional peers interested in a given topic
func (d *discover) Discover(topic string, opts ...discovery.Option) {
	if d.discovery == nil {
		return
	}

	d.discoverQ <- &discoverReq{topic, opts, make(chan struct{}, 1)}
}

// Bootstrap attempts to bootstrap to a given topic. Returns true if bootstrapped successfully, false otherwise.
func (d *discover) Bootstrap(ctx context.Context, topic string, ready RouterReady, opts ...discovery.Option) bool {
	if d.discovery == nil {
		return true
	}

	t := time.NewTimer(time.Hour)
	if !t.Stop() {
		<-t.C
	}
	defer t.Stop()

	for {
		// Check if ready for publishing
		bootstrapped := make(chan bool, 1)
		select {
		case d.p.eval <- func() {
			done, _ := ready(d.p.rt, topic)
			bootstrapped <- done
		}:
			if <-bootstrapped {
				return true
			}
		case <-d.p.ctx.Done():
			return false
		case <-ctx.Done():
			return false
		}

		// If not ready discover more peers
		disc := &discoverReq{topic, opts, make(chan struct{}, 1)}
		select {
		case d.discoverQ <- disc:
		case <-d.p.ctx.Done():
			return false
		case <-ctx.Done():
			return false
		}

		select {
		case <-disc.done:
		case <-d.p.ctx.Done():
			return false
		case <-ctx.Done():
			return false
		}

		t.Reset(time.Millisecond * 100)
		select {
		case <-t.C:
		case <-d.p.ctx.Done():
			return false
		case <-ctx.Done():
			return false
		}
	}
}

func (d *discover) handleDiscovery(ctx context.Context, topic string, opts []discovery.Option) {
	discoverCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	peerCh, err := d.discovery.FindPeers(discoverCtx, topic, opts...)
	if err != nil {
		log.Debugf("error finding peers for topic %s: %v", topic, err)
		return
	}

	d.connector.Connect(ctx, peerCh)
}

type discoverReq struct {
	topic string
	opts  []discovery.Option
	done  chan struct{}
}

type pubSubDiscovery struct {
	discovery.Discovery
	opts []discovery.Option
}

func (d *pubSubDiscovery) Advertise(ctx context.Context, ns string, opts ...discovery.Option) (time.Duration, error) {
	return d.Discovery.Advertise(ctx, "floodsub:"+ns, append(opts, d.opts...)...)
}

func (d *pubSubDiscovery) FindPeers(ctx context.Context, ns string, opts ...discovery.Option) (<-chan peer.AddrInfo, error) {
	return d.Discovery.FindPeers(ctx, "floodsub:"+ns, append(opts, d.opts...)...)
}

// WithDiscoveryOpts passes libp2p Discovery options into the PubSub discovery subsystem
func WithDiscoveryOpts(opts ...discovery.Option) DiscoverOpt {
	return func(d *discoverOptions) error {
		d.opts = opts
		return nil
	}
}

// BackoffConnectorFactory creates a BackoffConnector that is attached to a given host
type BackoffConnectorFactory func(host host.Host) (*discimpl.BackoffConnector, error)

// WithDiscoverConnector adds a custom connector that deals with how the discovery subsystem connects to peers
func WithDiscoverConnector(connFactory BackoffConnectorFactory) DiscoverOpt {
	return func(d *discoverOptions) error {
		d.connFactory = connFactory
		return nil
	}
}
