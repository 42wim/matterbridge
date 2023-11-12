package node

import (
	"crypto/ecdsa"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peerstore"
	basichost "github.com/libp2p/go-libp2p/p2p/host/basic"
	"github.com/libp2p/go-libp2p/p2p/muxer/mplex"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pwebtransport "github.com/libp2p/go-libp2p/p2p/transport/webtransport"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/waku-org/go-waku/waku/v2/protocol/filter"
	"github.com/waku-org/go-waku/waku/v2/protocol/legacy_filter"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/store"
	"github.com/waku-org/go-waku/waku/v2/rendezvous"
	"github.com/waku-org/go-waku/waku/v2/timesource"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Default userAgent
const userAgent string = "go-waku"

// Default minRelayPeersToPublish
const defaultMinRelayPeersToPublish = 0

type WakuNodeParameters struct {
	hostAddr       *net.TCPAddr
	clusterID      uint16
	dns4Domain     string
	advertiseAddrs []multiaddr.Multiaddr
	multiAddr      []multiaddr.Multiaddr
	addressFactory basichost.AddrsFactory
	privKey        *ecdsa.PrivateKey
	libP2POpts     []libp2p.Option
	peerstore      peerstore.Peerstore
	prometheusReg  prometheus.Registerer

	circuitRelayMinInterval time.Duration
	circuitRelayBootDelay   time.Duration

	enableNTP bool
	ntpURLs   []string

	enableWS  bool
	wsPort    int
	enableWSS bool
	wssPort   int
	tlsConfig *tls.Config

	logger   *zap.Logger
	logLevel logging.LogLevel

	enableRelay            bool
	enableLegacyFilter     bool
	isLegacyFilterFullNode bool
	enableFilterLightNode  bool
	enableFilterFullNode   bool
	legacyFilterOpts       []legacy_filter.Option
	filterOpts             []filter.Option
	pubsubOpts             []pubsub.Option

	minRelayPeersToPublish int
	maxMsgSizeBytes        int

	enableStore     bool
	messageProvider store.MessageProvider

	enableRendezvousPoint bool
	rendezvousDB          *rendezvous.DB

	maxPeerConnections int
	peerStoreCapacity  int

	enableDiscV5     bool
	udpPort          uint
	discV5bootnodes  []*enode.Node
	discV5autoUpdate bool

	enablePeerExchange bool

	enableRLN                    bool
	rlnRelayMemIndex             *uint
	rlnRelayDynamic              bool
	rlnSpamHandler               func(message *pb.WakuMessage, topic string) error
	rlnETHClientAddress          string
	keystorePath                 string
	keystorePassword             string
	rlnTreePath                  string
	rlnMembershipContractAddress common.Address

	keepAliveInterval time.Duration

	enableLightPush bool

	connStatusC chan<- ConnStatus
	connNotifCh chan<- PeerConnection

	storeFactory storeFactory
}

type WakuNodeOption func(*WakuNodeParameters) error

// Default options used in the libp2p node
var DefaultWakuNodeOptions = []WakuNodeOption{
	WithPrometheusRegisterer(prometheus.NewRegistry()),
	WithMaxPeerConnections(50),
	WithCircuitRelayParams(2*time.Second, 3*time.Minute),
}

// MultiAddresses return the list of multiaddresses configured in the node
func (w WakuNodeParameters) MultiAddresses() []multiaddr.Multiaddr {
	return w.multiAddr
}

// Identity returns a libp2p option containing the identity used by the node
func (w WakuNodeParameters) Identity() config.Option {
	return libp2p.Identity(*w.GetPrivKey())
}

// TLSConfig returns the TLS config used for setting up secure websockets
func (w WakuNodeParameters) TLSConfig() *tls.Config {
	return w.tlsConfig
}

// AddressFactory returns the address factory used by the node's host
func (w WakuNodeParameters) AddressFactory() basichost.AddrsFactory {
	return w.addressFactory
}

// WithLogger is a WakuNodeOption that adds a custom logger
func WithLogger(l *zap.Logger) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.logger = l
		logging.SetPrimaryCore(l.Core())
		return nil
	}
}

