package nat

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"

	"github.com/libp2p/go-nat"
)

// ErrNoMapping signals no mapping exists for an address
var ErrNoMapping = errors.New("mapping not established")

var log = logging.Logger("nat")

// MappingDuration is a default port mapping duration.
// Port mappings are renewed every (MappingDuration / 3)
const MappingDuration = time.Minute

// CacheTime is the time a mapping will cache an external address for
const CacheTime = 15 * time.Second

type entry struct {
	protocol string
	port     int
}

// so we can mock it in tests
var discoverGateway = nat.DiscoverGateway

// DiscoverNAT looks for a NAT device in the network and returns an object that can manage port mappings.
func DiscoverNAT(ctx context.Context) (*NAT, error) {
	natInstance, err := discoverGateway(ctx)
	if err != nil {
		return nil, err
	}
	var extAddr netip.Addr
	extIP, err := natInstance.GetExternalAddress()
	if err == nil {
		extAddr, _ = netip.AddrFromSlice(extIP)
	}

	// Log the device addr.
	addr, err := natInstance.GetDeviceAddress()
	if err != nil {
		log.Debug("DiscoverGateway address error:", err)
	} else {
		log.Debug("DiscoverGateway address:", addr)
	}

	ctx, cancel := context.WithCancel(context.Background())
	nat := &NAT{
		nat:       natInstance,
		extAddr:   extAddr,
		mappings:  make(map[entry]int),
		ctx:       ctx,
		ctxCancel: cancel,
	}
	nat.refCount.Add(1)
	go func() {
		defer nat.refCount.Done()
		nat.background()
	}()
	return nat, nil
}

// NAT is an object that manages address port mappings in
// NATs (Network Address Translators). It is a long-running
// service that will periodically renew port mappings,
// and keep an up-to-date list of all the external addresses.
type NAT struct {
	natmu sync.Mutex
	nat   nat.NAT
	// External IP of the NAT. Will be renewed periodically (every CacheTime).
	extAddr netip.Addr

	refCount  sync.WaitGroup
	ctx       context.Context
	ctxCancel context.CancelFunc

	mappingmu sync.RWMutex // guards mappings
	closed    bool
	mappings  map[entry]int
}

// Close shuts down all port mappings. NAT can no longer be used.
func (nat *NAT) Close() error {
	nat.mappingmu.Lock()
	nat.closed = true
	nat.mappingmu.Unlock()

	nat.ctxCancel()
	nat.refCount.Wait()
	return nil
}

func (nat *NAT) GetMapping(protocol string, port int) (addr netip.AddrPort, found bool) {
	nat.mappingmu.Lock()
	defer nat.mappingmu.Unlock()

	if !nat.extAddr.IsValid() {
		return netip.AddrPort{}, false
	}
	extPort, found := nat.mappings[entry{protocol: protocol, port: port}]
	if !found {
		return netip.AddrPort{}, false
	}
	return netip.AddrPortFrom(nat.extAddr, uint16(extPort)), true
}

// AddMapping attempts to construct a mapping on protocol and internal port.
// It blocks until a mapping was established. Once added, it periodically renews the mapping.
//
// May not succeed, and mappings may change over time;
// NAT devices may not respect our port requests, and even lie.
func (nat *NAT) AddMapping(ctx context.Context, protocol string, port int) error {
	switch protocol {
	case "tcp", "udp":
	default:
		return fmt.Errorf("invalid protocol: %s", protocol)
	}

	nat.mappingmu.Lock()
	defer nat.mappingmu.Unlock()

	if nat.closed {
		return errors.New("closed")
	}

	// do it once synchronously, so first mapping is done right away, and before exiting,
	// allowing users -- in the optimistic case -- to use results right after.
	extPort := nat.establishMapping(ctx, protocol, port)
	nat.mappings[entry{protocol: protocol, port: port}] = extPort
	return nil
}

// RemoveMapping removes a port mapping.
// It blocks until the NAT has removed the mapping.
func (nat *NAT) RemoveMapping(ctx context.Context, protocol string, port int) error {
	nat.mappingmu.Lock()
	defer nat.mappingmu.Unlock()

	switch protocol {
	case "tcp", "udp":
		e := entry{protocol: protocol, port: port}
		if _, ok := nat.mappings[e]; ok {
			delete(nat.mappings, e)
			return nat.nat.DeletePortMapping(ctx, protocol, port)
		}
		return errors.New("unknown mapping")
	default:
		return fmt.Errorf("invalid protocol: %s", protocol)
	}
}

func (nat *NAT) background() {
	const mappingUpdate = MappingDuration / 3

	now := time.Now()
	nextMappingUpdate := now.Add(mappingUpdate)
	nextAddrUpdate := now.Add(CacheTime)

	t := time.NewTimer(minTime(nextMappingUpdate, nextAddrUpdate).Sub(now)) // don't use a ticker here. We don't know how long establishing the mappings takes.
	defer t.Stop()

	var in []entry
	var out []int // port numbers
	for {
		select {
		case now := <-t.C:
			if now.After(nextMappingUpdate) {
				in = in[:0]
				out = out[:0]
				nat.mappingmu.Lock()
				for e := range nat.mappings {
					in = append(in, e)
				}
				nat.mappingmu.Unlock()
				// Establishing the mapping involves network requests.
				// Don't hold the mutex, just save the ports.
				for _, e := range in {
					out = append(out, nat.establishMapping(nat.ctx, e.protocol, e.port))
				}
				nat.mappingmu.Lock()
				for i, p := range in {
					if _, ok := nat.mappings[p]; !ok {
						continue // entry might have been deleted
					}
					nat.mappings[p] = out[i]
				}
				nat.mappingmu.Unlock()
				nextMappingUpdate = time.Now().Add(mappingUpdate)
			}
			if now.After(nextAddrUpdate) {
				var extAddr netip.Addr
				extIP, err := nat.nat.GetExternalAddress()
				if err == nil {
					extAddr, _ = netip.AddrFromSlice(extIP)
				}
				nat.extAddr = extAddr
				nextAddrUpdate = time.Now().Add(CacheTime)
			}
			t.Reset(time.Until(minTime(nextAddrUpdate, nextMappingUpdate)))
		case <-nat.ctx.Done():
			nat.mappingmu.Lock()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			for e := range nat.mappings {
				delete(nat.mappings, e)
				nat.nat.DeletePortMapping(ctx, e.protocol, e.port)
			}
			nat.mappingmu.Unlock()
			return
		}
	}
}

func (nat *NAT) establishMapping(ctx context.Context, protocol string, internalPort int) (externalPort int) {
	log.Debugf("Attempting port map: %s/%d", protocol, internalPort)
	const comment = "libp2p"

	nat.natmu.Lock()
	var err error
	externalPort, err = nat.nat.AddPortMapping(ctx, protocol, internalPort, comment, MappingDuration)
	if err != nil {
		// Some hardware does not support mappings with timeout, so try that
		externalPort, err = nat.nat.AddPortMapping(ctx, protocol, internalPort, comment, 0)
	}
	nat.natmu.Unlock()

	if err != nil || externalPort == 0 {
		// TODO: log.Event
		if err != nil {
			log.Warnf("failed to establish port mapping: %s", err)
		} else {
			log.Warnf("failed to establish port mapping: newport = 0")
		}
		// we do not close if the mapping failed,
		// because it may work again next time.
		return 0
	}

	log.Debugf("NAT Mapping: %d --> %d (%s)", externalPort, internalPort, protocol)
	return externalPort
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
