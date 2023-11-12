package rendezvous

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	inet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio/pbio"

	pb "github.com/waku-org/go-libp2p-rendezvous/pb"
)

var (
	DiscoverAsyncInterval = 2 * time.Minute
)

type RendezvousPoint interface {
	Register(ctx context.Context, ns string, ttl int) (time.Duration, error)
	Unregister(ctx context.Context, ns string) error
	Discover(ctx context.Context, ns string, limit int, cookie []byte) ([]Registration, []byte, error)
	DiscoverAsync(ctx context.Context, ns string) (<-chan Registration, error)
}

type Registration struct {
	Peer peer.AddrInfo
	Ns   string
	Ttl  int
}

type RendezvousClient interface {
	Register(ctx context.Context, ns string, ttl int) (time.Duration, error)
	Unregister(ctx context.Context, ns string) error
	Discover(ctx context.Context, ns string, limit int, cookie []byte) ([]peer.AddrInfo, []byte, error)
	DiscoverAsync(ctx context.Context, ns string) (<-chan peer.AddrInfo, error)
}

func NewRendezvousPoint(host host.Host, p peer.ID, opts ...RendezvousPointOption) RendezvousPoint {
	cfg := defaultRendezvousPointConfig
	cfg.apply(opts...)
	return &rendezvousPoint{
		addrFactory: cfg.AddrsFactory,
		host:        host,
		p:           p,
	}
}

type rendezvousPoint struct {
	addrFactory AddrsFactory
	host        host.Host
	p           peer.ID
}

func NewRendezvousClient(host host.Host, rp peer.ID) RendezvousClient {
	return NewRendezvousClientWithPoint(NewRendezvousPoint(host, rp))
}

func NewRendezvousClientWithPoint(rp RendezvousPoint) RendezvousClient {
	return &rendezvousClient{rp: rp}
}

type rendezvousClient struct {
	rp RendezvousPoint
}

func (rp *rendezvousPoint) Register(ctx context.Context, ns string, ttl int) (time.Duration, error) {
	s, err := rp.host.NewStream(ctx, rp.p, RendezvousProto)
	if err != nil {
		return 0, err
	}
	defer s.Reset()

	r := pbio.NewDelimitedReader(s, inet.MessageSizeMax)
	w := pbio.NewDelimitedWriter(s)

	addrs := rp.addrFactory(rp.host.Addrs())
	if len(addrs) == 0 {
		return 0, fmt.Errorf("no addrs available to advertise: %s", ns)
	}

	log.Debugf("advertising on `%s` with: %v", ns, addrs)

	privKey := rp.host.Peerstore().PrivKey(rp.host.ID())
	req, err := newRegisterMessage(privKey, ns, peer.AddrInfo{ID: rp.host.ID(), Addrs: addrs}, ttl)
	if err != nil {
		return 0, err
	}

	err = w.WriteMsg(req)
	if err != nil {
		return 0, err
	}

	var res pb.Message
	err = r.ReadMsg(&res)
	if err != nil {
		return 0, err
	}

	if res.GetType() != pb.Message_REGISTER_RESPONSE {
		return 0, fmt.Errorf("unexpected response: %s", res.GetType().String())
	}

	response := res.GetRegisterResponse()
	status := response.GetStatus()
	if status != pb.Message_OK {
		return 0, RendezvousError{Status: status, Text: res.GetRegisterResponse().GetStatusText()}
	}

	responseTTL := int64(0)
	if response.Ttl != nil {
		responseTTL = int64(*response.Ttl)
	}

	return time.Duration(responseTTL) * time.Second, nil
}

func (rc *rendezvousClient) Register(ctx context.Context, ns string, ttl int) (time.Duration, error) {
	if ttl < 120 {
		return 0, fmt.Errorf("registration TTL is too short")
	}

	returnedTTL, err := rc.rp.Register(ctx, ns, ttl)
	if err != nil {
		return 0, err
	}

	go registerRefresh(ctx, rc.rp, ns, ttl)
	return returnedTTL, nil
}

func registerRefresh(ctx context.Context, rz RendezvousPoint, ns string, ttl int) {
	var refresh time.Duration
	errcount := 0

	for {
		if errcount > 0 {
			// do randomized exponential backoff, up to ~4 hours
			if errcount > 7 {
				errcount = 7
			}
			backoff := 2 << uint(errcount)
			refresh = 5*time.Minute + time.Duration(rand.Intn(backoff*60000))*time.Millisecond
		} else {
			refresh = time.Duration(ttl-30) * time.Second
		}

		select {
		case <-time.After(refresh):
		case <-ctx.Done():
			return
		}

		_, err := rz.Register(ctx, ns, ttl)
		if err != nil {
			log.Errorf("Error registering [%s]: %s", ns, err.Error())
			errcount++
		} else {
			errcount = 0
		}
	}
}

