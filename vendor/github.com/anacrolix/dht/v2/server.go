package dht

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime/pprof"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/anacrolix/log"
	"github.com/anacrolix/missinggo/v2"
	"github.com/anacrolix/sync"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/anacrolix/torrent/iplist"
	"github.com/anacrolix/torrent/logonce"
	"github.com/anacrolix/torrent/metainfo"

	"github.com/anacrolix/torrent/bencode"

	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/dht/v2/int160"
	"github.com/anacrolix/dht/v2/krpc"
	peer_store "github.com/anacrolix/dht/v2/peer-store"
	"github.com/anacrolix/dht/v2/traversal"
	"github.com/anacrolix/dht/v2/types"
)

// A Server defines parameters for a DHT node server that is able to send
// queries, and respond to the ones from the network. Each node has a globally
// unique identifier known as the "node ID." Node IDs are chosen at random
// from the same 160-bit space as BitTorrent infohashes and define the
// behaviour of the node. Zero valued Server does not have a valid ID and thus
// is unable to function properly. Use `NewServer(nil)` to initialize a
// default node.
type Server struct {
	id          int160.T
	socket      net.PacketConn
	resendDelay func() time.Duration

	mu           sync.RWMutex
	transactions map[transactionKey]*Transaction
	nextT        uint64 // unique "t" field for outbound queries
	table        table
	closed       missinggo.Event
	ipBlockList  iplist.Ranger
	tokenServer  tokenServer // Manages tokens we issue to our queriers.
	config       ServerConfig
	stats        ServerStats
	sendLimit    *rate.Limiter

	lastBootstrap    time.Time
	bootstrappingNow bool

	store *bep44.Wrapper
}

func (s *Server) numGoodNodes() (num int) {
	s.table.forNodes(func(n *node) bool {
		if s.IsGood(n) {
			num++
		}
		return true
	})
	return
}

func prettySince(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	d /= time.Second
	d *= time.Second
	return fmt.Sprintf("%s ago", d)
}

func (s *Server) WriteStatus(w io.Writer) {
	fmt.Fprintf(w, "Listening on %s\n", s.Addr())
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintf(w, "Nodes in table: %d good, %d total\n", s.numGoodNodes(), s.numNodes())
	fmt.Fprintf(w, "Ongoing transactions: %d\n", len(s.transactions))
	fmt.Fprintf(w, "Server node ID: %x\n", s.id.Bytes())
	for i, b := range s.table.buckets {
		if b.Len() == 0 && b.lastChanged.IsZero() {
			continue
		}
		fmt.Fprintf(w,
			"b# %v: %v nodes, last updated: %v\n",
			i, b.Len(), prettySince(b.lastChanged))
		if b.Len() > 0 {
			tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
			fmt.Fprintf(tw, "  node id\taddr\tlast query\tlast response\trecv\tdiscard\tflags\n")
			b.EachNode(func(n *node) bool {
				var flags []string
				if s.IsQuestionable(n) {
					flags = append(flags, "q10e")
				}
				if s.nodeIsBad(n) {
					flags = append(flags, "bad")
				}
				if s.IsGood(n) {
					flags = append(flags, "good")
				}
				if n.IsSecure() {
					flags = append(flags, "sec")
				}
				fmt.Fprintf(tw, "  %x\t%s\t%s\t%s\t%d\t%v\t%v\n",
					n.Id.Bytes(),
					n.Addr,
					prettySince(n.lastGotQuery),
					prettySince(n.lastGotResponse),
					n.numReceivesFrom,
					n.failedLastQuestionablePing,
					strings.Join(flags, ","),
				)
				return true
			})
			tw.Flush()
		}
	}
	fmt.Fprintln(w)
}

func (s *Server) numNodes() (num int) {
	s.table.forNodes(func(n *node) bool {
		num++
		return true
	})
	return
}

// Stats returns statistics for the server.
func (s *Server) Stats() ServerStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	ss := s.stats
	ss.GoodNodes = s.numGoodNodes()
	ss.Nodes = s.numNodes()
	ss.OutstandingTransactions = len(s.transactions)
	return ss
}

// Addr returns the listen address for the server. Packets arriving to this address
// are processed by the server (unless aliens are involved).
func (s *Server) Addr() net.Addr {
	return s.socket.LocalAddr()
}

func NewDefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		NoSecurity:    true,
		StartingNodes: func() ([]Addr, error) { return GlobalBootstrapAddrs("udp") },
		DefaultWant:   []krpc.Want{krpc.WantNodes, krpc.WantNodes6},
		Store:         bep44.NewMemory(),
		Exp:           2 * time.Hour,
		SendLimiter:   DefaultSendLimiter,
	}
}

// If the NodeId hasn't been specified, generate one and secure it against the PublicIP if
// NoSecurity is not set.
func (c *ServerConfig) InitNodeId() {
	if missinggo.IsZeroValue(c.NodeId) {
		c.NodeId = RandomNodeID()
		if !c.NoSecurity && c.PublicIP != nil {
			SecureNodeId(&c.NodeId, c.PublicIP)
		}
	}
}

