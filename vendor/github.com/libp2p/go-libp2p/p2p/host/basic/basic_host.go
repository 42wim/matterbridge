package basichost

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/record"
	"github.com/libp2p/go-libp2p/core/transport"
	"github.com/libp2p/go-libp2p/p2p/host/autonat"
	"github.com/libp2p/go-libp2p/p2p/host/eventbus"
	"github.com/libp2p/go-libp2p/p2p/host/pstoremanager"
	"github.com/libp2p/go-libp2p/p2p/host/relaysvc"
	relayv2 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	libp2pwebtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/libp2p/go-netroute"

	logging "github.com/ipfs/go-log/v2"
	ma "github.com/multiformats/go-multiaddr"
	madns "github.com/multiformats/go-multiaddr-dns"
	manet "github.com/multiformats/go-multiaddr/net"
	msmux "github.com/multiformats/go-multistream"
)

// addrChangeTickrInterval is the interval between two address change ticks.
var addrChangeTickrInterval = 5 * time.Second

var log = logging.Logger("basichost")

var (
	// DefaultNegotiationTimeout is the default value for HostOpts.NegotiationTimeout.
	DefaultNegotiationTimeout = 10 * time.Second

	// DefaultAddrsFactory is the default value for HostOpts.AddrsFactory.
	DefaultAddrsFactory = func(addrs []ma.Multiaddr) []ma.Multiaddr { return addrs }
)

// AddrsFactory functions can be passed to New in order to override
// addresses returned by Addrs.
type AddrsFactory func([]ma.Multiaddr) []ma.Multiaddr

// BasicHost is the basic implementation of the host.Host interface. This
// particular host implementation:
//   - uses a protocol muxer to mux per-protocol streams
//   - uses an identity service to send + receive node information
//   - uses a nat service to establish NAT port mappings
type BasicHost struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	// ensures we shutdown ONLY once
	closeSync sync.Once
	// keep track of resources we need to wait on before shutting down
	refCount sync.WaitGroup

	network      network.Network
	psManager    *pstoremanager.PeerstoreManager
	mux          *msmux.MultistreamMuxer[protocol.ID]
	ids          identify.IDService
	hps          *holepunch.Service
	pings        *ping.PingService
	natmgr       NATManager
	maResolver   *madns.Resolver
	cmgr         connmgr.ConnManager
	eventbus     event.Bus
	relayManager *relaysvc.RelayManager

	AddrsFactory AddrsFactory

	negtimeout time.Duration

	emitters struct {
		evtLocalProtocolsUpdated event.Emitter
		evtLocalAddrsUpdated     event.Emitter
	}

	addrChangeChan chan struct{}

	addrMu                 sync.RWMutex
	filteredInterfaceAddrs []ma.Multiaddr
	allInterfaceAddrs      []ma.Multiaddr

	disableSignedPeerRecord bool
	signKey                 crypto.PrivKey
	caBook                  peerstore.CertifiedAddrBook

	autoNat autonat.AutoNAT
}

var _ host.Host = (*BasicHost)(nil)

// HostOpts holds options that can be passed to NewHost in order to
// customize construction of the *BasicHost.
type HostOpts struct {
	// EventBus sets the event bus. Will construct a new event bus if omitted.
	EventBus event.Bus

	// MultistreamMuxer is essential for the *BasicHost and will use a sensible default value if omitted.
	MultistreamMuxer *msmux.MultistreamMuxer[protocol.ID]

	// NegotiationTimeout determines the read and write timeouts on streams.
	// If 0 or omitted, it will use DefaultNegotiationTimeout.
	// If below 0, timeouts on streams will be deactivated.
	NegotiationTimeout time.Duration

	// AddrsFactory holds a function which can be used to override or filter the result of Addrs.
	// If omitted, there's no override or filtering, and the results of Addrs and AllAddrs are the same.
	AddrsFactory AddrsFactory

	// MultiaddrResolves holds the go-multiaddr-dns.Resolver used for resolving
	// /dns4, /dns6, and /dnsaddr addresses before trying to connect to a peer.
	MultiaddrResolver *madns.Resolver

	// NATManager takes care of setting NAT port mappings, and discovering external addresses.
	// If omitted, this will simply be disabled.
	NATManager func(network.Network) NATManager

	// ConnManager is a libp2p connection manager
	ConnManager connmgr.ConnManager

	// EnablePing indicates whether to instantiate the ping service
	EnablePing bool

	// EnableRelayService enables the circuit v2 relay (if we're publicly reachable).
	EnableRelayService bool
	// RelayServiceOpts are options for the circuit v2 relay.
	RelayServiceOpts []relayv2.Option

	// UserAgent sets the user-agent for the host.
	UserAgent string

	// ProtocolVersion sets the protocol version for the host.
	ProtocolVersion string

	// DisableSignedPeerRecord disables the generation of Signed Peer Records on this host.
	DisableSignedPeerRecord bool

	// EnableHolePunching enables the peer to initiate/respond to hole punching attempts for NAT traversal.
	EnableHolePunching bool
	// HolePunchingOptions are options for the hole punching service
	HolePunchingOptions []holepunch.Option

	// EnableMetrics enables the metrics subsystems
	EnableMetrics bool
	// PrometheusRegisterer is the PrometheusRegisterer used for metrics
	PrometheusRegisterer prometheus.Registerer
}

