/*
Package rcmgr is the resource manager for go-libp2p. This allows you to track
resources being used throughout your go-libp2p process. As well as making sure
that the process doesn't use more resources than what you define as your
limits. The resource manager only knows about things it is told about, so it's
the responsibility of the user of this library (either go-libp2p or a go-libp2p
user) to make sure they check with the resource manager before actually
allocating the resource.
*/
package rcmgr

import (
	"encoding/json"
	"io"
	"math"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Limit is an object that specifies basic resource limits.
type Limit interface {
	// GetMemoryLimit returns the (current) memory limit.
	GetMemoryLimit() int64
	// GetStreamLimit returns the stream limit, for inbound or outbound streams.
	GetStreamLimit(network.Direction) int
	// GetStreamTotalLimit returns the total stream limit
	GetStreamTotalLimit() int
	// GetConnLimit returns the connection limit, for inbound or outbound connections.
	GetConnLimit(network.Direction) int
	// GetConnTotalLimit returns the total connection limit
	GetConnTotalLimit() int
	// GetFDLimit returns the file descriptor limit.
	GetFDLimit() int
}

// Limiter is the interface for providing limits to the resource manager.
type Limiter interface {
	GetSystemLimits() Limit
	GetTransientLimits() Limit
	GetAllowlistedSystemLimits() Limit
	GetAllowlistedTransientLimits() Limit
	GetServiceLimits(svc string) Limit
	GetServicePeerLimits(svc string) Limit
	GetProtocolLimits(proto protocol.ID) Limit
	GetProtocolPeerLimits(proto protocol.ID) Limit
	GetPeerLimits(p peer.ID) Limit
	GetStreamLimits(p peer.ID) Limit
	GetConnLimits() Limit
}

// NewDefaultLimiterFromJSON creates a new limiter by parsing a json configuration,
// using the default limits for fallback.
func NewDefaultLimiterFromJSON(in io.Reader) (Limiter, error) {
	return NewLimiterFromJSON(in, DefaultLimits.AutoScale())
}

// NewLimiterFromJSON creates a new limiter by parsing a json configuration.
func NewLimiterFromJSON(in io.Reader, defaults ConcreteLimitConfig) (Limiter, error) {
	cfg, err := readLimiterConfigFromJSON(in, defaults)
	if err != nil {
		return nil, err
	}
	return &fixedLimiter{cfg}, nil
}

func readLimiterConfigFromJSON(in io.Reader, defaults ConcreteLimitConfig) (ConcreteLimitConfig, error) {
	var cfg PartialLimitConfig
	if err := json.NewDecoder(in).Decode(&cfg); err != nil {
		return ConcreteLimitConfig{}, err
	}
	return cfg.Build(defaults), nil
}

// fixedLimiter is a limiter with fixed limits.
type fixedLimiter struct {
	ConcreteLimitConfig
}

var _ Limiter = (*fixedLimiter)(nil)

func NewFixedLimiter(conf ConcreteLimitConfig) Limiter {
	log.Debugw("initializing new limiter with config", "limits", conf)
	return &fixedLimiter{conf}
}

// BaseLimit is a mixin type for basic resource limits.
type BaseLimit struct {
	Streams         int   `json:",omitempty"`
	StreamsInbound  int   `json:",omitempty"`
	StreamsOutbound int   `json:",omitempty"`
	Conns           int   `json:",omitempty"`
	ConnsInbound    int   `json:",omitempty"`
	ConnsOutbound   int   `json:",omitempty"`
	FD              int   `json:",omitempty"`
	Memory          int64 `json:",omitempty"`
}

func valueOrBlockAll(n int) LimitVal {
	if n == 0 {
		return BlockAllLimit
	} else if n == math.MaxInt {
		return Unlimited
	}
	return LimitVal(n)
}
func valueOrBlockAll64(n int64) LimitVal64 {
	if n == 0 {
		return BlockAllLimit64
	} else if n == math.MaxInt {
		return Unlimited64
	}
	return LimitVal64(n)
}

// ToResourceLimits converts the BaseLimit to a ResourceLimits
func (l BaseLimit) ToResourceLimits() ResourceLimits {
	return ResourceLimits{
		Streams:         valueOrBlockAll(l.Streams),
		StreamsInbound:  valueOrBlockAll(l.StreamsInbound),
		StreamsOutbound: valueOrBlockAll(l.StreamsOutbound),
		Conns:           valueOrBlockAll(l.Conns),
		ConnsInbound:    valueOrBlockAll(l.ConnsInbound),
		ConnsOutbound:   valueOrBlockAll(l.ConnsOutbound),
		FD:              valueOrBlockAll(l.FD),
		Memory:          valueOrBlockAll64(l.Memory),
	}
}

// Apply overwrites all zero-valued limits with the values of l2
// Must not use a pointer receiver.
func (l *BaseLimit) Apply(l2 BaseLimit) {
	if l.Streams == 0 {
		l.Streams = l2.Streams
	}
	if l.StreamsInbound == 0 {
		l.StreamsInbound = l2.StreamsInbound
	}
	if l.StreamsOutbound == 0 {
		l.StreamsOutbound = l2.StreamsOutbound
	}
	if l.Conns == 0 {
		l.Conns = l2.Conns
	}
	if l.ConnsInbound == 0 {
		l.ConnsInbound = l2.ConnsInbound
	}
	if l.ConnsOutbound == 0 {
		l.ConnsOutbound = l2.ConnsOutbound
	}
	if l.Memory == 0 {
		l.Memory = l2.Memory
	}
	if l.FD == 0 {
		l.FD = l2.FD
	}
}

// BaseLimitIncrease is the increase per GiB of allowed memory.
type BaseLimitIncrease struct {
	Streams         int `json:",omitempty"`
	StreamsInbound  int `json:",omitempty"`
	StreamsOutbound int `json:",omitempty"`
	Conns           int `json:",omitempty"`
	ConnsInbound    int `json:",omitempty"`
	ConnsOutbound   int `json:",omitempty"`
	// Memory is in bytes. Values over 1>>30 (1GiB) don't make sense.
	Memory int64 `json:",omitempty"`
	// FDFraction is expected to be >= 0 and <= 1.
	FDFraction float64 `json:",omitempty"`
}

// Apply overwrites all zero-valued limits with the values of l2
// Must not use a pointer receiver.
func (l *BaseLimitIncrease) Apply(l2 BaseLimitIncrease) {
	if l.Streams == 0 {
		l.Streams = l2.Streams
	}
	if l.StreamsInbound == 0 {
		l.StreamsInbound = l2.StreamsInbound
	}
	if l.StreamsOutbound == 0 {
		l.StreamsOutbound = l2.StreamsOutbound
	}
	if l.Conns == 0 {
		l.Conns = l2.Conns
	}
	if l.ConnsInbound == 0 {
		l.ConnsInbound = l2.ConnsInbound
	}
	if l.ConnsOutbound == 0 {
		l.ConnsOutbound = l2.ConnsOutbound
	}
	if l.Memory == 0 {
		l.Memory = l2.Memory
	}
	if l.FDFraction == 0 {
		l.FDFraction = l2.FDFraction
	}
}

func (l BaseLimit) GetStreamLimit(dir network.Direction) int {
	if dir == network.DirInbound {
		return l.StreamsInbound
	} else {
		return l.StreamsOutbound
	}
}

func (l BaseLimit) GetStreamTotalLimit() int {
	return l.Streams
}

func (l BaseLimit) GetConnLimit(dir network.Direction) int {
	if dir == network.DirInbound {
		return l.ConnsInbound
	} else {
		return l.ConnsOutbound
	}
}

func (l BaseLimit) GetConnTotalLimit() int {
	return l.Conns
}

func (l BaseLimit) GetFDLimit() int {
	return l.FD
}

func (l BaseLimit) GetMemoryLimit() int64 {
	return l.Memory
}

func (l *fixedLimiter) GetSystemLimits() Limit {
	return &l.system
}

func (l *fixedLimiter) GetTransientLimits() Limit {
	return &l.transient
}

func (l *fixedLimiter) GetAllowlistedSystemLimits() Limit {
	return &l.allowlistedSystem
}

func (l *fixedLimiter) GetAllowlistedTransientLimits() Limit {
	return &l.allowlistedTransient
}

func (l *fixedLimiter) GetServiceLimits(svc string) Limit {
	sl, ok := l.service[svc]
	if !ok {
		return &l.serviceDefault
	}
	return &sl
}

func (l *fixedLimiter) GetServicePeerLimits(svc string) Limit {
	pl, ok := l.servicePeer[svc]
	if !ok {
		return &l.servicePeerDefault
	}
	return &pl
}

func (l *fixedLimiter) GetProtocolLimits(proto protocol.ID) Limit {
	pl, ok := l.protocol[proto]
	if !ok {
		return &l.protocolDefault
	}
	return &pl
}

func (l *fixedLimiter) GetProtocolPeerLimits(proto protocol.ID) Limit {
	pl, ok := l.protocolPeer[proto]
	if !ok {
		return &l.protocolPeerDefault
	}
	return &pl
}

func (l *fixedLimiter) GetPeerLimits(p peer.ID) Limit {
	pl, ok := l.peer[p]
	if !ok {
		return &l.peerDefault
	}
	return &pl
}

func (l *fixedLimiter) GetStreamLimits(_ peer.ID) Limit {
	return &l.stream
}

func (l *fixedLimiter) GetConnLimits() Limit {
	return &l.conn
}
