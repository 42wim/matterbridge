package torrent

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"expvar"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anacrolix/chansync/events"
	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/log"
	"github.com/anacrolix/missinggo/perf"
	"github.com/anacrolix/missinggo/pubsub"
	"github.com/anacrolix/missinggo/v2"
	"github.com/anacrolix/missinggo/v2/bitmap"
	"github.com/anacrolix/missinggo/v2/pproffd"
	"github.com/anacrolix/sync"
	request_strategy "github.com/anacrolix/torrent/request-strategy"
	"github.com/davecgh/go-spew/spew"
	"github.com/dustin/go-humanize"
	"github.com/google/btree"
	"github.com/pion/datachannel"
	"golang.org/x/time/rate"

	"github.com/anacrolix/chansync"

	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/internal/limiter"
	"github.com/anacrolix/torrent/iplist"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/mse"
	pp "github.com/anacrolix/torrent/peer_protocol"
	"github.com/anacrolix/torrent/storage"
	"github.com/anacrolix/torrent/tracker"
	"github.com/anacrolix/torrent/webtorrent"
)

// Clients contain zero or more Torrents. A Client manages a blocklist, the
// TCP/UDP protocol ports, and DHT as desired.
type Client struct {
	// An aggregate of stats over all connections. First in struct to ensure 64-bit alignment of
	// fields. See #262.
	stats ConnStats

	_mu    lockWithDeferreds
	event  sync.Cond
	closed chansync.SetOnce

	config *ClientConfig
	logger log.Logger

	peerID         PeerID
	defaultStorage *storage.Client
	onClose        []func()
	dialers        []Dialer
	listeners      []Listener
	dhtServers     []DhtServer
	ipBlockList    iplist.Ranger

	// Set of addresses that have our client ID. This intentionally will
	// include ourselves if we end up trying to connect to our own address
	// through legitimate channels.
	dopplegangerAddrs map[string]struct{}
	badPeerIPs        map[string]struct{}
	torrents          map[InfoHash]*Torrent
	pieceRequestOrder map[interface{}]*request_strategy.PieceRequestOrder

	acceptLimiter   map[ipStr]int
	dialRateLimiter *rate.Limiter
	numHalfOpen     int

	websocketTrackers websocketTrackers

	activeAnnounceLimiter limiter.Instance
	webseedHttpClient     *http.Client
}

type ipStr string

func (cl *Client) BadPeerIPs() (ips []string) {
	cl.rLock()
	ips = cl.badPeerIPsLocked()
	cl.rUnlock()
	return
}

func (cl *Client) badPeerIPsLocked() (ips []string) {
	ips = make([]string, len(cl.badPeerIPs))
	i := 0
	for k := range cl.badPeerIPs {
		ips[i] = k
		i += 1
	}
	return
}

func (cl *Client) PeerID() PeerID {
	return cl.peerID
}

// Returns the port number for the first listener that has one. No longer assumes that all port
// numbers are the same, due to support for custom listeners. Returns zero if no port number is
// found.
func (cl *Client) LocalPort() (port int) {
	for i := 0; i < len(cl.listeners); i += 1 {
		if port = addrPortOrZero(cl.listeners[i].Addr()); port != 0 {
			return
		}
	}
	return
}

func writeDhtServerStatus(w io.Writer, s DhtServer) {
	dhtStats := s.Stats()
	fmt.Fprintf(w, " ID: %x\n", s.ID())
	spew.Fdump(w, dhtStats)
}

// Writes out a human readable status of the client, such as for writing to a
// HTTP status page.
func (cl *Client) WriteStatus(_w io.Writer) {
	cl.rLock()
	defer cl.rUnlock()
	w := bufio.NewWriter(_w)
	defer w.Flush()
	fmt.Fprintf(w, "Listen port: %d\n", cl.LocalPort())
	fmt.Fprintf(w, "Peer ID: %+q\n", cl.PeerID())
	fmt.Fprintf(w, "Extension bits: %v\n", cl.config.Extensions)
	fmt.Fprintf(w, "Announce key: %x\n", cl.announceKey())
	fmt.Fprintf(w, "Banned IPs: %d\n", len(cl.badPeerIPsLocked()))
	cl.eachDhtServer(func(s DhtServer) {
		fmt.Fprintf(w, "%s DHT server at %s:\n", s.Addr().Network(), s.Addr().String())
		writeDhtServerStatus(w, s)
	})
	spew.Fdump(w, &cl.stats)
	torrentsSlice := cl.torrentsAsSlice()
	fmt.Fprintf(w, "# Torrents: %d\n", len(torrentsSlice))
	fmt.Fprintln(w)
	sort.Slice(torrentsSlice, func(l, r int) bool {
		return torrentsSlice[l].infoHash.AsString() < torrentsSlice[r].infoHash.AsString()
	})
	for _, t := range torrentsSlice {
		if t.name() == "" {
			fmt.Fprint(w, "<unknown name>")
		} else {
			fmt.Fprint(w, t.name())
		}
		fmt.Fprint(w, "\n")
		if t.info != nil {
			fmt.Fprintf(
				w,
				"%f%% of %d bytes (%s)",
				100*(1-float64(t.bytesMissingLocked())/float64(t.info.TotalLength())),
				*t.length,
				humanize.Bytes(uint64(*t.length)))
		} else {
			w.WriteString("<missing metainfo>")
		}
		fmt.Fprint(w, "\n")
		t.writeStatus(w)
		fmt.Fprintln(w)
	}
}

func (cl *Client) initLogger() {
	logger := cl.config.Logger
	if logger.IsZero() {
		logger = log.Default
	}
	if cl.config.Debug {
		logger = logger.FilterLevel(log.Debug)
	}
	cl.logger = logger.WithValues(cl)
}

func (cl *Client) announceKey() int32 {
	return int32(binary.BigEndian.Uint32(cl.peerID[16:20]))
}

// Initializes a bare minimum Client. *Client and *ClientConfig must not be nil.
func (cl *Client) init(cfg *ClientConfig) {
	cl.config = cfg
	cl.dopplegangerAddrs = make(map[string]struct{})
	cl.torrents = make(map[metainfo.Hash]*Torrent)
	cl.dialRateLimiter = rate.NewLimiter(10, 10)
	cl.activeAnnounceLimiter.SlotsPerKey = 2
	cl.event.L = cl.locker()
	cl.ipBlockList = cfg.IPBlocklist
	cl.webseedHttpClient = &http.Client{
		Transport: &http.Transport{
			Proxy:           cfg.HTTPProxy,
			MaxConnsPerHost: 10,
		},
	}
}

