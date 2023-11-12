package libp2pquic

import (
	"sync"

	tpt "github.com/libp2p/go-libp2p/core/transport"
	"github.com/libp2p/go-libp2p/p2p/transport/quicreuse"

	ma "github.com/multiformats/go-multiaddr"
	"github.com/quic-go/quic-go"
)

const acceptBufferPerVersion = 4

// virtualListener is a listener that exposes a single multiaddr but uses another listener under the hood
type virtualListener struct {
	*listener
	udpAddr       string
	version       quic.VersionNumber
	t             *transport
	acceptRunnner *acceptLoopRunner
	acceptChan    chan acceptVal
}

var _ tpt.Listener = &virtualListener{}

func (l *virtualListener) Multiaddr() ma.Multiaddr {
	return l.listener.localMultiaddrs[l.version]
}

func (l *virtualListener) Close() error {
	l.acceptRunnner.RmAcceptForVersion(l.version, tpt.ErrListenerClosed)
	return l.t.CloseVirtualListener(l)
}

func (l *virtualListener) Accept() (tpt.CapableConn, error) {
	return l.acceptRunnner.Accept(l.listener, l.version, l.acceptChan)
}

type acceptVal struct {
	conn tpt.CapableConn
	err  error
}

type acceptLoopRunner struct {
	acceptSem chan struct{}

	muxerMu     sync.Mutex
	muxer       map[quic.VersionNumber]chan acceptVal
	muxerClosed bool
}

func (r *acceptLoopRunner) AcceptForVersion(v quic.VersionNumber) chan acceptVal {
	r.muxerMu.Lock()
	defer r.muxerMu.Unlock()

	ch := make(chan acceptVal, acceptBufferPerVersion)

	if _, ok := r.muxer[v]; ok {
		panic("unexpected chan already found in accept muxer")
	}

	r.muxer[v] = ch
	return ch
}

func (r *acceptLoopRunner) RmAcceptForVersion(v quic.VersionNumber, err error) {
	r.muxerMu.Lock()
	defer r.muxerMu.Unlock()

	if r.muxerClosed {
		// Already closed, all versions are removed
		return
	}

	ch, ok := r.muxer[v]
	if !ok {
		panic("expected chan in accept muxer")
	}
	ch <- acceptVal{err: err}
	delete(r.muxer, v)
}

func (r *acceptLoopRunner) sendErrAndClose(err error) {
	r.muxerMu.Lock()
	defer r.muxerMu.Unlock()
	r.muxerClosed = true
	for k, ch := range r.muxer {
		select {
		case ch <- acceptVal{err: err}:
		default:
		}
		delete(r.muxer, k)
		close(ch)
	}
}

// innerAccept is the inner logic of the Accept loop. Assume caller holds the
// acceptSemaphore. May return both a nil conn and nil error if it didn't find a
// conn with the expected version
func (r *acceptLoopRunner) innerAccept(l *listener, expectedVersion quic.VersionNumber, bufferedConnChan chan acceptVal) (tpt.CapableConn, error) {
	select {
	// Check if we have a buffered connection first from an earlier Accept call
	case v, ok := <-bufferedConnChan:
		if !ok {
			return nil, tpt.ErrListenerClosed
		}
		return v.conn, v.err
	default:
	}

	conn, err := l.Accept()

	if err != nil {
		r.sendErrAndClose(err)
		return nil, err
	}

	_, version, err := quicreuse.FromQuicMultiaddr(conn.RemoteMultiaddr())
	if err != nil {
		r.sendErrAndClose(err)
		return nil, err
	}

	if version == expectedVersion {
		return conn, nil
	}

	// This wasn't the version we were expecting, lets queue it up for a
	// future Accept call with a different version
	r.muxerMu.Lock()
	ch, ok := r.muxer[version]
	r.muxerMu.Unlock()

	if !ok {
		// Nothing to handle this connection version. Close it
		conn.Close()
		return nil, nil
	}

	// Non blocking
	select {
	case ch <- acceptVal{conn: conn}:
	default:
		// accept queue filled up, drop the connection
		conn.Close()
		log.Warn("Accept queue filled. Dropping connection.")
	}

	return nil, nil
}

func (r *acceptLoopRunner) Accept(l *listener, expectedVersion quic.VersionNumber, bufferedConnChan chan acceptVal) (tpt.CapableConn, error) {
	for {
		var conn tpt.CapableConn
		var err error
		select {
		case r.acceptSem <- struct{}{}:
			conn, err = r.innerAccept(l, expectedVersion, bufferedConnChan)
			<-r.acceptSem

			if conn == nil && err == nil {
				// Didn't find a conn for the expected version and there was no error, lets try again
				continue
			}
		case v, ok := <-bufferedConnChan:
			if !ok {
				return nil, tpt.ErrListenerClosed
			}
			conn = v.conn
			err = v.err
		}
		return conn, err
	}
}