// WithLogLevel is a WakuNodeOption that sets the log level for go-waku
func WithLogLevel(lvl zapcore.Level) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.logLevel = logging.LogLevel(lvl)
		logging.SetAllLoggers(params.logLevel)
		return nil
	}
}

// WithPrometheusRegisterer configures go-waku to use reg as the Registerer for all metrics subsystems
func WithPrometheusRegisterer(reg prometheus.Registerer) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		if reg == nil {
			return errors.New("registerer cannot be nil")
		}

		params.prometheusReg = reg
		return nil
	}
}

// WithDNS4Domain is a WakuNodeOption that adds a custom domain name to listen
func WithDNS4Domain(dns4Domain string) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.dns4Domain = dns4Domain
		previousAddrFactory := params.addressFactory
		params.addressFactory = func(inputAddr []multiaddr.Multiaddr) (addresses []multiaddr.Multiaddr) {
			addresses = append(addresses, inputAddr...)

			hostAddrMA, err := multiaddr.NewMultiaddr("/dns4/" + params.dns4Domain)
			if err != nil {
				panic(fmt.Sprintf("invalid dns4 address: %s", err.Error()))
			}

			tcp, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%d", params.hostAddr.Port))

			addresses = append(addresses, hostAddrMA.Encapsulate(tcp))

			if params.enableWS || params.enableWSS {
				if params.enableWSS {
					// WSS is deprecated in https://github.com/multiformats/multiaddr/pull/109
					wss, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%d/wss", params.wssPort))
					addresses = append(addresses, hostAddrMA.Encapsulate(wss))
					tlsws, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%d/tls/ws", params.wssPort))
					addresses = append(addresses, hostAddrMA.Encapsulate(tlsws))
				} else {
					ws, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%d/ws", params.wsPort))
					addresses = append(addresses, hostAddrMA.Encapsulate(ws))
				}
			}

			if previousAddrFactory != nil {
				return previousAddrFactory(addresses)
			}

			return addresses
		}

		return nil
	}
}

// WithHostAddress is a WakuNodeOption that configures libp2p to listen on a specific address
func WithHostAddress(hostAddr *net.TCPAddr) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.hostAddr = hostAddr
		hostAddrMA, err := manet.FromNetAddr(hostAddr)
		if err != nil {
			return err
		}
		params.multiAddr = append(params.multiAddr, hostAddrMA)

		return nil
	}
}

// WithAdvertiseAddresses is a WakuNodeOption that allows overriding the address used in the waku node with custom value
func WithAdvertiseAddresses(advertiseAddrs ...multiaddr.Multiaddr) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.advertiseAddrs = advertiseAddrs
		return WithMultiaddress(advertiseAddrs...)(params)
	}
}

// WithExternalIP is a WakuNodeOption that allows overriding the advertised external IP used in the waku node with custom value
func WithExternalIP(ip net.IP) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		oldAddrFactory := params.addressFactory
		params.addressFactory = func(inputAddr []multiaddr.Multiaddr) (addresses []multiaddr.Multiaddr) {
			addresses = append(addresses, inputAddr...)

			ipType := "/ip4/"
			if utils.IsIPv6(ip.String()) {
				ipType = "/ip6/"
			}

			hostAddrMA, err := multiaddr.NewMultiaddr(ipType + ip.String())
			if err != nil {
				panic("Could not build external IP")
			}

			addrSet := make(map[string]multiaddr.Multiaddr)
			for _, addr := range inputAddr {
				_, rest := multiaddr.SplitFirst(addr)

				addr := hostAddrMA.Encapsulate(rest)

				addrSet[addr.String()] = addr
			}

			for _, addr := range addrSet {
				addresses = append(addresses, addr)
			}

			if oldAddrFactory != nil {
				return oldAddrFactory(addresses)
			} else {
				return addresses
			}
		}
		return nil
	}
}

