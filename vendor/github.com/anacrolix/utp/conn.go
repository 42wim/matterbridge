package utp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/anacrolix/missinggo"
)

// Conn is a uTP stream and implements net.Conn. It owned by a Socket, which
// handles dispatching packets to and from Conns.
type Conn struct {
	recv_id, send_id uint16
	seq_nr, ack_nr   uint16
	lastAck          uint16
	lastTimeDiff     uint32
	peerWndSize      uint32
	cur_window       uint32
	connKey          connKey

	// Data waiting to be Read.
	readBuf         []byte
	readBufNotEmpty missinggo.Event

	socket           *Socket
	remoteSocketAddr net.Addr
	// The uTP timestamp.
	startTimestamp uint32
	// When the conn was allocated.
	created time.Time

	synAcked  bool // Syn is acked by the acceptor. Initiator also tracks it.
	gotFin    missinggo.Event
	wroteFin  missinggo.Event
	err       error
	closed    missinggo.Event
	destroyed missinggo.Event
	canWrite  missinggo.Event

	unackedSends []*send
	// Inbound payloads, the first is ack_nr+1.
	inbound    []recv
	inboundWnd int
	connDeadlines
	latencies []time.Duration

	// We need to send state packet.
	pendingSendState              bool
	sendPendingSendSendStateTimer *time.Timer
	// Send state is being delayed until sendStateTimer fires, which may have
	// been set at the beginning of a batch of received packets.
	sendPendingSendStateTimerActive bool

	// This timer fires when no packet has been received for a period.
	packetReadTimeoutTimer *time.Timer
}

var (
	_ net.Conn = &Conn{}
)

func (c *Conn) age() time.Duration {
	return time.Since(c.created)
}

func (c *Conn) timestamp() uint32 {
	return nowTimestamp() - c.startTimestamp
}

func (c *Conn) sendPendingSendStateTimerCallback() {
	mu.Lock()
	defer mu.Unlock()
	c.sendPendingSendStateTimerActive = false
	c.sendPendingSendSendStateTimer.Stop()
	c.sendPendingState()
}

// Send a state packet, if one is needed.
func (c *Conn) sendPendingState() {
	if c.destroyed.IsSet() {
		c.sendReset()
	} else {
		c.sendState()
	}
}

// So far as the spec makes clear, this is how many more, as-yet-unacked bytes
// we can fit into our receive buffers.
func (c *Conn) wndSize() uint32 {
	if len(c.readBuf)+c.inboundWnd > readBufferLen {
		return 0
	}
	return uint32(readBufferLen - len(c.readBuf) - c.inboundWnd)
}

func (c *Conn) makePacket(_type st, connID, seqNr uint16, payload []byte) (p []byte) {
	var selAck selectiveAckBitmask
	for i := 1; i < len(c.inbound); i++ {
		if c.inbound[i].seen {
			selAck.SetBit(i - 1)
		}
	}
	h := header{
		Type:          _type,
		Version:       1,
		ConnID:        connID,
		SeqNr:         seqNr,
		AckNr:         c.ack_nr,
		WndSize:       c.wndSize(),
		Timestamp:     c.timestamp(),
		TimestampDiff: c.lastTimeDiff,
	}
	if len(selAck.Bytes) != 0 {
		// The spec requires the number of bytes for a selective ACK to be at
		// least 4, and a multiple of 4.
		if len(selAck.Bytes)%4 != 0 {
			panic(len(selAck.Bytes))
		}
		h.Extensions = append(h.Extensions, extensionField{
			Type:  extensionTypeSelectiveAck,
			Bytes: selAck.Bytes,
		})
	}
	p = sendBufferPool.Get().([]byte)[:0:minMTU]
	n := h.Marshal(p)
	p = p[:n]
	// Extension headers are currently fixed in size.
	if n > maxHeaderSize {
		panic("header has unexpected size")
	}
	p = append(p, payload...)
	return
}

// Send the given payload with an up to date header.
func (c *Conn) send(_type st, connID uint16, payload []byte, seqNr uint16) (err error) {
	p := c.makePacket(_type, connID, seqNr, payload)
	n1, err := c.socket.writeTo(p, c.remoteSocketAddr)
	sendBufferPool.Put(p[:0:minMTU])
	if err != nil {
		return
	}
	if n1 != len(p) {
		panic(n1)
	}
	if c.unpendSendState() && _type != stState {
		// We needed to send a state packet, but this packet suppresses that
		// need.
		unsentStatePackets.Add(1)
	}
	return
}

