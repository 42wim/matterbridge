package utp

/*
#include "utp.h"
#include <stdbool.h>

struct utp_process_udp_args {
	const byte *buf;
	size_t len;
	const struct sockaddr *sa;
	socklen_t sal;
};

void process_received_messages(utp_context *ctx, struct utp_process_udp_args *args, size_t argslen)
{
	bool gotUtp = false;
	size_t i;
	for (i = 0; i < argslen; i++) {
		struct utp_process_udp_args *a = &args[i];
		//if (!a->len) continue;
		if (utp_process_udp(ctx, a->buf, a->len, a->sa, a->sal)) {
			gotUtp = true;
		}
	}
	if (gotUtp) {
		utp_issue_deferred_acks(ctx);
		utp_check_timeouts(ctx);
	}
}
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"syscall"
	"time"
	"unsafe"

	"github.com/anacrolix/log"
	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/missinggo/inproc"
	"github.com/anacrolix/mmsg"
)

const (
	utpCheckTimeoutInterval   = 500 * time.Millisecond
	issueDeferredUtpAcksDelay = 1000 * time.Microsecond
)

type Socket struct {
	pc            net.PacketConn
	ctx           *C.utp_context
	backlog       chan *Conn
	closed        bool
	conns         map[*C.utp_socket]*Conn
	nonUtpReads   chan packet
	writeDeadline time.Time
	readDeadline  time.Time

	// This is called without the package mutex, without knowing if the result will be needed.
	asyncFirewallCallback FirewallCallback
	// Whether the next accept is to be blocked.
	asyncBlock bool

	// This is called with the package mutex, and preferred.
	syncFirewallCallback FirewallCallback

	acksScheduled bool
	ackTimer      *time.Timer

	utpTimeoutChecker *time.Timer

	logger log.Logger
}

// A firewall callback returns true if an incoming connection request should be ignored. This is
// better than just accepting and closing, as it means no acknowledgement packet is sent.
type FirewallCallback func(net.Addr) bool

var (
	_               net.PacketConn = (*Socket)(nil)
	_               net.Listener   = (*Socket)(nil)
	errSocketClosed                = errors.New("Socket closed")
)

type packet struct {
	b    []byte
	from net.Addr
}

func listenPacket(network, addr string) (pc net.PacketConn, err error) {
	if network == "inproc" {
		return inproc.ListenPacket(network, addr)
	}
	return net.ListenPacket(network, addr)
}

type NewSocketOpt func(s *Socket)

func WithLogger(l log.Logger) NewSocketOpt {
	return func(s *Socket) {
		s.logger = l
	}
}

func NewSocket(network, addr string, opts ...NewSocketOpt) (*Socket, error) {
	pc, err := listenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	s := &Socket{
		pc:          pc,
		backlog:     make(chan *Conn, 5),
		conns:       make(map[*C.utp_socket]*Conn),
		nonUtpReads: make(chan packet, 100),
		logger:      Logger,
	}
	s.ackTimer = time.AfterFunc(math.MaxInt64, s.ackTimerFunc)
	s.ackTimer.Stop()

	for _, opt := range opts {
		opt(s)
	}

	func() {
		mu.Lock()
		defer mu.Unlock()
		ctx := C.utp_init(2)
		if ctx == nil {
			panic(ctx)
		}
		s.ctx = ctx
		ctx.setCallbacks()
		if utpLogging {
			ctx.setOption(C.UTP_LOG_NORMAL, 1)
			ctx.setOption(C.UTP_LOG_MTU, 1)
			ctx.setOption(C.UTP_LOG_DEBUG, 1)
		}
		libContextToSocket[ctx] = s
		s.utpTimeoutChecker = time.AfterFunc(0, s.timeoutCheckerTimerFunc)
	}()
	go s.packetReader()
	return s, nil
}

func (s *Socket) onLibSocketDestroyed(ls *C.utp_socket) {
	delete(s.conns, ls)
}

func (s *Socket) newConn(us *C.utp_socket) *Conn {
	c := &Conn{
		s:         s,
		us:        us,
		localAddr: s.pc.LocalAddr(),
	}
	c.cond.L = &mu
	s.conns[us] = c
	c.writeDeadlineTimer = time.AfterFunc(-1, c.cond.Broadcast)
	c.readDeadlineTimer = time.AfterFunc(-1, c.cond.Broadcast)
	return c
}

const maxNumBuffers = 16

func (s *Socket) packetReader() {
	mc := mmsg.NewConn(s.pc)
	// Increasing the messages increases the memory use, but also means we can
	// reduces utp_issue_deferred_acks and syscalls which should improve
	// efficiency. On the flip side, not all OSs implement batched reads.
	ms := make([]mmsg.Message, func() int {
		if mc.Err() == nil {
			return maxNumBuffers
		} else {
			return 1
		}
	}())
	for i := range ms {
		// The IPv4 UDP limit is allegedly about 64 KiB, and this message has
		// been seen on receiving on Windows with just 0x1000: wsarecvfrom: A
		// message sent on a datagram socket was larger than the internal
		// message buffer or some other network limit, or the buffer used to
		// receive a datagram into was smaller than the datagram itself.
		ms[i].Buffers = [][]byte{make([]byte, 0x10000)}
	}
	// Some crap OSs like Windoze will raise errors in Reads that don't
	// actually mean we should stop.
	consecutiveErrors := 0
	for {
		// In C, all the reads are processed and when it threatens to block,
		// we're supposed to call utp_issue_deferred_acks.
		n, err := mc.RecvMsgs(ms)
		if n == 1 {
			singleMsgRecvs.Add(1)
		}
		if n > 1 {
			multiMsgRecvs.Add(1)
		}
		if err != nil {
			mu.Lock()
			closed := s.closed
			mu.Unlock()
			if closed {
				// We don't care.
				return
			}
			// See https://github.com/anacrolix/torrent/issues/83. If we get
			// an endless stream of errors (such as the PacketConn being
			// Closed outside of our control, this work around may need to be
			// reconsidered.
			s.logger.Printf("ignoring socket read error: %s", err)
			consecutiveErrors++
			if consecutiveErrors >= 100 {
				s.logger.Print("too many consecutive errors, closing socket")
				s.Close()
				return
			}
			continue
		}
		consecutiveErrors = 0
		expMap.Add("successful mmsg receive calls", 1)
		expMap.Add("received messages", int64(n))
		s.processReceivedMessages(ms[:n])
	}
}

func (s *Socket) processReceivedMessages(ms []mmsg.Message) {
	mu.Lock()
	defer mu.Unlock()
	if s.closed {
		return
	}
	if processPacketsInC {
		var args [maxNumBuffers]C.struct_utp_process_udp_args
		for i, m := range ms {
			a := &args[i]
			a.buf = (*C.byte)(&m.Buffers[0][0])
			a.len = C.size_t(m.N)
			var rsa syscall.RawSockaddrAny
			rsa, a.sal = netAddrToLibSockaddr(m.Addr)
			a.sa = (*C.struct_sockaddr)(unsafe.Pointer(&rsa))
		}
		C.process_received_messages(s.ctx, &args[0], C.size_t(len(ms)))
	} else {
		gotUtp := false
		for _, m := range ms {
			gotUtp = s.processReceivedMessage(m.Buffers[0][:m.N], m.Addr) || gotUtp
		}
		if gotUtp && !s.closed {
			s.afterReceivingUtpMessages()
		}
	}
}

func (s *Socket) afterReceivingUtpMessages() {
	if s.acksScheduled {
		return
	}
	s.ackTimer.Reset(issueDeferredUtpAcksDelay)
	s.acksScheduled = true
}

func (s *Socket) issueDeferredAcks() {
	expMap.Add("utp_issue_deferred_acks calls", 1)
	C.utp_issue_deferred_acks(s.ctx)
}

func (s *Socket) checkUtpTimeouts() {
	expMap.Add("utp_check_timeouts calls", 1)
	C.utp_check_timeouts(s.ctx)
}

func (s *Socket) ackTimerFunc() {
	mu.Lock()
	defer mu.Unlock()
	if !s.acksScheduled || s.ctx == nil {
		return
	}
	s.acksScheduled = false
	s.issueDeferredAcks()
}

func (s *Socket) processReceivedMessage(b []byte, addr net.Addr) (utp bool) {
	if s.utpProcessUdp(b, addr) {
		socketUtpPacketsReceived.Add(1)
		return true
	} else {
		s.onReadNonUtp(b, addr)
		return false
	}
}

// Process packet batches entirely from C, reducing CGO overhead. Currently
// requires GODEBUG=cgocheck=0.
const processPacketsInC = false

var staticRsa syscall.RawSockaddrAny

// Wraps libutp's utp_process_udp, returning relevant information.
func (s *Socket) utpProcessUdp(b []byte, addr net.Addr) (utp bool) {
	if len(b) == 0 {
		// The implementation of utp_process_udp rejects null buffers, and
		// anything smaller than the UTP header size. It's also prone to
		// assert on those, which we don't want to trigger.
		return false
	}
	if missinggo.AddrPort(addr) == 0 {
		return false
	}
	mu.Unlock()
	// TODO: If it's okay to call the firewall callback without the package lock, aren't we assuming
	// that the next UDP packet to be processed by libutp has to be the one we've just used the
	// callback for? Why can't we assign directly to Socket.asyncBlock?
	asyncBlock := func() bool {
		if s.asyncFirewallCallback == nil || s.syncFirewallCallback != nil {
			return false
		}
		return s.asyncFirewallCallback(addr)
	}()
	mu.Lock()
	s.asyncBlock = asyncBlock
	if s.closed {
		return false
	}
	var sal C.socklen_t
	staticRsa, sal = netAddrToLibSockaddr(addr)
	ret := C.utp_process_udp(s.ctx, (*C.byte)(&b[0]), C.size_t(len(b)), (*C.struct_sockaddr)(unsafe.Pointer(&staticRsa)), sal)
	switch ret {
	case 1:
		return true
	case 0:
		return false
	default:
		panic(ret)
	}
}

func (s *Socket) timeoutCheckerTimerFunc() {
	mu.Lock()
	ok := s.ctx != nil
	if ok {
		s.checkUtpTimeouts()
	}
	if ok {
		s.utpTimeoutChecker.Reset(utpCheckTimeoutInterval)
	}
	mu.Unlock()
}

func (s *Socket) Close() error {
	mu.Lock()
	defer mu.Unlock()
	return s.closeLocked()
}

func (s *Socket) closeLocked() error {
	if s.closed {
		return nil
	}
	// Calling this deletes the pointer. It must not be referred to after
	// this.
	C.utp_destroy(s.ctx)
	s.ctx = nil
	s.pc.Close()
	close(s.backlog)
	close(s.nonUtpReads)
	s.closed = true
	s.ackTimer.Stop()
	s.utpTimeoutChecker.Stop()
	s.acksScheduled = false
	return nil
}

func (s *Socket) Addr() net.Addr {
	return s.pc.LocalAddr()
}

func (s *Socket) LocalAddr() net.Addr {
	return s.pc.LocalAddr()
}

func (s *Socket) Accept() (net.Conn, error) {
	nc, ok := <-s.backlog
	if !ok {
		return nil, errors.New("closed")
	}
	return nc, nil
}

func (s *Socket) Dial(addr string) (net.Conn, error) {
	return s.DialTimeout(addr, 0)
}

func (s *Socket) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	ctx := context.Background()
	if timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	return s.DialContext(ctx, "", addr)
}

func (s *Socket) resolveAddr(network, addr string) (net.Addr, error) {
	if network == "" {
		network = s.Addr().Network()
	}
	return resolveAddr(network, addr)
}

func resolveAddr(network, addr string) (net.Addr, error) {
	switch network {
	case "inproc":
		return inproc.ResolveAddr(network, addr)
	default:
		return net.ResolveUDPAddr(network, addr)
	}
}

// Passing an empty network will use the network of the Socket's listener.
func (s *Socket) DialContext(ctx context.Context, network, addr string) (_ net.Conn, err error) {
	if network == "" {
		network = s.pc.LocalAddr().Network()
	}
	ua, err := resolveAddr(network, addr)
	if err != nil {
		return nil, fmt.Errorf("error resolving address: %v", err)
	}
	sa, sl := netAddrToLibSockaddr(ua)
	mu.Lock()
	defer mu.Unlock()
	if s.closed {
		return nil, errSocketClosed
	}
	utpSock := utpCreateSocketAndConnect(s.ctx, sa, sl)
	c := s.newConn(utpSock)
	c.setRemoteAddr()
	err = c.waitForConnect(ctx)
	if err != nil {
		c.close()
		return
	}
	return c, err
}

func (s *Socket) pushBacklog(c *Conn) {
	select {
	case s.backlog <- c:
	default:
		c.close()
	}
}

func (s *Socket) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	p, ok := <-s.nonUtpReads
	if !ok {
		err = errors.New("closed")
		return
	}
	n = copy(b, p.b)
	addr = p.from
	return
}

func (s *Socket) onReadNonUtp(b []byte, from net.Addr) {
	if s.closed {
		return
	}
	socketNonUtpPacketsReceived.Add(1)
	select {
	case s.nonUtpReads <- packet{append([]byte(nil), b...), from}:
	default:
		// log.Printf("dropped non utp packet: no room in buffer")
		nonUtpPacketsDropped.Add(1)
	}
}

func (s *Socket) SetReadDeadline(t time.Time) error {
	panic("not implemented")
}

func (s *Socket) SetWriteDeadline(t time.Time) error {
	panic("not implemented")
}

func (s *Socket) SetDeadline(t time.Time) error {
	panic("not implemented")
}

func (s *Socket) WriteTo(b []byte, addr net.Addr) (int, error) {
	return s.pc.WriteTo(b, addr)
}

func (s *Socket) ReadBufferLen() int {
	mu.Lock()
	defer mu.Unlock()
	return int(C.utp_context_get_option(s.ctx, C.UTP_RCVBUF))
}

func (s *Socket) WriteBufferLen() int {
	mu.Lock()
	defer mu.Unlock()
	return int(C.utp_context_get_option(s.ctx, C.UTP_SNDBUF))
}

func (s *Socket) SetWriteBufferLen(len int) {
	mu.Lock()
	defer mu.Unlock()
	i := C.utp_context_set_option(s.ctx, C.UTP_SNDBUF, C.int(len))
	if i != 0 {
		panic(i)
	}
}

func (s *Socket) SetOption(opt Option, val int) int {
	mu.Lock()
	defer mu.Unlock()
	return int(C.utp_context_set_option(s.ctx, opt, C.int(val)))
}

// The callback is used before each packet is processed by libutp without the this package's mutex
// being held. libutp may not actually need the result as the packet might not be a connection
// attempt. If the callback function is expensive, it may be worth setting a synchronous callback
// using SetSyncFirewallCallback.
func (s *Socket) SetFirewallCallback(f FirewallCallback) {
	mu.Lock()
	s.asyncFirewallCallback = f
	mu.Unlock()
}

// SetSyncFirewallCallback sets a synchronous firewall callback. It's only called as needed by
// libutp. It is called with the package-wide mutex held. Any locks acquired by the callback should
// not also be held by code that might use this package.
func (s *Socket) SetSyncFirewallCallback(f FirewallCallback) {
	mu.Lock()
	s.syncFirewallCallback = f
	mu.Unlock()
}
