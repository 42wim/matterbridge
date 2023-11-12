package webtransport

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/quic-go/quicvarint"
)

type Dialer struct {
	// If not set, reasonable defaults will be used.
	// In order for WebTransport to function, this implementation will:
	// * overwrite the StreamHijacker and UniStreamHijacker
	// * enable datagram support
	// * set the MaxIncomingStreams to 100 on the quic.Config, if unset
	*http3.RoundTripper

	// StreamReorderingTime is the time an incoming WebTransport stream that cannot be associated
	// with a session is buffered.
	// This can happen if the response to a CONNECT request (that creates a new session) is reordered,
	// and arrives after the first WebTransport stream(s) for that session.
	// Defaults to 5 seconds.
	StreamReorderingTimeout time.Duration

	ctx       context.Context
	ctxCancel context.CancelFunc

	initOnce sync.Once

	conns sessionManager
}

func (d *Dialer) init() {
	timeout := d.StreamReorderingTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	d.conns = *newSessionManager(timeout)
	d.ctx, d.ctxCancel = context.WithCancel(context.Background())
	if d.RoundTripper == nil {
		d.RoundTripper = &http3.RoundTripper{}
	}
	d.RoundTripper.EnableDatagrams = true
	if d.RoundTripper.AdditionalSettings == nil {
		d.RoundTripper.AdditionalSettings = make(map[uint64]uint64)
	}
	d.RoundTripper.StreamHijacker = func(ft http3.FrameType, conn quic.Connection, str quic.Stream, e error) (hijacked bool, err error) {
		if isWebTransportError(e) {
			return true, nil
		}
		if ft != webTransportFrameType {
			return false, nil
		}
		id, err := quicvarint.Read(quicvarint.NewReader(str))
		if err != nil {
			if isWebTransportError(err) {
				return true, nil
			}
			return false, err
		}
		d.conns.AddStream(conn, str, sessionID(id))
		return true, nil
	}
	d.RoundTripper.UniStreamHijacker = func(st http3.StreamType, conn quic.Connection, str quic.ReceiveStream, err error) (hijacked bool) {
		if st != webTransportUniStreamType && !isWebTransportError(err) {
			return false
		}
		d.conns.AddUniStream(conn, str)
		return true
	}
	if d.QuicConfig == nil {
		d.QuicConfig = &quic.Config{}
	}
	if d.QuicConfig.MaxIncomingStreams == 0 {
		d.QuicConfig.MaxIncomingStreams = 100
	}
}

func (d *Dialer) Dial(ctx context.Context, urlStr string, reqHdr http.Header) (*http.Response, *Session, error) {
	d.initOnce.Do(func() { d.init() })

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, nil, err
	}
	if reqHdr == nil {
		reqHdr = http.Header{}
	}
	reqHdr.Add(webTransportDraftOfferHeaderKey, "1")
	req := &http.Request{
		Method: http.MethodConnect,
		Header: reqHdr,
		Proto:  "webtransport",
		Host:   u.Host,
		URL:    u,
	}
	req = req.WithContext(ctx)

	rsp, err := d.RoundTripper.RoundTripOpt(req, http3.RoundTripOpt{DontCloseRequestStream: true})
	if err != nil {
		return nil, nil, err
	}
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return rsp, nil, fmt.Errorf("received status %d", rsp.StatusCode)
	}
	str := rsp.Body.(http3.HTTPStreamer).HTTPStream()
	conn := d.conns.AddSession(
		rsp.Body.(http3.Hijacker).StreamCreator(),
		sessionID(str.StreamID()),
		str,
	)
	return rsp, conn, nil
}

func (d *Dialer) Close() error {
	d.ctxCancel()
	return nil
}