// NewHost constructs a new *BasicHost and activates it by attaching its stream and connection handlers to the given inet.Network.
func NewHost(n network.Network, opts *HostOpts) (*BasicHost, error) {
	if opts == nil {
		opts = &HostOpts{}
	}
	if opts.EventBus == nil {
		opts.EventBus = eventbus.NewBus()
	}

	psManager, err := pstoremanager.NewPeerstoreManager(n.Peerstore(), opts.EventBus)
	if err != nil {
		return nil, err
	}
	hostCtx, cancel := context.WithCancel(context.Background())

	h := &BasicHost{
		network:                 n,
		psManager:               psManager,
		mux:                     msmux.NewMultistreamMuxer[protocol.ID](),
		negtimeout:              DefaultNegotiationTimeout,
		AddrsFactory:            DefaultAddrsFactory,
		maResolver:              madns.DefaultResolver,
		eventbus:                opts.EventBus,
		addrChangeChan:          make(chan struct{}, 1),
		ctx:                     hostCtx,
		ctxCancel:               cancel,
		disableSignedPeerRecord: opts.DisableSignedPeerRecord,
	}

	h.updateLocalIpAddr()

	if h.emitters.evtLocalProtocolsUpdated, err = h.eventbus.Emitter(&event.EvtLocalProtocolsUpdated{}, eventbus.Stateful); err != nil {
		return nil, err
	}
	if h.emitters.evtLocalAddrsUpdated, err = h.eventbus.Emitter(&event.EvtLocalAddressesUpdated{}, eventbus.Stateful); err != nil {
		return nil, err
	}

	if !h.disableSignedPeerRecord {
		cab, ok := peerstore.GetCertifiedAddrBook(n.Peerstore())
		if !ok {
			return nil, errors.New("peerstore should also be a certified address book")
		}
		h.caBook = cab

		h.signKey = h.Peerstore().PrivKey(h.ID())
		if h.signKey == nil {
			return nil, errors.New("unable to access host key")
		}

		// persist a signed peer record for self to the peerstore.
		rec := peer.PeerRecordFromAddrInfo(peer.AddrInfo{
			ID:    h.ID(),
			Addrs: h.Addrs(),
		})
		ev, err := record.Seal(rec, h.signKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create signed record for self: %w", err)
		}
		if _, err := cab.ConsumePeerRecord(ev, peerstore.PermanentAddrTTL); err != nil {
			return nil, fmt.Errorf("failed to persist signed record to peerstore: %w", err)
		}
	}

	if opts.MultistreamMuxer != nil {
		h.mux = opts.MultistreamMuxer
	}

	idOpts := []identify.Option{
		identify.UserAgent(opts.UserAgent),
		identify.ProtocolVersion(opts.ProtocolVersion),
	}

	// we can't set this as a default above because it depends on the *BasicHost.
	if h.disableSignedPeerRecord {
		idOpts = append(idOpts, identify.DisableSignedPeerRecord())
	}
	if opts.EnableMetrics {
		idOpts = append(idOpts,
			identify.WithMetricsTracer(
				identify.NewMetricsTracer(identify.WithRegisterer(opts.PrometheusRegisterer))))
	}

	h.ids, err = identify.NewIDService(h, idOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Identify service: %s", err)
	}

	if opts.EnableHolePunching {
		if opts.EnableMetrics {
			hpOpts := []holepunch.Option{
				holepunch.WithMetricsTracer(holepunch.NewMetricsTracer(holepunch.WithRegisterer(opts.PrometheusRegisterer)))}
			opts.HolePunchingOptions = append(hpOpts, opts.HolePunchingOptions...)

		}
		h.hps, err = holepunch.NewService(h, h.ids, opts.HolePunchingOptions...)
		if err != nil {
			return nil, fmt.Errorf("failed to create hole punch service: %w", err)
		}
	}

	if uint64(opts.NegotiationTimeout) != 0 {
		h.negtimeout = opts.NegotiationTimeout
	}

	if opts.AddrsFactory != nil {
		h.AddrsFactory = opts.AddrsFactory
	}

	if opts.NATManager != nil {
		h.natmgr = opts.NATManager(n)
	}

	if opts.MultiaddrResolver != nil {
		h.maResolver = opts.MultiaddrResolver
	}

	if opts.ConnManager == nil {
		h.cmgr = &connmgr.NullConnMgr{}
	} else {
		h.cmgr = opts.ConnManager
		n.Notify(h.cmgr.Notifee())
	}

	if opts.EnableRelayService {
		if opts.EnableMetrics {
			// Prefer explicitly provided metrics tracer
			metricsOpt := []relayv2.Option{
				relayv2.WithMetricsTracer(
					relayv2.NewMetricsTracer(relayv2.WithRegisterer(opts.PrometheusRegisterer)))}
			opts.RelayServiceOpts = append(metricsOpt, opts.RelayServiceOpts...)
		}
		h.relayManager = relaysvc.NewRelayManager(h, opts.RelayServiceOpts...)
	}

	if opts.EnablePing {
		h.pings = ping.NewPingService(h)
	}

	n.SetStreamHandler(h.newStreamHandler)

	// register to be notified when the network's listen addrs change,
	// so we can update our address set and push events if needed
	listenHandler := func(network.Network, ma.Multiaddr) {
		h.SignalAddressChange()
	}
	n.Notify(&network.NotifyBundle{
		ListenF:      listenHandler,
		ListenCloseF: listenHandler,
	})

	return h, nil
}

