package udp

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/anacrolix/dht/v2/krpc"
)

// Client interacts with UDP trackers via its Writer and Dispatcher. It has no knowledge of
// connection specifics.
type Client struct {
	mu           sync.Mutex
	connId       ConnectionId
	connIdIssued time.Time
	Dispatcher   *Dispatcher
	Writer       io.Writer
}

func (cl *Client) Announce(
	ctx context.Context, req AnnounceRequest, opts Options,
	// Decides whether the response body is IPv6 or IPv4, see BEP 15.
	ipv6 func(net.Addr) bool,
) (
	respHdr AnnounceResponseHeader,
	// A slice of krpc.NodeAddr, likely wrapped in an appropriate unmarshalling wrapper.
	peers AnnounceResponsePeers,
	err error,
) {
	respBody, addr, err := cl.request(ctx, ActionAnnounce, append(mustMarshal(req), opts.Encode()...))
	if err != nil {
		return
	}
	r := bytes.NewBuffer(respBody)
	err = Read(r, &respHdr)
	if err != nil {
		err = fmt.Errorf("reading response header: %w", err)
		return
	}
	if ipv6(addr) {
		peers = &krpc.CompactIPv6NodeAddrs{}
	} else {
		peers = &krpc.CompactIPv4NodeAddrs{}
	}
	err = peers.UnmarshalBinary(r.Bytes())
	if err != nil {
		err = fmt.Errorf("reading response peers: %w", err)
	}
	return
}

// There's no way to pass options in a scrape, since we don't when the request body ends.
func (cl *Client) Scrape(
	ctx context.Context, ihs []InfoHash,
) (
	out ScrapeResponse, err error,
) {
	respBody, _, err := cl.request(ctx, ActionScrape, mustMarshal(ScrapeRequest(ihs)))
	if err != nil {
		return
	}
	r := bytes.NewBuffer(respBody)
	for r.Len() != 0 {
		var item ScrapeInfohashResult
		err = Read(r, &item)
		if err != nil {
			return
		}
		out = append(out, item)
	}
	if len(out) > len(ihs) {
		err = fmt.Errorf("got %v results but expected %v", len(out), len(ihs))
		return
	}
	return
}

func (cl *Client) connect(ctx context.Context) (err error) {
	// We could get fancier here and use RWMutex, and even fire off the connection asynchronously
	// and provide a grace period while it resolves.
	cl.mu.Lock()
	defer cl.mu.Unlock()
	if !cl.connIdIssued.IsZero() && time.Since(cl.connIdIssued) < time.Minute {
		return nil
	}
	respBody, _, err := cl.request(ctx, ActionConnect, nil)
	if err != nil {
		return err
	}
	var connResp ConnectionResponse
	err = binary.Read(bytes.NewReader(respBody), binary.BigEndian, &connResp)
	if err != nil {
		return
	}
	cl.connId = connResp.ConnectionId
	cl.connIdIssued = time.Now()
	return
}

func (cl *Client) connIdForRequest(ctx context.Context, action Action) (id ConnectionId, err error) {
	if action == ActionConnect {
		id = ConnectRequestConnectionId
		return
	}
	err = cl.connect(ctx)
	if err != nil {
		return
	}
	id = cl.connId
	return
}

func (cl *Client) requestWriter(ctx context.Context, action Action, body []byte, tId TransactionId) (err error) {
	var buf bytes.Buffer
	for n := 0; ; n++ {
		var connId ConnectionId
		connId, err = cl.connIdForRequest(ctx, action)
		if err != nil {
			return
		}
		buf.Reset()
		err = binary.Write(&buf, binary.BigEndian, RequestHeader{
			ConnectionId:  connId,
			Action:        action,
			TransactionId: tId,
		})
		if err != nil {
			panic(err)
		}
		buf.Write(body)
		_, err = cl.Writer.Write(buf.Bytes())
		if err != nil {
			return
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout(n)):
		}
	}
}

func (cl *Client) request(ctx context.Context, action Action, body []byte) (respBody []byte, addr net.Addr, err error) {
	respChan := make(chan DispatchedResponse, 1)
	t := cl.Dispatcher.NewTransaction(func(dr DispatchedResponse) {
		respChan <- dr
	})
	defer t.End()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	writeErr := make(chan error, 1)
	go func() {
		writeErr <- cl.requestWriter(ctx, action, body, t.Id())
	}()
	select {
	case dr := <-respChan:
		if dr.Header.Action == action {
			respBody = dr.Body
			addr = dr.Addr
		} else if dr.Header.Action == ActionError {
			// I've seen "Connection ID mismatch.^@" in less and other tools, I think they're just
			// not handling a trailing \x00 nicely.
			err = fmt.Errorf("error response: %#q", dr.Body)
		} else {
			err = fmt.Errorf("unexpected response action %v", dr.Header.Action)
		}
	case err = <-writeErr:
		err = fmt.Errorf("write error: %w", err)
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}
