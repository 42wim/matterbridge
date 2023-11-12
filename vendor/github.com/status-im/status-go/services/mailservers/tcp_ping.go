package mailservers

import (
	"context"
	"fmt"
	"net"
	"time"

	multiaddr "github.com/multiformats/go-multiaddr"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/status-im/status-go/rtt"
)

type PingQuery struct {
	Addresses []string `json:"addresses"`
	TimeoutMs int      `json:"timeoutMs"`
}

type PingResult struct {
	Address string  `json:"address"`
	RTTMs   *int    `json:"rttMs"`
	Err     *string `json:"error"`
}

type parseFn func(string) (string, error)

func (pr *PingResult) Update(rttMs int, err error) {
	if err != nil {
		errStr := err.Error()
		pr.Err = &errStr
	}
	if rttMs >= 0 {
		pr.RTTMs = &rttMs
	} else {
		pr.RTTMs = nil
	}
}

func EnodeToAddr(node *enode.Node) (string, error) {
	var ip4 enr.IPv4
	err := node.Load(&ip4)
	if err != nil {
		return "", err
	}
	var tcp enr.TCP
	err = node.Load(&tcp)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", net.IP(ip4).String(), tcp), nil
}

func EnodeStringToAddr(enodeAddr string) (string, error) {
	node, err := enode.ParseV4(enodeAddr)
	if err != nil {
		return "", err
	}
	return EnodeToAddr(node)
}

func parse(addresses []string, fn parseFn) (map[string]*PingResult, []string) {
	results := make(map[string]*PingResult, len(addresses))
	var toPing []string

	for i := range addresses {
		addr, err := fn(addresses[i])
		if err != nil {
			errStr := err.Error()
			results[addresses[i]] = &PingResult{Address: addresses[i], Err: &errStr}
			continue
		}
		results[addr] = &PingResult{Address: addresses[i]}
		toPing = append(toPing, addr)
	}
	return results, toPing
}

func mapValues(m map[string]*PingResult) []*PingResult {
	rval := make([]*PingResult, 0, len(m))
	for _, value := range m {
		rval = append(rval, value)
	}
	return rval
}

func DoPing(ctx context.Context, addresses []string, timeoutMs int, p parseFn) ([]*PingResult, error) {
	timeout := time.Duration(timeoutMs) * time.Millisecond

	resultsMap, toPing := parse(addresses, p)

	// run the checks concurrently
	results, err := rtt.CheckHosts(toPing, timeout)
	if err != nil {
		return nil, err
	}

	// set ping results
	for i := range results {
		r := results[i]
		pr := resultsMap[r.Addr]
		if pr == nil {
			continue
		}
		pr.Update(r.RTTMs, r.Err)
	}

	return mapValues(resultsMap), nil
}

func (a *API) Ping(ctx context.Context, pq PingQuery) ([]*PingResult, error) {
	return DoPing(ctx, pq.Addresses, pq.TimeoutMs, EnodeStringToAddr)
}

func MultiAddressToAddress(multiAddr string) (string, error) {

	ma, err := multiaddr.NewMultiaddr(multiAddr)
	if err != nil {
		return "", err
	}

	dns4, err := ma.ValueForProtocol(multiaddr.P_DNS4)
	if err != nil && err != multiaddr.ErrProtocolNotFound {
		return "", err
	}

	ip4 := ""
	if dns4 != "" {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ip4, err = net.DefaultResolver.LookupCNAME(ctx, dns4)
		if err != nil {
			return "", err
		}
	} else {
		ip4, err = ma.ValueForProtocol(multiaddr.P_IP4)
		if err != nil {
			return "", err
		}
	}

	tcp, err := ma.ValueForProtocol(multiaddr.P_TCP)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", ip4, tcp), nil
}

func (a *API) MultiAddressPing(ctx context.Context, pq PingQuery) ([]*PingResult, error) {
	return DoPing(ctx, pq.Addresses, pq.TimeoutMs, MultiAddressToAddress)
}