func (h *BasicHost) updateLocalIpAddr() {
	h.addrMu.Lock()
	defer h.addrMu.Unlock()

	h.filteredInterfaceAddrs = nil
	h.allInterfaceAddrs = nil

	// Try to use the default ipv4/6 addresses.

	if r, err := netroute.New(); err != nil {
		log.Debugw("failed to build Router for kernel's routing table", "error", err)
	} else {
		if _, _, localIPv4, err := r.Route(net.IPv4zero); err != nil {
			log.Debugw("failed to fetch local IPv4 address", "error", err)
		} else if localIPv4.IsGlobalUnicast() {
			maddr, err := manet.FromIP(localIPv4)
			if err == nil {
				h.filteredInterfaceAddrs = append(h.filteredInterfaceAddrs, maddr)
			}
		}

		if _, _, localIPv6, err := r.Route(net.IPv6unspecified); err != nil {
			log.Debugw("failed to fetch local IPv6 address", "error", err)
		} else if localIPv6.IsGlobalUnicast() {
			maddr, err := manet.FromIP(localIPv6)
			if err == nil {
				h.filteredInterfaceAddrs = append(h.filteredInterfaceAddrs, maddr)
			}
		}
	}

	// Resolve the interface addresses
	ifaceAddrs, err := manet.InterfaceMultiaddrs()
	if err != nil {
		// This usually shouldn't happen, but we could be in some kind
		// of funky restricted environment.
		log.Errorw("failed to resolve local interface addresses", "error", err)

		// Add the loopback addresses to the filtered addrs and use them as the non-filtered addrs.
		// Then bail. There's nothing else we can do here.
		h.filteredInterfaceAddrs = append(h.filteredInterfaceAddrs, manet.IP4Loopback, manet.IP6Loopback)
		h.allInterfaceAddrs = h.filteredInterfaceAddrs
		return
	}

	for _, addr := range ifaceAddrs {
		// Skip link-local addrs, they're mostly useless.
		if !manet.IsIP6LinkLocal(addr) {
			h.allInterfaceAddrs = append(h.allInterfaceAddrs, addr)
		}
	}

	// If netroute failed to get us any interface addresses, use all of
	// them.
	if len(h.filteredInterfaceAddrs) == 0 {
		// Add all addresses.
		h.filteredInterfaceAddrs = h.allInterfaceAddrs
	} else {
		// Only add loopback addresses. Filter these because we might
		// not _have_ an IPv6 loopback address.
		for _, addr := range h.allInterfaceAddrs {
			if manet.IsIPLoopback(addr) {
				h.filteredInterfaceAddrs = append(h.filteredInterfaceAddrs, addr)
			}
		}
	}
}

