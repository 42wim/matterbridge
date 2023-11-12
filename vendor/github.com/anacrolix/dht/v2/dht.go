package dht

import (
	"context"
	"crypto"
	_ "crypto/sha1"
	"errors"
	"math/rand"
	"net"
	"time"

	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/dht/v2/krpc"
	peer_store "github.com/anacrolix/dht/v2/peer-store"
	"github.com/anacrolix/log"
	"github.com/anacrolix/missinggo/v2"
	"github.com/anacrolix/torrent/iplist"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/time/rate"
)

func defaultQueryResendDelay() time.Duration {
	// This should be the highest reasonable UDP latency an end-user might have.
	return 2 * time.Second
}

// Uniquely identifies a transaction to us.
type transactionKey struct {
	RemoteAddr string // host:port
	T          string // The KRPC transaction ID.
}

type StartingNodesGetter func() ([]Addr, error)

// ServerConfig allows setting up a  configuration of the `Server` instance to be created with
// NewServer.
type ServerConfig struct {
	// Set NodeId Manually. Caller must ensure that if NodeId does not conform
	// to DHT Security Extensions, that NoSecurity is also set.
	NodeId krpc.ID
	Conn   net.PacketConn
	// Don't respond to queries from other nodes.
	Passive bool
	// Whether to wait for rate limiting to allow us to reply.
	WaitToReply bool

	StartingNodes StartingNodesGetter
	// Disable the DHT security extension: http://www.libtorrent.org/dht_sec.html.
	NoSecurity bool
	// Initial IP blocklist to use. Applied before serving and bootstrapping
	// begins.
	IPBlocklist iplist.Ranger
	// Used to secure the server's ID. Defaults to the Conn's LocalAddr(). Set to the IP that remote
	// nodes will see, as that IP is what they'll use to validate our ID.
	PublicIP net.IP

	// Hook received queries. Return false if you don't want to propagate to the default handlers.
	OnQuery func(query *krpc.Msg, source net.Addr) (propagate bool)
	// Called when a peer successfully announces to us.
	OnAnnouncePeer func(infoHash metainfo.Hash, ip net.IP, port int, portOk bool)
	// How long to wait before resending queries that haven't received a response. Defaults to a
	// random value between 4.5 and 5.5s.
	QueryResendDelay func() time.Duration
	// TODO: Expose Peers, to return NodeInfo for received get_peers queries.
	PeerStore peer_store.Interface
	// BEP-44: Storing arbitrary data in the DHT. If not store provided, a default in-memory
	// implementation will be used.
	Store bep44.Store
	// BEP-44: expiration time with non-announced items. Two hours by default
	Exp time.Duration

	// If no Logger is provided, log.Default is used and log.Debug messages are filtered out. Note
	// that all messages without a log.Level, have log.Debug added to them before being passed to
	// this Logger.
	Logger log.Logger

	DefaultWant []krpc.Want

	SendLimiter *rate.Limiter
}

// ServerStats instance is returned by Server.Stats() and stores Server metrics
type ServerStats struct {
	// Count of nodes in the node table that responded to our last query or
	// haven't yet been queried.
	GoodNodes int
	// Count of nodes in the node table.
	Nodes int
	// Transactions awaiting a response.
	OutstandingTransactions int
	// Individual announce_peer requests that got a success response.
	SuccessfulOutboundAnnouncePeerQueries int64
	// Nodes that have been blocked.
	BadNodes                 uint
	OutboundQueriesAttempted int64
}

func jitterDuration(average time.Duration, plusMinus time.Duration) time.Duration {
	return average - plusMinus/2 + time.Duration(rand.Int63n(int64(plusMinus)))
}

type Peer = krpc.NodeAddr

var DefaultGlobalBootstrapHostPorts = []string{
	"router.utorrent.com:6881",
	"router.bittorrent.com:6881",
	"dht.transmissionbt.com:6881",
	"dht.aelitis.com:6881",     // Vuze
	"router.silotis.us:6881",   // IPv6
	"dht.libtorrent.org:25401", // @arvidn's
	"dht.anacrolix.link:42069",
	"router.bittorrent.cloud:42069",
}

func GlobalBootstrapAddrs(network string) (addrs []Addr, err error) {
	initDnsResolver()
	for _, s := range DefaultGlobalBootstrapHostPorts {
		host, port, err := net.SplitHostPort(s)
		if err != nil {
			panic(err)
		}
		hostAddrs, err := dnsResolver.LookupHost(context.Background(), host)
		if err != nil {
			// log.Default.WithDefaultLevel(log.Debug).Printf("error looking up %q: %v", s, err)
			continue
		}
		for _, a := range hostAddrs {
			ua, err := net.ResolveUDPAddr("udp", net.JoinHostPort(a, port))
			if err != nil {
				log.Printf("error resolving %q: %v", a, err)
				continue
			}
			addrs = append(addrs, NewAddr(ua))
		}
	}
	if len(addrs) == 0 {
		err = errors.New("nothing resolved")
	}
	return
}

// Deprecated: Use function from krpc.
func RandomNodeID() (id krpc.ID) {
	return krpc.RandomNodeID()
}

func MakeDeterministicNodeID(public net.Addr) (id krpc.ID) {
	h := crypto.SHA1.New()
	h.Write([]byte(public.String()))
	h.Sum(id[:0:20])
	SecureNodeId(&id, missinggo.AddrIP(public))
	return
}