func NewClient(cfg *ClientConfig) (cl *Client, err error) {
	if cfg == nil {
		cfg = NewDefaultClientConfig()
		cfg.ListenPort = 0
	}
	var client Client
	client.init(cfg)
	cl = &client
	go cl.acceptLimitClearer()
	cl.initLogger()
	defer func() {
		if err != nil {
			cl.Close()
			cl = nil
		}
	}()

	storageImpl := cfg.DefaultStorage
	if storageImpl == nil {
		// We'd use mmap by default but HFS+ doesn't support sparse files.
		storageImplCloser := storage.NewFile(cfg.DataDir)
		cl.onClose = append(cl.onClose, func() {
			if err := storageImplCloser.Close(); err != nil {
				cl.logger.Printf("error closing default storage: %s", err)
			}
		})
		storageImpl = storageImplCloser
	}
	cl.defaultStorage = storage.NewClient(storageImpl)

	if cfg.PeerID != "" {
		missinggo.CopyExact(&cl.peerID, cfg.PeerID)
	} else {
		o := copy(cl.peerID[:], cfg.Bep20)
		_, err = rand.Read(cl.peerID[o:])
		if err != nil {
			panic("error generating peer id")
		}
	}

	sockets, err := listenAll(cl.listenNetworks(), cl.config.ListenHost, cl.config.ListenPort, cl.firewallCallback, cl.logger)
	if err != nil {
		return
	}

	// Check for panics.
	cl.LocalPort()

	for _, _s := range sockets {
		s := _s // Go is fucking retarded.
		cl.onClose = append(cl.onClose, func() { go s.Close() })
		if peerNetworkEnabled(parseNetworkString(s.Addr().Network()), cl.config) {
			cl.dialers = append(cl.dialers, s)
			cl.listeners = append(cl.listeners, s)
			if cl.config.AcceptPeerConnections {
				go cl.acceptConnections(s)
			}
		}
	}

	go cl.forwardPort()
	if !cfg.NoDHT {
		for _, s := range sockets {
			if pc, ok := s.(net.PacketConn); ok {
				ds, err := cl.NewAnacrolixDhtServer(pc)
				if err != nil {
					panic(err)
				}
				cl.dhtServers = append(cl.dhtServers, AnacrolixDhtServerWrapper{ds})
				cl.onClose = append(cl.onClose, func() { ds.Close() })
			}
		}
	}

	cl.websocketTrackers = websocketTrackers{
		PeerId: cl.peerID,
		Logger: cl.logger,
		GetAnnounceRequest: func(event tracker.AnnounceEvent, infoHash [20]byte) (tracker.AnnounceRequest, error) {
			cl.lock()
			defer cl.unlock()
			t, ok := cl.torrents[infoHash]
			if !ok {
				return tracker.AnnounceRequest{}, errors.New("torrent not tracked by client")
			}
			return t.announceRequest(event), nil
		},
		Proxy: cl.config.HTTPProxy,
		OnConn: func(dc datachannel.ReadWriteCloser, dcc webtorrent.DataChannelContext) {
			cl.lock()
			defer cl.unlock()
			t, ok := cl.torrents[dcc.InfoHash]
			if !ok {
				cl.logger.WithDefaultLevel(log.Warning).Printf(
					"got webrtc conn for unloaded torrent with infohash %x",
					dcc.InfoHash,
				)
				dc.Close()
				return
			}
			go t.onWebRtcConn(dc, dcc)
		},
	}

	return
}

func (cl *Client) AddDhtServer(d DhtServer) {
	cl.dhtServers = append(cl.dhtServers, d)
}

// Adds a Dialer for outgoing connections. All Dialers are used when attempting to connect to a
// given address for any Torrent.
func (cl *Client) AddDialer(d Dialer) {
	cl.lock()
	defer cl.unlock()
	cl.dialers = append(cl.dialers, d)
	for _, t := range cl.torrents {
		t.openNewConns()
	}
}

func (cl *Client) Listeners() []Listener {
	return cl.listeners
}

// Registers a Listener, and starts Accepting on it. You must Close Listeners provided this way
// yourself.
func (cl *Client) AddListener(l Listener) {
	cl.listeners = append(cl.listeners, l)
	if cl.config.AcceptPeerConnections {
		go cl.acceptConnections(l)
	}
}

func (cl *Client) firewallCallback(net.Addr) bool {
	cl.rLock()
	block := !cl.wantConns() || !cl.config.AcceptPeerConnections
	cl.rUnlock()
	if block {
		torrent.Add("connections firewalled", 1)
	} else {
		torrent.Add("connections not firewalled", 1)
	}
	return block
}

func (cl *Client) listenOnNetwork(n network) bool {
	if n.Ipv4 && cl.config.DisableIPv4 {
		return false
	}
	if n.Ipv6 && cl.config.DisableIPv6 {
		return false
	}
	if n.Tcp && cl.config.DisableTCP {
		return false
	}
	if n.Udp && cl.config.DisableUTP && cl.config.NoDHT {
		return false
	}
	return true
}

func (cl *Client) listenNetworks() (ns []network) {
	for _, n := range allPeerNetworks {
		if cl.listenOnNetwork(n) {
			ns = append(ns, n)
		}
	}
	return
}

// Creates an anacrolix/dht Server, as would be done internally in NewClient, for the given conn.
func (cl *Client) NewAnacrolixDhtServer(conn net.PacketConn) (s *dht.Server, err error) {
	cfg := dht.ServerConfig{
		IPBlocklist:    cl.ipBlockList,
		Conn:           conn,
		OnAnnouncePeer: cl.onDHTAnnouncePeer,
		PublicIP: func() net.IP {
			if connIsIpv6(conn) && cl.config.PublicIp6 != nil {
				return cl.config.PublicIp6
			}
			return cl.config.PublicIp4
		}(),
		StartingNodes: cl.config.DhtStartingNodes(conn.LocalAddr().Network()),
		OnQuery:       cl.config.DHTOnQuery,
		Logger:        cl.logger.WithContextText(fmt.Sprintf("dht server on %v", conn.LocalAddr().String())),
	}
	if f := cl.config.ConfigureAnacrolixDhtServer; f != nil {
		f(&cfg)
	}
	s, err = dht.NewServer(&cfg)
	if err == nil {
		go func() {
			ts, err := s.Bootstrap()
			if err != nil {
				cl.logger.Printf("error bootstrapping dht: %s", err)
			}
			log.Fstr("%v completed bootstrap (%+v)", s, ts).AddValues(s, ts).Log(cl.logger)
		}()
	}
	return
}

func (cl *Client) Closed() events.Done {
	return cl.closed.Done()
}

func (cl *Client) eachDhtServer(f func(DhtServer)) {
	for _, ds := range cl.dhtServers {
		f(ds)
	}
}

// Stops the client. All connections to peers are closed and all activity will come to a halt.
func (cl *Client) Close() (errs []error) {
	var closeGroup sync.WaitGroup // For concurrent cleanup to complete before returning
	cl.lock()
	for _, t := range cl.torrents {
		err := t.close(&closeGroup)
		if err != nil {
			errs = append(errs, err)
		}
	}
	for i := range cl.onClose {
		cl.onClose[len(cl.onClose)-1-i]()
	}
	cl.closed.Set()
	cl.unlock()
	cl.event.Broadcast()
	closeGroup.Wait() // defer is LIFO. We want to Wait() after cl.unlock()
	return
}