// Start starts background tasks in the host
func (h *BasicHost) Start() {
	h.psManager.Start()
	h.refCount.Add(1)
	h.ids.Start()
	go h.background()
}

// newStreamHandler is the remote-opened stream handler for network.Network
// TODO: this feels a bit wonky
func (h *BasicHost) newStreamHandler(s network.Stream) {
	before := time.Now()

	if h.negtimeout > 0 {
		if err := s.SetDeadline(time.Now().Add(h.negtimeout)); err != nil {
			log.Debug("setting stream deadline: ", err)
			s.Reset()
			return
		}
	}

	protoID, handle, err := h.Mux().Negotiate(s)
	took := time.Since(before)
	if err != nil {
		if err == io.EOF {
			logf := log.Debugf
			if took > time.Second*10 {
				logf = log.Warnf
			}
			logf("protocol EOF: %s (took %s)", s.Conn().RemotePeer(), took)
		} else {
			log.Debugf("protocol mux failed: %s (took %s)", err, took)
		}
		s.Reset()
		return
	}

	if h.negtimeout > 0 {
		if err := s.SetDeadline(time.Time{}); err != nil {
			log.Debugf("resetting stream deadline: ", err)
			s.Reset()
			return
		}
	}

	if err := s.SetProtocol(protoID); err != nil {
		log.Debugf("error setting stream protocol: %s", err)
		s.Reset()
		return
	}

	log.Debugf("negotiated: %s (took %s)", protoID, took)

	go handle(protoID, s)
}

// SignalAddressChange signals to the host that it needs to determine whether our listen addresses have recently
// changed.
// Warning: this interface is unstable and may disappear in the future.
func (h *BasicHost) SignalAddressChange() {
	select {
	case h.addrChangeChan <- struct{}{}:
	default:
	}
}

func makeUpdatedAddrEvent(prev, current []ma.Multiaddr) *event.EvtLocalAddressesUpdated {
	prevmap := make(map[string]ma.Multiaddr, len(prev))
	evt := event.EvtLocalAddressesUpdated{Diffs: true}
	addrsAdded := false

	for _, addr := range prev {
		prevmap[string(addr.Bytes())] = addr
	}
	for _, addr := range current {
		_, ok := prevmap[string(addr.Bytes())]
		updated := event.UpdatedAddress{Address: addr}
		if ok {
			updated.Action = event.Maintained
		} else {
			updated.Action = event.Added
			addrsAdded = true
		}
		evt.Current = append(evt.Current, updated)
		delete(prevmap, string(addr.Bytes()))
	}
	for _, addr := range prevmap {
		updated := event.UpdatedAddress{Action: event.Removed, Address: addr}
		evt.Removed = append(evt.Removed, updated)
	}

	if !addrsAdded && len(evt.Removed) == 0 {
		return nil
	}

	return &evt
}

func (h *BasicHost) makeSignedPeerRecord(evt *event.EvtLocalAddressesUpdated) (*record.Envelope, error) {
	current := make([]ma.Multiaddr, 0, len(evt.Current))
	for _, a := range evt.Current {
		current = append(current, a.Address)
	}

	rec := peer.PeerRecordFromAddrInfo(peer.AddrInfo{
		ID:    h.ID(),
		Addrs: current,
	})
	return record.Seal(rec, h.signKey)
}

func (h *BasicHost) background() {
	defer h.refCount.Done()
	var lastAddrs []ma.Multiaddr

	emitAddrChange := func(currentAddrs []ma.Multiaddr, lastAddrs []ma.Multiaddr) {
		// nothing to do if both are nil..defensive check
		if currentAddrs == nil && lastAddrs == nil {
			return
		}

		changeEvt := makeUpdatedAddrEvent(lastAddrs, currentAddrs)

		if changeEvt == nil {
			return
		}

		if !h.disableSignedPeerRecord {
			// add signed peer record to the event
			sr, err := h.makeSignedPeerRecord(changeEvt)
			if err != nil {
				log.Errorf("error creating a signed peer record from the set of current addresses, err=%s", err)
				return
			}
			changeEvt.SignedPeerRecord = sr

			// persist the signed record to the peerstore
			if _, err := h.caBook.ConsumePeerRecord(sr, peerstore.PermanentAddrTTL); err != nil {
				log.Errorf("failed to persist signed peer record in peer store, err=%s", err)
				return
			}
		}

		// emit addr change event on the bus
		if err := h.emitters.evtLocalAddrsUpdated.Emit(*changeEvt); err != nil {
			log.Warnf("error emitting event for updated addrs: %s", err)
		}
	}

	// periodically schedules an IdentifyPush to update our peers for changes
	// in our address set (if needed)
	ticker := time.NewTicker(addrChangeTickrInterval)
	defer ticker.Stop()

	for {
		if len(h.network.ListenAddresses()) > 0 {
			h.updateLocalIpAddr()
		}
		// Request addresses anyways because, technically, address filters still apply.
		// The underlying AllAddrs call is effectivley a no-op.
		curr := h.Addrs()
		emitAddrChange(curr, lastAddrs)
		lastAddrs = curr

		select {
		case <-ticker.C:
		case <-h.addrChangeChan:
		case <-h.ctx.Done():
			return
		}
	}
}