// NewServer initializes a new DHT node server.
func NewServer(c *ServerConfig) (s *Server, err error) {
	if c == nil {
		c = NewDefaultServerConfig()
	}
	if c.Conn == nil {
		c.Conn, err = net.ListenPacket("udp", ":0")
		if err != nil {
			return
		}
	}
	c.InitNodeId()
	// If Logger is empty, emulate the old behaviour: Everything is logged to the default location,
	// and there are no debug messages.
	if c.Logger.IsZero() {
		c.Logger = log.Default.FilterLevel(log.Info)
	}
	// Add log.Debug by default.
	c.Logger = c.Logger.WithDefaultLevel(log.Debug)

	if c.Store == nil {
		c.Store = bep44.NewMemory()
	}
	if c.SendLimiter == nil {
		c.SendLimiter = DefaultSendLimiter
	}

	s = &Server{
		config:      *c,
		ipBlockList: c.IPBlocklist,
		tokenServer: tokenServer{
			maxIntervalDelta: 2,
			interval:         5 * time.Minute,
			secret:           make([]byte, 20),
		},
		transactions: make(map[transactionKey]*Transaction),
		table: table{
			k: 8,
		},
		store: bep44.NewWrapper(c.Store, c.Exp),
	}
	rand.Read(s.tokenServer.secret)
	s.socket = c.Conn
	s.id = int160.FromByteArray(c.NodeId)
	s.table.rootID = s.id
	s.resendDelay = s.config.QueryResendDelay
	if s.resendDelay == nil {
		s.resendDelay = defaultQueryResendDelay
	}
	go s.serveUntilClosed()
	return
}

func (s *Server) serveUntilClosed() {
	err := s.serve()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.IsSet() {
		return
	}
	if err != nil {
		panic(err)
	}
}

// Returns a description of the Server.
func (s *Server) String() string {
	return fmt.Sprintf("dht server on %s (node id %v)", s.socket.LocalAddr(), s.id)
}

// Packets to and from any address matching a range in the list are dropped.
func (s *Server) SetIPBlockList(list iplist.Ranger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ipBlockList = list
}

func (s *Server) IPBlocklist() iplist.Ranger {
	return s.ipBlockList
}