func (cl *Client) ipBlockRange(ip net.IP) (r iplist.Range, blocked bool) {
	if cl.ipBlockList == nil {
		return
	}
	return cl.ipBlockList.Lookup(ip)
}

func (cl *Client) ipIsBlocked(ip net.IP) bool {
	_, blocked := cl.ipBlockRange(ip)
	return blocked
}

func (cl *Client) wantConns() bool {
	if cl.config.AlwaysWantConns {
		return true
	}
	for _, t := range cl.torrents {
		if t.wantConns() {
			return true
		}
	}
	return false
}

// TODO: Apply filters for non-standard networks, particularly rate-limiting.
func (cl *Client) rejectAccepted(conn net.Conn) error {
	if !cl.wantConns() {
		return errors.New("don't want conns right now")
	}
	ra := conn.RemoteAddr()
	if rip := addrIpOrNil(ra); rip != nil {
		if cl.config.DisableIPv4Peers && rip.To4() != nil {
			return errors.New("ipv4 peers disabled")
		}
		if cl.config.DisableIPv4 && len(rip) == net.IPv4len {
			return errors.New("ipv4 disabled")
		}
		if cl.config.DisableIPv6 && len(rip) == net.IPv6len && rip.To4() == nil {
			return errors.New("ipv6 disabled")
		}
		if cl.rateLimitAccept(rip) {
			return errors.New("source IP accepted rate limited")
		}
		if cl.badPeerIPPort(rip, missinggo.AddrPort(ra)) {
			return errors.New("bad source addr")
		}
	}
	return nil
}

func (cl *Client) acceptConnections(l Listener) {
	for {
		conn, err := l.Accept()
		torrent.Add("client listener accepts", 1)
		conn = pproffd.WrapNetConn(conn)
		cl.rLock()
		closed := cl.closed.IsSet()
		var reject error
		if !closed && conn != nil {
			reject = cl.rejectAccepted(conn)
		}
		cl.rUnlock()
		if closed {
			if conn != nil {
				conn.Close()
			}
			return
		}
		if err != nil {
			log.Fmsg("error accepting connection: %s", err).LogLevel(log.Debug, cl.logger)
			continue
		}
		go func() {
			if reject != nil {
				torrent.Add("rejected accepted connections", 1)
				cl.logger.LazyLog(log.Debug, func() log.Msg {
					return log.Fmsg("rejecting accepted conn: %v", reject)
				})
				conn.Close()
			} else {
				go cl.incomingConnection(conn)
			}
			cl.logger.LazyLog(log.Debug, func() log.Msg {
				return log.Fmsg("accepted %q connection at %q from %q",
					l.Addr().Network(),
					conn.LocalAddr(),
					conn.RemoteAddr(),
				)
			})
			torrent.Add(fmt.Sprintf("accepted conn remote IP len=%d", len(addrIpOrNil(conn.RemoteAddr()))), 1)
			torrent.Add(fmt.Sprintf("accepted conn network=%s", conn.RemoteAddr().Network()), 1)
			torrent.Add(fmt.Sprintf("accepted on %s listener", l.Addr().Network()), 1)
		}()
	}
}

// Creates the PeerConn.connString for a regular net.Conn PeerConn.
func regularNetConnPeerConnConnString(nc net.Conn) string {
	return fmt.Sprintf("%s-%s", nc.LocalAddr(), nc.RemoteAddr())
}

func (cl *Client) incomingConnection(nc net.Conn) {
	defer nc.Close()
	if tc, ok := nc.(*net.TCPConn); ok {
		tc.SetLinger(0)
	}
	c := cl.newConnection(nc, false, nc.RemoteAddr(), nc.RemoteAddr().Network(),
		regularNetConnPeerConnConnString(nc))
	defer func() {
		cl.lock()
		defer cl.unlock()
		c.close()
	}()
	c.Discovery = PeerSourceIncoming
	cl.runReceivedConn(c)
}

// Returns a handle to the given torrent, if it's present in the client.
func (cl *Client) Torrent(ih metainfo.Hash) (t *Torrent, ok bool) {
	cl.lock()
	defer cl.unlock()
	t, ok = cl.torrents[ih]
	return
}

func (cl *Client) torrent(ih metainfo.Hash) *Torrent {
	return cl.torrents[ih]
}

type DialResult struct {
	Conn   net.Conn
	Dialer Dialer
}

func countDialResult(err error) {
	if err == nil {
		torrent.Add("successful dials", 1)
	} else {
		torrent.Add("unsuccessful dials", 1)
	}
}

func reducedDialTimeout(minDialTimeout, max time.Duration, halfOpenLimit, pendingPeers int) (ret time.Duration) {
	ret = max / time.Duration((pendingPeers+halfOpenLimit)/halfOpenLimit)
	if ret < minDialTimeout {
		ret = minDialTimeout
	}
	return
}

// Returns whether an address is known to connect to a client with our own ID.
func (cl *Client) dopplegangerAddr(addr string) bool {
	_, ok := cl.dopplegangerAddrs[addr]
	return ok
}

// Returns a connection over UTP or TCP, whichever is first to connect.
func (cl *Client) dialFirst(ctx context.Context, addr string) (res DialResult) {
	return DialFirst(ctx, addr, cl.dialers)
}

// Returns a connection over UTP or TCP, whichever is first to connect.
func DialFirst(ctx context.Context, addr string, dialers []Dialer) (res DialResult) {
	{
		t := perf.NewTimer(perf.CallerName(0))
		defer func() {
			if res.Conn == nil {
				t.Mark(fmt.Sprintf("returned no conn (context: %v)", ctx.Err()))
			} else {
				t.Mark("returned conn over " + res.Dialer.DialerNetwork())
			}
		}()
	}
	ctx, cancel := context.WithCancel(ctx)
	// As soon as we return one connection, cancel the others.
	defer cancel()
	left := 0
	resCh := make(chan DialResult, left)
	for _, _s := range dialers {
		left++
		s := _s
		go func() {
			resCh <- DialResult{
				dialFromSocket(ctx, s, addr),
				s,
			}
		}()
	}
	// Wait for a successful connection.
	func() {
		defer perf.ScopeTimer()()
		for ; left > 0 && res.Conn == nil; left-- {
			res = <-resCh
		}
	}()
	// There are still incompleted dials.
	go func() {
		for ; left > 0; left-- {
			conn := (<-resCh).Conn
			if conn != nil {
				conn.Close()
			}
		}
	}()
	if res.Conn != nil {
		go torrent.Add(fmt.Sprintf("network dialed first: %s", res.Conn.RemoteAddr().Network()), 1)
	}
	return res
}