func (rp *rendezvousPoint) Unregister(ctx context.Context, ns string) error {
	s, err := rp.host.NewStream(ctx, rp.p, RendezvousProto)
	if err != nil {
		return err
	}
	defer s.Close()

	w := pbio.NewDelimitedWriter(s)
	req := newUnregisterMessage(ns, rp.host.ID())
	return w.WriteMsg(req)
}

func (rc *rendezvousClient) Unregister(ctx context.Context, ns string) error {
	return rc.rp.Unregister(ctx, ns)
}

func (rp *rendezvousPoint) Discover(ctx context.Context, ns string, limit int, cookie []byte) ([]Registration, []byte, error) {
	s, err := rp.host.NewStream(ctx, rp.p, RendezvousProto)
	if err != nil {
		return nil, nil, err
	}
	defer s.Reset()

	r := pbio.NewDelimitedReader(s, inet.MessageSizeMax)
	w := pbio.NewDelimitedWriter(s)

	return discoverQuery(ns, limit, cookie, r, w)
}

func discoverQuery(ns string, limit int, cookie []byte, r pbio.Reader, w pbio.Writer) ([]Registration, []byte, error) {
	req := newDiscoverMessage(ns, limit, cookie)
	err := w.WriteMsg(req)
	if err != nil {
		return nil, nil, err
	}

	var res pb.Message
	err = r.ReadMsg(&res)
	if err != nil {
		return nil, nil, err
	}

	if res.GetType() != pb.Message_DISCOVER_RESPONSE {
		return nil, nil, fmt.Errorf("unexpected response: %s", res.GetType().String())
	}

	status := res.GetDiscoverResponse().GetStatus()
	if status != pb.Message_OK {
		return nil, nil, RendezvousError{Status: status, Text: res.GetDiscoverResponse().GetStatusText()}
	}

	regs := res.GetDiscoverResponse().GetRegistrations()
	result := make([]Registration, 0, len(regs))
	for _, reg := range regs {
		pi, err := pbToPeerRecord(reg.SignedPeerRecord)
		if err != nil {
			log.Errorf("Invalid peer info: %s", err.Error())
			continue
		}
		result = append(result, Registration{Peer: pi, Ns: reg.GetNs(), Ttl: int(reg.GetTtl())})
	}

	return result, res.GetDiscoverResponse().GetCookie(), nil
}

func (rp *rendezvousPoint) DiscoverAsync(ctx context.Context, ns string) (<-chan Registration, error) {
	s, err := rp.host.NewStream(ctx, rp.p, RendezvousProto)
	if err != nil {
		return nil, err
	}

	ch := make(chan Registration)
	go discoverAsync(ctx, ns, s, ch)
	return ch, nil
}

func discoverAsync(ctx context.Context, ns string, s inet.Stream, ch chan Registration) {
	defer s.Reset()
	defer close(ch)

	r := pbio.NewDelimitedReader(s, inet.MessageSizeMax)
	w := pbio.NewDelimitedWriter(s)

	const batch = 200

	var (
		cookie []byte
		regs   []Registration
		err    error
	)

	for {
		regs, cookie, err = discoverQuery(ns, batch, cookie, r, w)
		if err != nil {
			// TODO robust error recovery
			//      - handle closed streams with backoff + new stream, preserving the cookie
			//      - handle E_INVALID_COOKIE errors in that case to restart the discovery
			log.Errorf("Error in discovery [%s]: %s", ns, err.Error())
			return
		}

		for _, reg := range regs {
			select {
			case ch <- reg:
			case <-ctx.Done():
				return
			}
		}

		if len(regs) < batch {
			// TODO adaptive backoff for heavily loaded rendezvous points
			select {
			case <-time.After(DiscoverAsyncInterval):
			case <-ctx.Done():
				return
			}
		}
	}
}

func (rc *rendezvousClient) Discover(ctx context.Context, ns string, limit int, cookie []byte) ([]peer.AddrInfo, []byte, error) {
	regs, cookie, err := rc.rp.Discover(ctx, ns, limit, cookie)
	if err != nil {
		return nil, nil, err
	}

	pinfos := make([]peer.AddrInfo, len(regs))
	for i, reg := range regs {
		pinfos[i] = reg.Peer
	}

	return pinfos, cookie, nil
}

func (rc *rendezvousClient) DiscoverAsync(ctx context.Context, ns string) (<-chan peer.AddrInfo, error) {
	rch, err := rc.rp.DiscoverAsync(ctx, ns)
	if err != nil {
		return nil, err
	}

	ch := make(chan peer.AddrInfo)
	go discoverPeersAsync(ctx, rch, ch)
	return ch, nil
}

func discoverPeersAsync(ctx context.Context, rch <-chan Registration, ch chan peer.AddrInfo) {
	defer close(ch)
	for {
		select {
		case reg, ok := <-rch:
			if !ok {
				return
			}

			select {
			case ch <- reg.Peer:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