// ID returns the (local) peer.ID associated with this Host
func (h *BasicHost) ID() peer.ID {
	return h.Network().LocalPeer()
}

// Peerstore returns the Host's repository of Peer Addresses and Keys.
func (h *BasicHost) Peerstore() peerstore.Peerstore {
	return h.Network().Peerstore()
}

// Network returns the Network interface of the Host
func (h *BasicHost) Network() network.Network {
	return h.network
}

// Mux returns the Mux multiplexing incoming streams to protocol handlers
func (h *BasicHost) Mux() protocol.Switch {
	return h.mux
}

// IDService returns
func (h *BasicHost) IDService() identify.IDService {
	return h.ids
}

func (h *BasicHost) EventBus() event.Bus {
	return h.eventbus
}

// SetStreamHandler sets the protocol handler on the Host's Mux.
// This is equivalent to:
//
//	host.Mux().SetHandler(proto, handler)
//
// (Thread-safe)
func (h *BasicHost) SetStreamHandler(pid protocol.ID, handler network.StreamHandler) {
	h.Mux().AddHandler(pid, func(p protocol.ID, rwc io.ReadWriteCloser) error {
		is := rwc.(network.Stream)
		handler(is)
		return nil
	})
	h.emitters.evtLocalProtocolsUpdated.Emit(event.EvtLocalProtocolsUpdated{
		Added: []protocol.ID{pid},
	})
}

// SetStreamHandlerMatch sets the protocol handler on the Host's Mux
// using a matching function to do protocol comparisons
func (h *BasicHost) SetStreamHandlerMatch(pid protocol.ID, m func(protocol.ID) bool, handler network.StreamHandler) {
	h.Mux().AddHandlerWithFunc(pid, m, func(p protocol.ID, rwc io.ReadWriteCloser) error {
		is := rwc.(network.Stream)
		handler(is)
		return nil
	})
	h.emitters.evtLocalProtocolsUpdated.Emit(event.EvtLocalProtocolsUpdated{
		Added: []protocol.ID{pid},
	})
}

// RemoveStreamHandler returns ..
func (h *BasicHost) RemoveStreamHandler(pid protocol.ID) {
	h.Mux().RemoveHandler(pid)
	h.emitters.evtLocalProtocolsUpdated.Emit(event.EvtLocalProtocolsUpdated{
		Removed: []protocol.ID{pid},
	})
}