func dialFromSocket(ctx context.Context, s Dialer, addr string) net.Conn {
	c, err := s.Dial(ctx, addr)
	// This is a bit optimistic, but it looks non-trivial to thread this through the proxy code. Set
	// it now in case we close the connection forthwith.
	if tc, ok := c.(*net.TCPConn); ok {
		tc.SetLinger(0)
	}
	countDialResult(err)
	return c
}

func forgettableDialError(err error) bool {
	return strings.Contains(err.Error(), "no suitable address found")
}

func (cl *Client) noLongerHalfOpen(t *Torrent, addr string) {
	if _, ok := t.halfOpen[addr]; !ok {
		panic("invariant broken")
	}
	delete(t.halfOpen, addr)
	cl.numHalfOpen--
	for _, t := range cl.torrents {
		t.openNewConns()
	}
}

// Performs initiator handshakes and returns a connection. Returns nil *connection if no connection
// for valid reasons.
func (cl *Client) initiateProtocolHandshakes(
	ctx context.Context,
	nc net.Conn,
	t *Torrent,
	outgoing, encryptHeader bool,
	remoteAddr PeerRemoteAddr,
	network, connString string,
) (
	c *PeerConn, err error,
) {
	c = cl.newConnection(nc, outgoing, remoteAddr, network, connString)
	c.headerEncrypted = encryptHeader
	ctx, cancel := context.WithTimeout(ctx, cl.config.HandshakesTimeout)
	defer cancel()
	dl, ok := ctx.Deadline()
	if !ok {
		panic(ctx)
	}
	err = nc.SetDeadline(dl)
	if err != nil {
		panic(err)
	}
	err = cl.initiateHandshakes(c, t)
	return
}

// Returns nil connection and nil error if no connection could be established for valid reasons.
func (cl *Client) establishOutgoingConnEx(t *Torrent, addr PeerRemoteAddr, obfuscatedHeader bool) (*PeerConn, error) {
	dialCtx, cancel := context.WithTimeout(context.Background(), func() time.Duration {
		cl.rLock()
		defer cl.rUnlock()
		return t.dialTimeout()
	}())
	defer cancel()
	dr := cl.dialFirst(dialCtx, addr.String())
	nc := dr.Conn
	if nc == nil {
		if dialCtx.Err() != nil {
			return nil, fmt.Errorf("dialing: %w", dialCtx.Err())
		}
		return nil, errors.New("dial failed")
	}
	c, err := cl.initiateProtocolHandshakes(context.Background(), nc, t, true, obfuscatedHeader, addr, dr.Dialer.DialerNetwork(), regularNetConnPeerConnConnString(nc))
	if err != nil {
		nc.Close()
	}
	return c, err
}

// Returns nil connection and nil error if no connection could be established
// for valid reasons.
func (cl *Client) establishOutgoingConn(t *Torrent, addr PeerRemoteAddr) (c *PeerConn, err error) {
	torrent.Add("establish outgoing connection", 1)
	obfuscatedHeaderFirst := cl.config.HeaderObfuscationPolicy.Preferred
	c, err = cl.establishOutgoingConnEx(t, addr, obfuscatedHeaderFirst)
	if err == nil {
		torrent.Add("initiated conn with preferred header obfuscation", 1)
		return
	}
	// cl.logger.Printf("error establishing connection to %s (obfuscatedHeader=%t): %v", addr, obfuscatedHeaderFirst, err)
	if cl.config.HeaderObfuscationPolicy.RequirePreferred {
		// We should have just tried with the preferred header obfuscation. If it was required,
		// there's nothing else to try.
		return
	}
	// Try again with encryption if we didn't earlier, or without if we did.
	c, err = cl.establishOutgoingConnEx(t, addr, !obfuscatedHeaderFirst)
	if err == nil {
		torrent.Add("initiated conn with fallback header obfuscation", 1)
	}
	// cl.logger.Printf("error establishing fallback connection to %v: %v", addr, err)
	return
}

// Called to dial out and run a connection. The addr we're given is already
// considered half-open.
func (cl *Client) outgoingConnection(t *Torrent, addr PeerRemoteAddr, ps PeerSource, trusted bool) {
	cl.dialRateLimiter.Wait(context.Background())
	c, err := cl.establishOutgoingConn(t, addr)
	if err == nil {
		c.conn.SetWriteDeadline(time.Time{})
	}
	cl.lock()
	defer cl.unlock()
	// Don't release lock between here and addPeerConn, unless it's for
	// failure.
	cl.noLongerHalfOpen(t, addr.String())
	if err != nil {
		if cl.config.Debug {
			cl.logger.Printf("error establishing outgoing connection to %v: %v", addr, err)
		}
		return
	}
	defer c.close()
	c.Discovery = ps
	c.trusted = trusted
	t.runHandshookConnLoggingErr(c)
}

// The port number for incoming peer connections. 0 if the client isn't listening.
func (cl *Client) incomingPeerPort() int {
	return cl.LocalPort()
}

func (cl *Client) initiateHandshakes(c *PeerConn, t *Torrent) error {
	if c.headerEncrypted {
		var rw io.ReadWriter
		var err error
		rw, c.cryptoMethod, err = mse.InitiateHandshake(
			struct {
				io.Reader
				io.Writer
			}{c.r, c.w},
			t.infoHash[:],
			nil,
			cl.config.CryptoProvides,
		)
		c.setRW(rw)
		if err != nil {
			return fmt.Errorf("header obfuscation handshake: %w", err)
		}
	}
	ih, err := cl.connBtHandshake(c, &t.infoHash)
	if err != nil {
		return fmt.Errorf("bittorrent protocol handshake: %w", err)
	}
	if ih != t.infoHash {
		return errors.New("bittorrent protocol handshake: peer infohash didn't match")
	}
	return nil
}

// Calls f with any secret keys. Note that it takes the Client lock, and so must be used from code
// that won't also try to take the lock. This saves us copying all the infohashes everytime.
func (cl *Client) forSkeys(f func([]byte) bool) {
	cl.rLock()
	defer cl.rUnlock()
	if false { // Emulate the bug from #114
		var firstIh InfoHash
		for ih := range cl.torrents {
			firstIh = ih
			break
		}
		for range cl.torrents {
			if !f(firstIh[:]) {
				break
			}
		}
		return
	}
	for ih := range cl.torrents {
		if !f(ih[:]) {
			break
		}
	}
}

func (cl *Client) handshakeReceiverSecretKeys() mse.SecretKeyIter {
	if ret := cl.config.Callbacks.ReceiveEncryptedHandshakeSkeys; ret != nil {
		return ret
	}
	return cl.forSkeys
}