func (c *Conn) unpendSendState() (wasPending bool) {
	wasPending = c.pendingSendState
	c.pendingSendState = false
	c.sendPendingSendSendStateTimer.Stop()
	c.sendPendingSendStateTimerActive = false
	return
}

func (c *Conn) pendSendState() {
	if c.pendingSendState {
		// A state packet is pending but hasn't been sent, and we want to send
		// another.
		unsentStatePackets.Add(1)
	}
	c.pendingSendState = true
	if !c.sendPendingSendStateTimerActive {
		c.sendPendingSendSendStateTimer.Reset(pendingSendStateDelay)
		c.sendPendingSendStateTimerActive = true
	}
}

func (me *Conn) writeSyn() {
	me.write(stSyn, me.recv_id, nil, me.seq_nr)
	return
}

func (c *Conn) write(_type st, connID uint16, payload []byte, seqNr uint16) (n int, err error) {
	switch _type {
	case stSyn, stFin, stData:
	default:
		panic(_type)
	}
	if c.wroteFin.IsSet() {
		panic("can't write after fin")
	}
	if len(payload) > maxPayloadSize {
		payload = payload[:maxPayloadSize]
	}
	err = c.send(_type, connID, payload, seqNr)
	if err != nil {
		c.destroy(fmt.Errorf("error sending packet: %s", err))
		return
	}
	n = len(payload)
	// Copy payload so caller to write can continue to use the buffer.
	if payload != nil {
		payload = append(sendBufferPool.Get().([]byte)[:0:minMTU], payload...)
	}
	send := &send{
		payloadSize: uint32(len(payload)),
		started:     missinggo.MonotonicNow(),
		_type:       _type,
		connID:      connID,
		payload:     payload,
		seqNr:       seqNr,
		conn:        c,
	}
	send.resendTimer = time.AfterFunc(c.resendTimeout(), send.timeoutResend)
	c.unackedSends = append(c.unackedSends, send)
	c.cur_window += send.payloadSize
	c.updateCanWrite()
	c.seq_nr++
	return
}

// TODO: Introduce a minimum latency.
func (c *Conn) latency() (ret time.Duration) {
	if len(c.latencies) == 0 {
		return initialLatency
	}
	for _, l := range c.latencies {
		ret += l
	}
	ret = (ret + time.Duration(len(c.latencies)) - 1) / time.Duration(len(c.latencies))
	return
}

func (c *Conn) sendState() {
	c.send(stState, c.send_id, nil, c.seq_nr)
	sentStatePackets.Add(1)
}

func (c *Conn) sendReset() {
	c.send(stReset, c.send_id, nil, c.seq_nr)
}

func (c *Conn) addLatency(l time.Duration) {
	c.latencies = append(c.latencies, l)
	if len(c.latencies) > 10 {
		c.latencies = c.latencies[len(c.latencies)-10:]
	}
}

// Ack our send with the given sequence number.
func (c *Conn) ack(nr uint16) {
	if !seqLess(c.lastAck, nr) {
		// Already acked.
		return
	}
	i := nr - c.lastAck - 1
	if int(i) >= len(c.unackedSends) {
		// Remote has acknowledged receipt of packets we haven't even sent.
		acksReceivedAheadOfSyn.Add(1)
		// log.Printf("got ack ahead of syn (%x > %x)", nr, c.seq_nr-1)
		return
	}
	s := c.unackedSends[i]
	latency, first := s.Ack()
	if first {
		c.cur_window -= s.payloadSize
		c.updateCanWrite()
		c.addLatency(latency)
	}
	// Trim sends that aren't needed anymore.
	for len(c.unackedSends) != 0 {
		if !c.unackedSends[0].acked.IsSet() {
			// Can't trim unacked sends any further.
			return
		}
		// Trim the front of the unacked sends.
		c.unackedSends = c.unackedSends[1:]
		c.updateCanWrite()
		c.lastAck++
	}
}

func (c *Conn) ackTo(nr uint16) {
	if !seqLess(nr, c.seq_nr) {
		return
	}
	for seqLess(c.lastAck, nr) {
		c.ack(c.lastAck + 1)
	}
}

// Return the send state for the sequence number. Returns nil if there's no
// outstanding send for that sequence number.
func (c *Conn) seqSend(seqNr uint16) *send {
	if !seqLess(c.lastAck, seqNr) {
		// Presumably already acked.
		return nil
	}
	i := int(seqNr - c.lastAck - 1)
	if i >= len(c.unackedSends) {
		// No such send.
		return nil
	}
	return c.unackedSends[i]
}