// NewStream opens a new stream to given peer p, and writes a p2p/protocol
// header with given protocol.ID. If there is no connection to p, attempts
// to create one. If ProtocolID is "", writes no header.
// (Thread-safe)
func (h *BasicHost) NewStream(ctx context.Context, p peer.ID, pids ...protocol.ID) (network.Stream, error) {
	// Ensure we have a connection, with peer addresses resolved by the routing system (#207)
	// It is not sufficient to let the underlying host connect, it will most likely not have
	// any addresses for the peer without any prior connections.
	// If the caller wants to prevent the host from dialing, it should use the NoDial option.
	if nodial, _ := network.GetNoDial(ctx); !nodial {
		err := h.Connect(ctx, peer.AddrInfo{ID: p})
		if err != nil {
			return nil, err
		}
	}

	s, err := h.Network().NewStream(network.WithNoDial(ctx, "already dialed"), p)
	if err != nil {
		// TODO: It would be nicer to get the actual error from the swarm,
		// but this will require some more work.
		if errors.Is(err, network.ErrNoConn) {
			return nil, errors.New("connection failed")
		}
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	// Wait for any in-progress identifies on the connection to finish. This
	// is faster than negotiating.
	//
	// If the other side doesn't support identify, that's fine. This will
	// just be a no-op.
	select {
	case <-h.ids.IdentifyWait(s.Conn()):
	case <-ctx.Done():
		_ = s.Reset()
		return nil, fmt.Errorf("identify failed to complete: %w", ctx.Err())
	}

	pref, err := h.preferredProtocol(p, pids)
	if err != nil {
		_ = s.Reset()
		return nil, err
	}

	if pref != "" {
		s.SetProtocol(pref)
		lzcon := msmux.NewMSSelect(s, pref)
		return &streamWrapper{
			Stream: s,
			rw:     lzcon,
		}, nil
	}

	// Negotiate the protocol in the background, obeying the context.
	var selected protocol.ID
	errCh := make(chan error, 1)
	go func() {
		selected, err = msmux.SelectOneOf(pids, s)
		errCh <- err
	}()
	select {
	case err = <-errCh:
		if err != nil {
			s.Reset()
			return nil, fmt.Errorf("failed to negotiate protocol: %w", err)
		}
	case <-ctx.Done():
		s.Reset()
		// wait for `SelectOneOf` to error out because of resetting the stream.
		<-errCh
		return nil, fmt.Errorf("failed to negotiate protocol: %w", ctx.Err())
	}

	s.SetProtocol(selected)
	h.Peerstore().AddProtocols(p, selected)
	return s, nil
}

func (h *BasicHost) preferredProtocol(p peer.ID, pids []protocol.ID) (protocol.ID, error) {
	supported, err := h.Peerstore().SupportsProtocols(p, pids...)
	if err != nil {
		return "", err
	}

	var out protocol.ID
	if len(supported) > 0 {
		out = supported[0]
	}
	return out, nil
}

// Connect ensures there is a connection between this host and the peer with
// given peer.ID. If there is not an active connection, Connect will issue a
// h.Network.Dial, and block until a connection is open, or an error is returned.
// Connect will absorb the addresses in pi into its internal peerstore.
// It will also resolve any /dns4, /dns6, and /dnsaddr addresses.
func (h *BasicHost) Connect(ctx context.Context, pi peer.AddrInfo) error {
	// absorb addresses into peerstore
	h.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.TempAddrTTL)

	forceDirect, _ := network.GetForceDirectDial(ctx)
	if !forceDirect {
		if h.Network().Connectedness(pi.ID) == network.Connected {
			return nil
		}
	}

	return h.dialPeer(ctx, pi.ID)
}