// Do encryption and bittorrent handshakes as receiver.
func (cl *Client) receiveHandshakes(c *PeerConn) (t *Torrent, err error) {
	defer perf.ScopeTimerErr(&err)()
	var rw io.ReadWriter
	rw, c.headerEncrypted, c.cryptoMethod, err = handleEncryption(c.rw(), cl.handshakeReceiverSecretKeys(), cl.config.HeaderObfuscationPolicy, cl.config.CryptoSelector)
	c.setRW(rw)
	if err == nil || err == mse.ErrNoSecretKeyMatch {
		if c.headerEncrypted {
			torrent.Add("handshakes received encrypted", 1)
		} else {
			torrent.Add("handshakes received unencrypted", 1)
		}
	} else {
		torrent.Add("handshakes received with error while handling encryption", 1)
	}
	if err != nil {
		if err == mse.ErrNoSecretKeyMatch {
			err = nil
		}
		return
	}
	if cl.config.HeaderObfuscationPolicy.RequirePreferred && c.headerEncrypted != cl.config.HeaderObfuscationPolicy.Preferred {
		err = errors.New("connection does not have required header obfuscation")
		return
	}
	ih, err := cl.connBtHandshake(c, nil)
	if err != nil {
		return nil, fmt.Errorf("during bt handshake: %w", err)
	}
	cl.lock()
	t = cl.torrents[ih]
	cl.unlock()
	return
}

var successfulPeerWireProtocolHandshakePeerReservedBytes expvar.Map

func init() {
	torrent.Set(
		"successful_peer_wire_protocol_handshake_peer_reserved_bytes",
		&successfulPeerWireProtocolHandshakePeerReservedBytes)
}

func (cl *Client) connBtHandshake(c *PeerConn, ih *metainfo.Hash) (ret metainfo.Hash, err error) {
	res, err := pp.Handshake(c.rw(), ih, cl.peerID, cl.config.Extensions)
	if err != nil {
		return
	}
	successfulPeerWireProtocolHandshakePeerReservedBytes.Add(res.PeerExtensionBits.String(), 1)
	ret = res.Hash
	c.PeerExtensionBytes = res.PeerExtensionBits
	c.PeerID = res.PeerID
	c.completedHandshake = time.Now()
	if cb := cl.config.Callbacks.CompletedHandshake; cb != nil {
		cb(c, res.Hash)
	}
	return
}

func (cl *Client) runReceivedConn(c *PeerConn) {
	err := c.conn.SetDeadline(time.Now().Add(cl.config.HandshakesTimeout))
	if err != nil {
		panic(err)
	}
	t, err := cl.receiveHandshakes(c)
	if err != nil {
		cl.logger.LazyLog(log.Debug, func() log.Msg {
			return log.Fmsg(
				"error receiving handshakes on %v: %s", c, err,
			).Add(
				"network", c.Network,
			)
		})
		torrent.Add("error receiving handshake", 1)
		cl.lock()
		cl.onBadAccept(c.RemoteAddr)
		cl.unlock()
		return
	}
	if t == nil {
		torrent.Add("received handshake for unloaded torrent", 1)
		cl.logger.LazyLog(log.Debug, func() log.Msg {
			return log.Fmsg("received handshake for unloaded torrent")
		})
		cl.lock()
		cl.onBadAccept(c.RemoteAddr)
		cl.unlock()
		return
	}
	torrent.Add("received handshake for loaded torrent", 1)
	c.conn.SetWriteDeadline(time.Time{})
	cl.lock()
	defer cl.unlock()
	t.runHandshookConnLoggingErr(c)
}

// Client lock must be held before entering this.
func (cl *Client) runHandshookConn(c *PeerConn, t *Torrent) error {
	c.setTorrent(t)
	for i, b := range cl.config.MinPeerExtensions {
		if c.PeerExtensionBytes[i]&b != b {
			return fmt.Errorf("peer did not meet minimum peer extensions: %x", c.PeerExtensionBytes[:])
		}
	}
	if c.PeerID == cl.peerID {
		if c.outgoing {
			connsToSelf.Add(1)
			addr := c.RemoteAddr.String()
			cl.dopplegangerAddrs[addr] = struct{}{}
		} /* else {
			// Because the remote address is not necessarily the same as its client's torrent listen
			// address, we won't record the remote address as a doppleganger. Instead, the initiator
			// can record *us* as the doppleganger.
		} */
		t.logger.Levelf(log.Debug, "local and remote peer ids are the same")
		return nil
	}
	c.r = deadlineReader{c.conn, c.r}
	completedHandshakeConnectionFlags.Add(c.connectionFlags(), 1)
	if connIsIpv6(c.conn) {
		torrent.Add("completed handshake over ipv6", 1)
	}
	if err := t.addPeerConn(c); err != nil {
		return fmt.Errorf("adding connection: %w", err)
	}
	defer t.dropConnection(c)
	c.startWriter()
	cl.sendInitialMessages(c, t)
	c.initUpdateRequestsTimer()
	err := c.mainReadLoop()
	if err != nil {
		return fmt.Errorf("main read loop: %w", err)
	}
	return nil
}

const check = false

func (p *Peer) initUpdateRequestsTimer() {
	if check {
		if p.updateRequestsTimer != nil {
			panic(p.updateRequestsTimer)
		}
	}
	p.updateRequestsTimer = time.AfterFunc(math.MaxInt64, p.updateRequestsTimerFunc)
}

const peerUpdateRequestsTimerReason = "updateRequestsTimer"

func (c *Peer) updateRequestsTimerFunc() {
	c.locker().Lock()
	defer c.locker().Unlock()
	if c.closed.IsSet() {
		return
	}
	if c.isLowOnRequests() {
		// If there are no outstanding requests, then a request update should have already run.
		return
	}
	if d := time.Since(c.lastRequestUpdate); d < updateRequestsTimerDuration {
		// These should be benign, Timer.Stop doesn't guarantee that its function won't run if it's
		// already been fired.
		torrent.Add("spurious timer requests updates", 1)
		return
	}
	c.updateRequests(peerUpdateRequestsTimerReason)
}

// Maximum pending requests we allow peers to send us. If peer requests are buffered on read, this
// instructs the amount of memory that might be used to cache pending writes. Assuming 512KiB
// (1<<19) cached for sending, for 16KiB (1<<14) chunks.
const localClientReqq = 1 << 5