func (s *Server) processPacket(b []byte, addr Addr) {
	// log.Printf("got packet %q", b)
	if len(b) < 2 || b[0] != 'd' {
		// KRPC messages are bencoded dicts.
		readNotKRPCDict.Add(1)
		return
	}
	var d krpc.Msg
	err := bencode.Unmarshal(b, &d)
	if _, ok := err.(bencode.ErrUnusedTrailingBytes); ok {
		// log.Printf("%s: received message packet with %d trailing bytes: %q", s, _err.NumUnusedBytes, b[len(b)-_err.NumUnusedBytes:])
		expvars.Add("processed packets with trailing bytes", 1)
	} else if err != nil {
		readUnmarshalError.Add(1)
		// log.Printf("%s: received bad krpc message from %s: %s: %+q", s, addr, err, b)
		func() {
			if se, ok := err.(*bencode.SyntaxError); ok {
				// The message was truncated.
				if int(se.Offset) == len(b) {
					return
				}
				// Some messages seem to drop to nul chars abrubtly.
				if int(se.Offset) < len(b) && b[se.Offset] == 0 {
					return
				}
				// The message isn't bencode from the first.
				if se.Offset == 0 {
					return
				}
			}
			// if missinggo.CryHeard() {
			log.Printf("%s: received bad krpc message from %s: %s: %+q", s, addr, err, b)
			// }
		}()
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed.IsSet() {
		return
	}
	if d.Y == "q" {
		expvars.Add("received queries", 1)
		s.logger().Printf("received query %q from %v", d.Q, addr)
		s.handleQuery(addr, d)
		return
	}
	tk := transactionKey{
		RemoteAddr: addr.String(),
		T:          d.T,
	}
	t, ok := s.transactions[tk]
	if !ok {
		s.logger().Printf("received response for untracked transaction %q from %v", d.T, addr)
		return
	}
	// s.logger().Printf("received response for transaction %q from %v", d.T, addr)
	go t.handleResponse(d)
	s.updateNode(addr, d.SenderID(), !d.ReadOnly, func(n *node) {
		n.lastGotResponse = time.Now()
		n.failedLastQuestionablePing = false
		n.numReceivesFrom++
	})
	// Ensure we don't provide more than one response to a transaction.
	s.deleteTransaction(tk)
}

func (s *Server) serve() error {
	var b [0x10000]byte
	for {
		n, addr, err := s.socket.ReadFrom(b[:])
		if err != nil {
			return err
		}
		expvars.Add("packets read", 1)
		if n == len(b) {
			logonce.Stderr.Printf("received dht packet exceeds buffer size")
			continue
		}
		if missinggo.AddrPort(addr) == 0 {
			readZeroPort.Add(1)
			continue
		}
		blocked, err := func() (bool, error) {
			s.mu.RLock()
			defer s.mu.RUnlock()
			if s.closed.IsSet() {
				return false, errors.New("server is closed")
			}
			return s.ipBlocked(missinggo.AddrIP(addr)), nil
		}()
		if err != nil {
			return err
		}
		if blocked {
			readBlocked.Add(1)
			continue
		}
		s.processPacket(b[:n], NewAddr(addr))
	}
}

func (s *Server) ipBlocked(ip net.IP) (blocked bool) {
	if s.ipBlockList == nil {
		return
	}
	_, blocked = s.ipBlockList.Lookup(ip)
	return
}

// Adds directly to the node table.
func (s *Server) AddNode(ni krpc.NodeInfo) error {
	id := int160.FromByteArray(ni.ID)
	if id.IsZero() {
		go s.Ping(ni.Addr.UDP())
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateNode(NewAddr(ni.Addr.UDP()), (*krpc.ID)(&ni.ID), true, func(*node) {})
}

func wantsContain(ws []krpc.Want, w krpc.Want) bool {
	for _, _w := range ws {
		if _w == w {
			return true
		}
	}
	return false
}

func shouldReturnNodes(queryWants []krpc.Want, querySource net.IP) bool {
	if len(queryWants) != 0 {
		return wantsContain(queryWants, krpc.WantNodes)
	}
	// Is it possible to be over IPv6 with IPv4 endpoints?
	return querySource.To4() != nil
}

func shouldReturnNodes6(queryWants []krpc.Want, querySource net.IP) bool {
	if len(queryWants) != 0 {
		return wantsContain(queryWants, krpc.WantNodes6)
	}
	return querySource.To4() == nil
}

func (s *Server) makeReturnNodes(target int160.T, filter func(krpc.NodeAddr) bool) []krpc.NodeInfo {
	return s.closestGoodNodeInfos(8, target, filter)
}

var krpcErrMissingArguments = krpc.Error{
	Code: krpc.ErrorCodeProtocolError,
	Msg:  "missing arguments dict",
}

// Filters peers per BEP 32 to return in the values field to a get_peers query.
func filterPeers(querySourceIp net.IP, queryWants []krpc.Want, allPeers []krpc.NodeAddr) (filtered []krpc.NodeAddr) {
	// The logic here is common with nodes, see BEP 32.
	retain4 := shouldReturnNodes(queryWants, querySourceIp)
	retain6 := shouldReturnNodes6(queryWants, querySourceIp)
	for _, peer := range allPeers {
		if ip, ok := func(ip net.IP) (net.IP, bool) {
			as4 := peer.IP.To4()
			as16 := peer.IP.To16()
			switch {
			case retain4 && len(ip) == net.IPv4len:
				return ip, true
			case retain6 && len(ip) == net.IPv6len:
				return ip, true
			case retain4 && as4 != nil:
				// Is it possible that we're converting to an IPv4 address when the transport in use
				// is IPv6?
				return as4, true
			case retain6 && as16 != nil:
				// Couldn't any IPv4 address be converted to IPv6, but isn't listening over IPv6?
				return as16, true
			default:
				return nil, false
			}
		}(peer.IP); ok {
			filtered = append(filtered, krpc.NodeAddr{IP: ip, Port: peer.Port})
		}
	}
	return
}

func (s *Server) setReturnNodes(r *krpc.Return, queryMsg krpc.Msg, querySource Addr) *krpc.Error {
	if queryMsg.A == nil {
		return &krpcErrMissingArguments
	}
	target := int160.FromByteArray(queryMsg.A.InfoHash)
	if shouldReturnNodes(queryMsg.A.Want, querySource.IP()) {
		r.Nodes = s.makeReturnNodes(target, func(na krpc.NodeAddr) bool { return na.IP.To4() != nil })
	}
	if shouldReturnNodes6(queryMsg.A.Want, querySource.IP()) {
		r.Nodes6 = s.makeReturnNodes(target, func(krpc.NodeAddr) bool { return true })
	}
	return nil
}

func (s *Server) handleQuery(source Addr, m krpc.Msg) {
	go func() {
		expvars.Add(fmt.Sprintf("received query %q", m.Q), 1)
		if a := m.A; a != nil {
			if a.NoSeed != 0 {
				expvars.Add("received argument noseed", 1)
			}
			if a.Scrape != 0 {
				expvars.Add("received argument scrape", 1)
			}
		}
	}()
	s.updateNode(source, m.SenderID(), !m.ReadOnly, func(n *node) {
		n.lastGotQuery = time.Now()
		n.numReceivesFrom++
	})
	if s.config.OnQuery != nil {
		propagate := s.config.OnQuery(&m, source.Raw())
		if !propagate {
			return
		}
	}
	// Don't respond.
	if s.config.Passive {
		return
	}
	// TODO: Should we disallow replying to ourself?
	args := m.A
	switch m.Q {
	case "ping":
		s.reply(source, m.T, krpc.Return{})
	case "get_peers":
		// Check for the naked m.A.Want deref below.
		if m.A == nil {
			s.sendError(source, m.T, krpcErrMissingArguments)
			break
		}
		var r krpc.Return
		if ps := s.config.PeerStore; ps != nil {
			r.Values = filterPeers(source.IP(), m.A.Want, ps.GetPeers(peer_store.InfoHash(args.InfoHash)))
			r.Token = func() *string {
				t := s.createToken(source)
				return &t
			}()
		}
		if len(r.Values) == 0 {
			if err := s.setReturnNodes(&r, m, source); err != nil {
				s.sendError(source, m.T, *err)
				break
			}
		}
		s.reply(source, m.T, r)
	case "find_node":
		var r krpc.Return
		if err := s.setReturnNodes(&r, m, source); err != nil {
			s.sendError(source, m.T, *err)
			break
		}
		s.reply(source, m.T, r)
	case "announce_peer":
		if !s.validToken(args.Token, source) {
			expvars.Add("received announce_peer with invalid token", 1)
			return
		}
		expvars.Add("received announce_peer with valid token", 1)

		var port int
		portOk := false
		if args.Port != nil {
			port = *args.Port
			portOk = true
		}
		if args.ImpliedPort {
			expvars.Add("received announce_peer with implied_port", 1)
			port = source.Port()
			portOk = true
		}
		if !portOk {
			expvars.Add("received announce_peer with no derivable port", 1)
		}

		if h := s.config.OnAnnouncePeer; h != nil {
			go h(metainfo.Hash(args.InfoHash), source.IP(), port, portOk)
		}
		if ps := s.config.PeerStore; ps != nil {
			go ps.AddPeer(
				peer_store.InfoHash(args.InfoHash),
				krpc.NodeAddr{IP: source.IP(), Port: port},
			)
		}

		s.reply(source, m.T, krpc.Return{})
	case "put":
		if !s.validToken(args.Token, source) {
			expvars.Add("received put with invalid token", 1)
			return
		}
		expvars.Add("received put with valid token", 1)

		i := &bep44.Item{
			V:    args.V,
			K:    args.K,
			Salt: args.Salt,
			Sig:  args.Sig,
			Cas:  args.Cas,
			Seq:  *args.Seq,
		}

		if err := s.store.Put(i); err != nil {
			kerr, ok := err.(krpc.Error)
			if !ok {
				s.sendError(source, m.T, krpc.ErrorMethodUnknown)
			}

			s.sendError(source, m.T, kerr)
			break
		}

		s.reply(source, m.T, krpc.Return{
			ID: s.ID(),
		})
	case "get":
		var r krpc.Return
		if err := s.setReturnNodes(&r, m, source); err != nil {
			s.sendError(source, m.T, *err)
			break
		}

		t := s.createToken(source)
		r.Token = &t

		item, err := s.store.Get(bep44.Target(args.Target))
		if err == bep44.ErrItemNotFound {
			s.reply(source, m.T, r)
			break
		}

		if kerr, ok := err.(krpc.Error); ok {
			s.sendError(source, m.T, kerr)
			break
		}

		if err != nil {
			s.sendError(source, m.T, krpc.Error{
				Code: krpc.ErrorCodeGenericError,
				Msg:  err.Error(),
			})
			break
		}

		r.Seq = &item.Seq

		if args.Seq != nil && item.Seq <= *args.Seq {
			s.reply(source, m.T, r)
			break
		}

		r.V = item.V
		r.K = item.K
		r.Sig = item.Sig

		s.reply(source, m.T, r)
	// case "sample_infohashes":
	// // Nodes supporting this extension should always include the samples field in the response,
	// // even when it is zero-length. This lets indexing nodes to distinguish nodes supporting this
	// // extension from those that respond to unknown query types which contain a target field [2].
	default:
		// TODO: http://libtorrent.org/dht_extensions.html#forward-compatibility
		s.sendError(source, m.T, krpc.ErrorMethodUnknown)
	}
}

func (s *Server) sendError(addr Addr, t string, e krpc.Error) {
	go func() {
		m := krpc.Msg{
			T: t,
			Y: "e",
			E: &e,
		}
		b, err := bencode.Marshal(m)
		if err != nil {
			panic(err)
		}
		s.logger().Printf("sending error to %q: %v", addr, e)
		_, err = s.writeToNode(context.Background(), b, addr, false, true)
		if err != nil {
			s.logger().Printf("error replying to %q: %v", addr, err)
		}
	}()
}

func (s *Server) reply(addr Addr, t string, r krpc.Return) {
	go func() {
		r.ID = s.id.AsByteArray()
		m := krpc.Msg{
			T:  t,
			Y:  "r",
			R:  &r,
			IP: addr.KRPC(),
		}
		b := bencode.MustMarshal(m)
		log.Fmsg("replying to %q", addr).Log(s.logger())
		wrote, err := s.writeToNode(context.Background(), b, addr, s.config.WaitToReply, true)
		if err != nil {
			s.config.Logger.Printf("error replying to %s: %s", addr, err)
		}
		if wrote {
			expvars.Add("replied to peer", 1)
		}
	}()
}

// Adds a node if appropriate.
func (s *Server) addNode(n *node) error {
	if s.nodeIsBad(n) {
		return errors.New("node is bad")
	}
	b := s.table.bucketForID(n.Id)
	if b.Len() >= s.table.k {
		if b.EachNode(func(bn *node) bool {
			// Replace bad and untested nodes with a good one.
			if s.nodeIsBad(bn) || (s.IsGood(n) && bn.lastGotResponse.IsZero()) {
				s.table.dropNode(bn)
			}
			return b.Len() >= s.table.k
		}) {
			return errors.New("no room in bucket")
		}
	}
	if err := s.table.addNode(n); err != nil {
		panic(fmt.Sprintf("expected to add node: %s", err))
	}
	return nil
}

func (s *Server) NodeRespondedToPing(addr Addr, id int160.T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id == s.id {
		return
	}
	b := s.table.bucketForID(id)
	if b.GetNode(addr, id) == nil {
		return
	}
	b.lastChanged = time.Now()
}

// Updates the node, adding it if appropriate.
func (s *Server) updateNode(addr Addr, id *krpc.ID, tryAdd bool, update func(*node)) error {
	if id == nil {
		return errors.New("id is nil")
	}
	int160Id := int160.FromByteArray(*id)
	n := s.table.getNode(addr, int160Id)
	missing := n == nil
	if missing {
		if !tryAdd {
			return errors.New("node not present and add flag false")
		}
		if int160Id == s.id {
			return errors.New("can't store own id in routing table")
		}
		n = &node{nodeKey: nodeKey{
			Id:   int160Id,
			Addr: addr,
		}}
	}
	update(n)
	if !missing {
		return nil
	}
	return s.addNode(n)
}

func (s *Server) nodeIsBad(n *node) bool {
	return s.nodeErr(n) != nil
}

func (s *Server) nodeErr(n *node) error {
	if n.Id == s.id {
		return errors.New("is self")
	}
	if n.Id.IsZero() {
		return errors.New("has zero id")
	}
	if !(s.config.NoSecurity || n.IsSecure()) {
		return errors.New("not secure")
	}
	if n.failedLastQuestionablePing {
		return errors.New("didn't respond to last questionable node ping")
	}
	return nil
}

func (s *Server) writeToNode(ctx context.Context, b []byte, node Addr, wait, rate bool) (wrote bool, err error) {
	func() {
		// This is a pain. It would be better if the blocklist returned an error if it was closed
		// instead.
		s.mu.RLock()
		defer s.mu.RUnlock()
		if s.closed.IsSet() {
			err = errors.New("server is closed")
			return
		}
		if list := s.ipBlockList; list != nil {
			if r, ok := list.Lookup(node.IP()); ok {
				err = fmt.Errorf("write to %v blocked by %v", node, r)
				return
			}
		}
	}()
	if err != nil {
		return
	}
	// s.config.Logger.WithValues(log.Debug).Printf("writing to %s: %q", node.String(), b)
	if rate {
		if wait {
			err = s.config.SendLimiter.Wait(ctx)
			if err != nil {
				err = fmt.Errorf("waiting for rate-limit token: %w", err)
				return false, err
			}
		} else {
			if !s.config.SendLimiter.Allow() {
				return false, errors.New("rate limit exceeded")
			}
		}
	}
	n, err := s.socket.WriteTo(b, node.Raw())
	writes.Add(1)
	if rate {
		expvars.Add("rated writes", 1)
	} else {
		expvars.Add("unrated writes", 1)
	}
	if err != nil {
		writeErrors.Add(1)
		if rate {
			// Give the token back. nfi if this will actually work.
			s.config.SendLimiter.AllowN(time.Now(), -1)
		}
		err = fmt.Errorf("error writing %d bytes to %s: %s", len(b), node, err)
		return
	}
	wrote = true
	if n != len(b) {
		err = io.ErrShortWrite
		return
	}
	return
}

func (s *Server) nextTransactionID() string {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(b[:], s.nextT)
	s.nextT++
	return string(b[:n])
}

func (s *Server) deleteTransaction(k transactionKey) {
	delete(s.transactions, k)
}

func (s *Server) addTransaction(k transactionKey, t *Transaction) {
	if _, ok := s.transactions[k]; ok {
		panic("transaction not unique")
	}
	s.transactions[k] = t
}

// ID returns the 20-byte server ID. This is the ID used to communicate with the
// DHT network.
func (s *Server) ID() [20]byte {
	return s.id.AsByteArray()
}

func (s *Server) createToken(addr Addr) string {
	return s.tokenServer.CreateToken(addr)
}

func (s *Server) validToken(token string, addr Addr) bool {
	return s.tokenServer.ValidToken(token, addr)
}

type numWrites int

func (s *Server) makeQueryBytes(q string, a krpc.MsgArgs, t string) []byte {
	a.ID = s.ID()
	m := krpc.Msg{
		T: t,
		Y: "q",
		Q: q,
		A: &a,
	}
	// BEP 43. Outgoing queries from passive nodes should contain "ro":1 in the top level
	// dictionary.
	if s.config.Passive {
		m.ReadOnly = true
	}
	b, err := bencode.Marshal(m)
	if err != nil {
		panic(err)
	}
	return b
}

type QueryResult struct {
	Reply  krpc.Msg
	Writes numWrites
	Err    error
}

func (qr QueryResult) ToError() error {
	if qr.Err != nil {
		return qr.Err
	}
	e := qr.Reply.Error()
	if e != nil {
		return e
	}
	return nil
}

// Converts a Server QueryResult to a traversal.QueryResult.
func (me QueryResult) TraversalQueryResult(addr krpc.NodeAddr) (ret traversal.QueryResult) {
	r := me.Reply.R
	if r == nil {
		return
	}
	ret.ResponseFrom = &krpc.NodeInfo{
		Addr: addr,
		ID:   r.ID,
	}
	ret.Nodes = r.Nodes
	ret.Nodes6 = r.Nodes6
	if r.Token != nil {
		ret.ClosestData = *r.Token
	}
	return
}

// Rate-limiting to be applied to writes for a given query. Queries occur inside transactions that
// will attempt to send several times. If the STM rate-limiting helpers are used, the first send is
// often already accounted for in the rate-limiting machinery before the query method that does the
// IO is invoked.
type QueryRateLimiting struct {
	// Don't rate-limit the first send for a query.
	NotFirst bool
	// Don't rate-limit any sends for a query. Note that there's still built-in waits before retries.
	NotAny        bool
	WaitOnRetries bool
	NoWaitFirst   bool
}

// The zero value for this uses reasonable/traditional defaults on Server methods.
type QueryInput struct {
	MsgArgs      krpc.MsgArgs
	RateLimiting QueryRateLimiting
	NumTries     int
}

// Performs an arbitrary query. `q` is the query value, defined by the DHT BEP. `a` should contain
// the appropriate argument values, if any. `a.ID` is clobbered by the Server. Responses to queries
// made this way are not interpreted by the Server. More specific methods like FindNode and GetPeers
// may make use of the response internally before passing it back to the caller.
func (s *Server) Query(ctx context.Context, addr Addr, q string, input QueryInput) (ret QueryResult) {
	if input.NumTries == 0 {
		input.NumTries = defaultMaxQuerySends
	}
	defer func(started time.Time) {
		s.logger().WithDefaultLevel(log.Debug).WithValues(q).Printf(
			"Query(%v) returned after %v (err=%v, reply.Y=%v, reply.E=%v, writes=%v)",
			q, time.Since(started), ret.Err, ret.Reply.Y, ret.Reply.E, ret.Writes)
	}(time.Now())
	replyChan := make(chan krpc.Msg, 1)
	t := &Transaction{
		onResponse: func(m krpc.Msg) {
			replyChan <- m
		},
	}
	tk := transactionKey{
		RemoteAddr: addr.String(),
	}
	s.mu.Lock()
	tid := s.nextTransactionID()
	s.stats.OutboundQueriesAttempted++
	tk.T = tid
	s.addTransaction(tk, t)
	s.mu.Unlock()
	// Receives a non-nil error from the sender, and closes when the sender completes.
	sendErr := make(chan error, 1)
	sendCtx, cancelSend := context.WithCancel(pprof.WithLabels(ctx, pprof.Labels("q", q)))
	go func() {
		err := s.transactionQuerySender(
			sendCtx,
			s.makeQueryBytes(q, input.MsgArgs, tid),
			&ret.Writes,
			addr,
			input.RateLimiting,
			input.NumTries)
		if err != nil {
			sendErr <- err
		}
		close(sendErr)
	}()
	expvars.Add(fmt.Sprintf("outbound %s queries", q), 1)
	select {
	case ret.Reply = <-replyChan:
	case <-ctx.Done():
		ret.Err = ctx.Err()
	case ret.Err = <-sendErr:
	}
	// Make sure the query sender stops.
	cancelSend()
	// Make sure the query sender has returned, it will either send an error that we didn't catch
	// above, or the channel will be closed by the sender completing.
	<-sendErr
	s.mu.Lock()
	s.deleteTransaction(tk)
	s.mu.Unlock()
	return
}

func (s *Server) transactionQuerySender(
	sendCtx context.Context,
	b []byte,
	writes *numWrites,
	addr Addr,
	rateLimiting QueryRateLimiting,
	numTries int,
) error {
	// log.Printf("sending %q", b)
	err := transactionSender(
		sendCtx,
		func() error {
			wrote, err := s.writeToNode(sendCtx, b, addr,
				// We only wait for the first write by default if rate-limiting is enabled for this
				// query.
				func() bool {
					if *writes == 0 {
						return !rateLimiting.NoWaitFirst
					} else {
						return rateLimiting.WaitOnRetries
					}
				}(),
				func() bool {
					if rateLimiting.NotAny {
						return false
					}
					if *writes == 0 {
						return !rateLimiting.NotFirst
					}
					return true
				}(),
			)
			if wrote {
				*writes++
			}
			return err
		},
		s.resendDelay,
		numTries,
	)
	if err != nil {
		return err
	}
	select {
	case <-sendCtx.Done():
		err = sendCtx.Err()
	case <-time.After(s.resendDelay()):
		err = TransactionTimeout
	}
	return fmt.Errorf("after %v tries: %w", numTries, err)
}

// Sends a ping query to the address given.
func (s *Server) PingQueryInput(node *net.UDPAddr, qi QueryInput) QueryResult {
	addr := NewAddr(node)
	res := s.Query(context.TODO(), addr, "ping", qi)
	if res.Err == nil {
		id := res.Reply.SenderID()
		if id != nil {
			s.NodeRespondedToPing(addr, id.Int160())
		}
	}
	return res
}

// Sends a ping query to the address given.
func (s *Server) Ping(node *net.UDPAddr) QueryResult {
	return s.PingQueryInput(node, QueryInput{})
}

// Put adds a new item to node. You need to call Get first for a write token.
func (s *Server) Put(ctx context.Context, node Addr, i bep44.Put, token string, rl QueryRateLimiting) QueryResult {
	if err := s.store.Put(i.ToItem()); err != nil {
		return QueryResult{
			Err: err,
		}
	}
	qi := QueryInput{
		MsgArgs: krpc.MsgArgs{
			Cas:   i.Cas,
			ID:    s.ID(),
			Salt:  i.Salt,
			Seq:   &i.Seq,
			Sig:   i.Sig,
			Token: token,
			V:     i.V,
		},
	}
	if i.K != nil {
		qi.MsgArgs.K = *i.K
	}
	return s.Query(ctx, node, "put", qi)
}

func (s *Server) announcePeer(
	ctx context.Context,
	node Addr, infoHash int160.T, port int, token string, impliedPort bool, rl QueryRateLimiting,
) (
	ret QueryResult,
) {
	if port == 0 && !impliedPort {
		ret.Err = errors.New("no port specified")
		return
	}
	ret = s.Query(
		ctx, node, "announce_peer",
		QueryInput{
			MsgArgs: krpc.MsgArgs{
				ImpliedPort: impliedPort,
				InfoHash:    infoHash.AsByteArray(),
				Port:        &port,
				Token:       token,
			},
			RateLimiting: rl,
		})
	if ret.Err != nil {
		return
	}
	if krpcError := ret.Reply.Error(); krpcError != nil {
		announceErrors.Add(1)
		ret.Err = krpcError
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.SuccessfulOutboundAnnouncePeerQueries++
	return
}

// Sends a find_node query to addr. targetID is the node we're looking for. The Server makes use of
// some of the response fields.
func (s *Server) FindNode(addr Addr, targetID int160.T, rl QueryRateLimiting) (ret QueryResult) {
	ret = s.Query(context.TODO(), addr, "find_node", QueryInput{
		MsgArgs: krpc.MsgArgs{
			Target: targetID.AsByteArray(),
			Want:   s.config.DefaultWant,
		},
		RateLimiting: rl,
	})
	return
}

// Returns how many nodes are in the node table.
func (s *Server) NumNodes() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.numNodes()
}

// Returns non-bad nodes from the routing table.
func (s *Server) Nodes() (nis []krpc.NodeInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.notBadNodes()
}

// Returns non-bad nodes from the routing table.
func (s *Server) notBadNodes() (nis []krpc.NodeInfo) {
	s.table.forNodes(func(n *node) bool {
		if s.nodeIsBad(n) {
			return true
		}
		nis = append(nis, krpc.NodeInfo{
			Addr: n.Addr.KRPC(),
			ID:   n.Id.AsByteArray(),
		})
		return true
	})
	return
}

// Stops the server network activity. This is all that's required to clean-up a Server.
func (s *Server) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed.Set()
	go s.socket.Close()
}

func (s *Server) GetPeers(ctx context.Context, addr Addr, infoHash int160.T, scrape bool, rl QueryRateLimiting) (ret QueryResult) {
	args := krpc.MsgArgs{
		InfoHash: infoHash.AsByteArray(),
		// TODO: Maybe IPv4-only Servers won't want IPv6 nodes?
		Want: s.config.DefaultWant,
	}
	if scrape {
		args.Scrape = 1
	}
	ret = s.Query(ctx, addr, "get_peers", QueryInput{
		MsgArgs:      args,
		RateLimiting: rl,
	})
	m := ret.Reply
	if m.R != nil {
		if m.R.Token == nil {
			expvars.Add("get_peers responses with no token", 1)
		} else if len(*m.R.Token) == 0 {
			expvars.Add("get_peers responses with empty token", 1)
		} else {
			expvars.Add("get_peers responses with token", 1)
		}
	}
	return
}

// Get gets item information from a specific target ID. If seq is set to a specific value,
// only items with seq bigger than the one provided will return a V, K and Sig, if any.
// Get must be used to get a Put write token, when you want to write an item instead of read it.
func (s *Server) Get(ctx context.Context, addr Addr, target bep44.Target, seq *int64, rl QueryRateLimiting) QueryResult {
	return s.Query(ctx, addr, "get", QueryInput{
		MsgArgs: krpc.MsgArgs{
			ID:     s.ID(),
			Target: target,
			Seq:    seq,
			Want:   []krpc.Want{krpc.WantNodes, krpc.WantNodes6},
		},
		RateLimiting: rl,
	})
}

func (s *Server) closestGoodNodeInfos(
	k int,
	targetID int160.T,
	filter func(krpc.NodeAddr) bool,
) (
	ret []krpc.NodeInfo,
) {
	for _, n := range s.closestNodes(k, targetID, func(n *node) bool {
		return s.IsGood(n) && filter(n.NodeInfo().Addr)
	}) {
		ret = append(ret, n.NodeInfo())
	}
	return
}

func (s *Server) closestNodes(k int, target int160.T, filter func(*node) bool) []*node {
	return s.table.closestNodes(k, target, filter)
}

func (s *Server) TraversalStartingNodes() (nodes []addrMaybeId, err error) {
	s.mu.RLock()
	s.table.forNodes(func(n *node) bool {
		nodes = append(nodes, addrMaybeId{Addr: n.Addr.KRPC(), Id: &n.Id})
		return true
	})
	s.mu.RUnlock()
	if len(nodes) > 0 {
		return
	}
	if s.config.StartingNodes != nil {
		// There seems to be floods on this call on occasion, which may cause a barrage of DNS
		// resolution attempts. This would require that we're unable to get replies because we can't
		// resolve, transmit or receive on the network. Nodes currently don't get expired from the
		// table, so once we have some entries, we should never have to fallback.
		s.logger().Levelf(log.Debug, "falling back on starting nodes")
		addrs, err := s.config.StartingNodes()
		if err != nil {
			return nil, errors.Wrap(err, "getting starting nodes")
		} else {
			// log.Printf("resolved %v addresses", len(addrs))
		}
		for _, a := range addrs {
			nodes = append(nodes, addrMaybeId{Addr: a.KRPC(), Id: nil})
		}
	}
	if len(nodes) == 0 {
		err = errors.New("no initial nodes")
	}
	return
}

func (s *Server) AddNodesFromFile(fileName string) (added int, err error) {
	ns, err := ReadNodesFromFile(fileName)
	if err != nil {
		return
	}
	for _, n := range ns {
		if s.AddNode(n) == nil {
			added++
		}
	}
	return
}

func (s *Server) logger() log.Logger {
	return s.config.Logger
}

func (s *Server) PeerStore() peer_store.Interface {
	return s.config.PeerStore
}

func (s *Server) getQuestionableNode() (ret *node) {
	s.table.forNodes(func(n *node) bool {
		if s.IsQuestionable(n) {
			ret = n
			return false
		}
		return true
	})
	return
}

func (s *Server) shouldStopRefreshingBucket(bucketIndex int) bool {
	b := &s.table.buckets[bucketIndex]
	// Stop if the bucket is full, and none of the nodes are bad.
	return b.Len() == s.table.K() && b.EachNode(func(n *node) bool {
		return !s.nodeIsBad(n)
	})
}

func (s *Server) refreshBucket(bucketIndex int) *traversal.Stats {
	s.mu.RLock()
	id := s.table.randomIdForBucket(bucketIndex)
	op := traversal.Start(traversal.OperationInput{
		Target: id.AsByteArray(),
		Alpha:  3,
		// Running this to completion with K matching the full-bucket size should result in a good,
		// full bucket, since the Server will add nodes that respond to its table to replace the bad
		// ones we're presumably refreshing. It might be possible to terminate the traversal early
		// as soon as the bucket is good.
		K: s.table.K(),
		DoQuery: func(ctx context.Context, addr krpc.NodeAddr) traversal.QueryResult {
			res := s.FindNode(NewAddr(addr.UDP()), id, QueryRateLimiting{})
			return res.TraversalQueryResult(addr)
		},
		NodeFilter: s.TraversalNodeFilter,
	})
	defer func() {
		s.mu.RUnlock()
		op.Stop()
		<-op.Stopped()
	}()
	b := &s.table.buckets[bucketIndex]
wait:
	for {
		if s.shouldStopRefreshingBucket(bucketIndex) {
			break wait
		}
		op.AddNodes(types.AddrMaybeIdSliceFromNodeInfoSlice(s.notBadNodes()))
		bucketChanged := b.changed.Signaled()
		s.mu.RUnlock()
		select {
		case <-op.Stalled():
			s.mu.RLock()
			break wait
		case <-bucketChanged:
		}
		s.mu.RLock()
	}
	return op.Stats()
}

func (s *Server) shouldBootstrap() bool {
	return s.lastBootstrap.IsZero() || time.Since(s.lastBootstrap) > 30*time.Minute
}

func (s *Server) shouldBootstrapUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shouldBootstrap()
}