// WithMultiaddress is a WakuNodeOption that configures libp2p to listen on a list of multiaddresses
func WithMultiaddress(addresses ...multiaddr.Multiaddr) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.multiAddr = append(params.multiAddr, addresses...)
		return nil
	}
}

// WithPrivateKey is used to set an ECDSA private key in a libp2p node
func WithPrivateKey(privKey *ecdsa.PrivateKey) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.privKey = privKey
		return nil
	}
}

// WithClusterID is used to set the node's ClusterID
func WithClusterID(clusterID uint16) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.clusterID = clusterID
		return nil
	}
}

// WithNTP is used to use ntp for any operation that requires obtaining time
// A list of ntp servers can be passed but if none is specified, some defaults
// will be used
func WithNTP(ntpURLs ...string) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		if len(ntpURLs) == 0 {
			ntpURLs = timesource.DefaultServers
		}

		params.enableNTP = true
		params.ntpURLs = ntpURLs
		return nil
	}
}

// GetPrivKey returns the private key used in the node
func (w *WakuNodeParameters) GetPrivKey() *crypto.PrivKey {
	privKey := crypto.PrivKey(utils.EcdsaPrivKeyToSecp256k1PrivKey(w.privKey))
	return &privKey
}

// WithLibP2POptions is a WakuNodeOption used to configure the libp2p node.
// This can potentially override any libp2p config that was set with other
// WakuNodeOption
func WithLibP2POptions(opts ...libp2p.Option) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.libP2POpts = opts
		return nil
	}
}

func WithPeerStore(ps peerstore.Peerstore) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.peerstore = ps
		return nil
	}
}

// WithWakuRelay enables the Waku V2 Relay protocol. This WakuNodeOption
// accepts a list of WakuRelay gossipsub option to setup the protocol
func WithWakuRelay(opts ...pubsub.Option) WakuNodeOption {
	return WithWakuRelayAndMinPeers(defaultMinRelayPeersToPublish, opts...)
}

// WithWakuRelayAndMinPeers enables the Waku V2 Relay protocol. This WakuNodeOption
// accepts a min peers require to publish and a list of WakuRelay gossipsub option to setup the protocol
func WithWakuRelayAndMinPeers(minRelayPeersToPublish int, opts ...pubsub.Option) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableRelay = true
		params.pubsubOpts = opts
		params.minRelayPeersToPublish = minRelayPeersToPublish
		return nil
	}
}

func WithMaxMsgSize(maxMsgSizeBytes int) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.maxMsgSizeBytes = maxMsgSizeBytes
		return nil
	}
}

func WithMaxPeerConnections(maxPeers int) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.maxPeerConnections = maxPeers
		return nil
	}
}

func WithPeerStoreCapacity(capacity int) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.peerStoreCapacity = capacity
		return nil
	}
}

// WithDiscoveryV5 is a WakuOption used to enable DiscV5 peer discovery
func WithDiscoveryV5(udpPort uint, bootnodes []*enode.Node, autoUpdate bool) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableDiscV5 = true
		params.udpPort = udpPort
		params.discV5bootnodes = bootnodes
		params.discV5autoUpdate = autoUpdate
		return nil
	}
}

// WithPeerExchange is a WakuOption used to enable Peer Exchange
func WithPeerExchange() WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enablePeerExchange = true
		return nil
	}
}

// WithLegacyWakuFilter enables the legacy Waku Filter protocol. This WakuNodeOption
// accepts a list of WakuFilter gossipsub options to setup the protocol
func WithLegacyWakuFilter(fullnode bool, filterOpts ...legacy_filter.Option) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableLegacyFilter = true
		params.isLegacyFilterFullNode = fullnode
		params.legacyFilterOpts = filterOpts
		return nil
	}
}