// See the order given in Transmission's tr_peerMsgsNew.
func (cl *Client) sendInitialMessages(conn *PeerConn, torrent *Torrent) {
	if conn.PeerExtensionBytes.SupportsExtended() && cl.config.Extensions.SupportsExtended() {
		conn.write(pp.Message{
			Type:       pp.Extended,
			ExtendedID: pp.HandshakeExtendedID,
			ExtendedPayload: func() []byte {
				msg := pp.ExtendedHandshakeMessage{
					M: map[pp.ExtensionName]pp.ExtensionNumber{
						pp.ExtensionNameMetadata: metadataExtendedId,
					},
					V:            cl.config.ExtendedHandshakeClientVersion,
					Reqq:         localClientReqq,
					YourIp:       pp.CompactIp(conn.remoteIp()),
					Encryption:   cl.config.HeaderObfuscationPolicy.Preferred || !cl.config.HeaderObfuscationPolicy.RequirePreferred,
					Port:         cl.incomingPeerPort(),
					MetadataSize: torrent.metadataSize(),
					// TODO: We can figured these out specific to the socket
					// used.
					Ipv4: pp.CompactIp(cl.config.PublicIp4.To4()),
					Ipv6: cl.config.PublicIp6.To16(),
				}
				if !cl.config.DisablePEX {
					msg.M[pp.ExtensionNamePex] = pexExtendedId
				}
				return bencode.MustMarshal(msg)
			}(),
		})
	}
	func() {
		if conn.fastEnabled() {
			if torrent.haveAllPieces() {
				conn.write(pp.Message{Type: pp.HaveAll})
				conn.sentHaves.AddRange(0, bitmap.BitRange(conn.t.NumPieces()))
				return
			} else if !torrent.haveAnyPieces() {
				conn.write(pp.Message{Type: pp.HaveNone})
				conn.sentHaves.Clear()
				return
			}
		}
		conn.postBitfield()
	}()
	if conn.PeerExtensionBytes.SupportsDHT() && cl.config.Extensions.SupportsDHT() && cl.haveDhtServer() {
		conn.write(pp.Message{
			Type: pp.Port,
			Port: cl.dhtPort(),
		})
	}
}

func (cl *Client) dhtPort() (ret uint16) {
	if len(cl.dhtServers) == 0 {
		return
	}
	return uint16(missinggo.AddrPort(cl.dhtServers[len(cl.dhtServers)-1].Addr()))
}

func (cl *Client) haveDhtServer() bool {
	return len(cl.dhtServers) > 0
}

// Process incoming ut_metadata message.
func (cl *Client) gotMetadataExtensionMsg(payload []byte, t *Torrent, c *PeerConn) error {
	var d pp.ExtendedMetadataRequestMsg
	err := bencode.Unmarshal(payload, &d)
	if _, ok := err.(bencode.ErrUnusedTrailingBytes); ok {
	} else if err != nil {
		return fmt.Errorf("error unmarshalling bencode: %s", err)
	}
	piece := d.Piece
	switch d.Type {
	case pp.DataMetadataExtensionMsgType:
		c.allStats(add(1, func(cs *ConnStats) *Count { return &cs.MetadataChunksRead }))
		if !c.requestedMetadataPiece(piece) {
			return fmt.Errorf("got unexpected piece %d", piece)
		}
		c.metadataRequests[piece] = false
		begin := len(payload) - d.PieceSize()
		if begin < 0 || begin >= len(payload) {
			return fmt.Errorf("data has bad offset in payload: %d", begin)
		}
		t.saveMetadataPiece(piece, payload[begin:])
		c.lastUsefulChunkReceived = time.Now()
		err = t.maybeCompleteMetadata()
		if err != nil {
			// Log this at the Torrent-level, as we don't partition metadata by Peer yet, so we
			// don't know who to blame. TODO: Also errors can be returned here that aren't related
			// to verifying metadata, which should be fixed. This should be tagged with metadata, so
			// log consumers can filter for this message.
			t.logger.WithDefaultLevel(log.Warning).Printf("error completing metadata: %v", err)
		}
		return err
	case pp.RequestMetadataExtensionMsgType:
		if !t.haveMetadataPiece(piece) {
			c.write(t.newMetadataExtensionMessage(c, pp.RejectMetadataExtensionMsgType, d.Piece, nil))
			return nil
		}
		start := (1 << 14) * piece
		c.logger.WithDefaultLevel(log.Debug).Printf("sending metadata piece %d", piece)
		c.write(t.newMetadataExtensionMessage(c, pp.DataMetadataExtensionMsgType, piece, t.metadataBytes[start:start+t.metadataPieceSize(piece)]))
		return nil
	case pp.RejectMetadataExtensionMsgType:
		return nil
	default:
		return errors.New("unknown msg_type value")
	}
}

func (cl *Client) badPeerAddr(addr PeerRemoteAddr) bool {
	if ipa, ok := tryIpPortFromNetAddr(addr); ok {
		return cl.badPeerIPPort(ipa.IP, ipa.Port)
	}
	return false
}

func (cl *Client) badPeerIPPort(ip net.IP, port int) bool {
	if port == 0 {
		return true
	}
	if cl.dopplegangerAddr(net.JoinHostPort(ip.String(), strconv.FormatInt(int64(port), 10))) {
		return true
	}
	if _, ok := cl.ipBlockRange(ip); ok {
		return true
	}
	if _, ok := cl.badPeerIPs[ip.String()]; ok {
		return true
	}
	return false
}

// Return a Torrent ready for insertion into a Client.
func (cl *Client) newTorrent(ih metainfo.Hash, specStorage storage.ClientImpl) (t *Torrent) {
	return cl.newTorrentOpt(AddTorrentOpts{
		InfoHash: ih,
		Storage:  specStorage,
	})
}

// Return a Torrent ready for insertion into a Client.
func (cl *Client) newTorrentOpt(opts AddTorrentOpts) (t *Torrent) {
	// use provided storage, if provided
	storageClient := cl.defaultStorage
	if opts.Storage != nil {
		storageClient = storage.NewClient(opts.Storage)
	}

	t = &Torrent{
		cl:       cl,
		infoHash: opts.InfoHash,
		peers: prioritizedPeers{
			om: btree.New(32),
			getPrio: func(p PeerInfo) peerPriority {
				ipPort := p.addr()
				return bep40PriorityIgnoreError(cl.publicAddr(ipPort.IP), ipPort)
			},
		},
		conns: make(map[*PeerConn]struct{}, 2*cl.config.EstablishedConnsPerTorrent),

		halfOpen:          make(map[string]PeerInfo),
		pieceStateChanges: pubsub.NewPubSub(),

		storageOpener:       storageClient,
		maxEstablishedConns: cl.config.EstablishedConnsPerTorrent,

		metadataChanged: sync.Cond{
			L: cl.locker(),
		},
		webSeeds:     make(map[string]*Peer),
		gotMetainfoC: make(chan struct{}),
	}
	t.networkingEnabled.Set()
	t.logger = cl.logger.WithContextValue(t).WithNames("torrent", t.infoHash.HexString())
	if opts.ChunkSize == 0 {
		opts.ChunkSize = defaultChunkSize
	}
	t.setChunkSize(opts.ChunkSize)
	return
}

// A file-like handle to some torrent data resource.
type Handle interface {
	io.Reader
	io.Seeker
	io.Closer
	io.ReaderAt
}

func (cl *Client) AddTorrentInfoHash(infoHash metainfo.Hash) (t *Torrent, new bool) {
	return cl.AddTorrentInfoHashWithStorage(infoHash, nil)
}