func (s *Server) pingQuestionableNodesInBucket(bucketIndex int) {
	b := &s.table.buckets[bucketIndex]
	var wg sync.WaitGroup
	b.EachNode(func(n *node) bool {
		if s.IsQuestionable(n) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := s.questionableNodePing(context.TODO(), n.Addr, n.Id.AsByteArray()).Err
				if err != nil {
					log.Printf("error pinging questionable node in bucket %v: %v", bucketIndex, err)
				}
			}()
		}
		return true
	})
	s.mu.RUnlock()
	wg.Wait()
	s.mu.RLock()
}

// A routine that maintains the Server's routing table, by pinging questionable nodes, and
// refreshing buckets. This should be invoked on a running Server when the caller is satisfied with
// having set it up. It is not necessary to explicitly Bootstrap the Server once this routine has
// started.
func (s *Server) TableMaintainer() {
	for {
		if s.shouldBootstrapUnlocked() {
			stats, err := s.Bootstrap()
			if err != nil {
				log.Printf("error bootstrapping during bucket refresh: %v", err)
			}
			log.Printf("bucket refresh bootstrap stats: %v", stats)
		}
		s.mu.RLock()
		for i := range s.table.buckets {
			s.pingQuestionableNodesInBucket(i)
			// if time.Since(b.lastChanged) < 15*time.Minute {
			//	continue
			// }
			if s.shouldStopRefreshingBucket(i) {
				continue
			}
			s.logger().Levelf(log.Info, "refreshing bucket %v", i)
			s.mu.RUnlock()
			stats := s.refreshBucket(i)
			s.logger().Levelf(log.Info, "finished refreshing bucket %v: %v", i, stats)
			s.mu.RLock()
			if !s.shouldStopRefreshingBucket(i) {
				// Presumably we couldn't fill the bucket anymore, so assume we're as deep in the
				// available node space as we can go.
				break
			}
		}
		s.mu.RUnlock()
		select {
		case <-s.closed.LockedChan(&s.mu):
			return
		case <-time.After(time.Minute):
		}
	}
}