// WithWakuFilter enables the Waku Filter V2 protocol for lightnode functionality
func WithWakuFilterLightNode() WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableFilterLightNode = true
		return nil
	}
}

// WithWakuFilterFullNode enables the Waku Filter V2 protocol full node functionality.
// This WakuNodeOption accepts a list of WakuFilter options to setup the protocol
func WithWakuFilterFullNode(filterOpts ...filter.Option) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableFilterFullNode = true
		params.filterOpts = filterOpts
		return nil
	}
}

// WithWakuStore enables the Waku V2 Store protocol and if the messages should
// be stored or not in a message provider.
func WithWakuStore() WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableStore = true
		return nil
	}
}

// WithWakuStoreFactory is used to replace the default WakuStore with a custom
// implementation that implements the store.Store interface
func WithWakuStoreFactory(factory storeFactory) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.storeFactory = factory

		return nil
	}
}

// WithMessageProvider is a WakuNodeOption that sets the MessageProvider
// used to store and retrieve persisted messages
func WithMessageProvider(s store.MessageProvider) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		if s == nil {
			return errors.New("message provider can't be nil")
		}
		params.messageProvider = s
		return nil
	}
}

// WithLightPush is a WakuNodeOption that enables the lightpush protocol
func WithLightPush() WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableLightPush = true
		return nil
	}
}

// WithKeepAlive is a WakuNodeOption used to set the interval of time when
// each peer will be ping to keep the TCP connection alive
func WithKeepAlive(t time.Duration) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.keepAliveInterval = t
		return nil
	}
}

// WithConnectionStatusChannel is a WakuNodeOption used to set a channel where the
// connection status changes will be pushed to. It's useful to identify when peer
// connections and disconnections occur
func WithConnectionStatusChannel(connStatus chan ConnStatus) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.connStatusC = connStatus
		return nil
	}
}

func WithConnectionNotification(ch chan<- PeerConnection) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.connNotifCh = ch
		return nil
	}
}

// WithWebsockets is a WakuNodeOption used to enable websockets support
func WithWebsockets(address string, port int) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableWS = true
		params.wsPort = port

		wsMa, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d/%s", address, port, "ws"))
		if err != nil {
			return err
		}

		params.multiAddr = append(params.multiAddr, wsMa)

		return nil
	}
}

// WithRendezvous is a WakuOption used to set the node as a rendezvous
// point, using an specific storage for the peer information
func WithRendezvous(db *rendezvous.DB) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableRendezvousPoint = true
		params.rendezvousDB = db
		return nil
	}
}

// WithSecureWebsockets is a WakuNodeOption used to enable secure websockets support
func WithSecureWebsockets(address string, port int, certPath string, keyPath string) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.enableWSS = true
		params.wssPort = port

		wsMa, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d/%s", address, port, "wss"))
		if err != nil {
			return err
		}
		params.multiAddr = append(params.multiAddr, wsMa)

		certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return err
		}
		params.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS12,
		}

		return nil
	}
}

func WithCircuitRelayParams(minInterval time.Duration, bootDelay time.Duration) WakuNodeOption {
	return func(params *WakuNodeParameters) error {
		params.circuitRelayBootDelay = bootDelay
		params.circuitRelayMinInterval = minInterval
		return nil
	}
}

// Default options used in the libp2p node
var DefaultLibP2POptions = []libp2p.Option{
	libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(quic.NewTransport),
		libp2p.Transport(libp2pwebtransport.New),
	),
	libp2p.UserAgent(userAgent),
	libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	),
	libp2p.EnableNATService(),
	libp2p.ConnectionManager(newConnManager(200, 300, connmgr.WithGracePeriod(0))),
	libp2p.EnableHolePunching(),
}

func newConnManager(lo int, hi int, opts ...connmgr.Option) *connmgr.BasicConnMgr {
	mgr, err := connmgr.NewConnManager(lo, hi, opts...)
	if err != nil {
		panic("could not create ConnManager: " + err.Error())
	}
	return mgr
}
