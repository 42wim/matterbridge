package pubsub

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/libp2p/go-msgio/protoio"
)

var TraceBufferSize = 1 << 16 // 64K ought to be enough for everyone; famous last words.
var MinTraceBatchSize = 16

// rejection reasons
const (
	RejectBlacklstedPeer      = "blacklisted peer"
	RejectBlacklistedSource   = "blacklisted source"
	RejectMissingSignature    = "missing signature"
	RejectUnexpectedSignature = "unexpected signature"
	RejectUnexpectedAuthInfo  = "unexpected auth info"
	RejectInvalidSignature    = "invalid signature"
	RejectValidationQueueFull = "validation queue full"
	RejectValidationThrottled = "validation throttled"
	RejectValidationFailed    = "validation failed"
	RejectValidationIgnored   = "validation ignored"
	RejectSelfOrigin          = "self originated message"
)

type basicTracer struct {
	ch     chan struct{}
	mx     sync.Mutex
	buf    []*pb.TraceEvent
	lossy  bool
	closed bool
}

func (t *basicTracer) Trace(evt *pb.TraceEvent) {
	t.mx.Lock()
	defer t.mx.Unlock()

	if t.closed {
		return
	}

	if t.lossy && len(t.buf) > TraceBufferSize {
		log.Debug("trace buffer overflow; dropping trace event")
	} else {
		t.buf = append(t.buf, evt)
	}

	select {
	case t.ch <- struct{}{}:
	default:
	}
}

func (t *basicTracer) Close() {
	t.mx.Lock()
	defer t.mx.Unlock()
	if !t.closed {
		t.closed = true
		close(t.ch)
	}
}

// JSONTracer is a tracer that writes events to a file, encoded in ndjson.
type JSONTracer struct {
	basicTracer
	w io.WriteCloser
}

// NewJsonTracer creates a new JSONTracer writing traces to file.
func NewJSONTracer(file string) (*JSONTracer, error) {
	return OpenJSONTracer(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

// OpenJSONTracer creates a new JSONTracer, with explicit control of OpenFile flags and permissions.
func OpenJSONTracer(file string, flags int, perm os.FileMode) (*JSONTracer, error) {
	f, err := os.OpenFile(file, flags, perm)
	if err != nil {
		return nil, err
	}

	tr := &JSONTracer{w: f, basicTracer: basicTracer{ch: make(chan struct{}, 1)}}
	go tr.doWrite()

	return tr, nil
}

func (t *JSONTracer) doWrite() {
	var buf []*pb.TraceEvent
	enc := json.NewEncoder(t.w)
	for {
		_, ok := <-t.ch

		t.mx.Lock()
		tmp := t.buf
		t.buf = buf[:0]
		buf = tmp
		t.mx.Unlock()

		for i, evt := range buf {
			err := enc.Encode(evt)
			if err != nil {
				log.Warnf("error writing event trace: %s", err.Error())
			}
			buf[i] = nil
		}

		if !ok {
			t.w.Close()
			return
		}
	}
}

var _ EventTracer = (*JSONTracer)(nil)

// PBTracer is a tracer that writes events to a file, as delimited protobufs.
type PBTracer struct {
	basicTracer
	w io.WriteCloser
}

func NewPBTracer(file string) (*PBTracer, error) {
	return OpenPBTracer(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
}

// OpenPBTracer creates a new PBTracer, with explicit control of OpenFile flags and permissions.
func OpenPBTracer(file string, flags int, perm os.FileMode) (*PBTracer, error) {
	f, err := os.OpenFile(file, flags, perm)
	if err != nil {
		return nil, err
	}

	tr := &PBTracer{w: f, basicTracer: basicTracer{ch: make(chan struct{}, 1)}}
	go tr.doWrite()

	return tr, nil
}

func (t *PBTracer) doWrite() {
	var buf []*pb.TraceEvent
	w := protoio.NewDelimitedWriter(t.w)
	for {
		_, ok := <-t.ch

		t.mx.Lock()
		tmp := t.buf
		t.buf = buf[:0]
		buf = tmp
		t.mx.Unlock()

		for i, evt := range buf {
			err := w.WriteMsg(evt)
			if err != nil {
				log.Warnf("error writing event trace: %s", err.Error())
			}
			buf[i] = nil
		}

		if !ok {
			t.w.Close()
			return
		}
	}
}

var _ EventTracer = (*PBTracer)(nil)

const RemoteTracerProtoID = protocol.ID("/libp2p/pubsub/tracer/1.0.0")

// RemoteTracer is a tracer that sends trace events to a remote peer
type RemoteTracer struct {
	basicTracer
	ctx  context.Context
	host host.Host
	peer peer.ID
}

// NewRemoteTracer constructs a RemoteTracer, tracing to the peer identified by pi
func NewRemoteTracer(ctx context.Context, host host.Host, pi peer.AddrInfo) (*RemoteTracer, error) {
	tr := &RemoteTracer{ctx: ctx, host: host, peer: pi.ID, basicTracer: basicTracer{ch: make(chan struct{}, 1), lossy: true}}
	host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
	go tr.doWrite()
	return tr, nil
}

func (t *RemoteTracer) doWrite() {
	var buf []*pb.TraceEvent

	s, err := t.openStream()
	if err != nil {
		log.Debugf("error opening remote tracer stream: %s", err.Error())
		return
	}

	var batch pb.TraceEventBatch

	gzipW := gzip.NewWriter(s)
	w := protoio.NewDelimitedWriter(gzipW)

	for {
		_, ok := <-t.ch

		// deadline for batch accumulation
		deadline := time.Now().Add(time.Second)

		t.mx.Lock()
		for len(t.buf) < MinTraceBatchSize && time.Now().Before(deadline) {
			t.mx.Unlock()
			time.Sleep(100 * time.Millisecond)
			t.mx.Lock()
		}

		tmp := t.buf
		t.buf = buf[:0]
		buf = tmp
		t.mx.Unlock()

		if len(buf) == 0 {
			goto end
		}

		batch.Batch = buf

		err = w.WriteMsg(&batch)
		if err != nil {
			log.Debugf("error writing trace event batch: %s", err)
			goto end
		}

		err = gzipW.Flush()
		if err != nil {
			log.Debugf("error flushin gzip stream: %s", err)
			goto end
		}

	end:
		// nil out the buffer to gc consumed events
		for i := range buf {
			buf[i] = nil
		}

		if !ok {
			if err != nil {
				s.Reset()
			} else {
				gzipW.Close()
				s.Close()
			}
			return
		}

		if err != nil {
			s.Reset()
			s, err = t.openStream()
			if err != nil {
				log.Debugf("error opening remote tracer stream: %s", err.Error())
				return
			}

			gzipW.Reset(s)
		}
	}
}

func (t *RemoteTracer) openStream() (network.Stream, error) {
	for {
		ctx, cancel := context.WithTimeout(t.ctx, time.Minute)
		s, err := t.host.NewStream(ctx, t.peer, RemoteTracerProtoID)
		cancel()
		if err != nil {
			if t.ctx.Err() != nil {
				return nil, err
			}

			// wait a minute and try again, to account for transient server downtime
			select {
			case <-time.After(time.Minute):
				continue
			case <-t.ctx.Done():
				return nil, t.ctx.Err()
			}
		}

		return s, nil
	}
}

var _ EventTracer = (*RemoteTracer)(nil)