func (c *Conn) resendTimeout() time.Duration {
	l := c.latency()
	ret := missinggo.JitterDuration(3*l, l)
	return ret
}

func (c *Conn) ackSkipped(seqNr uint16) {
	send := c.seqSend(seqNr)
	if send == nil {
		return
	}
	send.acksSkipped++
	if send.acked.IsSet() {
		return
	}
	switch send.acksSkipped {
	case 3, 60:
		ackSkippedResends.Add(1)
		send.resend()
		send.resendTimer.Reset(c.resendTimeout() * time.Duration(send.numResends))
	default:
	}
}

// Handle a packet destined for this connection.
func (c *Conn) receivePacket(h header, payload []byte) {
	c.packetReadTimeoutTimer.Reset(packetReadTimeout)
	c.processDelivery(h, payload)
}

func (c *Conn) receivePacketTimeoutCallback() {
	mu.Lock()
	c.destroy(errors.New("no packet read timeout"))
	mu.Unlock()
}

func (c *Conn) lazyDestroy() {
	if c.wroteFin.IsSet() && len(c.unackedSends) <= 1 && (c.gotFin.IsSet() || c.closed.IsSet()) {
		c.destroy(errors.New("lazily destroyed"))
	}
}

func (c *Conn) processDelivery(h header, payload []byte) {
	deliveriesProcessed.Add(1)
	defer c.lazyDestroy()
	c.assertHeader(h)
	c.peerWndSize = h.WndSize
	c.applyAcks(h)
	if h.Timestamp == 0 {
		c.lastTimeDiff = 0
	} else {
		c.lastTimeDiff = c.timestamp() - h.Timestamp
	}

	if h.Type == stReset {
		c.destroy(errors.New("peer reset"))
		return
	}
	if !c.synAcked {
		if h.Type != stState {
			return
		}
		c.synAcked = true
		c.updateCanWrite()
		c.ack_nr = h.SeqNr - 1
		return
	}
	if h.Type == stState {
		return
	}
	// Even if we didn't need or want this packet, we need to inform the peer
	// what our state is, in case they missed something.
	c.pendSendState()
	if !seqLess(c.ack_nr, h.SeqNr) {
		// Already received this packet.
		return
	}
	inboundIndex := int(h.SeqNr - c.ack_nr - 1)
	if inboundIndex < len(c.inbound) && c.inbound[inboundIndex].seen {
		// Already received this packet.
		return
	}
	// Derived from running in production:
	// grep -oP '(?<=packet out of order, index=)\d+' log | sort -n | uniq -c
	// 64 should correspond to 8 bytes of selective ack.
	if inboundIndex >= maxUnackedInbound {
		// Discard packet too far ahead.
		if logLevel >= 1 {
			log.Printf("received packet from %s %d ahead of next seqnr (%x > %x)", c.remoteSocketAddr, inboundIndex, h.SeqNr, c.ack_nr+1)
		}
		return
	}
	// Extend inbound so the new packet has a place.
	for inboundIndex >= len(c.inbound) {
		c.inbound = append(c.inbound, recv{})
	}
	c.inbound[inboundIndex] = recv{true, payload, h.Type}
	c.inboundWnd += len(payload)
	c.processInbound()
}

func (c *Conn) applyAcks(h header) {
	c.ackTo(h.AckNr)
	for _, ext := range h.Extensions {
		switch ext.Type {
		case extensionTypeSelectiveAck:
			c.ackSkipped(h.AckNr + 1)
			bitmask := selectiveAckBitmask{ext.Bytes}
			for i := 0; i < bitmask.NumBits(); i++ {
				if bitmask.BitIsSet(i) {
					nr := h.AckNr + 2 + uint16(i)
					// log.Printf("selectively acked %d", nr)
					c.ack(nr)
				} else {
					c.ackSkipped(h.AckNr + 2 + uint16(i))
				}
			}
		}
	}
}

func (c *Conn) assertHeader(h header) {
	if h.Type == stSyn {
		if h.ConnID != c.send_id {
			panic(fmt.Sprintf("%d != %d", h.ConnID, c.send_id))
		}
	} else {
		if h.ConnID != c.recv_id {
			panic("erroneous delivery")
		}
	}
}

func (c *Conn) updateReadBufNotEmpty() {
	c.readBufNotEmpty.SetBool(len(c.readBuf) != 0)
}

