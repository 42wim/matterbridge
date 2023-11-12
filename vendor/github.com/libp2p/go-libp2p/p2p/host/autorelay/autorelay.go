package autorelay

import (
	"context"
	"errors"
	"sync"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	basic "github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"

	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
)

var log = logging.Logger("autorelay")

type AutoRelay struct {
	refCount  sync.WaitGroup
	ctx       context.Context
	ctxCancel context.CancelFunc

	conf *config

	mx     sync.Mutex
	status network.Reachability

	relayFinder *relayFinder

	host   host.Host
	addrsF basic.AddrsFactory

	metricsTracer MetricsTracer
}

func NewAutoRelay(bhost *basic.BasicHost, opts ...Option) (*AutoRelay, error) {
	r := &AutoRelay{
		host:   bhost,
		addrsF: bhost.AddrsFactory,
		status: network.ReachabilityUnknown,
	}
	conf := defaultConfig
	for _, opt := range opts {
		if err := opt(&conf); err != nil {
			return nil, err
		}
	}
	r.ctx, r.ctxCancel = context.WithCancel(context.Background())
	r.conf = &conf
	r.relayFinder = newRelayFinder(bhost, conf.peerSource, &conf)
	r.metricsTracer = &wrappedMetricsTracer{conf.metricsTracer}
	bhost.AddrsFactory = r.hostAddrs

	return r, nil
}

func (r *AutoRelay) Start() {
	r.refCount.Add(1)
	go func() {
		defer r.refCount.Done()
		r.background()
	}()
}

func (r *AutoRelay) background() {
	subReachability, err := r.host.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged), eventbus.Name("autorelay (background)"))
	if err != nil {
		log.Debug("failed to subscribe to the EvtLocalReachabilityChanged")
		return
	}
	defer subReachability.Close()

	for {
		select {
		case <-r.ctx.Done():
			return
		case ev, ok := <-subReachability.Out():
			if !ok {
				return
			}
			// TODO: push changed addresses
			evt := ev.(event.EvtLocalReachabilityChanged)
			switch evt.Reachability {
			case network.ReachabilityPrivate, network.ReachabilityUnknown:
				err := r.relayFinder.Start()
				if errors.Is(err, errAlreadyRunning) {
					log.Debug("tried to start already running relay finder")
				} else if err != nil {
					log.Errorw("failed to start relay finder", "error", err)
				} else {
					r.metricsTracer.RelayFinderStatus(true)
				}
			case network.ReachabilityPublic:
				r.relayFinder.Stop()
				r.metricsTracer.RelayFinderStatus(false)
			}
			r.mx.Lock()
			r.status = evt.Reachability
			r.mx.Unlock()
		}
	}
}

func (r *AutoRelay) hostAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	return r.relayAddrs(r.addrsF(addrs))
}

func (r *AutoRelay) relayAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.status != network.ReachabilityPrivate {
		return addrs
	}
	return r.relayFinder.relayAddrs(addrs)
}

func (r *AutoRelay) Close() error {
	r.ctxCancel()
	err := r.relayFinder.Stop()
	r.refCount.Wait()
	return err
}