// Adds a torrent by InfoHash with a custom Storage implementation.
// If the torrent already exists then this Storage is ignored and the
// existing torrent returned with `new` set to `false`
func (cl *Client) AddTorrentInfoHashWithStorage(infoHash metainfo.Hash, specStorage storage.ClientImpl) (t *Torrent, new bool) {
	cl.lock()
	defer cl.unlock()
	t, ok := cl.torrents[infoHash]
	if ok {
		return
	}
	new = true

	t = cl.newTorrent(infoHash, specStorage)
	cl.eachDhtServer(func(s DhtServer) {
		if cl.config.PeriodicallyAnnounceTorrentsToDht {
			go t.dhtAnnouncer(s)
		}
	})
	cl.torrents[infoHash] = t
	cl.clearAcceptLimits()
	t.updateWantPeersEvent()
	// Tickle Client.waitAccept, new torrent may want conns.
	cl.event.Broadcast()
	return
}

// Adds a torrent by InfoHash with a custom Storage implementation.
// If the torrent already exists then this Storage is ignored and the
// existing torrent returned with `new` set to `false`
func (cl *Client) AddTorrentOpt(opts AddTorrentOpts) (t *Torrent, new bool) {
	infoHash := opts.InfoHash
	cl.lock()
	defer cl.unlock()
	t, ok := cl.torrents[infoHash]
	if ok {
		return
	}
	new = true

	t = cl.newTorrentOpt(opts)
	cl.eachDhtServer(func(s DhtServer) {
		if cl.config.PeriodicallyAnnounceTorrentsToDht {
			go t.dhtAnnouncer(s)
		}
	})
	cl.torrents[infoHash] = t
	cl.clearAcceptLimits()
	t.updateWantPeersEvent()
	// Tickle Client.waitAccept, new torrent may want conns.
	cl.event.Broadcast()
	return
}

type AddTorrentOpts struct {
	InfoHash  InfoHash
	Storage   storage.ClientImpl
	ChunkSize pp.Integer
}

// Add or merge a torrent spec. Returns new if the torrent wasn't already in the client. See also
// Torrent.MergeSpec.
func (cl *Client) AddTorrentSpec(spec *TorrentSpec) (t *Torrent, new bool, err error) {
	t, new = cl.AddTorrentOpt(AddTorrentOpts{
		InfoHash:  spec.InfoHash,
		Storage:   spec.Storage,
		ChunkSize: spec.ChunkSize,
	})
	modSpec := *spec
	if new {
		// ChunkSize was already applied by adding a new Torrent, and MergeSpec disallows changing
		// it.
		modSpec.ChunkSize = 0
	}
	err = t.MergeSpec(&modSpec)
	if err != nil && new {
		t.Drop()
	}
	return
}

type stringAddr string

var _ net.Addr = stringAddr("")

func (stringAddr) Network() string   { return "" }
func (me stringAddr) String() string { return string(me) }

// The trackers will be merged with the existing ones. If the Info isn't yet known, it will be set.
// spec.DisallowDataDownload/Upload will be read and applied
// The display name is replaced if the new spec provides one. Note that any `Storage` is ignored.
func (t *Torrent) MergeSpec(spec *TorrentSpec) error {
	if spec.DisplayName != "" {
		t.SetDisplayName(spec.DisplayName)
	}
	t.initialPieceCheckDisabled = spec.DisableInitialPieceCheck
	if spec.InfoBytes != nil {
		err := t.SetInfoBytes(spec.InfoBytes)
		if err != nil {
			return err
		}
	}
	cl := t.cl
	cl.AddDhtNodes(spec.DhtNodes)
	cl.lock()
	defer cl.unlock()
	useTorrentSources(spec.Sources, t)
	for _, url := range spec.Webseeds {
		t.addWebSeed(url)
	}
	for _, peerAddr := range spec.PeerAddrs {
		t.addPeer(PeerInfo{
			Addr:    stringAddr(peerAddr),
			Source:  PeerSourceDirect,
			Trusted: true,
		})
	}
	if spec.ChunkSize != 0 {
		panic("chunk size cannot be changed for existing Torrent")
	}
	t.addTrackers(spec.Trackers)
	t.maybeNewConns()
	t.dataDownloadDisallowed.SetBool(spec.DisallowDataDownload)
	t.dataUploadDisallowed = spec.DisallowDataUpload
	return nil
}

func useTorrentSources(sources []string, t *Torrent) {
	// TODO: bind context to the lifetime of *Torrent so that it's cancelled if the torrent closes
	ctx := context.Background()
	for i := 0; i < len(sources); i += 1 {
		s := sources[i]
		go func() {
			if err := useTorrentSource(ctx, s, t); err != nil {
				t.logger.WithDefaultLevel(log.Warning).Printf("using torrent source %q: %v", s, err)
			} else {
				t.logger.Printf("successfully used source %q", s)
			}
		}()
	}
}

func useTorrentSource(ctx context.Context, source string, t *Torrent) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		select {
		case <-t.GotInfo():
		case <-t.Closed():
		case <-ctx.Done():
		}
		cancel()
	}()
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, source, nil); err != nil {
		panic(err)
	}
	var resp *http.Response
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	var mi metainfo.MetaInfo
	err = bencode.NewDecoder(resp.Body).Decode(&mi)
	resp.Body.Close()
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return
	}
	return t.MergeSpec(TorrentSpecFromMetaInfo(&mi))
}

func (cl *Client) dropTorrent(infoHash metainfo.Hash, wg *sync.WaitGroup) (err error) {
	t, ok := cl.torrents[infoHash]
	if !ok {
		err = fmt.Errorf("no such torrent")
		return
	}
	err = t.close(wg)
	if err != nil {
		panic(err)
	}
	delete(cl.torrents, infoHash)
	return
}

func (cl *Client) allTorrentsCompleted() bool {
	for _, t := range cl.torrents {
		if !t.haveInfo() {
			return false
		}
		if !t.haveAllPieces() {
			return false
		}
	}
	return true
}

// Returns true when all torrents are completely downloaded and false if the
// client is stopped before that.
func (cl *Client) WaitAll() bool {
	cl.lock()
	defer cl.unlock()
	for !cl.allTorrentsCompleted() {
		if cl.closed.IsSet() {
			return false
		}
		cl.event.Wait()
	}
	return true
}

// Returns handles to all the torrents loaded in the Client.
func (cl *Client) Torrents() []*Torrent {
	cl.lock()
	defer cl.unlock()
	return cl.torrentsAsSlice()
}

func (cl *Client) torrentsAsSlice() (ret []*Torrent) {
	for _, t := range cl.torrents {
		ret = append(ret, t)
	}
	return
}