func (s *Server) questionableNodePing(ctx context.Context, addr Addr, id krpc.ID) QueryResult {
	// A ping query that will be certain to try at least 3 times.
	res := s.Query(ctx, addr, "ping", QueryInput{
		RateLimiting: QueryRateLimiting{
			WaitOnRetries: true,
		},
		NumTries: 3,
	})
	if res.Err == nil && res.Reply.R != nil {
		s.NodeRespondedToPing(addr, res.Reply.R.ID.Int160())
	} else {
		s.mu.Lock()
		s.updateNode(addr, &id, false, func(n *node) {
			n.failedLastQuestionablePing = true
		})
		s.mu.Unlock()
	}
	return res
}

// Whether we should consider a node for contact based on its address and possible ID.
func (s *Server) TraversalNodeFilter(node addrMaybeId) bool {
	if !validNodeAddr(node.Addr.UDP()) {
		return false
	}
	if s.ipBlocked(node.Addr.IP) {
		return false
	}
	if node.Id == nil {
		return true
	}
	return s.config.NoSecurity || NodeIdSecure(node.Id.AsByteArray(), node.Addr.IP)
}

func validNodeAddr(addr net.Addr) bool {
	// At least for UDP addresses, we know what doesn't work.
	ua := addr.(*net.UDPAddr)
	if ua.Port == 0 {
		return false
	}
	if ip4 := ua.IP.To4(); ip4 != nil && ip4[0] == 0 {
		// Why?
		return false
	}
	return true
}

// func (s *Server) refreshBucket(bucketIndex int) {
//	targetId := s.table.randomIdForBucket(bucketIndex)
// }