func (c *Conn) processInbound() {
	// Consume consecutive next packets.
	for !c.gotFin.IsSet() && len(c.inbound) > 0 && c.inbound[0].seen && len(c.readBuf) < readBufferLen {
		c.ack_nr++
		p := c.inbound[0]
		c.inbound = c.inbound[1:]
		c.inboundWnd -= len(p.data)
		c.readBuf = append(c.readBuf, p.data...)
		c.updateReadBufNotEmpty()
		if p.Type == stFin {
			c.gotFin.Set()
		}
	}
}

func (c *Conn) waitAck(seq uint16) {
	send := c.seqSend(seq)
	if send == nil {
		return
	}
	missinggo.WaitEvents(&mu, &send.acked, &c.destroyed)
	return
}

// Waits for sent SYN to be ACKed. Returns any errors.
func (c *Conn) recvSynAck() (err error) {
	mu.Lock()
	defer mu.Unlock()
	c.waitAck(1)
	if c.err != nil {
		err = c.err
	}
	c.synAcked = true
	c.updateCanWrite()
	return err
}

func (c *Conn) writeFin() {
	if c.wroteFin.IsSet() {
		return
	}
	c.write(stFin, c.send_id, nil, c.seq_nr)
	c.wroteFin.Set()
	return
}

func (c *Conn) destroy(reason error) {
	c.destroyed.Set()
	if c.err == nil {
		c.err = reason
	}
	c.detach()
}

func (c *Conn) closeNow() (err error) {
	c.closed.Set()
	c.writeFin()
	c.destroy(errors.New("destroyed"))
	return
}

func (c *Conn) Close() (err error) {
	mu.Lock()
	defer mu.Unlock()
	c.closed.Set()
	c.writeFin()
	c.lazyDestroy()
	return
}

func (c *Conn) LocalAddr() net.Addr {
	return addr{c.socket.Addr()}
}

func (c *Conn) Read(b []byte) (n int, err error) {
	mu.Lock()
	defer mu.Unlock()
	for {
		n = copy(b, c.readBuf)
		c.readBuf = c.readBuf[n:]
		c.updateReadBufNotEmpty()
		if n != 0 {
			// Inbound packets are backed up when the read buffer is too big.
			c.processInbound()
			return
		}
		if c.gotFin.IsSet() || c.closed.IsSet() {
			err = io.EOF
			return
		}
		if c.destroyed.IsSet() {
			if c.err == nil {
				panic("closed without receiving fin, and no error")
			}
			err = c.err
			return
		}
		if c.connDeadlines.read.passed.IsSet() {
			err = errTimeout
			return
		}
		missinggo.WaitEvents(&mu,
			&c.gotFin,
			&c.closed,
			&c.destroyed,
			&c.connDeadlines.read.passed,
			&c.readBufNotEmpty)
	}
}

func (c *Conn) RemoteAddr() net.Addr {
	return addr{c.remoteSocketAddr}
}

func (c *Conn) String() string {
	return fmt.Sprintf("<UTPConn %s-%s (%d)>", c.LocalAddr(), c.RemoteAddr(), c.recv_id)
}

func (c *Conn) updateCanWrite() {
	c.canWrite.SetBool(c.synAcked &&
		len(c.unackedSends) < maxUnackedSends &&
		c.cur_window <= c.peerWndSize)
}

func (c *Conn) Write(p []byte) (n int, err error) {
	mu.Lock()
	defer mu.Unlock()
	for len(p) != 0 {
		if c.wroteFin.IsSet() || c.closed.IsSet() {
			err = errClosed
			return
		}
		if c.destroyed.IsSet() {
			err = c.err
			return
		}
		if c.connDeadlines.write.passed.IsSet() {
			err = errTimeout
			return
		}
		// If peerWndSize is 0, we still want to send something, so don't
		// block until we exceed it.
		if c.canWrite.IsSet() {
			var n1 int
			n1, err = c.write(stData, c.send_id, p, c.seq_nr)
			n += n1
			if err != nil {
				break
			}
			if n1 == 0 {
				panic(len(p))
			}
			p = p[n1:]
			continue
		}
		missinggo.WaitEvents(&mu,
			&c.wroteFin,
			&c.closed,
			&c.destroyed,
			&c.connDeadlines.write.passed,
			&c.canWrite)
	}
	return
}

func (c *Conn) detach() {
	s := c.socket
	_, ok := s.conns[c.connKey]
	if !ok {
		return
	}
	delete(s.conns, c.connKey)
	s.lazyDestroy()
}