func (cl *Client) AddMagnet(uri string) (T *Torrent, err error) {
	spec, err := TorrentSpecFromMagnetUri(uri)
	if err != nil {
		return
	}
	T, _, err = cl.AddTorrentSpec(spec)
	return
}

func (cl *Client) AddTorrent(mi *metainfo.MetaInfo) (T *Torrent, err error) {
	ts, err := TorrentSpecFromMetaInfoErr(mi)
	if err != nil {
		return
	}
	T, _, err = cl.AddTorrentSpec(ts)
	return
}

func (cl *Client) AddTorrentFromFile(filename string) (T *Torrent, err error) {
	mi, err := metainfo.LoadFromFile(filename)
	if err != nil {
		return
	}
	return cl.AddTorrent(mi)
}

func (cl *Client) DhtServers() []DhtServer {
	return cl.dhtServers
}

func (cl *Client) AddDhtNodes(nodes []string) {
	for _, n := range nodes {
		hmp := missinggo.SplitHostMaybePort(n)
		ip := net.ParseIP(hmp.Host)
		if ip == nil {
			cl.logger.Printf("won't add DHT node with bad IP: %q", hmp.Host)
			continue
		}
		ni := krpc.NodeInfo{
			Addr: krpc.NodeAddr{
				IP:   ip,
				Port: hmp.Port,
			},
		}
		cl.eachDhtServer(func(s DhtServer) {
			s.AddNode(ni)
		})
	}
}

func (cl *Client) banPeerIP(ip net.IP) {
	cl.logger.Printf("banning ip %v", ip)
	if cl.badPeerIPs == nil {
		cl.badPeerIPs = make(map[string]struct{})
	}
	cl.badPeerIPs[ip.String()] = struct{}{}
}

func (cl *Client) newConnection(nc net.Conn, outgoing bool, remoteAddr PeerRemoteAddr, network, connString string) (c *PeerConn) {
	if network == "" {
		panic(remoteAddr)
	}
	c = &PeerConn{
		Peer: Peer{
			outgoing:        outgoing,
			choking:         true,
			peerChoking:     true,
			PeerMaxRequests: 250,

			RemoteAddr: remoteAddr,
			Network:    network,
			callbacks:  &cl.config.Callbacks,
		},
		connString: connString,
		conn:       nc,
	}
	c.peerImpl = c
	c.logger = cl.logger.WithDefaultLevel(log.Warning).WithContextValue(c)
	c.setRW(connStatsReadWriter{nc, c})
	c.r = &rateLimitedReader{
		l: cl.config.DownloadRateLimiter,
		r: c.r,
	}
	c.logger.WithDefaultLevel(log.Debug).Printf("initialized with remote %v over network %v (outgoing=%t)", remoteAddr, network, outgoing)
	for _, f := range cl.config.Callbacks.NewPeer {
		f(&c.Peer)
	}
	return
}

func (cl *Client) onDHTAnnouncePeer(ih metainfo.Hash, ip net.IP, port int, portOk bool) {
	cl.lock()
	defer cl.unlock()
	t := cl.torrent(ih)
	if t == nil {
		return
	}
	t.addPeers([]PeerInfo{{
		Addr:   ipPortAddr{ip, port},
		Source: PeerSourceDhtAnnouncePeer,
	}})
}

func firstNotNil(ips ...net.IP) net.IP {
	for _, ip := range ips {
		if ip != nil {
			return ip
		}
	}
	return nil
}

func (cl *Client) eachListener(f func(Listener) bool) {
	for _, s := range cl.listeners {
		if !f(s) {
			break
		}
	}
}

func (cl *Client) findListener(f func(Listener) bool) (ret Listener) {
	for i := 0; i < len(cl.listeners); i += 1 {
		if ret = cl.listeners[i]; f(ret) {
			return
		}
	}
	return nil
}

func (cl *Client) publicIp(peer net.IP) net.IP {
	// TODO: Use BEP 10 to determine how peers are seeing us.
	if peer.To4() != nil {
		return firstNotNil(
			cl.config.PublicIp4,
			cl.findListenerIp(func(ip net.IP) bool { return ip.To4() != nil }),
		)
	}

	return firstNotNil(
		cl.config.PublicIp6,
		cl.findListenerIp(func(ip net.IP) bool { return ip.To4() == nil }),
	)
}

func (cl *Client) findListenerIp(f func(net.IP) bool) net.IP {
	l := cl.findListener(
		func(l Listener) bool {
			return f(addrIpOrNil(l.Addr()))
		},
	)
	if l == nil {
		return nil
	}
	return addrIpOrNil(l.Addr())
}

// Our IP as a peer should see it.
func (cl *Client) publicAddr(peer net.IP) IpPort {
	return IpPort{IP: cl.publicIp(peer), Port: uint16(cl.incomingPeerPort())}
}

// ListenAddrs addresses currently being listened to.
func (cl *Client) ListenAddrs() (ret []net.Addr) {
	cl.lock()
	ret = make([]net.Addr, len(cl.listeners))
	for i := 0; i < len(cl.listeners); i += 1 {
		ret[i] = cl.listeners[i].Addr()
	}
	cl.unlock()
	return
}

func (cl *Client) onBadAccept(addr PeerRemoteAddr) {
	ipa, ok := tryIpPortFromNetAddr(addr)
	if !ok {
		return
	}
	ip := maskIpForAcceptLimiting(ipa.IP)
	if cl.acceptLimiter == nil {
		cl.acceptLimiter = make(map[ipStr]int)
	}
	cl.acceptLimiter[ipStr(ip.String())]++
}

func maskIpForAcceptLimiting(ip net.IP) net.IP {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.Mask(net.CIDRMask(24, 32))
	}
	return ip
}

func (cl *Client) clearAcceptLimits() {
	cl.acceptLimiter = nil
}

func (cl *Client) acceptLimitClearer() {
	for {
		select {
		case <-cl.closed.Done():
			return
		case <-time.After(15 * time.Minute):
			cl.lock()
			cl.clearAcceptLimits()
			cl.unlock()
		}
	}
}

func (cl *Client) rateLimitAccept(ip net.IP) bool {
	if cl.config.DisableAcceptRateLimiting {
		return false
	}
	return cl.acceptLimiter[ipStr(maskIpForAcceptLimiting(ip).String())] > 0
}

func (cl *Client) rLock() {
	cl._mu.RLock()
}

func (cl *Client) rUnlock() {
	cl._mu.RUnlock()
}

func (cl *Client) lock() {
	cl._mu.Lock()
}

func (cl *Client) unlock() {
	cl._mu.Unlock()
}

func (cl *Client) locker() *lockWithDeferreds {
	return &cl._mu
}

func (cl *Client) String() string {
	return fmt.Sprintf("<%[1]T %[1]p>", cl)
}

// Returns connection-level aggregate stats at the Client level. See the comment on
// TorrentStats.ConnStats.
func (cl *Client) ConnStats() ConnStats {
	return cl.stats.Copy()
}