// dialPeer opens a connection to peer, and makes sure to identify
// the connection once it has been opened.
func (h *BasicHost) dialPeer(ctx context.Context, p peer.ID) error {
	log.Debugf("host %s dialing %s", h.ID(), p)
	c, err := h.Network().DialPeer(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	// TODO: Consider removing this? On one hand, it's nice because we can
	// assume that things like the agent version are usually set when this
	// returns. On the other hand, we don't _really_ need to wait for this.
	//
	// This is mostly here to preserve existing behavior.
	select {
	case <-h.ids.IdentifyWait(c):
	case <-ctx.Done():
		return fmt.Errorf("identify failed to complete: %w", ctx.Err())
	}

	log.Debugf("host %s finished dialing %s", h.ID(), p)
	return nil
}

func (h *BasicHost) ConnManager() connmgr.ConnManager {
	return h.cmgr
}

// Addrs returns listening addresses that are safe to announce to the network.
// The output is the same as AllAddrs, but processed by AddrsFactory.
func (h *BasicHost) Addrs() []ma.Multiaddr {
	// This is a temporary workaround/hack that fixes #2233. Once we have a
	// proper address pipeline, rework this. See the issue for more context.
	type transportForListeninger interface {
		TransportForListening(a ma.Multiaddr) transport.Transport
	}

	type addCertHasher interface {
		AddCertHashes(m ma.Multiaddr) (ma.Multiaddr, bool)
	}

	addrs := h.AddrsFactory(h.AllAddrs())

	s, ok := h.Network().(transportForListeninger)
	if !ok {
		return addrs
	}

	// Copy addrs slice since we'll be modifying it.
	addrsOld := addrs
	addrs = make([]ma.Multiaddr, len(addrsOld))
	copy(addrs, addrsOld)

	for i, addr := range addrs {
		if ok, n := libp2pwebtransport.IsWebtransportMultiaddr(addr); ok && n == 0 {
			t := s.TransportForListening(addr)
			tpt, ok := t.(addCertHasher)
			if !ok {
				continue
			}
			addrWithCerthash, added := tpt.AddCertHashes(addr)
			addrs[i] = addrWithCerthash
			if !added {
				log.Debug("Couldn't add certhashes to webtransport multiaddr because we aren't listening on webtransport")
			}
		}
	}
	return addrs
}

// NormalizeMultiaddr returns a multiaddr suitable for equality checks.
// If the multiaddr is a webtransport component, it removes the certhashes.
func (h *BasicHost) NormalizeMultiaddr(addr ma.Multiaddr) ma.Multiaddr {
	if ok, n := libp2pwebtransport.IsWebtransportMultiaddr(addr); ok && n > 0 {
		out := addr
		for i := 0; i < n; i++ {
			out, _ = ma.SplitLast(out)
		}
		return out
	}
	return addr
}

// AllAddrs returns all the addresses of BasicHost at this moment in time.
// It's ok to not include addresses if they're not available to be used now.
func (h *BasicHost) AllAddrs() []ma.Multiaddr {
	listenAddrs := h.Network().ListenAddresses()
	if len(listenAddrs) == 0 {
		return nil
	}

	h.addrMu.RLock()
	filteredIfaceAddrs := h.filteredInterfaceAddrs
	allIfaceAddrs := h.allInterfaceAddrs
	h.addrMu.RUnlock()

	// Iterate over all _unresolved_ listen addresses, resolving our primary
	// interface only to avoid advertising too many addresses.
	var finalAddrs []ma.Multiaddr
	if resolved, err := manet.ResolveUnspecifiedAddresses(listenAddrs, filteredIfaceAddrs); err != nil {
		// This can happen if we're listening on no addrs, or listening
		// on IPv6 addrs, but only have IPv4 interface addrs.
		log.Debugw("failed to resolve listen addrs", "error", err)
	} else {
		finalAddrs = append(finalAddrs, resolved...)
	}

	finalAddrs = ma.Unique(finalAddrs)

	// use nat mappings if we have them
	if h.natmgr != nil && h.natmgr.HasDiscoveredNAT() {
		// We have successfully mapped ports on our NAT. Use those
		// instead of observed addresses (mostly).
		// Next, apply this mapping to our addresses.
		for _, listen := range listenAddrs {
			extMaddr := h.natmgr.GetMapping(listen)
			if extMaddr == nil {
				// not mapped
				continue
			}

			// if the router reported a sane address
			if !manet.IsIPUnspecified(extMaddr) {
				// Add in the mapped addr.
				finalAddrs = append(finalAddrs, extMaddr)
			} else {
				log.Warn("NAT device reported an unspecified IP as it's external address")
			}

			// Did the router give us a routable public addr?
			if manet.IsPublicAddr(extMaddr) {
				// well done
				continue
			}

			// No.
			// in case the router gives us a wrong address or we're behind a double-NAT.
			// also add observed addresses
			resolved, err := manet.ResolveUnspecifiedAddress(listen, allIfaceAddrs)
			if err != nil {
				// This can happen if we try to resolve /ip6/::/...
				// without any IPv6 interface addresses.
				continue
			}

			for _, addr := range resolved {
				// Now, check if we have any observed addresses that
				// differ from the one reported by the router. Routers
				// don't always give the most accurate information.
				observed := h.ids.ObservedAddrsFor(addr)

				if len(observed) == 0 {
					continue
				}

				// Drop the IP from the external maddr
				_, extMaddrNoIP := ma.SplitFirst(extMaddr)

				for _, obsMaddr := range observed {
					// Extract a public observed addr.
					ip, _ := ma.SplitFirst(obsMaddr)
					if ip == nil || !manet.IsPublicAddr(ip) {
						continue
					}

					finalAddrs = append(finalAddrs, ma.Join(ip, extMaddrNoIP))
				}
			}
		}
	} else {
		var observedAddrs []ma.Multiaddr
		if h.ids != nil {
			observedAddrs = h.ids.OwnObservedAddrs()
		}
		finalAddrs = append(finalAddrs, observedAddrs...)
	}
	finalAddrs = ma.Unique(finalAddrs)
	finalAddrs = inferWebtransportAddrsFromQuic(finalAddrs)

	return finalAddrs
}

var wtComponent = ma.StringCast("/webtransport")

// inferWebtransportAddrsFromQuic infers more webtransport addresses from QUIC addresses.
// This is useful when we discover our public QUIC address, but haven't discovered our public WebTransport addrs.
// If we see that we are listening on the same port for QUIC and WebTransport,
// we can be pretty sure that the WebTransport addr will be reachable if the
// QUIC one is.
// We assume the input is deduped.
func inferWebtransportAddrsFromQuic(in []ma.Multiaddr) []ma.Multiaddr {
	// We need to check if we are listening on the same ip+port for QUIC and WebTransport.
	// If not, there's nothing to do since we can't infer anything.

	// Count the number of QUIC addrs, this will let us allocate just once at the beginning.
	quicAddrCount := 0
	for _, addr := range in {
		if _, lastComponent := ma.SplitLast(addr); lastComponent.Protocol().Code == ma.P_QUIC_V1 {
			quicAddrCount++
		}
	}
	quicOrWebtransportAddrs := make(map[string]struct{}, quicAddrCount)
	webtransportAddrs := make(map[string]struct{}, quicAddrCount)
	foundSameListeningAddr := false
	for _, addr := range in {
		isWebtransport, numCertHashes := libp2pwebtransport.IsWebtransportMultiaddr(addr)
		if isWebtransport {
			for i := 0; i < numCertHashes; i++ {
				// Remove certhashes
				addr, _ = ma.SplitLast(addr)
			}
			webtransportAddrs[addr.String()] = struct{}{}
			// Remove webtransport component, now it's a multiaddr that ends in /quic-v1
			addr, _ = ma.SplitLast(addr)
		}

		if _, lastComponent := ma.SplitLast(addr); lastComponent.Protocol().Code == ma.P_QUIC_V1 {
			addrStr := addr.String()
			if _, ok := quicOrWebtransportAddrs[addrStr]; ok {
				foundSameListeningAddr = true
			} else {
				quicOrWebtransportAddrs[addrStr] = struct{}{}
			}
		}
	}

	if !foundSameListeningAddr {
		return in
	}

	if len(webtransportAddrs) == 0 {
		// No webtransport addresses, we aren't listening on any webtransport
		// address, so we shouldn't add any.
		return in
	}

	out := make([]ma.Multiaddr, 0, len(in)+(quicAddrCount-len(webtransportAddrs)))
	for _, addr := range in {
		// Add all the original addresses
		out = append(out, addr)
		if _, lastComponent := ma.SplitLast(addr); lastComponent.Protocol().Code == ma.P_QUIC_V1 {
			// Convert quic to webtransport
			addr = addr.Encapsulate(wtComponent)
			if _, ok := webtransportAddrs[addr.String()]; ok {
				// We already have this address
				continue
			}
			// Add the new inferred address
			out = append(out, addr)
		}
	}

	return out
}

// SetAutoNat sets the autonat service for the host.
func (h *BasicHost) SetAutoNat(a autonat.AutoNAT) {
	h.addrMu.Lock()
	defer h.addrMu.Unlock()
	if h.autoNat == nil {
		h.autoNat = a
	}
}

// GetAutoNat returns the host's AutoNAT service, if AutoNAT is enabled.
func (h *BasicHost) GetAutoNat() autonat.AutoNAT {
	h.addrMu.Lock()
	defer h.addrMu.Unlock()
	return h.autoNat
}

// Close shuts down the Host's services (network, etc).
func (h *BasicHost) Close() error {
	h.closeSync.Do(func() {
		h.ctxCancel()
		if h.natmgr != nil {
			h.natmgr.Close()
		}
		if h.cmgr != nil {
			h.cmgr.Close()
		}
		if h.ids != nil {
			h.ids.Close()
		}
		if h.autoNat != nil {
			h.autoNat.Close()
		}
		if h.relayManager != nil {
			h.relayManager.Close()
		}
		if h.hps != nil {
			h.hps.Close()
		}

		_ = h.emitters.evtLocalProtocolsUpdated.Close()
		_ = h.emitters.evtLocalAddrsUpdated.Close()
		h.Network().Close()

		h.psManager.Close()
		if h.Peerstore() != nil {
			h.Peerstore().Close()
		}

		h.refCount.Wait()

		if h.Network().ResourceManager() != nil {
			h.Network().ResourceManager().Close()
		}
	})

	return nil
}

type streamWrapper struct {
	network.Stream
	rw io.ReadWriteCloser
}

func (s *streamWrapper) Read(b []byte) (int, error) {
	return s.rw.Read(b)
}

func (s *streamWrapper) Write(b []byte) (int, error) {
	return s.rw.Write(b)
}

func (s *streamWrapper) Close() error {
	return s.rw.Close()
}

func (s *streamWrapper) CloseWrite() error {
	// Flush the handshake before closing, but ignore the error. The other
	// end may have closed their side for reading.
	//
	// If something is wrong with the stream, the user will get on error on
	// read instead.
	if flusher, ok := s.rw.(interface{ Flush() error }); ok {
		_ = flusher.Flush()
	}
	return s.Stream.CloseWrite()
}
