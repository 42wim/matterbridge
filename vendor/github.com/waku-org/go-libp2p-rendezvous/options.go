package rendezvous

import (
	ma "github.com/multiformats/go-multiaddr"
)

type RendezvousPointOption func(cfg *rendezvousPointConfig)

type AddrsFactory func(addrs []ma.Multiaddr) []ma.Multiaddr

var DefaultAddrFactory = func(addrs []ma.Multiaddr) []ma.Multiaddr { return addrs }

var defaultRendezvousPointConfig = rendezvousPointConfig{
	AddrsFactory: DefaultAddrFactory,
}

type rendezvousPointConfig struct {
	AddrsFactory AddrsFactory
}

func (cfg *rendezvousPointConfig) apply(opts ...RendezvousPointOption) {
	for _, opt := range opts {
		opt(cfg)
	}
}

// AddrsFactory configures libp2p to use the given address factory.
func ClientWithAddrsFactory(factory AddrsFactory) RendezvousPointOption {
	return func(cfg *rendezvousPointConfig) {
		cfg.AddrsFactory = factory
	}
}
